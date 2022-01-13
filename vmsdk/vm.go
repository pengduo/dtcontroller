package vmsdk

import (
	"fmt"
	"log"

	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vapi/rest"
	"github.com/vmware/govmomi/vapi/vcenter"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
	"golang.org/x/net/context"

	logrus "github.com/sirupsen/logrus"
)

//虚拟机操作

// 开机
func VmPowerOn(ctx context.Context, vm *object.VirtualMachine) {
	_, err := vm.PowerOn(ctx)
	if err != nil {
		panic(err)
	}
}

// 关机
func VmShutdown(ctx context.Context, vm *object.VirtualMachine) {
	_, err := vm.PowerOff(ctx)
	if err != nil {
		panic(err)
	}
}

// 查找虚拟机
func GetVms(ctx context.Context, client *vim25.Client) []mo.VirtualMachine {
	m := view.NewManager(client)
	v, err := m.CreateContainerView(ctx, client.ServiceContent.RootFolder,
		[]string{"VirtualMachine"}, true)
	if err != nil {
		panic(err)
	}
	defer v.Destroy(ctx)
	var vms []mo.VirtualMachine
	err = v.Retrieve(ctx, []string{"VirtualMachine"},
		[]string{"summary", "runtime", "datastore"}, &vms)
	if err != nil {
		panic(err)
	}
	return vms
}

// 获取虚拟机信息
func GetVmInfo(ctx context.Context, c *vim25.Client, vm *mo.VirtualMachine) error {
	m := view.NewManager(c)
	v, _ := m.CreateContainerView(ctx, c.ServiceContent.RootFolder,
		[]string{"VirtualMachine"}, true)
	defer v.Destroy(ctx)
	err := v.RetrieveWithFilter(ctx, []string{"VirtualMachine"}, []string{"summary"},
		&vm, property.Filter{
			"name": vm.Name,
		})
	if err != nil {
		logrus.Info("虚拟机不存在")
		return err
	}
	return nil
}

// 使用OVF模版部署虚拟机
func OVF(ctx context.Context, c *vim25.Client, rc *rest.Client,
	name string, ds string, itemID string) error {
	finder := find.NewFinder(c)
	m := view.NewManager(c)
	v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, nil, true)
	defer v.Destroy(ctx)
	if err != nil {
		return err
	}
	// 检查虚拟机名称重复
	var vm mo.VirtualMachine
	err = v.RetrieveWithFilter(ctx, []string{"VirtualMachine"}, []string{"summary"},
		&vm, property.Filter{
			"name": name,
		})
	if err == nil {
		logrus.Info("虚拟机已经存在:", name)
		return err
	}

	pool, err := finder.ResourcePool(ctx, "Resources")
	if err != nil {
		logrus.Info("查找Resources", err)
		return err
	}

	// 检查ds是否存在
	var datastore mo.Datastore
	err = v.RetrieveWithFilter(ctx, []string{"Datastore"}, []string{"summary"},
		&datastore, property.Filter{
			"name": ds,
		})
	if err != nil {
		logrus.Info(err)
		return err
	}

	deploy := vcenter.Deploy{
		DeploymentSpec: vcenter.DeploymentSpec{
			Name:               name,
			DefaultDatastoreID: datastore.Reference().Value,
			AcceptAllEULA:      true,
		},
		Target: vcenter.Target{
			ResourcePoolID: pool.Reference().Value,
		},
	}

	_, err = vcenter.NewManager(rc).DeployLibraryItem(ctx, itemID, deploy)
	if err != nil {
		logrus.Info(err.Error())
		return err
	}

	return nil
}

// 克隆虚拟机
func CloneVm(ctx context.Context, exist string, new string,
	c *vim25.Client) (mo.VirtualMachine, error) {

	// 查找数据中心
	finder := find.NewFinder(c)
	dc, err := finder.DefaultDatacenter(ctx)
	if err != nil {
		logrus.Info("查找数据中心出错", err)
		return mo.VirtualMachine{}, err
	}
	finder.SetDatacenter(dc)
	folders, err := dc.Folders(ctx)
	if err != nil {
		logrus.Info(err)
		return mo.VirtualMachine{}, err
	}

	// 构建克隆参数
	spec := types.VirtualMachineCloneSpec{
		PowerOn: true,
	}
	// 检查被克隆的存在
	ovjectVm, err := finder.VirtualMachine(ctx, exist)
	if err != nil {
		logrus.Info("克隆对象不存在")
		return mo.VirtualMachine{}, err
	}
	// 克隆操作
	task, err := ovjectVm.Clone(ctx, folders.VmFolder, new, spec)
	if err != nil {
		logrus.Info("克隆失败", err)
		return mo.VirtualMachine{}, err
	}
	// 等待
	info, err := task.WaitForResult(ctx)
	if err != nil {
		logrus.Info("等待超时", err)
		return mo.VirtualMachine{}, err
	}
	// 查找虚拟机
	clone := object.NewVirtualMachine(c, info.Result.(types.ManagedObjectReference))
	name, err := clone.ObjectName(ctx)
	if err != nil {
		return mo.VirtualMachine{}, err
	}

	ovjectVm, err = finder.VirtualMachine(ctx, name)
	if err != nil {
		logrus.Info("克隆失败", err)
		return mo.VirtualMachine{}, err
	}
	var vm *mo.VirtualMachine

	vm, err = Object2Mo(ctx, ovjectVm, c)
	if err != nil {
		logrus.Info("转换出错", err)
		return mo.VirtualMachine{}, err
	}
	return *vm, nil
}

