package testrunner

import (
	"encoding/json"
	"os"
	"os/exec"
	"time"

	"code.cloudfoundry.org/bbs/cmd/bbs/config"
	"code.cloudfoundry.org/durationjson"

	. "github.com/onsi/gomega"
	ginkgomon "github.com/tedsuo/ifrit/ginkgomon_v2"
)

func New(binPath string, bbsConfig config.BBSConfig) *ginkgomon.Runner {
	if bbsConfig.ReportInterval == 0 {
		bbsConfig.ReportInterval = durationjson.Duration(time.Minute)
	}

	f, err := os.CreateTemp("", "bbs.config")
	Expect(err).NotTo(HaveOccurred())

	err = json.NewEncoder(f).Encode(bbsConfig)
	Expect(err).NotTo(HaveOccurred())

	return ginkgomon.New(ginkgomon.Config{
		Name:              "bbs",
		Command:           exec.Command(binPath, "-config", f.Name()),
		StartCheck:        "bbs.started",
		StartCheckTimeout: 20 * time.Second,
		Cleanup: func() {
			// do not use Expect otherwise a race condition will happen
			os.RemoveAll(f.Name())
		},
	})
}

func WaitForMigration(binPath string, bbsConfig config.BBSConfig) *ginkgomon.Runner {
	runner := New(binPath, bbsConfig)
	runner.StartCheck = "finished-migrations"
	return runner
}
