# config-watcher
- 一个基于分布式系统的etcd监听者

# 依赖
- etcd v3

# 功能
- 监听不丢失：
  - 利用etcd revision概念，即使网络不稳定，导致监听器重连，也不会丢失订阅消息
- 封装了etcd watch模块，用户只需关心：
  - 获取、设置、更改监听对象时，回调函数触发的相关业务逻辑
- 带有初始值设置功能：
  - 分布式系统环境下，利用etcd事务操作，多个节点只会有一个节点执行初始化，无需担忧重复执行，以及开发额外的代码去做初始化操作

# 用法
- 见[`main.go`](main.go)

# 预览
```shell
# put callback（初始化成功时触发）
$ etcdctl get /the/key/you/want/to/watch

$ ./etcd_watcher 
put callback: /the/key/you/want/to/watch=/the/value/you/want/to/init

$ etcdctl get /the/key/you/want/to/watch
/the/key/you/want/to/watch
/the/value/you/want/to/init

# get callback（初始化失败时触发，即已经设置过了）
$ etcdctl get /the/key/you/want/to/watch
/the/key/you/want/to/watch
/the/value/you/want/to/init

$ ./etcd_watcher 
get callback: /the/key/you/want/to/watch=/the/value/you/want/to/init

# change callback（监听对象变化时触发）
$ etcdctl put /the/key/you/want/to/watch 123

$ ./etcd_watcher 
...
change callback: /the/key/you/want/to/watch=123
```