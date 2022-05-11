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
	pipelinev1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
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
	l.Info("start reconciliation")

	var rpr rprv1alpha1.RetryablePipelineRun
	if err := r.Get(ctx, req.NamespacedName, &rpr); err != nil {
		l.Error(err, "unable to fetch RetryablePipelineRun")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if rpr.HasSucceeded() {
		l.Info("already succeeded, nothing to do any more.")
		return ctrl.Result{}, nil
	}

	if !rpr.HasStarted() {
		l.Info("initialization started")
		rpr.InitializeStatus()
		rpr.ReserveNextPipelineRunName()
		if err := r.Status().Update(ctx, &rpr); err != nil {
			l.Error(err, "unable to update RetryablePipelineRun Status")
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
		l.Info("initialization completed. Requeueing...")
		return ctrl.Result{Requeue: true}, nil
	}

	// TODO check retry annotation and reserve PipelineRun name

	l.Info("listing PipelineRun")
	var prs pipelinev1beta1.PipelineRunList
	if err := r.List(ctx, &prs, &client.ListOptions{
		LabelSelector: rpr.ChildLabels().AsSelector(),
		Namespace:     rpr.Namespace,
	}); err != nil {
		l.Error(err, "unable to list PipelineRun")
		return ctrl.Result{}, err
	}

	l.Info("update RetryablePipelineRun status")
	for _, pr := range prs.Items {
		s := rpr.PipelineRunStatus(pr.Name)
		if s == nil {
			l.Info("not belonging PipelineRun found", "name", pr.Name)
			continue
		}
		s.CopyFrom(&pr.Status)
	}
	rpr.AggregateChildrenResults()

	if err := r.Status().Update(ctx, &rpr); err != nil {
		l.Error(err, "unable to list RetryablePipelineRun Status")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	l.Info("updated RetryablePipelineRun status")

	name := rpr.NextPipelineRunName()
	if name != "" {
		l.Info("creating PipelineRun")
		// TODO generate new spec for retry
		pr := rpr.NewPipelineRun(name, rpr.Spec.PipelineRunSpec)

		if err := r.createPipelineRun(ctx, pr, &rpr); err != nil {
			l.Error(err, "unable to create PipelineRun")
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
		l.Info("created pipeline")
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RetryablePipelineRunReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&rprv1alpha1.RetryablePipelineRun{}).
		Owns(&pipelinev1beta1.PipelineRun{}).
		Complete(r)
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
