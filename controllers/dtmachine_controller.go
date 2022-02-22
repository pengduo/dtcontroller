package controllers

import (
	"context"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vapi/rest"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	appsv1 "dtcontroller/api/v1"
	"dtcontroller/util"
	"dtcontroller/vmsdk/esxi"

	logrus "github.com/sirupsen/logrus"
)

// DtMachineReconciler reconciles a Machine object
type DtMachineReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// 定义预删除标记
const dtMachineFinalizer = "dtmachine.finalizers.dtwave"

//+kubebuilder:rbac:groups=apps.dtwave.com,resources=dtmachines,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps.dtwave.com,resources=dtmachines/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps.dtwave.com,resources=dtmachines/finalizers,verbs=update

func (r *DtMachineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	var machine = &appsv1.DtMachine{}
	if err := r.Get(ctx, req.NamespacedName, machine); err != nil {
		logrus.Info("cannot find dtmachine")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	var dtClusterName = machine.Spec.DtCluster
	var dtCluster = &appsv1.DtCluster{}
	if err := r.Get(ctx, client.ObjectKey{Name: dtClusterName}, dtCluster); err != nil {
		logrus.Info("cannot find dtCluster ", err)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	var dtnode = &appsv1.DtNode{}
	if err := r.Get(ctx, client.ObjectKey{Name: dtCluster.Status.DtNode}, dtnode); err != nil {
		logrus.Info("cannot find dtnode ", err)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	var specNode = dtnode.Spec.Node
	var envNode = os.Getenv("MY_NODE_NAME")
	var podName = os.Getenv("MY_POD_NAME")
	logrus.Info("dtnode_controller", specNode, envNode, podName)
	//判断是否该任务归自己处理
	if specNode != envNode {
		logrus.Info("pod cannot handle this dtnode, pod is : ",
			podName, ",target node is : ", specNode)
		return ctrl.Result{}, nil
	}

	var modelName = machine.Spec.DtModel
	var model = &appsv1.DtModel{}
	if err := r.Get(ctx, client.ObjectKey{Name: modelName, Namespace: machine.Namespace}, model); err != nil {
		logrus.Info("cannot find dtmodel ", modelName, err)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	//更新dtmodel的状态
	r.Status().Update(ctx, model)

	// 预删除逻辑
	if machine.ObjectMeta.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(machine, dtMachineFinalizer) {
			controllerutil.AddFinalizer(machine, dtMachineFinalizer)
			if err := r.Update(ctx, machine); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		//删除逻辑
		if controllerutil.ContainsFinalizer(machine, dtMachineFinalizer) {
			if err := r.machineFinalizer(ctx, machine.Name, dtCluster); err != nil {
				return ctrl.Result{}, err
			}
			controllerutil.RemoveFinalizer(machine, dtMachineFinalizer)
			if err := r.Update(ctx, machine); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	err := assignMachine(ctx, machine, model, dtCluster)
	if err != nil {
		logrus.Info("部署出错了")
		machine.Status.Phase = "Failed"
	} else {
		machine.Status.Phase = "Ready"
	}

	r.Status().Update(ctx, machine)
	return ctrl.Result{}, nil
}

// assignMachine is desined to deploy a machine with target dtcluster configuration
func assignMachine(ctx context.Context, machine *appsv1.DtMachine, model *appsv1.DtModel, dtcluster *appsv1.DtCluster) error {
	var content = dtcluster.Spec.Content
	// if provider is set to esxi, do this
	if dtcluster.Spec.Provider == string(ESXI) {
		var ip = content["ip"]
		var username = content["username"]
		var password = content["password"]
		if ip == "" || username == "" || password == "" {
			return &util.Err{Msg: "check is there an error in ip or username or password"}
		}
		return assignMachineESXI(ctx, machine, model, ip, username, password)
	}
	return nil
}

// 删除逻辑
func (r *DtMachineReconciler) machineFinalizer(ctx context.Context, machineName string, dtCluster *appsv1.DtCluster) error {
	var content = dtCluster.Spec.Content
	var provider = dtCluster.Spec.Provider
	if provider == "esxi" {
		var ip = content["ip"]
		var username = content["username"]
		var password = content["password"]
		if ip == "" || username == "" || password == "" {
			return &util.Err{Msg: "check is there an error in ip username or password"}
		}
		return destoryMachineESXI(ctx, machineName, ip, username, password)
	}
	return nil
}

//从vcenter删除虚拟机
func destoryMachineESXI(ctx context.Context, machineName string,
	ip string, username string, password string) error {
	logrus.Info("开始删除机器实例")
	vURL := strings.Join([]string{"https://", username, ":",
		password, "@", ip, "/sdk"}, "")
	c, err := esxi.Vmclient(ctx, vURL, username, password)
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
			"name": machineName,
		})
	if err != nil {
		logrus.Info("无法找到目标虚拟机,结束删除操作")
		return nil
	}
	//先关机
	var objectVm = *esxi.Mo2object(c.Client, &vm)

	_, err = objectVm.PowerOff(ctx)
	if err != nil {
		logrus.Info("关机出错了")
		return err
	}

	err = esxi.VmDelete(ctx, &objectVm)
	if err != nil {
		logrus.Info("出错了", err.Error())
		return err
	}
	logrus.Info("删除虚拟机实例结束")
	return nil
}

// 在ESXI部署DtMachine
func assignMachineESXI(ctx context.Context, machine *appsv1.DtMachine, model *appsv1.DtModel, ip string, username string, password string) error {
	logrus.Info("创建machine", machine.Name)
	vURL := strings.Join([]string{"https://", username, ":",
		password, "@", ip, "/sdk"}, "")
	vmClient, err := esxi.Vmclient(ctx, vURL, username, password)
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
		machine.Status.CpuUsed = ""
		machine.Status.HostName = ""
		return nil
	}
	if model.Status.Phase != "Ready" {
		return &util.Err{Msg: "dtmodel is not ready"}
	}
	// 2. 判断部署类型
	switch model.Spec.Type {
	case "bare":
		assignMachineESXIBare(machine, model, vmClient.Client)
	case "clone":
		assignMachineESXIClone(machine, model, vmClient.Client)
	case "ovf":
		assignMachineESXIOVF(machine, model, vmClient.Client, username, password)
	default:
		logrus.Info("不支持的部署方式")
		return util.NewError("不支持的部署方式")
	}
	// 3. 检查部署结果
	err = esxi.GetVmInfo(ctx, vmClient.Client, &vm)
	if err != nil {
		logrus.Info("获取虚拟机信息出错", err)
		return err
	}

	return nil
}

// clone a machine with esxi sdk
func assignMachineESXIClone(machine *appsv1.DtMachine, dtmodel *appsv1.DtModel, c *vim25.Client) error {
	var content = dtmodel.Spec.Content
	var base = content["base"]
	_, err := esxi.CloneVm(context.Background(), base, machine.Name, c)
	if err != nil {
		logrus.Info("克隆失败", err)
		machine.Status.Phase = "Failed"
		return err
	} else {
		machine.Status.Phase = "Ready"
		logrus.Info("克隆机器成功")
	}
	return nil
}

// create machine through ovf on esxi
func assignMachineESXIOVF(machine *appsv1.DtMachine, dtmodel *appsv1.DtModel, c *vim25.Client, username string, password string) error {
	rc := rest.NewClient(c)
	if err := rc.Login(context.Background(), url.UserPassword(username, password)); err != nil {
		logrus.Info("登录出错", err)
		return err
	}

	var content = dtmodel.Spec.Content
	var library = content["library"]
	var ovf = content["ovf"]
	var ds = content["ds"]

	if ovf == "" || library == "" || ds == "" {
		return &util.Err{Msg: "ovf or library or os or ds is not set"}
	}
	//获取内容库
	item, err := esxi.GetLibraryItem(context.Background(), rc, library, "ovf", ovf)
	if err != nil {
		logrus.Info("内容库获取出错", err)
		return err
	}
	if err := esxi.OVF(context.Background(), c, rc, machine.Name, ds, item.ID); err != nil {
		logrus.Info("cannot create machine from ovf", err)
		return err
	}
	return nil
}

func assignMachineESXIBare(machine *appsv1.DtMachine, dtmodel *appsv1.DtModel, c *vim25.Client) error {
	var content = dtmodel.Spec.Content
	var ds = content["ds"]
	var cpu = content["cpu"]
	var memory = content["cpu"]
	var os = content["os"]

	cpu32, _ := strconv.ParseInt(cpu, 10, 32)
	memory64, _ := strconv.ParseInt(memory, 10, 64)
	vm, err := esxi.NewVirtualMachine(c, machine.Name,
		ds, int32(cpu32), memory64,
		os,
	)

	if err != nil {
		logrus.Info("部署失败")
		machine.Status.Phase = "failed"
		machine.Status.CpuUsed = "unknown"
		machine.Status.DiskUsed = "unknown"
		machine.Status.HostName = "unknown"
		machine.Status.Mac = "unknown"
		return err
	}
	logrus.Info("部署机器成功", vm.Summary.Config.Name)
	return nil
}

// 注册控制器
// SetupWithManager sets up the controller with the Manager.
func (r *DtMachineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.DtMachine{}).
		Complete(r)
}
