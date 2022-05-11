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

// InitializeConditions will set all conditions in condset to unknown for the RetryablePipelineRun
// and set the started time to the current time.
func (rpr *RetryablePipelineRun) InitializeConditions() {
	conditionManager := condSet.Manage(rpr.Status)
	conditionManager.InitializeConditions()
	initialCondition := conditionManager.GetCondition(apis.ConditionSucceeded)
	initialCondition.Reason = pipelinev1beta1.PipelineRunReasonStarted.String()
	conditionManager.SetCondition(*initialCondition)
	rpr.Status.StartTime = &initialCondition.LastTransitionTime.Inner
}

// GetCondition returns the Condition of RetryablePipelineRun.
func (s *RetryablePipelineRunStatus) GetCondition() *apis.Condition {
	return condSet.Manage(s).GetCondition(apis.ConditionSucceeded)
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
	return rpr.Status.GetCondition().IsTrue()
}

// HasDone returns true if the RetryablePipelineRun has been succeeded/completed/failed.
func (rpr *RetryablePipelineRun) HasDone() bool {
	return !rpr.Status.GetCondition().IsUnknown()
}

// HasCancelled returns true if the RetryablePipelineRun has been cancelled.
func (rpr *RetryablePipelineRun) HasCancelled() bool {
	cond := rpr.Status.GetCondition()
	return cond.IsFalse() && cond.Reason == pipelinev1beta1.PipelineRunReasonCancelled.String()
}

// SetCondition sets the condition, unsetting previous conditions with the same type.
func (s *RetryablePipelineRunStatus) SetCondition(newCond *apis.Condition) {
	if newCond != nil {
		condSet.Manage(s).SetCondition(*newCond)
	}
}

// MarkRunning changes the Succeeded condition to Unknown with the provided reason and message.
func (s *RetryablePipelineRunStatus) MarkRunning(reason, format string, msgs ...interface{}) {
	condSet.Manage(s).MarkUnknown(apis.ConditionSucceeded, reason, format, msgs...)
}

// MarkSucceeded changes the Succeeded condition to True with the provided reason and message.
func (s *RetryablePipelineRunStatus) MarkSucceeded(reason, format string, msgs ...interface{}) {
	condSet.Manage(s).MarkTrueWithReason(apis.ConditionSucceeded, reason, format, msgs...)
	cond := s.GetCondition()
	s.CompletionTime = &cond.LastTransitionTime.Inner
}

// MarkFailed changes the Succeeded condition to False with the provided reason and message.
func (s *RetryablePipelineRunStatus) MarkFailed(reason, format string, msgs ...interface{}) {
	condSet.Manage(s).MarkFalse(apis.ConditionSucceeded, reason, format, msgs...)
	cond := s.GetCondition()
	s.CompletionTime = &cond.LastTransitionTime.Inner
}

// GetCondition returns the Condition of PartialPipelineRunStatus.
func (s *PartialPipelineRunStatus) GetCondition() *apis.Condition {
	return condSet.Manage(s).GetCondition(apis.ConditionSucceeded)
}
