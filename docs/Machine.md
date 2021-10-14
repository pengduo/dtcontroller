# Machine

## 资源解释

Machine是一种全局性的资源，其不属于任何一个名称空间。Machine代表分配出来的计算实例，场景类似于vmware的虚拟机分配。Machine可以分配的资源来源于DtNode添加的资源池，且只有Pysical类型的DtNode可以再次做虚拟化，其他类型的无法实现。

## 字段解释

* DtNode: 主动填写/程序回填，关联DtNode信息
* HostName: 程序回填，分配的主机名
* Ip: 程序回填，分配的IP
* Mac: 程序回填，分配的Mac地址
* User: 主动填写/程序回填，分配的用户名
* Password: 主动填写/程序回填，分配的密码
* Cpu: 主动填写，申请的机器配置信息
* Memory: 主动填写，申请的机器配置
* Disk: 主动填写，申请的的机器配置
* Command: 主动填写，初始化的时候执行的脚本
* Labels: 主动填写，标签

## 资源操作

针对Machine资源而言，需要实现以下功能：
1. Machine资源的申请
2. Machine资源的修改
3. Machine资源的退还
4. Machine资源的查询

### 申请资源

申请Machine，即用户通过声明式文件，申请一个Machine类型的计算资源。该操作会在已经管理的DtNode资源上，进行资源的划分，给出一个计算资源实例。实例：
```yaml 
apiVersion: apps.dtwave.com/v1
kind: Machine
metadata:
  name: machine-sample
spec:
  command: "sum=0; for i in `seq 1 100`; do sum=$[$i+$sum]; done; echo $sum"
  user: deploy
  dtnode: node1
  password: deploy@1298
  cpu: 2
  memory: 1
  disk: 100
  labels:
    dtwave-env: "prd"
```

