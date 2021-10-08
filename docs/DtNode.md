# DtNode

## 资源解释

DtNode是一种全局性的资源，其不属于任一名称空间。DtNode代表节点资源，例如物理机资源、虚拟机资源以及ECS资源，kubernetes集群中只存储，DtNode的元数据信息，包含：Ip, Mac, Type, User, Password, Cpu, Memory, Hostname, Disk等信息。

## 资源操作

针对DtNode资源而言，k8s具备以下作用：

* 数据库
* 资源管理平台

### 增加资源

增加资源，即通过声明式文件，注册一个已经存在的DtNode到kubernetes集群中，该操作可以是外部已有的虚拟机资源，纳入k8s管理。
示例：
```yaml 
apiVersion: apps.dtwave.com/v1
kind: DtNode
metadata:
  name: dtnode-sample
spec:
  # Add fields here
  foo: bar
```

### 修改资源

修改声明式文件，并提交到k8s集群，由集群修改相关数据

### 查找资源

通过API查询指定资源

### 删除资源

删除不需要的DtNode资源

### 使用资源

使用资源的时候，需要声明成Machine类型的资源，k8s将已经托管的DtNode重新分配成Machine使用