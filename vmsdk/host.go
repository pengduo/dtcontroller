package vmsdk

import (
	"strconv"
	"strings"

	appsv1 "dtcontroller/api/v1"

	"github.com/sirupsen/logrus"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"golang.org/x/net/context"
)

var ctx = context.Background()

//主机操作

// 主机结构体
type Host struct {
	Name      string
	Ip        string
	Cpu       string
	Memory    string
	Disk      string
	State     string
	Processor string
}

// 主机集合
type Hosts struct {
	Hosts []Host
}

// 初始化结构体
func NewHosts() *Hosts {
	return &Hosts{
		Hosts: make([]Host, 10),
	}
}

// 增加主机
func (hosts *Hosts) AddHost(name string, ip string, cpu string,
	memory string, disk string, state string, processor string) {
	host := &Host{
		Name:      name,
		Ip:        ip,
		Cpu:       cpu,
		Memory:    memory,
		Disk:      disk,
		State:     state,
		Processor: processor,
	}
	hosts.Hosts = append(hosts.Hosts, *host)
}

// 查询主机ip
func (hosts *Hosts) SelectHost(name string) string {
	ip := "None"
	for _, hosts := range hosts.Hosts {
		if hosts.Name == name {
			ip = hosts.Ip
		}
	}
	return ip
}

// 读取主机信息
func GetVmHosts(ctx context.Context, client *vim25.Client, vmshosts *Hosts) {
	m := view.NewManager(client)
	v, err := m.CreateContainerView(ctx, client.ServiceContent.RootFolder,
		[]string{"HostSystem"}, true)
	if err != nil {
		panic(err)
	}
	defer v.Destroy(ctx)
	var hss []mo.HostSystem
	err = v.Retrieve(ctx, []string{"HostSystem"},
		[]string{"summary"}, &hss)
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
func GetDtNodeInfo(ctx context.Context, client *vim25.Client, dtNode *appsv1.DtNode) error {
	m := view.NewManager(client)
	v, err := m.CreateContainerView(ctx, client.ServiceContent.RootFolder,
		nil, true)
	if err != nil {
		logrus.Info(err)
	}
	defer v.Destroy(ctx)
	var hss []mo.HostSystem
	err = v.Retrieve(ctx, []string{"HostSystem"},
		[]string{"summary"}, &hss)
	if err != nil {
		logrus.Info("主机不存在", err)
		return err
	}

	//检查ds
	var datastore mo.Datastore
	err = v.RetrieveWithFilter(ctx, []string{"Datastore"}, []string{"summary"}, &datastore, property.Filter{
		"name": dtNode.Spec.Datastore,
	})

	if err != nil {
		logrus.Info("存储库不存在", dtNode.Spec.Datastore)
		return err
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
	return nil
}
