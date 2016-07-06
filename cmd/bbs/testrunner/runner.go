package testrunner

import (
	"os/exec"
	"strconv"
	"time"

	"github.com/tedsuo/ifrit/ginkgomon"
)

type Args struct {
	Address                    string
	AdvertiseURL               string
	AuctioneerAddress          string
	ConsulCluster              string
	DropsondePort              int
	EtcdCACert                 string
	EtcdClientCert             string
	EtcdClientKey              string
	EtcdClientSessionCacheSize int
	EtcdCluster                string
	EtcdMaxIdleConnsPerHost    int

	HealthAddress string

	DatabaseConnectionString string
	DatabaseDriver           string

	MetricsReportInterval time.Duration

	ActiveKeyLabel string
	EncryptionKeys []string

	RequireSSL bool
	CAFile     string
	KeyFile    string
	CertFile   string

	ConvergeRepeatInterval      time.Duration
	KickTaskDuration            time.Duration
	ExpireCompletedTaskDuration time.Duration
	ExpirePendingTaskDuration   time.Duration
}

func (args Args) ArgSlice() []string {
	arguments := []string{
		"-advertiseURL", args.AdvertiseURL,
		"-auctioneerAddress", args.AuctioneerAddress,
		"-consulCluster", args.ConsulCluster,
		"-dropsondePort", strconv.Itoa(args.DropsondePort),
		"-etcdCaFile", args.EtcdCACert,
		"-etcdCertFile", args.EtcdClientCert,
		"-etcdCluster", args.EtcdCluster,
		"-etcdKeyFile", args.EtcdClientKey,
		"-etcdClientSessionCacheSize", strconv.Itoa(args.EtcdClientSessionCacheSize),
		"-etcdMaxIdleConnsPerHost", strconv.Itoa(args.EtcdMaxIdleConnsPerHost),
		"-databaseConnectionString", args.DatabaseConnectionString,
		"-databaseDriver", args.DatabaseDriver,
		"-healthAddress", args.HealthAddress,
		"-listenAddress", args.Address,
		"-logLevel", "debug",
		"-metricsReportInterval", args.MetricsReportInterval.String(),
		"-activeKeyLabel", args.ActiveKeyLabel,
		"-requireSSL=" + strconv.FormatBool(args.RequireSSL),
		"-caFile", args.CAFile,
		"-certFile", args.CertFile,
		"-keyFile", args.KeyFile,
	}

	for _, key := range args.EncryptionKeys {
		arguments = append(arguments, "-encryptionKey="+key)
	}

	if args.ConvergeRepeatInterval > 0 {
		arguments = append(arguments, "-convergeRepeatInterval", args.ConvergeRepeatInterval.String())
	}

	return arguments
}

func New(binPath string, args Args) *ginkgomon.Runner {
	if args.MetricsReportInterval == 0 {
		args.MetricsReportInterval = time.Minute
	}

	return ginkgomon.New(ginkgomon.Config{
		Name:       "bbs",
		Command:    exec.Command(binPath, args.ArgSlice()...),
		StartCheck: "bbs.started",
	})
}

func WaitForMigration(binPath string, args Args) *ginkgomon.Runner {
	if args.MetricsReportInterval == 0 {
		args.MetricsReportInterval = time.Minute
	}

	return ginkgomon.New(ginkgomon.Config{
		Name:       "bbs",
		Command:    exec.Command(binPath, args.ArgSlice()...),
		StartCheck: "finished-migrations",
	})
}