// 原生创建虚拟机
func NewVirtualMachine(c *vim25.Client, vmName string,
	ds string, cpu int32, memory int64, guestId string) (mo.VirtualMachine, error) {
	m := view.NewManager(c)
	v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, nil, true)
	if err != nil {
		logrus.Info(err)
		return mo.VirtualMachine{}, err
	}
	defer v.Destroy(ctx)
	// 检查ds是否存在
	var datastore mo.Datastore
	err = v.RetrieveWithFilter(ctx, []string{"Datastore"}, []string{"summary"},
		&datastore, property.Filter{
			"name": ds,
		})
	if err != nil {
		logrus.Info(err)
		return mo.VirtualMachine{}, err
	}
	// 检查虚拟机名称重复
	var vm mo.VirtualMachine
	err = v.RetrieveWithFilter(ctx, []string{"VirtualMachine"}, []string{"summary"},
		&vm, property.Filter{
			"name": vmName,
		})
	if err == nil {
		logrus.Info("虚拟机已经存在:", vmName)
		return mo.VirtualMachine{}, err
	}
	// 开始新建虚拟机
	vmSpec := types.VirtualMachineConfigSpec{
		Name:    vmName,
		GuestId: guestId,
		Files: &types.VirtualMachineFileInfo{
			VmPathName: "[" + ds + "]",
		},
		NumCPUs:           cpu,
		MemoryMB:          memory,
		NpivOnNonRdmDisks: types.NewBool(true),
	}
	// 查找数据中心
	finder := find.NewFinder(c)
	dc, err := finder.DefaultDatacenter(ctx)
	if err != nil {
		logrus.Info(err)
		return mo.VirtualMachine{}, err
	}
	finder.SetDatacenter(dc)
	folders, err := dc.Folders(ctx)
	if err != nil {
		logrus.Info(err)
		return mo.VirtualMachine{}, err
	}
	// 查找资源池
	pool, err := finder.ResourcePool(ctx, "Resources")
	if err != nil {
		log.Panicln(err)
		return mo.VirtualMachine{}, err
	}
	task, err := folders.VmFolder.CreateVM(ctx, vmSpec, pool, nil)
	if err != nil {
		logrus.Info(err)
		return mo.VirtualMachine{}, err
	}
	info, err := task.WaitForResult(ctx)
	if err != nil {
		logrus.Info(err)
		return mo.VirtualMachine{}, err
	}
	logrus.Info(info)

	// 检索虚拟机是否创建成功
	vm = mo.VirtualMachine{}
	err = v.RetrieveWithFilter(ctx, []string{"VirtualMachine"}, []string{"summary"},
		&vm, property.Filter{
			"name": vmName,
		})
	if err != nil {
		logrus.Info("虚拟机创建失败:", vmName)
		return mo.VirtualMachine{}, err
	}
	return vm, nil
}

//清理孤立的虚拟机
func CleanOrphaned(ctx context.Context, c *vim25.Client, vms *[]mo.VirtualMachine) {
	for _, vm := range *vms {
		if vm.Summary.Runtime.ConnectionState == "orphaned" {
			fmt.Println("清理孤立虚拟机:", vm.Summary.Config.Name)
			ovm := Mo2object(c, &vm)
			VmDelete(ctx, ovm)
		}
	}
}

//mo类型的虚拟机转换成object类型的
func Mo2object(c *vim25.Client, mvm *mo.VirtualMachine) *object.VirtualMachine {
	vm := object.NewVirtualMachine(c, mvm.Reference())
	return vm
}

//object类型转mo
func Object2Mo(ctx context.Context, vm *object.VirtualMachine, c *vim25.Client) (*mo.VirtualMachine, error) {
	var mvm mo.VirtualMachine
	pc := property.DefaultCollector(c)
	err := pc.RetrieveOne(ctx, vm.Reference(), []string{"runtime.host", "config.uuid"}, &mvm)
	if err != nil {
		return nil, err
	}
	return &mvm, err
}

//删除虚拟机
func VmDelete(ctx context.Context, vm *object.VirtualMachine) error {
	task, err := vm.Destroy(ctx)
	if err != nil {
		return err
	}
	if task.Wait(ctx) != nil {
		return err
	}
	return nil
}
