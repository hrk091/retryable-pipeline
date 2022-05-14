/*
 Copyright 2022 Hiroki Okui.

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package v1alpha1

import (
	"github.com/hrk091/retryable-pipeline/pkg/pipelinerun"
	pipelinev1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"github.com/tektoncd/pipeline/pkg/names"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
)

// RetryablePipelineRunStatus defines the observed state of RetryablePipelineRun.
type RetryablePipelineRunStatus struct {
	duckv1beta1.Status `json:",inline"`

	// StartTime is the time the RetryablePipelineRun is actually started.
	// +optional
	StartTime *metav1.Time `json:"startTime,omitempty"`

	// CompletionTime is the time the RetryablePipelineRun completed.
	// +optional
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`

	// PinnedPipelineRun contains the exact spec used to instantiate the first run.
	PinnedPipelineRun *pipelinev1beta1.PipelineRun `json:"pinnedPipelineRun,omitempty"`

	// PipelineRuns is the list of child PipelineRun status copies with using its name as the merge key.
	// +optional
	PipelineRuns []*PartialPipelineRunStatus `json:"pipelineRuns,omitempty"`

	// TaskResults is the map of TaskResult with the PipelineTask name as the key.
	// +optional
	TaskResults map[string]*CompletedTaskResults `json:"taskResults,omitempty"`

	// PipelineResults is the list of results written out by the pipeline task's containers.
	// +optional
	// +listType=atomic
	PipelineResults []pipelinev1beta1.PipelineRunResult `json:"pipelineResults,omitempty"`

	// RetryKey is the copy of retry-key annotation to compare with its previous value.
	RetryKey string
}

// InitializeStatus initializes the RetryablePipelineRun status.
func (rpr *RetryablePipelineRun) InitializeStatus() {
	if rpr.Status == nil {
		rpr.Status = &RetryablePipelineRunStatus{
			PipelineRuns:    []*PartialPipelineRunStatus{},
			PipelineResults: []pipelinev1beta1.PipelineRunResult{},
		}
		rpr.Status.InitializeConditions()
	}
}

// PinPipelineRun copies resolved PipelineSpec and metadata to RetryablePipelineRunStatus.PinnedPipelineRun
// to pin spec and metadata of the first PipelineRun and enable to reuse them for building retry PipelineRun.
func (rpr *RetryablePipelineRun) PinPipelineRun(pr *pipelinev1beta1.PipelineRun) bool {
	if pr.Status.PipelineSpec == nil {
		return false
	}
	ppr := pr.DeepCopy()
	ppr.ObjectMeta = metav1.ObjectMeta{
		Namespace:   pr.Namespace,
		Labels:      pr.Labels,
		Annotations: pr.Annotations,
	}
	delete(ppr.Annotations, "kubectl.kubernetes.io/last-applied-configuration")
	ppr.Status = pipelinev1beta1.PipelineRunStatus{
		PipelineRunStatusFields: pipelinev1beta1.PipelineRunStatusFields{
			PipelineSpec: pr.Status.PipelineSpec,
		},
	}

	rpr.Status.PinnedPipelineRun = ppr
	return true
}

// AggregateChildrenResults aggregates belonging PipelineRun/TaskRun statuses
// and populates RetryablePipelineRunStatus with their latest statuses.
func (rpr *RetryablePipelineRun) AggregateChildrenResults() {
	m := map[string]*CompletedTaskResults{}
	for _, pr := range rpr.Status.PipelineRuns {
		for trName, tr := range pr.TaskRuns {
			if !tr.Status.GetCondition(apis.ConditionSucceeded).IsTrue() {
				continue
			}
			if _, ok := m[tr.PipelineTaskName]; ok {
				continue
			}
			m[tr.PipelineTaskName] = &CompletedTaskResults{
				FromPipelineRun: pr.Name,
				FromTaskRun:     trName,
				Results:         tr.Status.TaskRunResults,
			}
		}
	}
	rpr.Status.TaskResults = m

	lastPr := rpr.Status.PipelineRuns[len(rpr.Status.PipelineRuns)-1]
	lastCond := lastPr.GetCondition()
	// TODO create original condition instead of copying
	rpr.Status.SetCondition(lastCond.DeepCopy())
	if lastCond.IsTrue() {
		rpr.Status.PipelineResults = lastPr.PipelineResults
	}
}

// ResolvedResultRefs converts completed TaskRun results to the list of pipelinerun.ResolvedResultRef
// in order to utilize it for variable substitution.
func (s *RetryablePipelineRunStatus) ResolvedResultRefs() pipelinerun.ResolvedResultRefs {
	var refs pipelinerun.ResolvedResultRefs
	for ptname, results := range s.TaskResults {
		for _, r := range results.Results {
			refs = append(refs, &pipelinerun.ResolvedResultRef{
				Value: *pipelinev1beta1.NewArrayOrString(r.Value),
				ResultReference: pipelinev1beta1.ResultRef{
					PipelineTask: ptname,
					Result:       r.Name,
				},
				FromTaskRun: results.FromTaskRun,
			})
		}
	}
	return refs
}

// CompletedTaskResults contains the TaskRunResults generated by belonging PipelineRuns.
type CompletedTaskResults struct {
	// FromPipelineRun shows the name of PipelineRun where the task result is written.
	FromPipelineRun string `json:"fromPipelineRun,omitempty"`

	// FromTaskRun shows the name of TaskRun where the task result is written.
	FromTaskRun string `json:"fromTaskRun,omitempty"`

	// Results are the list of results written out by the task's containers
	// +optional
	// +listType=atomic
	Results []pipelinev1beta1.TaskRunResult `json:"results,omitempty"`
}

// PartialPipelineRunStatus contains the fields of belonging PipelineRuns' status.
type PartialPipelineRunStatus struct {
	duckv1beta1.Status `json:",inline"`

	// Name is the name of PipelineRun as well as the primary key of PartialPipelineRunStatus collection.
	Name string `json:"name"`

	// StartTime is the time the PipelineRun is actually started.
	// +optional
	StartTime *metav1.Time `json:"startTime,omitempty"`

	// CompletionTime is the time the PipelineRun completed.
	// +optional
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`

	// PipelineResults are the list of results written out by the pipeline task's containers
	// +optional
	// +listType=atomic
	PipelineResults []pipelinev1beta1.PipelineRunResult `json:"pipelineResults,omitempty"`

	// SkippedTasks is the list of tasks that were skipped due to when expressions evaluating to false.
	// +optional
	// +listType=atomic
	SkippedTasks []pipelinev1beta1.SkippedTask `json:"skippedTasks,omitempty"`

	// TaskRuns is the map of PipelineRunTaskRunStatus with the taskRun name as the key.
	// +optional
	TaskRuns map[string]*pipelinev1beta1.PipelineRunTaskRunStatus `json:"taskRuns,omitempty"`
}

// IsRetryKeyChanged returns true when a retry-key annotation is changed from the previous value.
func (rpr *RetryablePipelineRun) IsRetryKeyChanged() bool {
	return RetryKey(&rpr.ObjectMeta) != rpr.Status.RetryKey
}

// CopyRetryKey copies a retry-key annotation to status.retryKey field.
func (rpr *RetryablePipelineRun) CopyRetryKey() {
	rpr.Status.RetryKey = RetryKey(&rpr.ObjectMeta)
}

// CopyFrom copies the status of given PipelineRun to the list of PipelineRun statuses inside RetryablePipelineRun status.
func (s *PartialPipelineRunStatus) CopyFrom(pr *pipelinev1beta1.PipelineRunStatus) {
	s.StartTime = pr.StartTime
	s.CompletionTime = pr.CompletionTime
	s.Conditions = pr.Conditions
	s.SkippedTasks = pr.SkippedTasks
	s.TaskRuns = pr.TaskRuns
	s.PipelineResults = pr.PipelineResults
}

// PipelineRunStatus returns the latest status of given PipelineRun from the copies inside RetryablePipelineRun status.
func (rpr *RetryablePipelineRun) PipelineRunStatus(name string) *PartialPipelineRunStatus {
	for _, pr := range rpr.Status.PipelineRuns {
		if pr.Name == name {
			return pr
		}
	}
	return nil
}

// ReserveNextPipelineRunName reserves a name of PipelineRun to be created at the next reconcile loop.
func (rpr *RetryablePipelineRun) ReserveNextPipelineRunName() bool {
	for _, pr := range rpr.Status.PipelineRuns {
		if pr.StartTime == nil {
			return false
		}
	}
	rpr.Status.PipelineRuns = append(rpr.Status.PipelineRuns, &PartialPipelineRunStatus{
		Name: rpr.genPipelineRunName(),
	})
	return true
}

// NextPipelineRunName returns a name of PipelineRun that should be used for the next PipelineRun.
func (rpr *RetryablePipelineRun) NextPipelineRunName() string {
	for _, pr := range rpr.Status.PipelineRuns {
		if pr.StartTime == nil {
			return pr.Name
		}
	}
	return ""
}

// StartedPipelineRunCount returns the number of child PipelineRuns already started.
func (rpr *RetryablePipelineRun) StartedPipelineRunCount() int {
	c := 0
	for _, pr := range rpr.Status.PipelineRuns {
		if pr.StartTime != nil {
			c++
		}
	}
	return c
}

func (rpr *RetryablePipelineRun) genPipelineRunName() string {
	return names.SimpleNameGenerator.RestrictLengthWithRandomSuffix(rpr.Name)
}
