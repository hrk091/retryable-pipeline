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
	"github.com/hrk091/retryable-pipeline/internal"
	"github.com/hrk091/retryable-pipeline/pkg/pipelinerun"
	"github.com/stretchr/testify/assert"
	pipelinev1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
	"time"
)

func TestRetryablePipelineRun_AggregateChildrenResults(t *testing.T) {
	o := internal.DecodeRPR([]byte(`
apiVersion: tekton.hrk091.dev/v1alpha1
kind: RetryablePipelineRun
metadata:
  name: test
  annotations:
    tekton.hrk091.dev/retry-key: foo
status:
  pipelineRuns:
  - name: sample-lfvd6-48wlj
    conditions:
    - lastTransitionTime: "2022-05-14T03:23:21Z"
      message: 'Tasks Completed: 2 (Failed: 1, Cancelled 0), Skipped: 0'
      reason: Failed
      status: "False"
      type: Succeeded
    taskRuns:
      sample-lfvd6-48wlj-task1:
        pipelineTaskName: task1
        status:
          conditions:
          - lastTransitionTime: "2022-05-14T03:21:02Z"
            message: All Steps have completed executing
            reason: Succeeded
            status: "True"
            type: Succeeded
          taskResults:
          - name: out
            value: foo
      sample-lfvd6-48wlj-task2:
        pipelineTaskName: task2
        status:
          conditions:
          - lastTransitionTime: "2022-05-14T03:21:02Z"
            message: All Steps have completed executing
            reason: Failed
            status: "False"
            type: Succeeded
  - name: sample-lfvd6-gwfj8
    conditions:
    - lastTransitionTime: "2022-05-14T03:27:57Z"
      message: 'Tasks Completed: 1 (Failed: 0, Cancelled 0), Skipped: 1'
      reason: Completed
      status: "True"
      type: Succeeded
    taskRuns:
      sample-lfvd6-gwfj8-task2:
        pipelineTaskName: task2
        status:
          conditions:
          - lastTransitionTime: "2022-05-14T03:27:57Z"
            message: All Steps have completed executing
            reason: Succeeded
            status: "True"
            type: Succeeded
          taskResults:
          - name: out
            value: bar
    `))
	want := map[string]*v1alpha1.CompletedTaskResults{
		"task1": {
			FromPipelineRun: "sample-lfvd6-48wlj",
			FromTaskRun:     "sample-lfvd6-48wlj-task1",
			Results: []pipelinev1beta1.TaskRunResult{
				{Name: "out", Value: "foo"},
			},
		},
		"task2": {
			FromPipelineRun: "sample-lfvd6-gwfj8",
			FromTaskRun:     "sample-lfvd6-gwfj8-task2",
			Results: []pipelinev1beta1.TaskRunResult{
				{Name: "out", Value: "bar"},
			},
		},
	}

	o.AggregateChildrenResults()
	assert.Equal(t, want, o.Status.TaskResults)
}

func TestRetryablePipelineRunStatus_ResolvedResultRefs(t *testing.T) {
	s := v1alpha1.RetryablePipelineRunStatus{
		TaskResults: map[string]*v1alpha1.CompletedTaskResults{
			"task1": {
				FromPipelineRun: "sample-lfvd6-48wlj",
				FromTaskRun:     "sample-lfvd6-48wlj-task1",
				Results: []pipelinev1beta1.TaskRunResult{
					{Name: "out", Value: "foo"},
				},
			},
			"task2": {
				FromPipelineRun: "sample-lfvd6-gwfj8",
				FromTaskRun:     "sample-lfvd6-gwfj8-task2",
				Results: []pipelinev1beta1.TaskRunResult{
					{Name: "out", Value: "bar"},
				},
			},
		},
	}
	want := pipelinerun.ResolvedResultRefs{
		{
			ResultReference: pipelinev1beta1.ResultRef{
				PipelineTask: "task1",
				Result:       "out",
			},
			Value:       *pipelinev1beta1.NewArrayOrString("foo"),
			FromTaskRun: "sample-lfvd6-48wlj-task1",
		},
		{
			ResultReference: pipelinev1beta1.ResultRef{
				PipelineTask: "task2",
				Result:       "out",
			},
			Value:       *pipelinev1beta1.NewArrayOrString("bar"),
			FromTaskRun: "sample-lfvd6-gwfj8-task2",
		},
	}
	assert.Equal(t, want, s.ResolvedResultRefs())
}

func TestRetryablePipelineRun_IsRetryKeyChanged(t *testing.T) {
	o := internal.DecodeRPR([]byte(`
apiVersion: tekton.hrk091.dev/v1alpha1
kind: RetryablePipelineRun
metadata:
  name: test
  annotations:
    tekton.hrk091.dev/retry-key: foo
`))
	o.InitializeStatus()
	assert.True(t, o.IsRetryKeyChanged())

	o.CopyRetryKey()
	assert.False(t, o.IsRetryKeyChanged())

	o.Annotations[v1alpha1.AnnKeyRetryKey] = "bar"
	assert.True(t, o.IsRetryKeyChanged())

	o.CopyRetryKey()
	assert.False(t, o.IsRetryKeyChanged())
}

func TestRetryablePipelineRun_ReserveNextPipelineRunName(t *testing.T) {
	o := internal.DecodeRPR([]byte(`
apiVersion: tekton.hrk091.dev/v1alpha1
kind: RetryablePipelineRun
metadata:
  name: test
status:
  startTime: "2022-05-06T13:38:06Z"
  conditions:
    - reason: Started
      status: Unknown
      type: Succeeded
`))

	// init state
	assert.Empty(t, o.NextPipelineRunName())
	assert.Equal(t, 0, o.StartedPipelineRunCount())

	// first
	ok := o.ReserveNextPipelineRunName()
	assert.True(t, ok)
	assert.NotEmpty(t, o.NextPipelineRunName())
	assert.Equal(t, 0, o.StartedPipelineRunCount())

	// first (repeat before PipelineRun creation)
	ok = o.ReserveNextPipelineRunName()
	assert.False(t, ok)
	assert.NotEmpty(t, o.NextPipelineRunName())
	assert.Equal(t, 0, o.StartedPipelineRunCount())

	// first PipelineRun start
	o.Status.PipelineRuns[0].StartTime = &metav1.Time{Time: time.Now()}
	assert.Empty(t, o.NextPipelineRunName())
	assert.Equal(t, 1, o.StartedPipelineRunCount())

	// second
	ok = o.ReserveNextPipelineRunName()
	assert.True(t, ok)
	assert.NotEmpty(t, o.NextPipelineRunName())
	assert.Equal(t, 1, o.StartedPipelineRunCount())

	// second (repeat before PipelineRun creation)
	ok = o.ReserveNextPipelineRunName()
	assert.False(t, ok)
	assert.NotEmpty(t, o.NextPipelineRunName())
	assert.Equal(t, 1, o.StartedPipelineRunCount())

	// second PipelineRun start
	o.Status.PipelineRuns[1].StartTime = &metav1.Time{Time: time.Now()}
	assert.Empty(t, o.NextPipelineRunName())
	assert.Equal(t, 2, o.StartedPipelineRunCount())
}
