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

	logrus "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	appsv1 "dtcontroller/api/v1"
	"dtcontroller/util"
)

const (
	BARE  = "bare"
	CLONE = "clone"
	OVF   = "ovf"
)

// DtModelReconciler reconciles a DtModel object
type DtModelReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=apps.dtwave.com,resources=dtmodels,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps.dtwave.com,resources=dtmodels/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps.dtwave.com,resources=dtmodels/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the DtModel object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *DtModelReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	var dtmodel = &appsv1.DtModel{}

	if err := r.Get(ctx, req.NamespacedName, dtmodel); err != nil {
		logrus.Info("cannot find dtmodel")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	var dtmodelType = dtmodel.Spec.Type
	var dtmodelProvider = dtmodel.Spec.Provider
	var dtmodelContent = dtmodel.Spec.Content
	// 处理esxi的虚拟化方案
	if dtmodelProvider == "esxi" {
		switch dtmodelType {
		case OVF:
			if err := checkesxiovf(dtmodel.Name, dtmodelContent); err != nil {
				dtmodel.Status.Phase = "NotReady"
			} else {
				dtmodel.Status.Phase = "Ready"
			}
		case BARE:
			//todo
			break
		case CLONE:
			//todo
			break
		default:
			break
		}
	}
	//更新状态
	r.Status().Update(ctx, dtmodel)
	r.Update(ctx, dtmodel)

	return ctrl.Result{}, nil
}

// check ovf model if provider is set to esxi
func checkesxiovf(name string, content map[string]string) error {
	var library = content["library"]
	var ds = content["ds"]
	var ovf = content["ovf"]

	if library == "" || ds == "" || ovf == "" {
		return &util.Err{Msg: "library or os or ds or ovf is not set"}
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DtModelReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.DtModel{}).
		Complete(r)
}
