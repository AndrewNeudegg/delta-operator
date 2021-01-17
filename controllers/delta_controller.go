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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

const (
	// AnnotationOperatorManaged specifies that the the given resource is managed by this operator.
	AnnotationOperatorManaged = "delta.routing.andrewneudegg.com/managed"
	// AnnotationDeltaName specifies the associated name of the given delta crd.
	AnnotationDeltaName = "delta.routing.andrewneudegg.com/name"
)

// DeltaReconciler reconciles a Delta object
type DeltaReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

type DeploymentReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=routing.andrewneudegg.com,resources=delta,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=routing.andrewneudegg.com,resources=delta/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=routing.andrewneudegg.com,resources=delta/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps/v1,resources=deployment,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps/v1,resources=deployment/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps/v1,resources=deployment/finalizers,verbs=update

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
	log.Info("starting reconciliation for deltas")

	// test if this is a delete operation.
	// if its a delete we should remove the associated deployment.
	deltaList := &routingv1.DeltaList{}
	err := r.List(ctx, deltaList, &client.ListOptions{
		Namespace: req.Namespace,
	})
	if err != nil {
		log.Error(err, "stopping, could not list deltas in ", req.Namespace)
		return ctrl.Result{}, nil
	}

	var deltaItem *routingv1.Delta
	for _, dI := range deltaList.Items {
		if dI.Name == req.Name {
			deltaItem = &dI
		}
	}
	if deltaItem == nil {
		log.Info("was deleted, stopping")
		return ctrl.Result{}, nil
	}

	// get the delta that we are operating on.
	delta := &routingv1.Delta{}
	err = r.Get(ctx, req.NamespacedName, delta)
	if err != nil {
		log.Error(err, fmt.Sprintf("could not find '%s'", req.NamespacedName))
		return ctrl.Result{
			Requeue:      false,
			RequeueAfter: 0,
		}, nil
	}

	// list deployments.
	deploymentList := &appsv1.DeploymentList{}
	err = r.List(ctx, deploymentList, &client.ListOptions{
		Namespace: req.Namespace,
	})
	if err != nil {
		log.Error(err, "stopping, could not list deployments in ", req.Namespace)
		return ctrl.Result{}, nil
	}

	associatedDeployment := &appsv1.Deployment{}
	for _, deployment := range deploymentList.Items {
		if _, ok := deployment.Annotations[AnnotationOperatorManaged]; !ok {
			continue
		}

		if val, ok := deployment.Annotations[AnnotationDeltaName]; ok {
			if val == delta.Spec.Name {
				associatedDeployment = &deployment
				break
			}
		}
	}

	if associatedDeployment.Name == "" {
		deltaTemplate := delta.Spec.Template
		deltaTemplate.Annotations = map[string]string{
			AnnotationOperatorManaged: "true",
			AnnotationDeltaName:       delta.Spec.Name,
		}
		deltaTemplate.Labels = map[string]string{
			"delta.routing.andrewneudegg.com": "true",
		}
		deltaTemplate.Name = delta.Name
		deltaTemplate.ObjectMeta = metav1.ObjectMeta{
			Name:         delta.Spec.Name,
			GenerateName: delta.Spec.Name,
			Namespace:    req.Namespace,

			Labels: map[string]string{
				"delta.routing.andrewneudegg.com": "true",
			},
			Annotations: map[string]string{
				AnnotationOperatorManaged: "true",
				AnnotationDeltaName:       delta.Spec.Name,
			},
			OwnerReferences: []metav1.OwnerReference{},
		}

		newDeployment := appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:                       delta.Spec.Name,
				GenerateName:               delta.Spec.Name,
				Namespace:                  req.Namespace,
				DeletionGracePeriodSeconds: new(int64),
				Labels: map[string]string{
					"delta.routing.andrewneudegg.com": "true",
				},
				Annotations: map[string]string{
					AnnotationOperatorManaged: "true",
					AnnotationDeltaName:       delta.Spec.Name,
				},
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: &delta.Spec.Count,
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"delta.routing.andrewneudegg.com": "true",
					},
				},
				Template: deltaTemplate,
				Strategy: appsv1.DeploymentStrategy{
					Type: "RollingUpdate",
					RollingUpdate: &appsv1.RollingUpdateDeployment{
						MaxUnavailable: &intstr.IntOrString{IntVal: 1},
					},
				},
			},
		}

		err = r.Create(ctx, &newDeployment, &client.CreateOptions{})
		if err != nil {
			log.Error(err, "failed to create deployment")
			return ctrl.Result{}, nil
		}
	} else {
		log.Info("this operator does not support patching deployments at this time")
	}

	// if associatedDeployment == nil then we need to create this object,
	// else we should patch the object.

	return ctrl.Result{}, nil
}

// Reconcile deployments that this operator may be watching.
func (r *DeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("deployments", req.NamespacedName)
	log.Info("starting reconciliation for deployments")
	log.Info("TODO: do something on deployment update...")
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

// SetupWithManager sets up the controller with the Manager.
func (r *DeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.Deployment{}).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: 2,
		}).
		Complete(r)
}
