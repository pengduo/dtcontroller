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
	"strings"

	"github.com/vmware/govmomi/object"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/cri-api/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	appsv1 "dtcontroller/api/v1"
	"dtcontroller/vmsdk"
)

// finalizer tag
const machineFinalizer = "machine.finalizers.dtwave.com"

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
	log := log.FromContext(ctx)

	// your logic here
	instance := &appsv1.Machine{}
	err := r.Client.Get(context.TODO(), req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, client.IgnoreNotFound(err)
		}
		return reconcile.Result{}, err
	}
	dtnode := &appsv1.DtNode{}
	err = r.Client.Get(ctx, client.ObjectKey{Name: instance.Spec.DtNode}, dtnode)
	if err != nil {
		log.Info(err.Error())
	} else {
		assignMachine(*instance, *dtnode)
	}

	// pre finalizer
	if !instance.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, r.instanceFinalizer(ctx, instance, dtnode)
	}

	if !containsString(instance.Finalizers, machineFinalizer) {
		instance.Finalizers = append(instance.Finalizers, machineFinalizer)
		if err := r.Client.Update(ctx, instance); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MachineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.Machine{}).
		Complete(r)
}

//destoryMachine used to delete machine from vcenter
func destoryMachine(instance appsv1.Machine, dtnode appsv1.DtNode) *appsv1.Machine {
	log.Log.Info("开始删除机器实例")
	ctx := context.Background()
	vURL := strings.Join([]string{"https://", dtnode.Spec.User, ":",
		dtnode.Spec.Password, "@", dtnode.Spec.Ip, "/sdk"}, "")
	_, err := vmsdk.Vmclient(ctx, vURL, dtnode.Spec.User, dtnode.Spec.Password)
	if err != nil {
		log.Log.Info(err.Error())
	}
	vm := object.VirtualMachine{}

	vmsdk.VmDelete(ctx, &vm)
	log.Log.Info("删除实例结束")
	return &instance
}

//分配Machine资源处理方法
func assignMachine(instance appsv1.Machine, dtnode appsv1.DtNode) *appsv1.Machine {
	log.Log.Info("开始分配机器实例")
	ctx := context.Background()
	vURL := strings.Join([]string{"https://", dtnode.Spec.User, ":",
		dtnode.Spec.Password, "@", dtnode.Spec.Ip, "/sdk"}, "")

	vmClient, err := vmsdk.Vmclient(ctx, vURL, dtnode.Spec.User, dtnode.Spec.Password)
	if err != nil {
		log.Log.Info(err.Error())
	}
	switch instance.Spec.Type {
	case "bare":
		vmhost, err := vmsdk.DeployFromBare(context.Background(), vmClient.Client, instance.Name, "Datacenter", "Resources", "[datastore1]")
		if err != nil {
			log.Log.Info("部署失败")
		} else {
			log.Log.Info("分配机器成功", vmhost)
		}
	case "clone":
		_, err = vmsdk.CloneVm("test01", instance.Name, ctx, vmClient.Client, "Datacenter")
		if err != nil {
			log.Log.Info("部署失败")
		} else {
			log.Log.Info("分配机器成功")
		}
	}

	return &instance
}

func (r *MachineReconciler) instanceFinalizer(ctx context.Context,
	instance *appsv1.Machine, dtnode *appsv1.DtNode) error {
	fmt.Println("-------")
	fmt.Println(instance.Name, "\t", dtnode.Name)
	fmt.Println("-------")

	// 预删除执行完毕，移除 nodeFinalizer
	instance.Finalizers = removeString(instance.Finalizers, machineFinalizer)
	return r.Client.Update(ctx, instance)
}

func removeString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}

func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}
