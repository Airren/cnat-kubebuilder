/*
Copyright 2022 ByteGopher.

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

	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/pkg/apis/clientauthentication/install"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	cnatv1alpha1 "www.github.com/airren/cnat-kubebuilder/api/v1alpha1"
)

// AtReconciler reconciles a At object
type AtReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=cnat.bytegopher.com,resources=ats,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cnat.bytegopher.com,resources=ats/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cnat.bytegopher.com,resources=ats/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the At object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.1/pkg/reconcile
func (r *AtReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// TODO(user): your logic here

	reqLogger := log.Log.WithValues("namespace", req.Namespace, "at", req.Name)
	reqLogger.Info("=== Reconciling At")

	// Fetch the At instance
	instance := &cnatv1alpha1.At{}
	err := r.Get(context.TODO(), req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile
			// request-return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object-requeue the request:
		return reconcile.Result{}, err
	}

	// If no phase set, default to pending (the initial phase)
	if instance.Status.Phase == "" {
		instance.Status.Phase = cnatv1alpha1.PhasePending
	}

	// Now let's make the main case distinction: implementing the state diagram
	// PENDING -> RUNNING -> DONE
	switch instance.Status.Phase {
	case cnatv1alpha1.PhasePending:
		reqLogger.Info("Phase: PENDING")
		// As long as we haven't executed the command yet, we need to check if 
		// it's already time to act
		reqLogger.Info("Checking schedule","Target", instance.Spec.Schedule)
		// Check if it's already time to execute the command with a tolerance of 
		// 2 second
		d, err := timeUtilSchedule(instance.Spec.Schedule)
		if err != nil{
			reqLogger.Error(err, "Schedule parsing failure")
			// Error reading the schedule. Wait until it is fixed.
			return reconcile.Result{},err
		}
		reqLogger.Info("Schedule parsing done","Result","diff",fmt.Sprintf("%v",d))

		if d >0{
			// Not yet time to execute the command, wait until the schedule time
			return reconcile.Result{RequeueAfter: d},nil
		}

		reqLogger.Info("It's time!", "Ready to execute", instance.Spec.Command)
		instance.Status.Phase = cnatv1alpha1.PhaseRunning

	case cnatv1alpha1.PhaseRunning:
		reqLogger.Info("Phase: RUNNING")
		pod := newPodForCR(instance)
		// Set At instance as the owner and controller
		err := controllerutil.SetControllerReference(instance, pod, r.Scheme) 
		if err != nil{
			// requeue with error
			return reconcile.Result{},err
		}
		found := &corev1.Pod{}
		nsName := types.NamespacedName{
			Name: pod.Name,
			Namespace: pod.Namespace,
		}
		err := r.Get(context.TODO(), nsName, found)
		// Try to see if the pod already exist and if not
		// (which we expect) then create a noe-shot pod as per spec
		if err!= nil && errors.IsNotFound(err){
			err = r.Create(context.TODO(), pod)
			if err != nil{
				// requeue with error
				return reconcile.Result{},err
			}
		} else if err != nil{
			// requeue with error
			return reconcile.Result{},err
		} else if found.Status.Phase == corev1.PodFaild || found.Status.Phase == corev1.PodSucceed{
			reqLogger.Info("Container terminated","reason", found.Status.Reason,
		"message", found.Status.Message)
		} else{
			// 
		}



		 

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AtReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cnatv1alpha1.At{}).
		Complete(r)
}
