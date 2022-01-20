package controllers

import (
	"context"
	"net/url"
	"strconv"
	"strings"

	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vapi/rest"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

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
	_ = log.FromContext(ctx)

	var machine = &appsv1.Machine{}
	var dtnode = &appsv1.DtNode{}
	if err := r.Get(ctx, req.NamespacedName, machine); err != nil {
		logrus.Info("找不到Machine")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	dtnodeName := machine.Spec.DtNode
	if err := r.Get(ctx, client.ObjectKey{Name: dtnodeName}, dtnode); err != nil {
		logrus.Info("dtNode不存在", err)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	if dtnode.Status.Phase != "Ready" {
		logrus.Info("dtnode状态异常")
		machine.Status.Phase = "Failed"
		r.Status().Update(ctx, machine)
		return ctrl.Result{}, nil
	}

	// 预删除逻辑
	if machine.ObjectMeta.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(machine, machineFinalizer) {
			controllerutil.AddFinalizer(machine, machineFinalizer)
			if err := r.Update(ctx, machine); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		//删除逻辑
		if controllerutil.ContainsFinalizer(machine, machineFinalizer) {
			if err := r.machineFinalizer(ctx, machine, dtnode); err != nil {
				return ctrl.Result{}, err
			}

			controllerutil.RemoveFinalizer(machine, machineFinalizer)
			if err := r.Update(ctx, machine); err != nil {
				return ctrl.Result{}, err
			}
		}

		return ctrl.Result{}, nil
	}

	err := assignMachine(ctx, machine, *dtnode)
	if err != nil {
		logrus.Info("部署出错了")
		machine.Status.Phase = "Failed"
	}
	r.Status().Update(ctx, machine)
	return ctrl.Result{}, nil
}

// 删除逻辑
func (r *MachineReconciler) machineFinalizer(ctx context.Context,
	machine *appsv1.Machine, dtnode *appsv1.DtNode) error {
	err := destoryMachine(ctx, machine, dtnode)
	if err != nil {
		logrus.Info("删除虚拟机失败")
		return err
	}
	return nil
}

//从vcenter删除虚拟机
func destoryMachine(ctx context.Context, machine *appsv1.Machine,
	dtnode *appsv1.DtNode) error {
	logrus.Info("开始删除机器实例")
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
		logrus.Info("无法找到目标虚拟机，结束删除操作")
		return nil
	}
	//先关机
	var objectVm = *vmsdk.Mo2object(c.Client, &vm)

	_, err = objectVm.PowerOff(ctx)
	if err != nil {
		logrus.Info("关机出错了")
		return err
	}

	err = vmsdk.VmDelete(ctx, &objectVm)
	if err != nil {
		logrus.Info("出错了", err.Error())
		return err
	}
	logrus.Info("删除虚拟机实例结束")
	return nil
}

// 分配Machine资源处理方法
func assignMachine(ctx context.Context, machine *appsv1.Machine, dtnode appsv1.DtNode) error {
	logrus.Info("创建machine", machine.Name)
	vURL := strings.Join([]string{"https://", dtnode.Spec.User, ":",
		dtnode.Spec.Password, "@", dtnode.Spec.Ip, "/sdk"}, "")
	vmClient, err := vmsdk.Vmclient(ctx, vURL, dtnode.Spec.User, dtnode.Spec.Password)
	if err != nil {
		logrus.Info(err.Error())
		return err
	}
	// 1. 检查是否存在
	m := view.NewManager(vmClient.Client)
	v, err := m.CreateContainerView(ctx, vmClient.Client.ServiceContent.RootFolder, nil, true)
	if err != nil {
		logrus.Info(err)
		return err
	}
	defer v.Destroy(ctx)
	var vm mo.VirtualMachine
	err = v.RetrieveWithFilter(ctx, []string{"VirtualMachine"}, []string{"summary"},
		&vm, property.Filter{
			"name": machine.Name,
		})
	if err == nil {
		logrus.Info("已经存在该名称的虚拟机")
		machine.Status.Phase = "ready"
		machine.Status.Ip = vm.Summary.Guest.IpAddress
		machine.Status.CpuUsed = strings.Join([]string{strconv.Itoa(int(vm.Runtime.MaxCpuUsage)), strconv.Itoa(int(machine.Spec.Cpu))}, "/")
		machine.Status.HostName = strings.Join([]string{strconv.Itoa(int(vm.Runtime.MaxMemoryUsage)), strconv.Itoa(int(machine.Spec.Memory))}, "/")
		return nil
	}
	// 2. 判断部署类型
	switch machine.Spec.Type {
	case "bare":
		if machine.Status.Phase == "Ready" {
			break
		}
		vm, err := vmsdk.NewVirtualMachine(vmClient.Client, machine.Name,
			dtnode.Spec.Datastore, machine.Spec.Cpu, machine.Spec.Memory,
			string(types.VirtualMachineGuestOsIdentifierCentos7_64Guest),
		)
		if err != nil {
			logrus.Info("部署失败")
			machine.Status.Phase = "Failed"
			machine.Status.CpuUsed = "unknown"
			machine.Status.DiskUsed = "unknown"
			machine.Status.HostName = "unknown"
			machine.Status.Mac = "unknown"
			return err
		} else {
			logrus.Info("部署机器成功", vm.Summary.Config.Name)
		}
	case "clone":
		if machine.Status.Phase == "Ready" {
			break
		}
		_, err := vmsdk.CloneVm(ctx, machine.Spec.Base, machine.Name, vmClient.Client)
		if err != nil {
			logrus.Info("克隆失败", err)
			machine.Status.Phase = "Failed"
			return err
		} else {
			machine.Status.Phase = "Ready"
			logrus.Info("克隆机器成功")
		}
	case "ovf":
		rc := rest.NewClient(vmClient.Client)
		if err := rc.Login(ctx, url.UserPassword(dtnode.Spec.User, dtnode.Spec.Password)); err != nil {
			logrus.Info("登录出错", err)
			break
		}
		//获取内容库
		item, err := vmsdk.GetLibraryItem(ctx, rc, "library1", "ovf", "centos7")
		if err != nil {
			logrus.Info("内容库获取出错", err)
			return err
		}
		vmsdk.OVF(ctx, vmClient.Client, rc, machine.Name, dtnode.Spec.Datastore, item.ID)
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

	machine.Status.Phase = "Ready"
	machine.Status.Ip = vm.Summary.Guest.IpAddress
	machine.Status.CpuUsed = strings.Join([]string{strconv.Itoa(int(vm.Runtime.MaxCpuUsage)), strconv.Itoa(int(machine.Spec.Cpu))}, "/")
	machine.Status.HostName = strings.Join([]string{strconv.Itoa(int(vm.Runtime.MaxMemoryUsage)), strconv.Itoa(int(machine.Spec.Memory))}, "/")

	return nil
}

// 注册控制器
// SetupWithManager sets up the controller with the Manager.
func (r *MachineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.Machine{}).
		Complete(r)
}
