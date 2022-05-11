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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"testing"
	"time"
)

func TestConditions(t *testing.T) {

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

func TestReserveNextPipelineRunName(t *testing.T) {
	b := []byte(`
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
`)
	var o v1alpha1.RetryablePipelineRun
	if err := yaml.Unmarshal(b, &o); err != nil {
		panic(err)
	}

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
