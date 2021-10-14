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
	"os"
	"os/exec"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/cri-api/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	appsv1 "dtcontroller/api/v1"
)

var (
	output []byte
)

// MachineReconciler reconciles a Machine object
type MachineReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=apps.dtwave.com,resources=machines,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps.dtwave.com,resources=machines/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps.dtwave.com,resources=machines/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Machine object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *MachineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// your logic here
	instance := &appsv1.Machine{}
	r.Recorder.Event(instance, corev1.EventTypeNormal, "test", "test")
	err := r.Client.Get(context.TODO(), req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	err = r.Update(ctx, instance)
	if err != nil {
		return ctrl.Result{}, err
	}

	switch instance.Status.Phase {
	case "":
		instance.Status.Phase = "PENDING"
		instance = assignMachine(*instance)
	case "FAILED":
		log.Log.Info("分配失败")
	case "PENDING":
		instance.Status.Phase = "RUNNING"
	case "RUNNING":
		instance.Status.Phase = "COMPLATED"
	}

	err = r.Update(ctx, instance)
	if err != nil {
		instance.Status.Phase = "FAILED"
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MachineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.Machine{}).
		Complete(r)
}

//分配Machine资源处理方法
func assignMachine(instance appsv1.Machine) *appsv1.Machine {
	log.Log.Info("开始执行分配机器操作")
	var cmdResult string
	cmd := exec.Command("bash", "-c", instance.Spec.Command)
	if output, err := cmd.CombinedOutput(); err != nil {
		cmdResult = err.Error()
	} else {
		cmdResult = string(output)
	}
	//模拟真实操作
	instance.Spec.HostName = "localhost-" + instance.Name
	instance.Spec.User = "deploy"
	hostname, _ := os.Hostname()
	instance.Spec.DtNode = hostname
	instance.Spec.Cpu = "1c"
	instance.Spec.Ip = "192.168.23.23"
	instance.Spec.Mac = "6a:00:03:3d:c1:90"
	// instance.Spec.Password = "123456" + instance.Spec.CmdResult
	log.Log.Info(cmdResult)
	// instance.Spec.CmdResult = cmdResult
	log.Log.Info("分配机器完成")
	log.Log.Info(instance.Name)
	return &instance
}
