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
	pipelinev1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"github.com/tektoncd/pipeline/pkg/names"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
)

// RetryablePipelineRunStatus defines the observed state of RetryablePipelineRun
type RetryablePipelineRunStatus struct {
	duckv1beta1.Status `json:",inline"`

	// StartTime is the time the RetryablePipelineRun is actually started.
	// +optional
	StartTime *metav1.Time `json:"startTime,omitempty"`

	// CompletionTime is the time the RetryablePipelineRun completed.
	// +optional
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`

	// PipelineRuns is the list of PipelineRun statuses with using its name as the merge key.
	// +optional
	PipelineRuns []*PartialPipelineRunStatusFields `json:"pipelineRuns,omitempty"`

	// PipelineResults is the list of results written out by the pipeline task's containers.
	// +optional
	// +listType=atomic
	PipelineResults []pipelinev1beta1.PipelineRunResult `json:"pipelineResults,omitempty"`

	// PinnedPipelineRun contains the exact spec used to instantiate the first run.
	PinnedPipelineRun *pipelinev1beta1.PipelineRun `json:"pinnedPipelineRun,omitempty"`
}

// InitializeStatus will set all conditions in condSet to unknown for the RetryablePipelineRun
// and set the started time to the current time
func (rpr *RetryablePipelineRun) InitializeStatus() {
	justStarted := false
	if rpr.Status == nil {
		rpr.Status = &RetryablePipelineRunStatus{
			PipelineRuns:    []*PartialPipelineRunStatusFields{},
			PipelineResults: []pipelinev1beta1.PipelineRunResult{},
		}
		justStarted = true
	}
	if justStarted {
		conditionManager := condSet.Manage(rpr.Status)
		conditionManager.InitializeConditions()
		initialCondition := conditionManager.GetCondition(apis.ConditionSucceeded)
		initialCondition.Reason = pipelinev1beta1.PipelineRunReasonStarted.String()
		conditionManager.SetCondition(*initialCondition)
		rpr.Status.StartTime = &initialCondition.LastTransitionTime.Inner
	}
}

// PartialPipelineRunStatusFields contains the fields of child PipelineRuns' status.
type PartialPipelineRunStatusFields struct {
	// Name is the name of PipelineRun as well as the primary key of PartialPipelineRunStatusFields collection.
	Name string `json:"name"`

	// StartTime is the time the PipelineRun is actually started.
	// +optional
	StartTime *metav1.Time `json:"startTime,omitempty"`

	// CompletionTime is the time the PipelineRun completed.
	// +optional
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`

	// Conditions is the latest available observations of a resource's current state.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions duckv1beta1.Conditions `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`

	// SkippedTasks is the list of tasks that were skipped due to when expressions evaluating to false.
	// +optional
	// +listType=atomic
	SkippedTasks []pipelinev1beta1.SkippedTask `json:"skippedTasks,omitempty"`

	// TaskRuns is the map of PipelineRunTaskRunStatus with the taskRun name as the key.
	// +optional
	TaskRuns map[string]*pipelinev1beta1.PipelineRunTaskRunStatus `json:"taskRuns,omitempty"`
}

func (s *PartialPipelineRunStatusFields) SyncFrom(pr *pipelinev1beta1.PipelineRunStatus) {
	s.StartTime = pr.StartTime
	s.CompletionTime = pr.CompletionTime
	s.Conditions = pr.Conditions
	s.SkippedTasks = pr.SkippedTasks
	s.TaskRuns = pr.TaskRuns
}

func (rpr *RetryablePipelineRun) PipelineRunStatus(name string) *PartialPipelineRunStatusFields {
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
	rpr.Status.PipelineRuns = append(rpr.Status.PipelineRuns, &PartialPipelineRunStatusFields{
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
