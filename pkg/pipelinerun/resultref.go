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

package pipelinerun

import (
	"fmt"
	pipelinev1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
)

type ResolvedResultRefs []*ResolvedResultRef

// ResolvedResultRef represents a result ref reference that has been fully resolved (value has been populated).
// If the value is from a Result, then the ResultReference will be populated to point to the ResultReference
// which resulted in the value
type ResolvedResultRef struct {
	Value           pipelinev1beta1.ArrayOrString
	ResultReference pipelinev1beta1.ResultRef
	FromTaskRun     string
}

func (rs ResolvedResultRefs) getStringReplacements() map[string]string {
	replacements := map[string]string{}
	for _, r := range rs {
		for _, target := range r.getReplaceTarget() {
			replacements[target] = r.Value.StringVal
		}
	}
	return replacements
}

func (r *ResolvedResultRef) getReplaceTarget() []string {
	return []string{
		fmt.Sprintf("%s.%s.%s.%s", pipelinev1beta1.ResultTaskPart, r.ResultReference.PipelineTask, pipelinev1beta1.ResultResultPart, r.ResultReference.Result),
		fmt.Sprintf("%s.%s.%s[%q]", pipelinev1beta1.ResultTaskPart, r.ResultReference.PipelineTask, pipelinev1beta1.ResultResultPart, r.ResultReference.Result),
		fmt.Sprintf("%s.%s.%s['%s']", pipelinev1beta1.ResultTaskPart, r.ResultReference.PipelineTask, pipelinev1beta1.ResultResultPart, r.ResultReference.Result),
	}
}
