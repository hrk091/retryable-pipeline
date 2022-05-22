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
	"github.com/stretchr/testify/assert"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"knative.dev/pkg/apis"
	"testing"
)

func TestTaskRunStatusCount_IsRunning(t *testing.T) {
	var testcases = []struct {
		name  string
		given *v1alpha1.TaskRunStatusCount
		want  bool
	}{
		{
			name:  "when not initialized",
			given: nil,
			want:  false,
		},
		{
			name: "when some tasks are incomplete",
			given: &v1alpha1.TaskRunStatusCount{
				Incomplete: 1,
			},
			want: true,
		},
		{
			name: "when some tasks are not started",
			given: &v1alpha1.TaskRunStatusCount{
				NotStarted: 1,
			},
			want: true,
		},
		{
			name:  "when no tasks are incomplete",
			given: &v1alpha1.TaskRunStatusCount{},
			want:  false,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.given.IsRunning())
		})
	}
}

func TestTaskRunStatusCount_IsFailed(t *testing.T) {
	var testcases = []struct {
		name  string
		given *v1alpha1.TaskRunStatusCount
		want  bool
	}{
		{
			name:  "when not initialized",
			given: nil,
			want:  false,
		},
		{
			name: "when some tasks are incomplete",
			given: &v1alpha1.TaskRunStatusCount{
				Incomplete: 1,
			},
			want: false,
		},
		{
			name: "when some tasks are not started",
			given: &v1alpha1.TaskRunStatusCount{
				NotStarted: 1,
			},
			want: false,
		},
		{
			name: "when no tasks are incomplete but some tasks are failed",
			given: &v1alpha1.TaskRunStatusCount{
				Failed: 1,
			},
			want: true,
		},
		{
			name:  "when no tasks are incomplete and no tasks are failed",
			given: &v1alpha1.TaskRunStatusCount{},
			want:  false,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.given.IsFailed())
		})
	}
}

func TestTaskRunStatusCount_IsSucceeded(t *testing.T) {
	var testcases = []struct {
		name  string
		given *v1alpha1.TaskRunStatusCount
		want  bool
	}{
		{
			name:  "when not initialized",
			given: nil,
			want:  false,
		},
		{
			name: "when some tasks are incomplete",
			given: &v1alpha1.TaskRunStatusCount{
				Incomplete: 1,
			},
			want: false,
		},
		{
			name: "when some tasks are not started",
			given: &v1alpha1.TaskRunStatusCount{
				NotStarted: 1,
			},
			want: false,
		},
		{
			name: "when no tasks are incomplete but some tasks are failed",
			given: &v1alpha1.TaskRunStatusCount{
				Failed: 1,
			},
			want: false,
		},
		{
			name: "when no tasks are incomplete but some tasks are cancelled",
			given: &v1alpha1.TaskRunStatusCount{
				Cancelled: 1,
			},
			want: false,
		},
		{
			name:  "when no tasks are incomplete and no tasks are neither failed nor cancelled",
			given: &v1alpha1.TaskRunStatusCount{},
			want:  true,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.given.IsSucceeded())
		})
	}
}

func TestTaskRunStatusCount_IsCancelled(t *testing.T) {
	var testcases = []struct {
		name  string
		given *v1alpha1.TaskRunStatusCount
		want  bool
	}{
		{
			name:  "when not initialized",
			given: nil,
			want:  false,
		},
		{
			name: "when some tasks are incomplete",
			given: &v1alpha1.TaskRunStatusCount{
				Incomplete: 1,
			},
			want: false,
		},
		{
			name: "when some tasks are not started",
			given: &v1alpha1.TaskRunStatusCount{
				NotStarted: 1,
			},
			want: false,
		},
		{
			name: "when no tasks are incomplete but some tasks are cancelled",
			given: &v1alpha1.TaskRunStatusCount{
				Cancelled: 1,
			},
			want: true,
		},
		{
			name:  "when no tasks are incomplete and no tasks are cancelled",
			given: &v1alpha1.TaskRunStatusCount{},
			want:  false,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.given.IsCancelled())
		})
	}
}

func TestNewReducedPipelineRunCondition_notInitialized(t *testing.T) {
	rpr := NewRetryablePipelineRun("test", PipelineRef("sample"), InitStatus())
	c := v1alpha1.NewReducedPipelineRunCondition(rpr)
	assert.Nil(t, c)
}

func TestNewReducedPipelineRunCondition_noPipelineRunStatuses(t *testing.T) {
	rpr := internal.NewRetryablePipelineRunTestData("sample.rpr")
	rpr.Status.PipelineRuns = []*v1alpha1.PartialPipelineRunStatus{}
	c := v1alpha1.NewReducedPipelineRunCondition(rpr)

	assert.Nil(t, c.Tasks["task1"].Condition)
	assert.False(t, c.Tasks["task1"].Started)
	assert.False(t, c.Tasks["task1"].Skipped)

	assert.Nil(t, c.Tasks["task2"].Condition)
	assert.False(t, c.Tasks["task2"].Started)
	assert.False(t, c.Tasks["task2"].Skipped)
}

