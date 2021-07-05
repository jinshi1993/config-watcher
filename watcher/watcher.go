package watcher

import (
	"context"
	"time"

	"go.etcd.io/etcd/clientv3"
)

type EtcdWatcher struct {
	WatchKey      string
	watchRevision int64

	InitValue     string
	WatchGetCb    func([]byte)
	WatchPutCb    func()
	WatchChangeCb func([]byte)
}

type EtcdClient struct {
	etcdClient *clientv3.Client
	timeout    int64

	watchers []*EtcdWatcher
}

func NewEtcdClient(etcdAddress []string, timeoutMS int64) (*EtcdClient, error) {
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints:   etcdAddress,
		DialTimeout: time.Duration(timeoutMS) * time.Millisecond,
	})

	if err != nil {
		return nil, err
	}

	return &EtcdClient{
		etcdClient: etcdClient,
		timeout:    timeoutMS,
	}, nil
}

func (ec *EtcdClient) GetClient() *clientv3.Client {
	return ec.etcdClient
}

func (ec *EtcdClient) initWatcher(etcdWatcher *EtcdWatcher) error {
	contextTimeout, cancel := context.WithTimeout(
		context.Background(),
		time.Duration(ec.timeout)*time.Millisecond,
	)
	defer cancel()

	cmp := []clientv3.Cmp{}
	get := []clientv3.Op{}
	put := []clientv3.Op{}

	cmp = append(cmp, clientv3.Compare(clientv3.CreateRevision(etcdWatcher.WatchKey), ">", 0))
	get = append(get, clientv3.OpGet(etcdWatcher.WatchKey))
	if etcdWatcher.InitValue != "" {
		put = append(put, clientv3.OpPut(etcdWatcher.WatchKey, etcdWatcher.InitValue))
	}

	resp, err := ec.etcdClient.Txn(contextTimeout).If(cmp...).Then(get...).Else(put...).Commit()
	if err != nil {
		return err
	}

	etcdWatcher.watchRevision = resp.Header.Revision

	if resp.Succeeded {
		if etcdWatcher.WatchGetCb != nil {
			etcdWatcher.WatchGetCb(resp.Responses[0].GetResponseRange().Kvs[0].Value[:])
		}
	} else {
		if etcdWatcher.WatchPutCb != nil {
			etcdWatcher.WatchPutCb()
		}
	}

	return nil
}

func (ec *EtcdClient) AddWatcher(etcdWatcher *EtcdWatcher) error {
	if err := ec.initWatcher(etcdWatcher); err != nil {
		return err
	}
	ec.watchers = append(ec.watchers, etcdWatcher)
	return nil
}

func (ec *EtcdClient) Watch() {
	for _, etcdWatcher := range ec.watchers {
		ec.watch(etcdWatcher)
	}
}

func (ec *EtcdClient) watch(ew *EtcdWatcher) {
	go func() {
		watchCtx, watchCancel := context.WithCancel(context.TODO())
		watchChannel := ec.etcdClient.Watch(
			watchCtx,
			ew.WatchKey,
			clientv3.WithFilterDelete(),
			clientv3.WithRev(ew.watchRevision+1),
		)

		for w := range watchChannel {
			if w.Canceled {
				break
			} else {
				for _, ev := range w.Events {
					if ew.WatchChangeCb != nil {
						ew.WatchChangeCb(ev.Kv.Value[:])
					}
				}
			}
		}

		watchCancel()
		ec.initWatcher(ew)
		ec.watch(ew)
	}()
}
