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

package controllers

import (
	"errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
)

type EventEmitter struct {
	Recorder record.EventRecorder
}

// EmitInfo creates new normal Event to indicate info message.
func (r *EventEmitter) EmitInfo(obj runtime.Object, reason, message string) {
	r.Recorder.Event(obj, corev1.EventTypeNormal, reason, message)
}

// EmitError creates new warning Event to indicate Error.
func (r *EventEmitter) EmitError(obj runtime.Object, err error) {
	if err == nil {
		return
	}
	st := GenErrorStatus(err)
	r.Recorder.Event(obj, corev1.EventTypeWarning, string(st.Reason), st.Message)
	return
}

// GenErrorStatus makes metav1.Status from error.
func GenErrorStatus(err error) *metav1.Status {
	if err == nil {
		return nil
	}
	e := &apierrors.StatusError{}
	if errors.As(err, &e) {
		st := e.Status()
		return &st
	} else {
		return &metav1.Status{
			Status:  metav1.StatusFailure,
			Message: err.Error(),
			Reason:  metav1.StatusReasonInternalError,
			Code:    500,
		}
	}
}