func TestNewReducedPipelineRunCondition_withPipelineRunStatuses(t *testing.T) {
	rpr := internal.NewRetryablePipelineRunTestData("sample.rpr")
	c := v1alpha1.NewReducedPipelineRunCondition(rpr)

	assert.True(t, c.Tasks["task1"].Condition.IsTrue())
	assert.True(t, c.Tasks["task1"].Started)
	assert.False(t, c.Tasks["task1"].Skipped)

	assert.True(t, c.Tasks["task2"].Condition.IsTrue())
	assert.True(t, c.Tasks["task2"].Started)
	assert.False(t, c.Tasks["task2"].Skipped)
}

func TestNewReducedPipelineRunCondition_Stats(t *testing.T) {
	testCases := []struct {
		name  string
		given *v1alpha1.ReducedPipelineRunCondition
		want  *v1alpha1.TaskRunStatusCount
	}{
		{
			name:  "when not initialized",
			given: nil,
			want:  nil,
		},
		{
			name: "when running",
			given: &v1alpha1.ReducedPipelineRunCondition{
				Tasks: map[string]v1alpha1.ReducedTaskRunCondition{
					"task1": {
						Skipped: true,
					},
					"task2": {
						Condition: &apis.Condition{
							Type:   apis.ConditionSucceeded,
							Status: corev1.ConditionTrue,
							Reason: v1beta1.TaskRunReasonSuccessful.String(),
						},
						Started: true,
					},
					"task3": {
						Condition: &apis.Condition{
							Type:   apis.ConditionSucceeded,
							Status: corev1.ConditionUnknown,
							Reason: v1beta1.TaskRunReasonRunning.String(),
						},
						Started: true,
					},
					"task4": {},
				},
			},
			want: &v1alpha1.TaskRunStatusCount{
				All:        4,
				Skipped:    1,
				NotStarted: 1,
				Incomplete: 1,
				Succeeded:  1,
			},
		},
		{
			name: "when succeeded",
			given: &v1alpha1.ReducedPipelineRunCondition{
				Tasks: map[string]v1alpha1.ReducedTaskRunCondition{
					"task1": {
						Skipped: true,
					},
					"task2": {
						Condition: &apis.Condition{
							Type:   apis.ConditionSucceeded,
							Status: corev1.ConditionTrue,
							Reason: v1beta1.TaskRunReasonSuccessful.String(),
						},
						Started: true,
					},
					"task3": {
						Condition: &apis.Condition{
							Type:   apis.ConditionSucceeded,
							Status: corev1.ConditionTrue,
							Reason: v1beta1.TaskRunReasonSuccessful.String(),
						},
						Started: true,
					},
					"task4": {
						Skipped: true,
					},
				},
			},
			want: &v1alpha1.TaskRunStatusCount{
				All:       4,
				Skipped:   2,
				Succeeded: 2,
			},
		},
		{
			name: "when failed",
			given: &v1alpha1.ReducedPipelineRunCondition{
				Tasks: map[string]v1alpha1.ReducedTaskRunCondition{
					"task1": {
						Skipped: true,
					},
					"task2": {
						Condition: &apis.Condition{
							Type:   apis.ConditionSucceeded,
							Status: corev1.ConditionTrue,
							Reason: v1beta1.TaskRunReasonSuccessful.String(),
						},
						Started: true,
					},
					"task3": {
						Condition: &apis.Condition{
							Type:   apis.ConditionSucceeded,
							Status: corev1.ConditionFalse,
							Reason: v1beta1.TaskRunReasonFailed.String(),
						},
						Started: true,
					},
					"task4": {},
				},
			},
			want: &v1alpha1.TaskRunStatusCount{
				All:        4,
				Skipped:    1,
				NotStarted: 1,
				Succeeded:  1,
				Failed:     1,
			},
		},
		{
			name: "when cancelled",
			given: &v1alpha1.ReducedPipelineRunCondition{
				Tasks: map[string]v1alpha1.ReducedTaskRunCondition{
					"task1": {
						Skipped: true,
					},
					"task2": {
						Condition: &apis.Condition{
							Type:   apis.ConditionSucceeded,
							Status: corev1.ConditionTrue,
							Reason: v1beta1.TaskRunReasonSuccessful.String(),
						},
						Started: true,
					},
					"task3": {
						Condition: &apis.Condition{
							Type:   apis.ConditionSucceeded,
							Status: corev1.ConditionFalse,
							Reason: v1beta1.TaskRunReasonCancelled.String(),
						},
						Started: true,
					},
					"task4": {},
				},
			},
			want: &v1alpha1.TaskRunStatusCount{
				All:        4,
				Skipped:    1,
				NotStarted: 1,
				Succeeded:  1,
				Cancelled:  1,
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.given.Stats())
		})
	}
}
