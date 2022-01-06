package controllers

import (
	"context"
	"strings"

	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25"
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

//+kubebuilder:rbac:groups=apps.dtwave.com,resources=machines,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps.dtwave.com,resources=machines/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps.dtwave.com,resources=machines/finalizers,verbs=update

func (r *MachineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	machine := &appsv1.Machine{}
	err := r.Client.Get(context.TODO(), req.NamespacedName, machine)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, client.IgnoreNotFound(err)
		}
		return reconcile.Result{}, err
	}

	dtnode := &appsv1.DtNode{}
	dtnodeName := machine.Spec.DtNode
	err = r.Client.Get(ctx, client.ObjectKey{Name: dtnodeName}, dtnode)
	if err != nil {
		logrus.Info("dtNode可能不存在", err)
		return ctrl.Result{}, nil
	}
	err = AssignMachine(machine, *dtnode)
	if err != nil {
		log.Info("部署出错了")
		return ctrl.Result{}, nil
	}
	r.Status().Update(ctx, machine)
	return ctrl.Result{}, nil
}

//从vcenter删除虚拟机
func DestoryMachine(c *vim25.Client, machine appsv1.Machine, dtnode appsv1.DtNode) *appsv1.Machine {
	logrus.Info("开始删除机器实例")
	ctx := context.Background()
	vURL := strings.Join([]string{"https://", dtnode.Spec.User, ":",
		dtnode.Spec.Password, "@", dtnode.Spec.Ip, "/sdk"}, "")
	_, err := vmsdk.Vmclient(ctx, vURL, dtnode.Spec.User, dtnode.Spec.Password)
	if err != nil {
		logrus.Info(err.Error())
	}

	m := view.NewManager(c)
	v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, nil, true)
	if err != nil {
		logrus.Info(err.Error())
	}
	defer v.Destroy(ctx)
	var vm mo.VirtualMachine
	err = v.RetrieveWithFilter(ctx, []string{"VirtualMachine"}, []string{"summary"},
		&vm, property.Filter{
			"name": machine.Name,
		})
	if err != nil {
		logrus.Info("无法找到目标虚拟机，删除操作无效")
		return &machine
	}
	var objectVm = *vmsdk.Mo2object(c, &vm)
	err = vmsdk.VmDelete(ctx, &objectVm)
	if err != nil {
		logrus.Info("出错了", err.Error())
	}
	logrus.Info("删除实例结束")
	return &machine
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
