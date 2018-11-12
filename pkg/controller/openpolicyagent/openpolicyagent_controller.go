/*
Copyright 2018 Replicated.

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

package openpolicyagent

import (
	"context"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	controllersv1alpha1 "github.com/replicatedhq/gatekeeper/pkg/apis/controllers/v1alpha1"
	"github.com/replicatedhq/gatekeeper/pkg/logger"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// Add creates a new OpenPolicyAgent Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileOpenPolicyAgent{
		Client:     mgr.GetClient(),
		scheme:     mgr.GetScheme(),
		Logger:     logger.New(),
		RestConfig: mgr.GetConfig(),
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("openpolicyagent-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to OpenPolicyAgent
	err = c.Watch(&source.Kind{Type: &controllersv1alpha1.OpenPolicyAgent{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Uncomment watch a Deployment created by OpenPolicyAgent - change this for objects you create
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &controllersv1alpha1.OpenPolicyAgent{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileOpenPolicyAgent{}

// ReconcileOpenPolicyAgent reconciles a OpenPolicyAgent object
type ReconcileOpenPolicyAgent struct {
	client.Client
	scheme     *runtime.Scheme
	Logger     log.Logger
	RestConfig *rest.Config
}

// Reconcile reads that state of the cluster for a OpenPolicyAgent object and makes changes based on the state read
// and what is in the OpenPolicyAgent.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  The scaffolding writes
// a Deployment as an example
// Automatically generate RBAC rules to allow the Controller to read and write Deployments
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=controllers.replicated.com,resources=openpolicyagents,verbs=get;list;watch;create;update;patch;delete
func (r *ReconcileOpenPolicyAgent) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the OpenPolicyAgent instance
	instance := &controllersv1alpha1.OpenPolicyAgent{}
	err := r.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	if err := r.reconcileOpenPolicyAgent(instance); err != nil {
		level.Error(r.Logger).Log("event", "reconcileOpenPolicyAgent", "err", err)
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}
