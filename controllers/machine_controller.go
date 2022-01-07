package controllers

import (
	"context"
	"strings"

	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/cri-api/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	appsv1 "dtcontroller/api/v1"
	"dtcontroller/util"
	"dtcontroller/vmsdk"

	logrus "github.com/sirupsen/logrus"
)

// MachineReconciler reconciles a Machine object
type MachineReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// 定义预删除标记
const machineFinalizer = "machine.finalizers.dtwave"

//+kubebuilder:rbac:groups=apps.dtwave.com,resources=machines,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps.dtwave.com,resources=machines/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps.dtwave.com,resources=machines/finalizers,verbs=update

func (r *MachineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	machine := &appsv1.Machine{}
	dtnode := &appsv1.DtNode{}
	err := r.Client.Get(context.TODO(), req.NamespacedName, machine)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, client.IgnoreNotFound(err)
		}
		return reconcile.Result{}, err
	}

	dtnodeName := machine.Spec.DtNode
	err = r.Client.Get(ctx, client.ObjectKey{Name: dtnodeName}, dtnode)
	if err != nil {
		logrus.Info("dtNode可能不存在", err)
		return ctrl.Result{}, nil
	}

	// 预删除
	if !machine.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, r.machineFinalizer(ctx, machine, dtnode)
	}
	// 删除时间戳为空，则不需要删除，加入到资源中
	if !containsString(machine.Finalizers, machineFinalizer) {
		machine.Finalizers = append(machine.Finalizers, machineFinalizer)
		if err := r.Client.Update(ctx, machine); err != nil {
			return ctrl.Result{}, err
		}
	}

	err = AssignMachine(machine, *dtnode)
	if err != nil {
		log.Info("部署出错了")
		return ctrl.Result{}, nil
	}
	r.Status().Update(ctx, machine)
	return ctrl.Result{}, nil
}

// 删除逻辑
func (r *MachineReconciler) machineFinalizer(ctx context.Context, machine *appsv1.Machine, dtnode *appsv1.DtNode) error {
	err := destoryMachine(machine, dtnode)
	if err != nil {
		logrus.Info("删除虚拟机失败")
		return err
	}
	machine.Finalizers = removeString(machine.Finalizers, machineFinalizer)
	return r.Client.Update(ctx, machine)
}

//从vcenter删除虚拟机
func destoryMachine(machine *appsv1.Machine, dtnode *appsv1.DtNode) error {
	logrus.Info("开始删除机器实例")
	ctx := context.Background()
	vURL := strings.Join([]string{"https://", dtnode.Spec.User, ":",
		dtnode.Spec.Password, "@", dtnode.Spec.Ip, "/sdk"}, "")
	c, err := vmsdk.Vmclient(ctx, vURL, dtnode.Spec.User, dtnode.Spec.Password)
	if err != nil {
		logrus.Info(err.Error())
		return err
	}

	m := view.NewManager(c.Client)
	v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, nil, true)
	if err != nil {
		logrus.Info(err.Error())
		return err
	}
	defer v.Destroy(ctx)
	var vm mo.VirtualMachine
	err = v.RetrieveWithFilter(ctx, []string{"VirtualMachine"}, []string{"summary"},
		&vm, property.Filter{
			"name": machine.Name,
		})
	if err != nil {
		logrus.Info("无法找到目标虚拟机，执行删除操作")
		return nil
	}
	var objectVm = *vmsdk.Mo2object(c.Client, &vm)
	err = vmsdk.VmDelete(ctx, &objectVm)
	if err != nil {
		logrus.Info("出错了", err.Error())
		return err
	}
	logrus.Info("删除虚拟机实例结束")
	return nil
}

// 分配Machine资源处理方法
func AssignMachine(machine *appsv1.Machine, dtnode appsv1.DtNode) error {
	logrus.Info("创建machine")
	ctx := context.Background()
	vURL := strings.Join([]string{"https://", dtnode.Spec.User, ":",
		dtnode.Spec.Password, "@", dtnode.Spec.Ip, "/sdk"}, "")
	vmClient, err := vmsdk.Vmclient(ctx, vURL, dtnode.Spec.User, dtnode.Spec.Password)
	if err != nil {
		logrus.Info(err.Error())
		return err
	}
	// 1. 检查是否已经存在
	m := view.NewManager(vmClient.Client)
	v, err := m.CreateContainerView(ctx, vmClient.Client.ServiceContent.RootFolder, nil, true)
	if err != nil {
		logrus.Info(err.Error())
	}
	defer v.Destroy(ctx)
	var vm mo.VirtualMachine
	err = v.RetrieveWithFilter(ctx, []string{"VirtualMachine"}, []string{"summary"},
		&vm, property.Filter{
			"name": machine.Name,
		})
	if err == nil {
		logrus.Info("已经存在该名称的虚拟机")
		return util.NewError("已经存在该名称的虚拟机")
	}
	// 2. 判断部署类型
	switch machine.Spec.Type {
	case "bare":
		if machine.Status.Phase == "ready" {
			break
		}
		vmhost, err := vmsdk.DeployFromBare(context.Background(),
			vmClient.Client, machine.Name, "Datacenter", "Resources", "[datastore1]")
		if err != nil {
			logrus.Info("部署失败")
			machine.Status.Phase = "failed"
		} else {
			logrus.Info("部署机器成功", vmhost)
		}
	case "clone":
		if machine.Status.Phase == "ready" {
			break
		}
		_, err = vmsdk.CloneVm("test01", machine.Name, ctx, vmClient.Client, "Datacenter")
		if err != nil {
			logrus.Info("部署失败")
			machine.Status.Phase = "failed"
		} else {
			machine.Status.Phase = "ready"
			logrus.Info("部署机器成功")
		}
	default:
		logrus.Info("不支持的部署方式")
		return util.NewError("不支持的部署方式")
	}
	// 3. 检查部署结果
	err = vmsdk.GetVmInfo(ctx, vmClient.Client, &vm)
	if err != nil {
		logrus.Info("获取虚拟机信息出错", err)
		return err
	}
	return nil
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

// 注册控制器
// SetupWithManager sets up the controller with the Manager.
func (r *MachineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.Machine{}).
		Complete(r)
}
