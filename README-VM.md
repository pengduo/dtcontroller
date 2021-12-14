## vmsdk调用文档

示例：
[vm-practice](http://gitlab.dtwave-inc.com/shuqi-devops/vm-practice.git)

```golang
package main

import (
	"log"
	"net/url"
	"vm-practice/vmhost"

	"github.com/vmware/govmomi"
	"golang.org/x/net/context"
)

const (
	envURL          = "https://linjb@vsphere.local:LIN115jinbao!@192.168.123.138/sdk"
	envUserName     = "linjb@vsphere.local"
	envPassword     = "LIN115jinbao!"
	envInsecure     = "true"
	libraryName     = "library1"
	libraryItemName = "centos7"
	libraryItemType = "ovf"
)

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