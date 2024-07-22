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
	"io"
	"os"

	"github.com/gookit/goutil/dump"
	applicationv1alpha1 "github.com/ksctl/ksctl/ksctl-components/operators/application/api/v1alpha1"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/logger"
	"github.com/ksctl/ksctl/pkg/types"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	log          types.LoggerFactory
	LogVerbosity = map[string]int{
		"DEBUG": -1,
		"":      0,
	}

	LogWriter io.Writer = os.Stdout
)

const stackFinalizer = "ksctl.com/stack-finalizer"

// StackReconciler reconciles a Stack object
type StackReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=application.ksctl.com,resources=stacks,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=application.ksctl.com,resources=stacks/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=application.ksctl.com,resources=stacks/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// the Stack object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.2/pkg/reconcile
func (r *StackReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	ctx = context.WithValue(ctx, consts.KsctlModuleNameKey, "ksctl-app-stack-controller")
	ctx = context.WithValue(ctx, consts.KsctlContextUserID, "ksctl-app-stack-controller")

	log.Debug(ctx, "Triggered Reconciliation")

	stack := new(applicationv1alpha1.Stack)

	if err := r.Get(ctx, req.NamespacedName, stack); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	input := types.KsctlApp{
		StackName: stack.Spec.StackName,
	}

	if stack.Spec.Overrides != nil {
		input.Overrides = make(map[string]map[string]any, 0)
		for key, value := range stack.Spec.Overrides {
			if converted, err := convertDataToBeConsumable(ctx, value); err != nil {
				return ctrl.Result{}, err
			} else {
				input.Overrides[string(key)] = converted
			}
		}
	}

	println("=============")
	dump.Println(input)
	println("=============")

	log.Debug(ctx, "Debugging", "name", stack.Name, "namespace", stack.Namespace)
	log.Debug(ctx, "stack Spec", "spec", stack.Spec)

	// if stack.DeletionTimestamp.IsZero() {
	// 	if !containsString(stack.ObjectMeta.Finalizers, stackFinalizer) {
	//
	// 		log.Debug(ctx, "adding finalizer", "finalizer", stackFinalizer)
	//
	// 		stack.ObjectMeta.Finalizers = append(stack.ObjectMeta.Finalizers, stackFinalizer)
	//
	// 		if err := r.Update(context.Background(), stack); err != nil {
	// 			return ctrl.Result{}, err
	// 		}
	// 	} else {
	//
	// 		ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	// 		rpcClient, conn, err := NewClient(ctx)
	// 		defer cancel()
	//
	// 		if _, ok := helpers.IsContextPresent(ctx, consts.KsctlTestFlagKey); ok {
	// 			defer func() {
	// 				if err := conn.Close(); err != nil {
	// 					log.Error("Connection failed to close", "Reason", err)
	// 				}
	// 			}()
	// 		}
	//
	// 		if err != nil {
	// 			log.Error("New RPC Client", "Reason", err)
	// 			stack.Status.ReasonOfFailure = err.Error()
	//
	// 			if _err := r.Update(context.Background(), stack); _err != nil {
	// 				log.Error("update failed", "Reason", _err)
	// 				return ctrl.Result{}, _err
	// 			}
	// 			return ctrl.Result{
	// 				RequeueAfter: 30 * time.Second,
	// 				Requeue:      true,
	// 			}, err
	// 		}
	//
	// 		if _err := InstallApps(ctx, rpcClient, stack.Spec.Components); _err != nil {
	// 			log.Error("InstallApp", "Reason", _err)
	// 			stack.Status.Success = false
	// 			stack.Status.ReasonOfFailure = _err.Error()
	//
	// 			if __err := r.Update(context.Background(), stack); __err != nil {
	// 				log.Error("update failed", "Reason", _err)
	// 				return ctrl.Result{}, __err
	// 			}
	// 			return ctrl.Result{}, _err
	// 		}
	//
	// 		stack.Status.Success = true
	//
	// 		if _err := r.Update(context.Background(), stack); _err != nil {
	// 			log.Error("update failed", "Reason", _err)
	// 			return ctrl.Result{}, _err
	// 		}
	//
	// 		log.Success(ctx, "Install Application was successful")
	// 	}
	//
	// } else {
	// 	if containsString(stack.ObjectMeta.Finalizers, stackFinalizer) {
	//
	// 		ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	// 		rpcClient, conn, err := NewClient(ctx)
	// 		defer cancel()
	// 		if _, ok := helpers.IsContextPresent(ctx, consts.KsctlTestFlagKey); ok {
	// 			defer func() {
	// 				if err := conn.Close(); err != nil {
	// 					log.Error("Connection failed to close", "Reason", err)
	// 				}
	// 			}()
	// 		}
	//
	// 		if err != nil {
	// 			log.Error("New RPC Client", "Reason", err)
	// 			return ctrl.Result{
	// 				RequeueAfter: 30 * time.Second,
	// 				Requeue:      true,
	// 			}, err
	// 		}
	//
	// 		if _err := DeleteApps(ctx, rpcClient, stack.Spec.Components); _err != nil {
	// 			log.Error("UninstallApp", "Reason", _err)
	// 			return ctrl.Result{}, _err
	// 		}
	//
	// 		log.Success(ctx, "Uninstall Application was successful")
	//
	// 		stack.ObjectMeta.Finalizers = removeString(stack.ObjectMeta.Finalizers, stackFinalizer)
	// 		if err := r.Update(context.Background(), stack); err != nil {
	// 			return ctrl.Result{}, err
	// 		}
	// 	}
	// }

	return ctrl.Result{}, nil
}

func convertDataToBeConsumable(ctx context.Context, raw applicationv1alpha1.ComponentOverrides) (out map[string]any, err error) {
	if raw.Object == nil {
		return nil, nil
	}
	items := raw.Object
	out = make(map[string]any, 0)
	for key, unkownType := range items {
		switch v := unkownType.(type) {
		case map[string]interface{}:
			out[key] = v
		case []interface{}:
			out[key] = v
		case string:
			out[key] = v
		case int:
			out[key] = v
		case bool:
			out[key] = v
		default:
			return nil, log.NewError(ctx, "Unknown type", "type", v)
		}
	}

	return
}

// Helper functions to manage finalizers
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item != s {
			result = append(result, item)
		}
	}
	return
}

// SetupWithManager sets up the controller with the Manager.
func (r *StackReconciler) SetupWithManager(mgr ctrl.Manager) error {
	log = logger.NewStructuredLogger(
		LogVerbosity[os.Getenv("LOG_LEVEL")],
		LogWriter)

	return ctrl.NewControllerManagedBy(mgr).
		For(&applicationv1alpha1.Stack{}).
		Complete(r)
}
