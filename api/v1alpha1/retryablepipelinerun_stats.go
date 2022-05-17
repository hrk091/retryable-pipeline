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
	Pending   bool
	Skipped   bool
}

func NewReducedPipelineRunCondition(rpr *RetryablePipelineRunStatus) *ReducedPipelineRunCondition {
	c := &ReducedPipelineRunCondition{
		Tasks: map[string]ReducedTaskRunCondition{},
	}
	for _, pt := range rpr.PinnedPipelineRun.Status.PipelineSpec.Tasks {
		c.Tasks[pt.Name] = ReducedTaskRunCondition{
			Pending: true,
		}
	}
	for _, prs := range rpr.PipelineRuns {
		c.collectCondition(prs)
	}
	return c
}

func (c *ReducedPipelineRunCondition) collectCondition(prs *PartialPipelineRunStatus) {
	for _, st := range prs.SkippedTasks {
		if pt, ok := c.Tasks[st.Name]; ok && pt.Pending {
			c.Tasks[st.Name] = ReducedTaskRunCondition{
				Condition: nil,
				Pending:   false,
				Skipped:   true,
			}
		}
	}

	for _, tr := range prs.TaskRuns {
		c.Tasks[tr.PipelineTaskName] = ReducedTaskRunCondition{
			Condition: tr.Status.GetCondition(apis.ConditionSucceeded),
			Pending:   false,
			Skipped:   false,
		}
	}
}

func (c *ReducedPipelineRunCondition) Stats() TaskRunStatusCount {
	s := TaskRunStatusCount{
		Skipped:    0,
		Incomplete: 0,
		Succeeded:  0,
		Failed:     0,
		Cancelled:  0,
	}
	// TODO count any other not started tasks into incomplete tasks
	for _, t := range c.Tasks {
		cond := t.Condition
		switch {
		case t.Skipped:
			s.Skipped++
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
	Skipped    int
	Incomplete int
	Succeeded  int
	Failed     int
	Cancelled  int
}

func (c TaskRunStatusCount) Info() string {
	return fmt.Sprintf("Tasks Completed: %d (Failed: %d, Cancelled %d), Skipped: %d", c.Succeeded, c.Failed, c.Cancelled, c.Skipped)
}

func (c TaskRunStatusCount) IsRunning() bool {
	return c.Incomplete > 0
}

func (c TaskRunStatusCount) IsFailed() bool {
	return !c.IsRunning() && c.Failed > 0
}

func (c TaskRunStatusCount) IsCancelled() bool {
	return !c.IsRunning() && c.Cancelled > 0
}

func (c TaskRunStatusCount) IsSucceeded() bool {
	return !c.IsRunning() && c.Failed == 0 && c.Cancelled == 0
}
