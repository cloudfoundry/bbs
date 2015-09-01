package etcd

import "github.com/coreos/go-etcd/etcd"

type StoreClient interface {
	Get(key string, sort bool, recursive bool) (*etcd.Response, error)
	Set(key string, value []byte, ttl uint64) (*etcd.Response, error)
	Create(key string, value []byte, ttl uint64) (*etcd.Response, error)
	Delete(key string, recursive bool) (*etcd.Response, error)
	DeleteDir(key string) (*etcd.Response, error)
	CompareAndSwap(key string, value []byte, ttl uint64, prevIndex uint64) (*etcd.Response, error)
	CompareAndDelete(key string, prevIndex uint64) (*etcd.Response, error)
	Watch(prefix string, waitIndex uint64, recursive bool, receiver chan *etcd.Response, stop chan bool) (*etcd.Response, error)
}

type storeClient struct {
	client *etcd.Client
}

func NewStoreClient(client *etcd.Client) StoreClient {
	return &storeClient{client}
}

func (sc *storeClient) Get(key string, sort bool, recursive bool) (*etcd.Response, error) {
	return sc.client.Get(key, sort, recursive)
}

func (sc *storeClient) Set(key string, payload []byte, ttl uint64) (*etcd.Response, error) {
	return sc.client.Set(key, string(payload), ttl)
}

func (sc *storeClient) Create(key string, payload []byte, ttl uint64) (*etcd.Response, error) {
	return sc.client.Create(key, string(payload), ttl)
}

func (sc *storeClient) Delete(key string, recursive bool) (*etcd.Response, error) {
	return sc.client.Delete(key, recursive)
}

func (sc *storeClient) DeleteDir(key string) (*etcd.Response, error) {
	return sc.client.DeleteDir(key)
}

func (sc *storeClient) CompareAndSwap(key string, payload []byte, ttl uint64, prevIndex uint64) (*etcd.Response, error) {
	return sc.client.CompareAndSwap(key, string(payload), ttl, "", prevIndex)
}

func (sc *storeClient) CompareAndDelete(key string, prevIndex uint64) (*etcd.Response, error) {
	return sc.client.CompareAndDelete(key, "", prevIndex)
}

func (sc *storeClient) Watch(
	prefix string,
	waitIndex uint64,
	recursive bool,
	receiver chan *etcd.Response,
	stop chan bool,
) (*etcd.Response, error) {
	return sc.client.Watch(prefix, waitIndex, recursive, receiver, stop)
}
