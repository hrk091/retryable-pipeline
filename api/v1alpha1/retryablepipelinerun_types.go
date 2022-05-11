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
	"k8s.io/apimachinery/pkg/labels"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName="rpr"
//+kubebuilder:printcolumn:name="SUCCEEDED",type="string",JSONPath=`.status.conditions[?(@.type=="Succeeded")].status`
//+kubebuilder:printcolumn:name="REASON",type="string",JSONPath=`.status.conditions[?(@.type=="Succeeded")].reason`
//+kubebuilder:printcolumn:name="STARTTIME",type="date",JSONPath=".status.startTime"
//+kubebuilder:printcolumn:name="COMPLETIONTIME",type="date",JSONPath=".status.completionTime"

// RetryablePipelineRun is the Schema for the retryablepipelineruns API
type RetryablePipelineRun struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +optional
	Spec *RetryablePipelineRunSpec `json:"spec,omitempty"`

	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:Schemaless
	// +kubebuilder:validation:Type=object
	// +optional
	Status *RetryablePipelineRunStatus `json:"status,omitempty"`
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
	*pipelinev1beta1.PipelineRunSpec `json:",inline"`
}

func (rpr *RetryablePipelineRun) ChildLabels() labels.Set {
	return labels.Merge(rpr.Labels, map[string]string{LabelKeyRetryablePipelineRun: rpr.Name})
}

func (rpr *RetryablePipelineRun) NewPipelineRun(name string, spec *pipelinev1beta1.PipelineRunSpec) *pipelinev1beta1.PipelineRun {
	return &pipelinev1beta1.PipelineRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   rpr.Namespace,
			Labels:      rpr.ChildLabels(),
			Annotations: rpr.Annotations,
		},
		Spec: *spec,
	}
}
