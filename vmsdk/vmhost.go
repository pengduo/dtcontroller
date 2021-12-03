package vmsdk

import (
	"log"
	"strconv"
	"strings"

	appsv1 "dtcontroller/api/v1"

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

//分为主机操作和虚拟机操作
//vmhost和vm

// 主机结构体
type VmsHost struct {
	Name      string
	Ip        string
	Cpu       string
	Memory    string
	Disk      string
	State     string
	Processor string
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

func (vmshosts *VmsHosts) AddHost(name string, ip string, cpu string,
	memory string, disk string, state string, processor string) {
	host := &VmsHost{
		Name:      name,
		Ip:        ip,
		Cpu:       cpu,
		Memory:    memory,
		Disk:      disk,
		State:     state,
		Processor: processor,
	}
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
func VmDelete(ctx context.Context, vm *object.VirtualMachine) {
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
func GetVms(ctx context.Context, client *vim25.Client) []mo.VirtualMachine {
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
	return vms
}

// 读取主机信息
func GetVmHosts(ctx context.Context, client *vim25.Client, vmshosts *VmsHosts) {
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
		s := hs.Summary
		h := s.Hardware
		z := s.QuickStats
		ncpu := int32(h.NumCpuCores)
		cpu := strings.Join([]string{strconv.Itoa(int(z.OverallCpuUsage)), "/", strconv.Itoa(int(ncpu * h.CpuMhz))}, "")
		memory := strings.Join([]string{strconv.Itoa(int(z.OverallMemoryUsage)), "/", strconv.Itoa(int(h.MemorySize >> 20))}, "")
		// todo
		vmshosts.AddHost(hs.Name, "", cpu, memory, "", string(s.Runtime.ConnectionState), h.CpuModel)
	}
}

//回填DtNode信息
func GetDtNodeInfo(ctx context.Context, client *vim25.Client, dtNode *appsv1.DtNode) {
	m := view.NewManager(client)
	v, err := m.CreateContainerView(ctx, client.ServiceContent.RootFolder,
		[]string{"HostSystem"}, true)
	if err != nil {
		panic(err)
	}
	defer v.Destroy(ctx)
	var hss []mo.HostSystem
	err = v.Retrieve(ctx, []string{"HostSystem"}, []string{"summary"}, &hss)
	if err != nil {
		panic(err)
	}
	var cpuUsage int32
	var cpuAll int32
	var memoryUsage int32
	var memoryAll int32

	for _, hs := range hss {
		s := hs.Summary
		h := s.Hardware
		z := s.QuickStats
		ncpu := int32(h.NumCpuCores)
		cpuUsage = z.OverallCpuUsage + cpuUsage
		cpuAll = ncpu*h.CpuMhz + cpuAll
		memoryUsage = z.OverallMemoryUsage + memoryUsage
		memoryAll = memoryAll + int32(h.MemorySize)>>20
	}
	dtNode.Status.Cpu = strings.Join([]string{strconv.Itoa(int(cpuUsage)), "/", strconv.Itoa(int(cpuAll))}, "")
	dtNode.Status.Memory = strings.Join([]string{strconv.Itoa(int(memoryUsage)), "/", strconv.Itoa(int(memoryAll))}, "")
	dtNode.Status.Hosts = len(hss)
}

// 使用OVF模版部署虚拟机
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

// 原生部署一个虚拟机
func DeployFromBare(ctx context.Context, c *vim25.Client, name string, datacenter string,
	resourcepool string, datastore string) (VmsHost, error) {
	vmhost := VmsHost{}

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
