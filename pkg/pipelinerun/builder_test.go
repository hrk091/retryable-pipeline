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

package pipelinerun_test

import (
	"github.com/hrk091/retryable-pipeline/internal"
	"github.com/hrk091/retryable-pipeline/pkg/pipelinerun"
	"github.com/stretchr/testify/assert"
	pipelinev1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"testing"
)

func testObjectMeta() metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      "test",
		Namespace: "test-namespace",
	}
}

func expect(b []byte) *pipelinev1beta1.PipelineRun {
	var want pipelinev1beta1.PipelineRun
	if err := yaml.Unmarshal(b, &want); err != nil {
		panic(err)
	}
	return &want
}

func TestNewPipelineRun(t *testing.T) {
	want := expect([]byte(`
metadata:
  name: test
  namespace: test-namespace
`))
	assert.Equal(t, want, pipelinerun.NewPipelineRun(testObjectMeta()))
}

func TestAllSpec(t *testing.T) {
	want := expect([]byte(`
metadata:
  name: test
  namespace: test-namespace
spec:
  pipelineRef:
    name: sample-pipeline
`))

	runSpec := pipelinev1beta1.PipelineRunSpec{
		PipelineRef: &pipelinev1beta1.PipelineRef{
			Name: "sample-pipeline",
		},
	}
	pr := pipelinerun.NewPipelineRun(testObjectMeta(), pipelinerun.Spec(&runSpec))
	assert.Equal(t, want, pr)
}

func TestPipelineSpec(t *testing.T) {
	want := expect([]byte(`
metadata:
  name: test
  namespace: test-namespace
spec:
  pipelineSpec:
    params:
      - name: msg
        type: string
    results:
      - description: test result
        name: out
        value: $(tasks.task2.results.out)
    tasks:
      - name: task1
        params:
          - name: msg
            value: $(params.msg)
        taskRef:
          kind: Task
          name: sample-echo
      - name: task2
        params:
          - name: msg
            value: $(tasks.task1.results.out)
        runAfter:
          - task1
        taskRef:
          kind: Task
          name: sample-echo
`))
	p := internal.NewPipelineTestData("sample.pipeline")
	pr := pipelinerun.NewPipelineRun(testObjectMeta(), pipelinerun.PipelineSpec(&p.Spec))
	assert.Equal(t, want, pr)
}

func TestRemovePipelineRef(t *testing.T) {
	want := expect([]byte(`
metadata:
  name: test
  namespace: test-namespace
spec:
  params:
    - name: foo
      value: bar
`))

	runSpec := pipelinev1beta1.PipelineRunSpec{
		Params: []pipelinev1beta1.Param{
			{Name: "foo", Value: *pipelinev1beta1.NewArrayOrString("bar")},
		},
		PipelineRef: &pipelinev1beta1.PipelineRef{
			Name: "sample-pipeline",
		},
	}
	pr := pipelinerun.NewPipelineRun(testObjectMeta(), pipelinerun.Spec(&runSpec), pipelinerun.RemovePipelineRef())
	assert.Equal(t, want, pr)
}

func TestApplyResultsToPipelineTasks(t *testing.T) {
	want := expect([]byte(`
metadata:
  name: test
  namespace: test-namespace
spec:
  pipelineSpec:
    params:
      - name: msg
        type: string
    results:
      - description: test result
        name: out
        value: $(tasks.task2.results.out)
    tasks:
      - name: task1
        params:
          - name: msg
            value: $(params.msg)
        taskRef:
          kind: Task
          name: sample-echo
      - name: task2
        params:
          - name: msg
            value: REPLACED
        runAfter:
          - task1
        taskRef:
          kind: Task
          name: sample-echo
        when:
          - input: REPLACED
            operator: in
            values: 
              - REPLACED
              - "$(tasks.not-exist.results.out)"
`))
	p := internal.NewPipelineTestData("sample.pipeline")
	p.Spec.Tasks[1].WhenExpressions = pipelinev1beta1.WhenExpressions{
		pipelinev1beta1.WhenExpression{
			Input:    "$(tasks.task1.results.out)",
			Operator: "in",
			Values:   []string{"$(tasks.task1.results.out)", "$(tasks.not-exist.results.out)"},
		},
	}
	refs := pipelinerun.ResolvedResultRefs{
		&pipelinerun.ResolvedResultRef{
			ResultReference: pipelinev1beta1.ResultRef{
				PipelineTask: "task1",
				Result:       "out",
			},
			Value:       *pipelinev1beta1.NewArrayOrString("REPLACED"),
			FromTaskRun: "test-abcde-task1",
		},
	}
	pr := pipelinerun.NewPipelineRun(testObjectMeta(), pipelinerun.PipelineSpec(&p.Spec), pipelinerun.ApplyResultsToPipelineTasks(refs))
	assert.Equal(t, want, pr)
}

