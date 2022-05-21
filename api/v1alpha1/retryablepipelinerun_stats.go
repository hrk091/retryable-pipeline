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
	"fmt"
	pipelinev1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"knative.dev/pkg/apis"
)

type ReducedPipelineRunCondition struct {
	Tasks map[string]ReducedTaskRunCondition
}

type ReducedTaskRunCondition struct {
	Condition *apis.Condition
	Started   bool
	Skipped   bool
}

func NewReducedPipelineRunCondition(rpr *RetryablePipelineRun) *ReducedPipelineRunCondition {
	c := &ReducedPipelineRunCondition{
		Tasks: map[string]ReducedTaskRunCondition{},
	}
	ns, err := rpr.PipelineTaskNames()
	if err != nil {
		return nil
	}
	for _, n := range ns {
		c.Tasks[n] = ReducedTaskRunCondition{}
	}
	for _, prs := range rpr.Status.PipelineRuns {
		c.collectCondition(prs)
	}
	return c
}

func (c *ReducedPipelineRunCondition) collectCondition(prs *PartialPipelineRunStatus) {
	for _, st := range prs.SkippedTasks {
		if pt, ok := c.Tasks[st.Name]; ok && !pt.Started {
			c.Tasks[st.Name] = ReducedTaskRunCondition{
				Condition: nil,
				Skipped:   true,
			}
		}
	}

	for _, tr := range prs.TaskRuns {
		c.Tasks[tr.PipelineTaskName] = ReducedTaskRunCondition{
			Condition: tr.Status.GetCondition(apis.ConditionSucceeded),
			Started:   true,
		}
	}
}

func (c *ReducedPipelineRunCondition) Stats() *TaskRunStatusCount {
	if c == nil {
		return nil
	}
	s := &TaskRunStatusCount{}
	for _, t := range c.Tasks {
		s.All++
		cond := t.Condition
		switch {
		case t.Skipped:
			s.Skipped++
		case !t.Started:
			s.NotStarted++
		case cond.IsTrue():
			s.Succeeded++
		case cond.IsFalse() && cond.Reason == pipelinev1beta1.TaskRunReasonCancelled.String():
			s.Cancelled++
		case cond.IsFalse():
			s.Failed++
		default:
			s.Incomplete++
		}
	}
	return s
}

type TaskRunStatusCount struct {
	All        int
	Skipped    int
	NotStarted int
	Incomplete int
	Succeeded  int
	Failed     int
	Cancelled  int
}

func (c *TaskRunStatusCount) Info() string {
	if c == nil {
		return "Not initialized yet"
	}
	return fmt.Sprintf("Tasks Completed: %d (Failed: %d, Cancelled %d), Skipped: %d", c.Succeeded, c.Failed, c.Cancelled, c.Skipped)
}

func (c *TaskRunStatusCount) IsRunning() bool {
	if c == nil {
		return false
	}
	return c.NotStarted > 0 || c.Incomplete > 0
}

func (c *TaskRunStatusCount) IsFailed() bool {
	if c == nil {
		return false
	}
	return !c.IsRunning() && c.Failed > 0
}

func (c *TaskRunStatusCount) IsCancelled() bool {
	if c == nil {
		return false
	}
	return !c.IsRunning() && c.Cancelled > 0
}

func (c *TaskRunStatusCount) IsSucceeded() bool {
	if c == nil {
		return false
	}
	return !c.IsRunning() && c.Failed == 0 && c.Cancelled == 0
}
