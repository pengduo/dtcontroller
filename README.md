# 简介

resource-operator 是一个资源管控工具，提供了针对虚拟化和容器话的资源管控

1. Machine

2. DtNode

3. MachineGroup

项目使用前提：
* 安装ESXI
* 安装ESXI管理工具，Vcenter
* 部署一套k8s平台
* 安装kubectl工具

## 使用文档

### 部署DtNode

新建文件 `dtnode-sample.yaml`
```yaml
apiVersion: apps.dtwave.com/v1
kind: DtNode
metadata:
  name: dtnode-sample
spec:
  provider: vmware # 提供商
  ip: "192.168.123.138" #vcenter ip
  user: "linjb@vsphere.local" # vcenter 登录名
  password: "LIN115jinbao!" # 密码·
  datacenter: "sss" # dc名称
  datastore: "bigdata" # ds名称
```

```shell
kubectl apply -f dtnode-sample.yaml
```

以上完成了Dtnode的部署，Dtnode作为vmware的抽象，必须先于Machine部署

### 部署Machine

新建文件`machine-sample.yaml`
```yaml 
apiVersion: apps.dtwave.com/v1
kind: Machine
metadata:
  name: machine-sample
spec:
  type: bare ## 类型，可选bare clone ovf
  user: deploy ## 用户名
  dtnode: "dtnode-sample" ## 使用的dtnode
  password: deploy@1298 ## 密码
  cpu: 2 ## cpu 核
  memory: 256 ##内存 MB
  disk: "100" ## 磁盘GB
```

```shell
kubectl apply -f machine-sample.yaml
```

machine作为虚拟机实例的抽象，从Dtnode上分配出来

## 开发文档

新增一种资源：`kubebuilder create api --group apps --version v1 --kind DtNode`

生成crd文件：`make && make manifests`

## DtNode文档
[NtNode](docs/Dtnode.md)

## Machine文档
[Machine](docs/Machine.md)

## MachineGroup文档
[MachineGroup](docs/MachineGroup.md)

