package etcd

import (
	"github.com/cloudfoundry-incubator/bbs/db/codec"
	"github.com/coreos/go-etcd/etcd"
)

type StoreClient interface {
	Get(key string, sort bool, recursive bool) (*etcd.Response, error)
	Set(key string, value []byte, ttl uint64) (*etcd.Response, error)
	Create(key string, value []byte, ttl uint64) (*etcd.Response, error)
	Delete(key string, recursive bool) (*etcd.Response, error)
	CompareAndSwap(key string, value []byte, ttl uint64, prevIndex uint64) (*etcd.Response, error)
	CompareAndDelete(key string, prevIndex uint64) (*etcd.Response, error)
	Watch(prefix string, waitIndex uint64, recursive bool, receiver chan *etcd.Response, stop chan bool) (*etcd.Response, error)
}

type storeClient struct {
	client *etcd.Client
	codecs *codec.Codecs
}

func NewStoreClient(client *etcd.Client, defaultEncoding codec.Kind) StoreClient {
	return &storeClient{
		client: client,
		codecs: codec.NewCodecs(defaultEncoding),
	}
}

func (sc *storeClient) Get(key string, sort bool, recursive bool) (*etcd.Response, error) {
	response, err := sc.client.Get(key, sort, recursive)
	if err != nil {
		return nil, err
	}

	err = sc.decode(response.Node)
	if err != nil {
		return nil, err
	}

	return response, err
}

func (sc *storeClient) decode(node *etcd.Node) error {
	payload, err := sc.codecs.Decode([]byte(node.Value))
	if err != nil {
		return err
	}

	node.Value = string(payload)

	for _, n := range node.Nodes {
		if err := sc.decode(n); err != nil {
			return err
		}
	}

	return nil
}

func (sc *storeClient) Set(key string, value []byte, ttl uint64) (*etcd.Response, error) {
	return sc.client.Set(key, string(value), ttl)
}

func (sc *storeClient) Create(key string, value []byte, ttl uint64) (*etcd.Response, error) {
	return sc.client.Create(key, string(value), ttl)
}

func (sc *storeClient) Delete(key string, recursive bool) (*etcd.Response, error) {
	return sc.client.Delete(key, recursive)
}

func (sc *storeClient) CompareAndSwap(key string, value []byte, ttl uint64, prevIndex uint64) (*etcd.Response, error) {
	return sc.client.CompareAndSwap(key, string(value), ttl, "", prevIndex)
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
