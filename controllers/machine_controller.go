package controllers

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/mo"
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

// MachineReconciler reconciles a Machine object
type MachineReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=apps.dtwave.com,resources=machines,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps.dtwave.com,resources=machines/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps.dtwave.com,resources=machines/finalizers,verbs=update

func (r *MachineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	instance := &appsv1.Machine{}
	err := r.Client.Get(context.TODO(), req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, client.IgnoreNotFound(err)
		}
		return reconcile.Result{}, err
	}

	dtnode := &appsv1.DtNode{}
	dtnodeName := instance.Spec.DtNode
	err = r.Client.Get(ctx, client.ObjectKey{Name: dtnodeName}, dtnode)
	fmt.Println("dtnodeName", dtnodeName)
	if err != nil {
		log.Info(err.Error())
		return ctrl.Result{}, nil
	}

	instance, err = assignMachine(*instance, *dtnode)
	if err != nil {
		log.Info(err.Error())
		return ctrl.Result{}, nil
	}
	r.Status().Update(ctx, instance)
	return ctrl.Result{}, nil
}

//从vcenter删除虚拟机
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

// 分配Machine资源处理方法
func assignMachine(instance appsv1.Machine, dtnode appsv1.DtNode) (*appsv1.Machine, error) {
	log.Log.Info("开始分配机器实例")
	ctx := context.Background()
	vURL := strings.Join([]string{"https://", dtnode.Spec.User, ":",
		dtnode.Spec.Password, "@", dtnode.Spec.Ip, "/sdk"}, "")

	vmClient, err := vmsdk.Vmclient(ctx, vURL, dtnode.Spec.User, dtnode.Spec.Password)
	if err != nil {
		log.Log.Info(err.Error())
		return &appsv1.Machine{}, err
	}

	switch instance.Spec.Type {
	case "bare":
		if instance.Status.Phase == "ready" {
			break
		}
		vmhost, err := vmsdk.DeployFromBare(context.Background(),
			vmClient.Client, instance.Name, "Datacenter", "Resources", "[datastore1]")
		if err != nil {
			log.Log.Info("部署失败")
			instance.Status.Phase = "failed"
		} else {
			log.Log.Info("部署机器成功", vmhost)
		}
	case "clone":
		if instance.Status.Phase == "ready" {
			break
		}
		_, err = vmsdk.CloneVm("test01", instance.Name, ctx, vmClient.Client, "Datacenter")
		if err != nil {
			log.Log.Info("部署失败")
			instance.Status.Phase = "failed"
		} else {
			instance.Status.Phase = "ready"
			log.Log.Info("部署机器成功")
		}
	}
	var vm mo.VirtualMachine
	vmsdk.GetVmInfo(ctx, vmClient.Client, instance.Name, vm)

	if !reflect.DeepEqual(vm, mo.VirtualMachine{}) {
		if vm.Summary.Runtime.Host.Value != "" {
			instance.Status.HostName = vm.Summary.Runtime.Host.Value
		}
		if vm.Summary.Guest.IpAddress != "" {
			instance.Status.Ip = vm.Summary.Guest.IpAddress
		}
	}

	return &instance, nil
}

func RemoveString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}

func ContainsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// 注册控制器
// SetupWithManager sets up the controller with the Manager.
func (r *MachineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.Machine{}).
		Complete(r)
}
