# 简介

dtcontroller是一系列控制器的整合

其包含以下几种资源的控制：

1. Machine

2. Iot

3. Proxy

4. DtNode

## 开发文档

新增一种资源：`kubebuilder create api --group apps --version v1 --kind DtNode`

生成crd文件：`make manifests`

## DtNode文档
[NtNode](docs/Dtnode.md)

## Machine文档
[Machine](docs/Machine.md)

## Proxy文档
[Proxy](docs/Proxy.md)

## Iot文档
[Iot](docs/Iot.md)

