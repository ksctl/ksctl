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
	"time"

	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/logger"
	"github.com/ksctl/ksctl/pkg/types"

	storagev1alpha1 "github.com/ksctl/ksctl/ksctl-components/operators/storage/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	log types.LoggerFactory

	LogVerbosity = map[string]int{
		"DEBUG": -1,
		"":      0,
	}

	LogWriter io.Writer = os.Stdout

	ControllerTestSkip string = "CONTROLLER"
)

// ImportStateReconciler reconciles a ImportState object
type ImportStateReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=storage.ksctl.com,resources=importstates,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=storage.ksctl.com,resources=importstates/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=storage.ksctl.com,resources=importstates/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// the ImportState object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.2/pkg/reconcile
func (r *ImportStateReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	log.Debug("Triggered Reconciliation")

	exportedData := new(storagev1alpha1.ImportState)

	if err := r.Get(ctx, req.NamespacedName, exportedData); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if exportedData.Spec.Handled {
		log.Success("skipped!, already handled")
		return ctrl.Result{}, nil
	}

	log.Debug("Debugging", "name", exportedData.Name, "namespace", exportedData.Namespace)
	log.Debug("exported Spec", "status", exportedData.Status, "rawData", len(exportedData.Spec.RawExportedData))

	exportedData.Spec.Handled = true

	ctx, cancel := context.WithTimeout(ctx, time.Minute)

	rpcClient, conn, err := NewClient(ctx)
	defer cancel()

	if os.Getenv(string(consts.KsctlFakeFlag)) != ControllerTestSkip { // to ecape test
		defer func() {
			if err := conn.Close(); err != nil {
				log.Error("Connection failed to close", "Reason", err)
			}
		}()
	}

	if err != nil {
		log.Error("New RPC Client", "Reason", err)
		exportedData.Spec.Success = false
		exportedData.Spec.ReasonOfFailure = err.Error()

		if _err := r.Update(context.Background(), exportedData); _err != nil {
			log.Error("update failed", "Reason", _err)
			return ctrl.Result{}, _err
		}
		return ctrl.Result{}, err
	}

	if _err := ImportData(ctx, rpcClient, exportedData.Spec.RawExportedData); _err != nil {
		log.Error("ImportData", "Reason", _err)
		exportedData.Spec.Success = false
		exportedData.Spec.ReasonOfFailure = _err.Error()

		if __err := r.Update(context.Background(), exportedData); __err != nil {
			log.Error("update failed", "Reason", _err)
			return ctrl.Result{}, __err
		}
		return ctrl.Result{}, _err
	}

	exportedData.Spec.Success = true

	if _err := r.Update(context.Background(), exportedData); _err != nil {
		log.Error("update failed", "Reason", _err)
		return ctrl.Result{}, _err
	}

	log.Success("Import was successful")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ImportStateReconciler) SetupWithManager(mgr ctrl.Manager) error {
	log = logger.NewStructuredLogger(
		LogVerbosity[os.Getenv("LOG_LEVEL")],
		LogWriter)
	log.SetPackageName("ksctl-storage-importer")

	return ctrl.NewControllerManagedBy(mgr).
		For(&storagev1alpha1.ImportState{}).
		Complete(r)
}
