package etcd

import (
	"github.com/cloudfoundry-incubator/bbs/db/codec"
	"github.com/coreos/go-etcd/etcd"
)

type StoreClient interface {
	SupportsBinary() bool
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
	client   *etcd.Client
	codecs   *codec.Codecs
	encoding codec.Kind
}

func NewStoreClient(client *etcd.Client, encoding codec.Kind) StoreClient {
	return &storeClient{
		client:   client,
		encoding: encoding,
		codecs:   codec.NewCodecs(encoding),
	}
}

func (sc *storeClient) SupportsBinary() bool {
	return sc.encoding.SupportsBinary()
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

func (sc *storeClient) Set(key string, payload []byte, ttl uint64) (*etcd.Response, error) {
	data, err := sc.codecs.Encode(payload)
	if err != nil {
		return nil, err
	}

	return sc.client.Set(key, string(data), ttl)
}

func (sc *storeClient) Create(key string, payload []byte, ttl uint64) (*etcd.Response, error) {
	data, err := sc.codecs.Encode(payload)
	if err != nil {
		return nil, err
	}

	return sc.client.Create(key, string(data), ttl)
}

func (sc *storeClient) Delete(key string, recursive bool) (*etcd.Response, error) {
	return sc.client.Delete(key, recursive)
}

func (sc *storeClient) DeleteDir(key string) (*etcd.Response, error) {
	return sc.client.DeleteDir(key)
}

func (sc *storeClient) CompareAndSwap(key string, payload []byte, ttl uint64, prevIndex uint64) (*etcd.Response, error) {
	data, err := sc.codecs.Encode(payload)
	if err != nil {
		return nil, err
	}

	res, err := sc.client.CompareAndSwap(key, string(data), ttl, "", prevIndex)
	if err != nil {
		return nil, err
	}

	res.Node.Value = ""
	res.PrevNode.Value = ""

	return res, err
}

func (sc *storeClient) CompareAndDelete(key string, prevIndex uint64) (*etcd.Response, error) {
	res, err := sc.client.CompareAndDelete(key, "", prevIndex)
	if err != nil {
		return nil, err
	}

	res.Node.Value = ""
	res.PrevNode.Value = ""

	return res, err
}

func (sc *storeClient) Watch(
	prefix string,
	waitIndex uint64,
	recursive bool,
	receiver chan *etcd.Response,
	stop chan bool,
) (*etcd.Response, error) {
	var proxy chan *etcd.Response

	if receiver != nil {
		proxy = make(chan *etcd.Response)
		go func() {
			for response := range proxy {
				sc.decode(response.PrevNode)
				sc.decode(response.Node)
				receiver <- response
			}
			close(receiver)
		}()
	}

	response, err := sc.client.Watch(prefix, waitIndex, recursive, proxy, stop)
	if err != nil {
		return nil, err
	}
	sc.decode(response.PrevNode)
	sc.decode(response.Node)
	return response, nil
}

func (sc *storeClient) decode(node *etcd.Node) error {
	if node == nil {
		return nil
	}

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
