package testrunner

import (
	"os/exec"

	"github.com/tedsuo/ifrit/ginkgomon"
)

type Args struct {
	Address        string
	EtcdCluster    string
	EtcdClientCert string
	EtcdClientKey  string
	EtcdCACert     string
}

func (args Args) ArgSlice() []string {
	return []string{
		"-address", args.Address,
		"-etcdCluster", args.EtcdCluster,
		"-etcdCertFile", args.EtcdClientCert,
		"-etcdKeyFile", args.EtcdClientKey,
		"-etcdCaFile", args.EtcdCACert,
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
