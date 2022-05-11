package v1alpha1_test

import (
	"github.com/hrk091/retryable-pipeline/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/util/yaml"
	"testing"
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
