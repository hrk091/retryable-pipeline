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

package internal

import (
	"github.com/hrk091/retryable-pipeline/api/v1alpha1"
	pipelinev1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func DecodePR(b []byte) *pipelinev1beta1.PipelineRun {
	var o pipelinev1beta1.PipelineRun
	if err := yaml.Unmarshal(b, &o); err != nil {
		panic(err)
	}
	return &o
}

func DecodeRPR(b []byte) *v1alpha1.RetryablePipelineRun {
	var o v1alpha1.RetryablePipelineRun
	if err := yaml.Unmarshal(b, &o); err != nil {
		panic(err)
	}
	return &o
}
