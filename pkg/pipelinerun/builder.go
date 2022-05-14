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
	pipelinev1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"github.com/tektoncd/pipeline/pkg/substitution"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Transformer func(r *pipelinev1beta1.PipelineRun)

// NewPipelineRun returns a new PipelineRun with applying one or more transformers.
func NewPipelineRun(m metav1.ObjectMeta, transformers ...Transformer) *pipelinev1beta1.PipelineRun {
	pr := &pipelinev1beta1.PipelineRun{
		ObjectMeta: m,
	}
	for _, transform := range transformers {
		transform(pr)
	}
	return pr
}

// Spec returns a Transformer func to add a given PipelineRunSpec to PipelineRun.
func Spec(spec *pipelinev1beta1.PipelineRunSpec) Transformer {
	return func(r *pipelinev1beta1.PipelineRun) {
		r.Spec = *spec.DeepCopy()
	}
}

// RemovePipelineRef returns a Transformer func to remove PipelineRef.
func RemovePipelineRef() Transformer {
	return func(r *pipelinev1beta1.PipelineRun) {
		r.Spec.PipelineRef = nil
	}
}

// PipelineSpec returns a Transformer func to add a given PipelineSpec to PipelineRun.
func PipelineSpec(spec *pipelinev1beta1.PipelineSpec) Transformer {
	return func(r *pipelinev1beta1.PipelineRun) {
		r.Spec.PipelineSpec = spec.DeepCopy()
	}
}

// ApplyResultsToPipelineTask returns a Transformer func to apply variable substitution for given PipelineTask
// using already completed task results.
func ApplyResultsToPipelineTask(pipelineTaskName string, refs ResolvedResultRefs) Transformer {
	return func(r *pipelinev1beta1.PipelineRun) {
		stringReplacements := refs.getStringReplacements()
		// TODO replace condition
		for i, t := range r.Spec.PipelineSpec.Tasks {
			if t.Name == pipelineTaskName {
				pipelineTask := t.DeepCopy()
				pipelineTask.Params = replaceParamValues(pipelineTask.Params, stringReplacements, nil)
				pipelineTask.WhenExpressions = pipelineTask.WhenExpressions.ReplaceWhenExpressionsVariables(stringReplacements, nil)
				r.Spec.PipelineSpec.Tasks[i] = *pipelineTask
			}
		}
	}
}

// ApplyResultsToPipelineTasks returns a Transformer func to apply variable substitution for all PipelineTasks
// using already completed task results.
func ApplyResultsToPipelineTasks(refs ResolvedResultRefs) Transformer {
	return func(r *pipelinev1beta1.PipelineRun) {
		stringReplacements := refs.getStringReplacements()
		// TODO replace condition
		for i, t := range r.Spec.PipelineSpec.Tasks {
			pipelineTask := t.DeepCopy()
			pipelineTask.Params = replaceParamValues(pipelineTask.Params, stringReplacements, nil)
			pipelineTask.WhenExpressions = pipelineTask.WhenExpressions.ReplaceWhenExpressionsVariables(stringReplacements, nil)
			r.Spec.PipelineSpec.Tasks[i] = *pipelineTask
		}
	}
}

// ApplyResultsToPipelineResults returns a Transformer func to apply variable substitution for pipeline results
// using already completed task results.
func ApplyResultsToPipelineResults(refs ResolvedResultRefs) Transformer {
	return func(r *pipelinev1beta1.PipelineRun) {
		stringReplacements := refs.getStringReplacements()
		r.Spec.PipelineSpec.Results = replaceResultValues(r.Spec.PipelineSpec.Results, stringReplacements)
	}
}

// SkipTask returns a Transformer func to add WhenExpression which is always evaluated as false
// to a given PipelineTask to skip it.
func SkipTask(pipelineTaskName string) Transformer {
	return func(r *pipelinev1beta1.PipelineRun) {
		for i, t := range r.Spec.PipelineSpec.Tasks {
			if t.Name == pipelineTaskName {
				t.WhenExpressions = SkipWhenExpression()
			}
			r.Spec.PipelineSpec.Tasks[i] = t
		}
	}
}

// SkipWhenExpression creates v1beta1.WhenExpression which is always evaluated as false to skip tasks it is attached.
func SkipWhenExpression() pipelinev1beta1.WhenExpressions {
	return pipelinev1beta1.WhenExpressions{
		{
			Input:    "skipped",
			Operator: "in",
			Values:   []string{""},
		},
	}
}

func replaceParamValues(params []pipelinev1beta1.Param, stringReplacements map[string]string, arrayReplacements map[string][]string) []pipelinev1beta1.Param {
	for i := range params {
		params[i].Value.ApplyReplacements(stringReplacements, arrayReplacements)
	}
	return params
}

func replaceResultValues(results []pipelinev1beta1.PipelineResult, stringReplacements map[string]string) []pipelinev1beta1.PipelineResult {
	for i := range results {
		results[i].Value = substitution.ApplyReplacements(results[i].Value, stringReplacements)
	}
	return results
}
