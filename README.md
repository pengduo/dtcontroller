# 简介

resource-operator 是一个资源管控工具，提供了针对虚拟化和容器化的资源管控

1. DtCluster
2. DtNode
3. DtModel
4. DtMachine

项目使用前提：
* 部署完成虚拟化方案，支持ESXI, PVE
* 部署一套k8s平台
* 安装kubectl工具

## 使用文档

### 下载源码

```shell
git clone http://gitlab.dtwave-inc.com/meraks/resource-operator.git
```

## 部署主程序

```shell

make docker-build 

make docker-push 

make install 

make deploy 

```


### 部署DtCluster

```shell
kubectl apply -f config/sample/apps_v1_dtcluster_esxi.yaml
```

以上完成了DtCluster的部署，DtCluster作为DtNode的配置数据，必须先于DtNode部署

### 部署DtNode

```shell
kubectl apply -f config/sample/apps_v1_dtnode.yaml
```
DtNode是Node的扩展，其丰富了Node属性，与Node实行一一绑定。


## 开发文档

新增一种资源：`kubebuilder create api --group apps --version v1 --kind sample`

生成crd文件：`make && make manifests`

### DtCluster文档

[DtCluster](docs/dtcluster.md)

### DtNode文档

[DtNode](docs/dtnode.md)

### DtModel文档

[DtModel](docs/dtmodel.md)

### DtMachine文档

[DtMachine](docs/dtmachine.md)

## 注意事项

* 程序仅在kubernetes 1.20版本测试通过

## 待实现

1. 虚拟机加上标签，避免误删除
2. 网络问题