func TestApplyResultsToPipelineTask(t *testing.T) {
	want := expect([]byte(`
metadata:
  name: test
  namespace: test-namespace
spec:
  pipelineSpec:
    params:
      - name: msg
        type: string
    results:
      - description: test result
        name: out
        value: $(tasks.task2.results.out)
    tasks:
      - name: task1
        params:
          - name: msg
            value: $(params.msg)
        taskRef:
          kind: Task
          name: sample-echo
      - name: task2
        params:
          - name: msg
            value: REPLACED
        runAfter:
          - task1
        taskRef:
          kind: Task
          name: sample-echo
        when:
          - input: REPLACED
            operator: in
            values: 
              - REPLACED
              - "$(tasks.not-exist.results.out)"
`))
	p := internal.NewPipelineTestData("sample.pipeline")
	p.Spec.Tasks[1].WhenExpressions = pipelinev1beta1.WhenExpressions{
		pipelinev1beta1.WhenExpression{
			Input:    "$(tasks.task1.results.out)",
			Operator: "in",
			Values:   []string{"$(tasks.task1.results.out)", "$(tasks.not-exist.results.out)"},
		},
	}
	refs := pipelinerun.ResolvedResultRefs{
		&pipelinerun.ResolvedResultRef{
			ResultReference: pipelinev1beta1.ResultRef{
				PipelineTask: "task1",
				Result:       "out",
			},
			Value:       *pipelinev1beta1.NewArrayOrString("REPLACED"),
			FromTaskRun: "test-abcde-task1",
		},
	}
	pr := pipelinerun.NewPipelineRun(testObjectMeta(), pipelinerun.PipelineSpec(&p.Spec), pipelinerun.ApplyResultsToPipelineTask("task2", refs))
	assert.Equal(t, want, pr)
}

func TestApplyResultsToPipelineResults(t *testing.T) {
	want := expect([]byte(`
metadata:
  name: test
  namespace: test-namespace
spec:
  pipelineSpec:
    params:
      - name: msg
        type: string
    results:
      - description: test result
        name: out
        value: REPLACED
    tasks:
      - name: task1
        params:
          - name: msg
            value: $(params.msg)
        taskRef:
          kind: Task
          name: sample-echo
      - name: task2
        params:
          - name: msg
            value: $(tasks.task1.results.out)
        runAfter:
          - task1
        taskRef:
          kind: Task
          name: sample-echo
`))
	p := internal.NewPipelineTestData("sample.pipeline")
	refs := pipelinerun.ResolvedResultRefs{
		&pipelinerun.ResolvedResultRef{
			ResultReference: pipelinev1beta1.ResultRef{
				PipelineTask: "task2",
				Result:       "out",
			},
			Value:       *pipelinev1beta1.NewArrayOrString("REPLACED"),
			FromTaskRun: "test-abcde-task1",
		},
	}
	pr := pipelinerun.NewPipelineRun(testObjectMeta(), pipelinerun.PipelineSpec(&p.Spec), pipelinerun.ApplyResultsToPipelineResults(refs))
	assert.Equal(t, want, pr)
}

func TestSkipTask(t *testing.T) {
	want := expect([]byte(`
metadata:
  name: test
  namespace: test-namespace
spec:
  pipelineSpec:
    params:
      - name: msg
        type: string
    results:
      - description: test result
        name: out
        value: $(tasks.task2.results.out)
    tasks:
      - name: task1
        params:
          - name: msg
            value: $(params.msg)
        taskRef:
          kind: Task
          name: sample-echo
        when:
          - input: skipped
            operator: in
            values: [""]
      - name: task2
        params:
          - name: msg
            value: $(tasks.task1.results.out)
        runAfter:
          - task1
        taskRef:
          kind: Task
          name: sample-echo
`))
	p := internal.NewPipelineTestData("sample.pipeline")
	pr := pipelinerun.NewPipelineRun(testObjectMeta(), pipelinerun.PipelineSpec(&p.Spec), pipelinerun.SkipTask("task1"))
	assert.Equal(t, want, pr)
}
