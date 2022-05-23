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
	"context"
	"fmt"
	pipelinev1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	rprv1alpha1 "github.com/hrk091/retryable-pipeline/api/v1alpha1"
)

// RetryablePipelineRunReconciler reconciles a RetryablePipelineRun object
type RetryablePipelineRunReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	EventEmitter
}

//+kubebuilder:rbac:groups=tekton.hrk091.dev,resources=retryablepipelineruns,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=tekton.hrk091.dev,resources=retryablepipelineruns/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=tekton.hrk091.dev,resources=retryablepipelineruns/finalizers,verbs=update
//+kubebuilder:rbac:groups=tekton.dev,resources=pipelineruns,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=persistentvolumeclaims,verbs=create;delete
//+kubebuilder:rbac:groups=core,resources=events,verbs=create;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *RetryablePipelineRunReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)
	l.Info("reconciliation started")

	var rpr rprv1alpha1.RetryablePipelineRun
	if err := r.Get(ctx, req.NamespacedName, &rpr); err != nil {
		l.Error(err, "unable to fetch RetryablePipelineRun")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if rpr.Status.HasSucceeded() {
		l.Info("already succeeded, nothing to do any more.")
		return ctrl.Result{}, nil
	}

	if !rpr.Status.HasStarted() {
		l.Info("initialization started")
		rpr.InitializeStatus()
		rpr.ReserveNextPipelineRunName()
		if err := r.Status().Update(ctx, &rpr); err != nil {
			l.Error(err, "unable to update RetryablePipelineRun Status")
			r.EmitError(&rpr, err)
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
		l.Info("initialization completed. Requeueing...")
		return ctrl.Result{Requeue: true}, nil
	}

	if rpr.IsRetryKeyChanged() {
		l.Info("retry setup started")
		if rpr.ReserveNextPipelineRunName() {
			rpr.CopyRetryKey()
			if err := r.Status().Update(ctx, &rpr); err != nil {
				l.Error(err, "unable to update RetryablePipelineRun Status")
				r.EmitError(&rpr, err)
				return ctrl.Result{}, client.IgnoreNotFound(err)
			}
			l.Info("retry setup completed. Requeueing...")
			return ctrl.Result{Requeue: true}, nil
		} else {
			l.Info("there is still running PipelineRun. retry skipped.")
		}
	}

	if pr := r.makeNextPipelineRunIfReady(&rpr); pr != nil {
		l.Info("creating new PipelineRun", "PipelineRun", pr)
		if err := r.createPipelineRun(ctx, pr, &rpr); err == nil {
			r.EmitInfo(&rpr, "PipelineRun Created", fmt.Sprintf("Name: %s", pr.Name))
		} else if apierrors.IsAlreadyExists(err) {
			l.Info("PipelineRun is already created", "name", pr.Name)
		} else {
			l.Error(err, "unable to create PipelineRun")
			r.EmitError(&rpr, err)
			return ctrl.Result{}, err
		}
	}

	var prs pipelinev1beta1.PipelineRunList
	if err := r.List(ctx, &prs, &client.ListOptions{
		LabelSelector: rpr.ChildLabels().AsSelector(),
		Namespace:     rpr.Namespace,
	}); err != nil {
		l.Error(err, "unable to list PipelineRun")
		r.EmitError(&rpr, err)
		return ctrl.Result{}, err
	}

	if !rpr.IsPipelineSpecPinned() {
		l.Info("pin PipelineSpec")
		if len(prs.Items) == 0 {
			l.Info("PipelineRun is not created yet")
			return ctrl.Result{}, nil
		}
		if ok := rpr.PinPipelineSpecFrom(&prs.Items[0]); !ok {
			l.Info("PipelineSpec is not resolved yet")
			return ctrl.Result{}, nil
		}
	}
	for _, pt := range rpr.Status.PinnedPipelineRun.Status.PipelineSpec.Tasks {
		if ok, err := rpr.IsTaskSpecPinned(pt.Name); err != nil {
			l.Info("unable to pin TaskSpec since PipelineSpec is not resolved yet", "PipelineTask", pt.Name)
		} else if !ok {
			l.Info("pin TaskSpec from the newest PipelineRun", "PipelineTask", pt.Name)
			_, _ = rpr.PinTaskSpecFrom(&prs.Items[len(prs.Items)-1], pt.Name)
		}
	}

	oldStats := rprv1alpha1.NewReducedPipelineRunCondition(&rpr).Stats()
	for _, pr := range prs.Items {
		s := rpr.PipelineRunStatus(pr.Name)
		if s == nil {
			l.Info("not belonging PipelineRun found", "name", pr.Name)
			continue
		}
		s.CopyFrom(&pr.Status)
	}
	rpr.AggregateChildrenResults()

	newStats := rpr.UpdateCondition()
	if newStats.ChangedFrom(oldStats) {
		r.EmitInfo(&rpr, rpr.Status.GetCondition().Reason, newStats.Info())
	}

	if err := r.Status().Update(ctx, &rpr); err != nil {
		l.Error(err, "unable to update RetryablePipelineRun Status")
		r.EmitError(&rpr, err)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	l.Info("reconciliation ended")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RetryablePipelineRunReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&rprv1alpha1.RetryablePipelineRun{}).
		Owns(&pipelinev1beta1.PipelineRun{}).
		Complete(r)
}

func (r *RetryablePipelineRunReconciler) makeNextPipelineRunIfReady(rpr *rprv1alpha1.RetryablePipelineRun) *pipelinev1beta1.PipelineRun {
	name := rpr.NextPipelineRunName()
	if name == "" {
		return nil
	}
	if rpr.StartedPipelineRunCount() == 0 {
		return rpr.NewPipelineRun(name)
	}
	if !rpr.Status.HasDone() {
		return nil
	}
	return rpr.NewRetryPipelineRun(name)
}

func (r *RetryablePipelineRunReconciler) createPipelineRun(ctx context.Context, pr *pipelinev1beta1.PipelineRun, rpr *rprv1alpha1.RetryablePipelineRun) error {
	l := log.FromContext(ctx)
	if err := ctrl.SetControllerReference(rpr, pr, r.Scheme); err != nil {
		l.Error(err, "unable to set controller reference to PipelineRun")
		return err
	}
	if err := r.Create(ctx, pr); err != nil {
		l.Error(err, "unable to create PipelineRun")
		return err
	}
	l.Info("created a PipelineRun")
	return nil
}
