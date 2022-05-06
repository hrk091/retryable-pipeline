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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// RetryablePipelineRun is the Schema for the retryablepipelineruns API
type RetryablePipelineRun struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +optional
	Spec RetryablePipelineRunSpec `json:"spec,omitempty"`

	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:Schemaless
	// +kubebuilder:validation:Type=object
	// +optional
	Status RetryablePipelineRunStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RetryablePipelineRunList contains a list of RetryablePipelineRun
type RetryablePipelineRunList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RetryablePipelineRun `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RetryablePipelineRun{}, &RetryablePipelineRunList{})
}

// RetryablePipelineRunSpec defines the desired state of RetryablePipelineRun
type RetryablePipelineRunSpec struct {
	pipelinev1beta1.PipelineRunSpec `json:"spec,inline"`
}

// RetryablePipelineRunStatus defines the observed state of RetryablePipelineRun
type RetryablePipelineRunStatus struct {
	// StartTime is the time the RetryablePipelineRun is actually started.
	// +optional
	StartTime *metav1.Time `json:"startTime,omitempty"`

	// CompletionTime is the time the RetryablePipelineRun completed.
	// +optional
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`

	// PipelineRuns is the map of PipelineRunTaskRunStatus with the taskRun name as the key.
	// +optional
	PipelineRuns map[string]*pipelinev1beta1.PipelineRunStatus `json:"pipelineRuns,omitempty"`

	// PipelineResults is the list of results written out by the pipeline task's containers
	// +optional
	// +listType=atomic
	PipelineResults []pipelinev1beta1.PipelineRunResult `json:"pipelineResults,omitempty"`

	// PinnedPipelineRun contains the exact spec used to instantiate the first run.
	PinnedPipelineRun *pipelinev1beta1.PipelineRun `json:"pinnedPipelineRun,omitempty"`
}

// PartialPipelineRunStatusFields contain the fields of child PipelineRuns' status.
type PartialPipelineRunStatusFields struct {
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

	// ChildReferences is the list of TaskRun and Run names, PipelineTask names, and API versions/kinds for children of this PipelineRun.
	// +optional
	// +listType=atomic
	ChildReferences []pipelinev1beta1.ChildStatusReference `json:"childReferences,omitempty"`
}
