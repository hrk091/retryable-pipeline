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
	"github.com/stretchr/testify/assert"
	pipelinev1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"knative.dev/pkg/apis"
	"testing"
)

func TestRetryablePipelineRunStatus_Conditions(t *testing.T) {

	type condBools struct {
		hasStarted   bool
		hasSucceeded bool
		hasDone      bool
		hasCancelled bool
	}

	var testCases = []struct {
		name  string
		given []byte
		want  condBools
	}{
		{
			name: "when RetryablePipelineRun started",
			given: []byte(`
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
`),
			want: condBools{
				hasStarted:   true,
				hasSucceeded: false,
				hasDone:      false,
				hasCancelled: false,
			},
		},
		{
			name: "when RetryablePipelineRun running",
			given: []byte(`
apiVersion: tekton.hrk091.dev/v1alpha1
kind: RetryablePipelineRun
metadata:
  name: test
status:
  startTime: "2022-05-06T13:38:06Z"
  conditions:
    - reason: Running
      status: Unknown
      type: Succeeded
`),
			want: condBools{
				hasStarted:   true,
				hasSucceeded: false,
				hasDone:      false,
				hasCancelled: false,
			},
		},
		{
			name: "when RetryablePipelineRun succeeded",
			given: []byte(`
apiVersion: tekton.hrk091.dev/v1alpha1
kind: RetryablePipelineRun
metadata:
  name: test
status:
  startTime: "2022-05-06T13:38:06Z"
  conditions:
    - reason: Succeeded
      status: "True"
      type: Succeeded
`),
			want: condBools{
				hasStarted:   true,
				hasSucceeded: true,
				hasDone:      true,
				hasCancelled: false,
			},
		},
		{
			name: "when RetryablePipelineRun completed",
			given: []byte(`
apiVersion: tekton.hrk091.dev/v1alpha1
kind: RetryablePipelineRun
metadata:
  name: test
status:
  startTime: "2022-05-06T13:38:06Z"
  conditions:
    - reason: Completed
      status: "True"
      type: Succeeded
`),
			want: condBools{
				hasStarted:   true,
				hasSucceeded: true,
				hasDone:      true,
				hasCancelled: false,
			},
		},
		{
			name: "when RetryablePipelineRun failed",
			given: []byte(`
apiVersion: tekton.hrk091.dev/v1alpha1
kind: RetryablePipelineRun
metadata:
  name: test
status:
  startTime: "2022-05-06T13:38:06Z"
  conditions:
    - reason: Failed
      status: "False"
      type: Succeeded
`),
			want: condBools{
				hasStarted:   true,
				hasSucceeded: false,
				hasDone:      true,
				hasCancelled: false,
			},
		},
		{
			name: "when RetryablePipelineRun cancelled",
			given: []byte(`
apiVersion: tekton.hrk091.dev/v1alpha1
kind: RetryablePipelineRun
metadata:
  name: test
status:
  startTime: "2022-05-06T13:38:06Z"
  conditions:
    - reason: Cancelled
      status: "False"
      type: Succeeded
`),
			want: condBools{
				hasStarted:   true,
				hasSucceeded: false,
				hasDone:      true,
				hasCancelled: true,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var o v1alpha1.RetryablePipelineRun
			if err := yaml.Unmarshal(tc.given, &o); err != nil {
				panic(err)
			}

			assert.Equal(t, tc.want.hasStarted, o.HasStarted())
			assert.Equal(t, tc.want.hasSucceeded, o.HasSucceeded())
			assert.Equal(t, tc.want.hasDone, o.HasDone())
			assert.Equal(t, tc.want.hasCancelled, o.HasCancelled())
		})
	}
}

func TestRetryablePipelineRunStatus_SetCondition(t *testing.T) {
	r := NewRetryablePipelineRun("test", PipelineRef("sample"), InitStatus())
	cond := apis.Condition{
		Type:   apis.ConditionSucceeded,
		Status: corev1.ConditionUnknown,
		Reason: pipelinev1beta1.PipelineRunReasonRunning.String(),
	}

	r.Status.SetCondition(&cond)

	act := r.Status.GetCondition()
	assert.Equal(t, cond.Type, act.Type)
	assert.Equal(t, cond.Status, act.Status)
	assert.Equal(t, cond.Reason, act.Reason)
}

func TestRetryablePipelineRunStatus_MarkRunning(t *testing.T) {
	r := NewRetryablePipelineRun("test", PipelineRef("sample"), InitStatus())

	r.Status.MarkRunning(pipelinev1beta1.PipelineRunReasonRunning.String(), "some message")

	assert.True(t, r.HasStarted())
	assert.False(t, r.HasSucceeded())
	assert.False(t, r.HasDone())
}

func TestRetryablePipelineRunStatus_MarkSucceeded(t *testing.T) {
	r := NewRetryablePipelineRun("test", PipelineRef("sample"), InitStatus())

	r.Status.MarkSucceeded(pipelinev1beta1.PipelineRunReasonSuccessful.String(), "some message")

	assert.True(t, r.HasStarted())
	assert.True(t, r.HasSucceeded())
	assert.True(t, r.HasDone())
}

func TestRetryablePipelineRunStatus_MarkFailed(t *testing.T) {
	r := NewRetryablePipelineRun("test", PipelineRef("sample"), InitStatus())

	r.Status.MarkFailed(pipelinev1beta1.PipelineRunReasonFailed.String(), "some message")

	assert.True(t, r.HasStarted())
	assert.False(t, r.HasSucceeded())
	assert.True(t, r.HasDone())
}
