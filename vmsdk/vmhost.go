package vmsdk

import (
	"fmt"
	"log"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vapi/library"
	"github.com/vmware/govmomi/vapi/rest"
	"github.com/vmware/govmomi/vapi/vcenter"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
	"golang.org/x/net/context"
)

type VmsHost struct {
	Name string
	Ip   string
}
type VmsHosts struct {
	VmsHosts []VmsHost
}

// 初始化结构体
func NewVmsHosts() *VmsHosts {
	return &VmsHosts{
		VmsHosts: make([]VmsHost, 10),
	}
}

func (vmshosts *VmsHosts) AddHost(name string, ip string) {
	host := &VmsHost{name, ip}
	vmshosts.VmsHosts = append(vmshosts.VmsHosts, *host)
}

// 查询主机ip
func (vmshosts *VmsHosts) SelectHost(name string) string {
	ip := "None"
	for _, hosts := range vmshosts.VmsHosts {
		if hosts.Name == name {
			ip = hosts.Ip
		}
	}
	return ip
}

//开机
func VmPowerOn(ctx context.Context, vm *object.VirtualMachine) {
	_, err := vm.PowerOn(ctx)
	if err != nil {
		panic(err)
	}
}

//关机
func VmShutdown(ctx context.Context, vm *object.VirtualMachine) {
	// 第一个返回值是 task，我认为没必要处理，如果你要处理的话可以接收后处理
	_, err := vm.PowerOff(ctx)
	if err != nil {
		panic(err)
	}
}

//删除
func vmDelete(ctx context.Context, vm *object.VirtualMachine) {
	// task 可以处理，也可以不处理
	task, err := vm.Destroy(ctx)
	if err != nil {
		panic(err)
	}
	if task.Wait(ctx) != nil {
		panic(err)
	}
}

// 获取虚拟机
func GetVms(ctx context.Context, client *vim25.Client, vmshosts *VmsHosts) []mo.VirtualMachine {
	m := view.NewManager(client)
	v, err := m.CreateContainerView(ctx, client.ServiceContent.RootFolder, []string{"VirtualMachine"}, true)
	if err != nil {
		panic(err)
	}
	defer v.Destroy(ctx)
	var vms []mo.VirtualMachine
	err = v.Retrieve(ctx, []string{"VirtualMachine"}, []string{"summary", "runtime", "datastore"}, &vms)
	if err != nil {
		panic(err)
	}

	// 输出虚拟机信息到csv
	// file, _ := os.OpenFile("./vms.csv", os.O_WRONLY|os.O_CREATE, os.ModePerm)
	//防止中文乱码
	// file.WriteString("\xEF\xBB\xBF")
	// w := csv.NewWriter(file)
	// w.Write([]string{"宿主机", "虚拟机", "系统", "状态", "IP地址", "资源"})
	// w.Flush()
	// for _, vm := range vms {
	// 	//虚拟机资源信息
	// 	res := strconv.Itoa(int(vm.Summary.Config.MemorySizeMB)) + " MB " + strconv.Itoa(int(vm.Summary.Config.NumCpu)) + " vCPU(s) " + units.ByteSize(vm.Summary.Storage.Committed+vm.Summary.Storage.Uncommitted).String()
	// 	w.Write([]string{vmshosts.SelectHost(vm.Summary.Runtime.Host.Value), vm.Summary.Config.Name, vm.Summary.Config.GuestFullName, string(vm.Summary.Runtime.PowerState), vm.Summary.Guest.IpAddress, res})
	// 	w.Flush()
	// }
	// file.Close()

	return vms

}

// 读取主机信息
func GetHosts(ctx context.Context, client *vim25.Client, vmshosts *VmsHosts) {
	m := view.NewManager(client)
	v, err := m.CreateContainerView(ctx, client.ServiceContent.RootFolder, []string{"HostSystem"}, true)
	if err != nil {
		panic(err)
	}
	defer v.Destroy(ctx)
	var hss []mo.HostSystem
	err = v.Retrieve(ctx, []string{"HostSystem"}, []string{"summary"}, &hss)
	if err != nil {
		panic(err)
	}
	for _, hs := range hss {
		vmshosts.AddHost(hs.Summary.Host.Value, hs.Summary.Config.Name)
	}
}

// 使用OVF模版部署
func DeployFromOVF(ctx context.Context, c *govmomi.Client, rc *rest.Client, item library.Item, name string, datastoreID string, networkKey string, networkValue string, resourcePoolID string, folderID string) bool {
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
		log.Panicln(err.Error())
	}

	f := find.NewFinder(c.Client)
	obj, err := f.ObjectReference(ctx, *ref)
	if err != nil {
		log.Println(err.Error())
	}

	vm := obj.(*object.VirtualMachine)
	return vm != nil
}

// 全新创建虚拟机
func DeployFromBare(ctx context.Context, c *vim25.Client, name string, datacenter string, resourcepool string, datastore string) error {

	finder := find.NewFinder(c)
	dc, err := finder.Datacenter(ctx, datacenter)
	if err != nil {
		return err
	}

	finder.SetDatacenter(dc)

	folders, err := dc.Folders(ctx)
	if err != nil {
		return err
	}

	pool, err := finder.ResourcePool(ctx, resourcepool)
	if err != nil {
		return err
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
		return err
	}

	info, err := task.WaitForResult(ctx)
	if err != nil {
		return err
	}

	vm := object.NewVirtualMachine(c, info.Result.(types.ManagedObjectReference))
	_, err = vm.ObjectName(ctx)
	if err != nil {
		return err
	}

	return nil

}

// 克隆虚拟机
func CloneVm(exists string, new string, ctx context.Context, c *vim25.Client, datacenter string) error {
	finder := find.NewFinder(c)
	dc, err := finder.Datacenter(ctx, datacenter)
	if err != nil {
		return err
	}

	finder.SetDatacenter(dc)

	vm, err := finder.VirtualMachine(ctx, exists)
	if err != nil {
		return err
	}

	folders, err := dc.Folders(ctx)
	if err != nil {
		return err
	}

	spec := types.VirtualMachineCloneSpec{
		PowerOn: false,
	}

	task, err := vm.Clone(ctx, folders.VmFolder, new, spec)
	if err != nil {
		return err
	}

	info, err := task.WaitForResult(ctx)
	if err != nil {
		return err
	}

	clone := object.NewVirtualMachine(c, info.Result.(types.ManagedObjectReference))
	name, err := clone.ObjectName(ctx)
	if err != nil {
		return err
	}

	fmt.Println(name)
	return nil

}
