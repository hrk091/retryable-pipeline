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

package v1alpha1_test

import (
	"github.com/hrk091/retryable-pipeline/api/v1alpha1"
	pipelinev1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Transformer func(r *v1alpha1.RetryablePipelineRun)

func NewRetryablePipelineRun(name string, transformers ...Transformer) *v1alpha1.RetryablePipelineRun {
	r := &v1alpha1.RetryablePipelineRun{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	for _, transform := range transformers {
		transform(r)
	}
	return r
}

func PipelineRef(v string) Transformer {
	return func(r *v1alpha1.RetryablePipelineRun) {
		r.Spec = &v1alpha1.RetryablePipelineRunSpec{
			PipelineRunSpec: &pipelinev1beta1.PipelineRunSpec{
				PipelineRef: &pipelinev1beta1.PipelineRef{
					Name: v,
				},
			},
		}
	}
}

func InitStatus() Transformer {
	return func(r *v1alpha1.RetryablePipelineRun) {
		r.InitializeStatus()
	}
}
