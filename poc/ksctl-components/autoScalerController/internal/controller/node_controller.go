/*
Copyright 2024.

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

package controller

import (
	"context"
	"log/slog"

	"github.com/fatih/color"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// NodeReconciler reconciles a Node object
type NodeReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=core,resources=nodes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=nodes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core,resources=nodes/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Node object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *NodeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)
	slog.Info("### Triggered Reconcile (NODE) ###")

	var node corev1.Node
	if err := r.Get(ctx, req.NamespacedName, &node); err != nil {
		l.Error(err, "unable to fetch Node")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	for _, condition := range node.Status.Conditions {
		slog.Info("Node condition", "type", condition.Type, "status", condition.Status, "nodeName", node.ObjectMeta.Name)

		if condition.Type == corev1.NodeDiskPressure && condition.Status == corev1.ConditionTrue {
			// Detected Disk Pressure - Handle scale up
			// l.Info("Disk pressure detected, handling scale up...", "node", req.NamespacedName)
			color.HiRed("Disk pressure")
			// Implement your logic to create a ConfigMap here
		}
		if condition.Type == corev1.NodeMemoryPressure && condition.Status == corev1.ConditionTrue {
			// Detected Disk Pressure - Handle scale up
			// l.Info("Disk pressure detected, handling scale up...", "node", req.NamespacedName)
			color.HiRed("Memory pressure")
			// Implement your logic to create a ConfigMap here
		}
	}
	slog.Info("###########################")

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NodeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Node{}).
		Complete(r)
}
