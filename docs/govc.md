## govc使用

1. 创建虚拟机

```shell
govc vm.create -m 256 -c 1 -disk 10G -ds bigdata test1 
govc vm.create -m 256 -c 1 -disk 10G -ds bigdata -dc ha-datacenter test2
```
* `-ds`指定Datastore
* `-dc`指定DataCenter 可选
* `-c`指定CPU核
* `-m`指定内存
* `-disk`指定磁盘


