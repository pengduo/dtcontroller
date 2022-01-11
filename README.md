# 简介

dtcontroller是一系列控制器的整合

其包含以下几种资源的控制：

1. Machine

2. DtNode

3. MachineGroup


## 开发文档

新增一种资源：`kubebuilder create api --group apps --version v1 --kind DtNode`

生成crd文件：`make && make manifests`

## DtNode文档
[NtNode](docs/Dtnode.md)

## Machine文档
[Machine](docs/Machine.md)

## MachineGroup文档
[MachineGroup](docs/MachineGroup.md)


