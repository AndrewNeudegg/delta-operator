/*
Copyright 2021.

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

	routingv1 "github.com/andrewneudegg/delta-operator/api/v1"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

// DeltaReconciler reconciles a Delta object
type DeltaReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=routing.andrewneudegg.com,resources=delta,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=routing.andrewneudegg.com,resources=delta/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=routing.andrewneudegg.com,resources=delta/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Delta object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.0/pkg/reconcile
func (r *DeltaReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("delta", req.NamespacedName)

	deltaList := &routingv1.DeltaList{}
	err := r.List(ctx, deltaList, &client.ListOptions{
		LabelSelector: nil,
		FieldSelector: nil,
		Namespace:     req.Namespace,
		Limit:         0,
		Continue:      "",
	})
	if err != nil {
		log.Error(err, "could not list deltas in ", req.Namespace)
		return ctrl.Result{}, nil
	}

	var deltaItem *routingv1.Delta
	for _, dI := range deltaList.Items {
		if dI.Name == req.Name {
			deltaItem = &dI
		}
	}
	if deltaItem == nil {
		log.Info("was deleted")
		return ctrl.Result{}, nil
	}



	delta := &routingv1.Delta{}
	err = r.Get(ctx, req.NamespacedName, delta)
	if err != nil {
		log.Error(err, fmt.Sprintf("could not find '%s'", req.NamespacedName))
		return ctrl.Result{
			Requeue:      false,
			RequeueAfter: 0,
		}, nil
	}

	log.Info(delta.Spec.Config)
	// your logic here

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DeltaReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&routingv1.Delta{}).
		Owns(&appsv1.Deployment{}).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: 2,
		}).
		Complete(r)
}
