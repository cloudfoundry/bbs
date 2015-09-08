package testrunner

import (
	"os/exec"

	"github.com/tedsuo/ifrit/ginkgomon"
)

type Args struct {
	Address           string
	AdvertiseURL      string
	AuctioneerAddress string
	ConsulCluster     string
	EtcdCluster       string
	EtcdClientCert    string
	EtcdClientKey     string
	EtcdCACert        string
}

func (args Args) ArgSlice() []string {
	return []string{
		"-advertiseURL", args.AdvertiseURL,
		"-auctioneerAddress", args.AuctioneerAddress,
		"-consulCluster", args.ConsulCluster,
		"-etcdCaFile", args.EtcdCACert,
		"-etcdCertFile", args.EtcdClientCert,
		"-etcdCluster", args.EtcdCluster,
		"-etcdKeyFile", args.EtcdClientKey,
		"-listenAddress", args.Address,
		"-logLevel", "debug",
	}
}

func New(binPath string, args Args) *ginkgomon.Runner {
	return ginkgomon.New(ginkgomon.Config{
		Name:       "bbs",
		Command:    exec.Command(binPath, args.ArgSlice()...),
		StartCheck: "bbs.started",
	})
}
