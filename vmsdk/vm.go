package vmsdk

import (
	"fmt"
	"log"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vapi/library"
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
	// 第一个返回值是 task，我认为没必要处理，如果你要处理的话可以接收后处理
	_, err := vm.PowerOff(ctx)
	if err != nil {
		panic(err)
	}
}

// 获取虚拟机
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
func DeployFromOVF(ctx context.Context, c *govmomi.Client,
	rc *rest.Client, item library.Item, name string, datastoreID string,
	networkKey string, networkValue string, resourcePoolID string, folderID string) bool {
	deploy := vcenter.Deploy{
		DeploymentSpec: vcenter.DeploymentSpec{
			Name:               name,
			DefaultDatastoreID: datastoreID,
			AcceptAllEULA:      true,
			NetworkMappings: []vcenter.NetworkMapping{{
				Key:   networkKey,
				Value: networkValue,
			}},
		},
		Target: vcenter.Target{
			ResourcePoolID: resourcePoolID,
			FolderID:       folderID,
		},
	}

	ref, err := vcenter.NewManager(rc).DeployLibraryItem(ctx, item.ID, deploy)
	if err != nil {
		logrus.Panicln(err.Error())
	}

	f := find.NewFinder(c.Client)
	obj, err := f.ObjectReference(ctx, *ref)
	if err != nil {
		logrus.Println(err.Error())
	}

	vm := obj.(*object.VirtualMachine)
	return vm != nil
}

// 原生部署一个虚拟机
func DeployFromBare(ctx context.Context, c *vim25.Client,
	name string, datacenter string,
	resourcepool string, datastore string) (Host, error) {
	vmhost := Host{}

	finder := find.NewFinder(c)
	dc, err := finder.Datacenter(ctx, datacenter)
	if err != nil {
		return vmhost, err
	}
	finder.SetDatacenter(dc)
	folders, err := dc.Folders(ctx)
	if err != nil {
		return vmhost, err
	}
	pool, err := finder.ResourcePool(ctx, resourcepool)
	if err != nil {
		return vmhost, err
	}
	spec := types.VirtualMachineConfigSpec{
		Name:    name,
		GuestId: string(types.VirtualMachineGuestOsIdentifierCentos7_64Guest),
		Files: &types.VirtualMachineFileInfo{
			VmPathName: datastore,
		},
		NumCPUs:           1,
		MemoryMB:          256,
		NpivOnNonRdmDisks: types.NewBool(true),
	}
	task, err := folders.VmFolder.CreateVM(ctx, spec, pool, nil)
	if err != nil {
		return vmhost, err
	}
	info, err := task.WaitForResult(ctx)
	if err != nil {
		return vmhost, err
	}
	vm := object.NewVirtualMachine(c, info.Result.(types.ManagedObjectReference))
	_, err = vm.ObjectName(ctx)
	if err != nil {
		return vmhost, err
	}
	return vmhost, nil
}

// 克隆虚拟机
func CloneVm(exists string, new string, ctx context.Context, c *vim25.Client,
	datacenter string) (string, error) {
	var name string
	finder := find.NewFinder(c)
	dc, err := finder.Datacenter(ctx, datacenter)
	if err != nil {
		return name, err
	}

	finder.SetDatacenter(dc)

	vm, err := finder.VirtualMachine(ctx, exists)
	if err != nil {
		return name, err
	}

	folders, err := dc.Folders(ctx)
	if err != nil {
		return name, err
	}

	spec := types.VirtualMachineCloneSpec{
		PowerOn: false,
	}

	task, err := vm.Clone(ctx, folders.VmFolder, new, spec)
	if err != nil {
		return name, err
	}

	info, err := task.WaitForResult(ctx)
	if err != nil {
		return name, err
	}

	clone := object.NewVirtualMachine(c, info.Result.(types.ManagedObjectReference))
	name, err = clone.ObjectName(ctx)
	if err != nil {
		return name, err
	}

	return name, nil
}

//原生创建虚拟机
func NewVirtualMachine(c *vim25.Client, vmName string,
	ds string, cpu int32, memory int64, guestId string) (mo.VirtualMachine, error) {
	m := view.NewManager(c)
	v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, nil, true)
	if err != nil {
		logrus.Info(err)
		return mo.VirtualMachine{}, err
	}
	defer v.Destroy(ctx)
	//检查ds是否存在
	var datastore mo.Datastore
	err = v.RetrieveWithFilter(ctx, []string{"Datastore"}, []string{"summary"},
		&datastore, property.Filter{
			"name": ds,
		})
	if err != nil {
		logrus.Info(err)
		return mo.VirtualMachine{}, err
	}
	//检查虚拟机名称重复
	var vm mo.VirtualMachine
	err = v.RetrieveWithFilter(ctx, []string{"VirtualMachine"}, []string{"summary"},
		&vm, property.Filter{
			"name": vmName,
		})
	if err == nil {
		logrus.Info("虚拟机已经存在:", vmName)
		return mo.VirtualMachine{}, err
	}
	//开始新建虚拟机
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
	//查找数据中心
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
	//查找资源池
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

	//检索虚拟机是否创建成功
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
