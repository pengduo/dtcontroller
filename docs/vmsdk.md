## vmsdk调用文档

### 操作示例：
[vm-practice](http://gitlab.dtwave-inc.com/shuqi-devops/vm-practice.git)

### 例子1:建立连接

```golang
package main

import (
	"log"
	"net/url"
	"vm-practice/vmhost"

	"github.com/vmware/govmomi"
	"golang.org/x/net/context"
)
//环境变量
const (
	envURL          = "https://linjb@vsphere.local:LIN115jinbao!@192.168.123.138/sdk"
	envUserName     = "linjb@vsphere.local"
	envPassword     = "LIN115jinbao!"
	envInsecure     = "true"
	libraryName     = "library1"
	libraryItemName = "centos7"
	libraryItemType = "ovf"
)
//创建客户端连接
func client(ctx context.Context, vURL string, username string, password string) (client *govmomi.Client, err error) {
	u, err := url.Parse(vURL)
	if err != nil {
		log.Panicln(err.Error())
		return client, err
	}
	u.User = url.UserPassword(username, password)
	c, err := govmomi.NewClient(ctx, u, true)
	if err != nil {
		log.Panicln(err.Error())
		return client, err
	}
	return c, nil
}
func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c, err := client(ctx, envURL, envUserName, envPassword)
	if err != nil {
		log.Panicln(err.Error())
	}

	vmhost.DeployFromBare(ctx, c.Client, "vm03", "Datacenter", "Resources", "[datastore1]")
}

```

### 例子2:删除虚拟机
```golang
func vmDelete(ctx context.Context, vm *object.VirtualMachine) {
	task, err := vm.Destroy(ctx)
	if err != nil {
		panic(err)
	}
	if task.Wait(ctx) != nil {
		panic(err)
	}
}
```

### 例子3:转换虚拟机格式

```golang
govmomi库中的虚拟机有2种，一种是mo包下的，另外一种是object包下的，mo包下的会功能更简单一些，但是删除等功能时候，只能使用object类型的。
//mo类型的虚拟机转换成object类型的
func mo2object(c *vim25.Client, mvm *mo.VirtualMachine) *object.VirtualMachine {
	vm := object.NewVirtualMachine(c, mvm.Reference())
	return vm
}
//object类型转换成mo类型的
func x(ctx context.Context, vm *object.VirtualMachine, c *vim25.Client) {
	var mvm mo.VirtualMachine

	pc := property.DefaultCollector(c)
	// 如果想要全部属性，可以传一个空的字串切片
	err := pc.RetrieveOne(ctx, vm.Reference(), []string{"runtime.host", "config.uuid"}, &mvm)
}
```

### 例子4:清理孤立状态的虚拟机

调用sdk删除的虚拟机不会直接删除，而是变成孤立状态(orphaned)，需要需要调用删除方法彻底删除

```golang
func cleanOrphaned(c *vim25.Client, vms *[]mo.VirtualMachine) {
	for _, vm := range *vms {
		if vm.Summary.Runtime.ConnectionState == "orphaned" {
			fmt.Println("清理孤立虚拟机：", vm.Summary.Config.Name)
			ovm := mo2object(c, &vm)
			vmDelete(ctx, ovm)
		}
	}
}
```
