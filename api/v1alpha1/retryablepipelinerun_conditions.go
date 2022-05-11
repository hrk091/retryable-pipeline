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

package v1alpha1

import (
	pipelinev1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"knative.dev/pkg/apis"
)

var condSet = apis.NewBatchConditionSet()

func (s *RetryablePipelineRunStatus) GetCondition(t apis.ConditionType) *apis.Condition {
	return condSet.Manage(s).GetCondition(t)
}

// HasStarted returns whether RetryablePipelineRun has valid start time set in its status.
func (rpr *RetryablePipelineRun) HasStarted() bool {
	if rpr.Status == nil {
		return false
	}
	return rpr.Status.StartTime != nil && !rpr.Status.StartTime.IsZero()
}

// HasSucceeded returns true if the RetryablePipelineRun has been succeeded.
func (rpr *RetryablePipelineRun) HasSucceeded() bool {
	return rpr.Status.GetCondition(apis.ConditionSucceeded).IsTrue()
}

// HasDone returns true if the RetryablePipelineRun has been succeeded/completed/failed.
func (rpr *RetryablePipelineRun) HasDone() bool {
	return !rpr.Status.GetCondition(apis.ConditionSucceeded).IsUnknown()
}

// HasCancelled returns true if the RetryablePipelineRun has been cancelled.
func (rpr *RetryablePipelineRun) HasCancelled() bool {
	cond := rpr.Status.GetCondition(apis.ConditionSucceeded)
	return cond.IsFalse() && cond.Reason == pipelinev1beta1.PipelineRunReasonCancelled.String()
}
