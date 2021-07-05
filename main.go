package main

import (
	"etcd_watcher/watcher"
	"fmt"
)

const (
	watchKey  = "/the/key/you/want/to/watch"
	initValue = "/the/value/you/want/to/init"
)

func ScannerGetCallback(getValue []byte) {
	fmt.Printf("get callback: %s=%s\n", watchKey, string(getValue))
}

func ScannerPutCallback() {
	fmt.Printf("put callback: %s=%s\n", watchKey, initValue)
}

func ScannerChangeCallback(changeValue []byte) {
	fmt.Printf("change callback: %s=%s\n", watchKey, string(changeValue))
}

func main() {
	etcdClient, etcdErrMsg := watcher.NewEtcdClient(
		[]string{"127.0.0.1:2379"},
		2000,
	)
	if etcdClient == nil {
		fmt.Println(etcdErrMsg.Error())
		return
	}

	// add watcher
	etcdClient.AddWatcher(
		&watcher.EtcdWatcher{
			WatchKey:      watchKey,
			InitValue:     initValue,             // 如果不需要初始化，不设即可，watchPutCb也不用设置
			WatchGetCb:    ScannerGetCallback,    // 初始值获取回调
			WatchPutCb:    ScannerPutCallback,    // 配合initValue
			WatchChangeCb: ScannerChangeCallback, // 值变化回调函数
		})

	// start watching
	etcdClient.Watch()

	var ch chan struct{} = make(chan struct{})
	<-ch
}
