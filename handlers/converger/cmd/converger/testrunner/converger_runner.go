package testrunner

import (
	"os/exec"

	"github.com/onsi/gomega/gexec"
	"github.com/tedsuo/ifrit/ginkgomon"
)

type ConvergerRunner struct {
	binPath string
	Session *gexec.Session
	config  Config
}

type Config struct {
	BinPath                     string
	ConvergeRepeatInterval      string
	KickTaskDuration            string
	ExpirePendingTaskDuration   string
	ExpireCompletedTaskDuration string
	ConsulCluster               string
	LogLevel                    string
	BBSAddress                  string
}

func (c *Config) ArgSlice() []string {
	return []string{
		"-logLevel", c.LogLevel,
		"-convergeRepeatInterval", c.ConvergeRepeatInterval,
		"-kickTaskDuration", c.KickTaskDuration,
		"-expirePendingTaskDuration", c.ExpirePendingTaskDuration,
		"-expireCompletedTaskDuration", c.ExpireCompletedTaskDuration,
		"-lockRetryInterval", "1s",
		"-consulCluster", c.ConsulCluster,
		"-bbsAddress", c.BBSAddress,
	}
}

func New(config *Config) *ginkgomon.Runner {
	return ginkgomon.New(ginkgomon.Config{
		Name:          "converger",
		AnsiColorCode: "94m",
		Command:       exec.Command(config.BinPath, config.ArgSlice()...),
		StartCheck:    "acquiring-lock",
	})
}
