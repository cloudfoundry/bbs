package metrics

import (
	"os"
	"sync"
	"time"

	"code.cloudfoundry.org/clock"
	logging "code.cloudfoundry.org/diego-logging-client"
	"github.com/tedsuo/ifrit"
)

const (
	DefaultEmitMetricsFrequency = 15 * time.Second

	domainMetricPrefix = "Domain."

	ConvergenceLRPRunsMetric     = "ConvergenceLRPRuns"
	ConvergenceLRPDurationMetric = "ConvergenceLRPDuration"
	LRPsUnclaimedMetric          = "LRPsUnclaimed"
	LRPsClaimedMetric            = "LRPsClaimed"
	LRPsRunningMetric            = "LRPsRunning"
	CrashedActualLRPsMetric      = "CrashedActualLRPs"
	CrashingDesiredLRPsMetric    = "CrashingDesiredLRPs"
	LRPsDesiredMetric            = "LRPsDesired"
	LRPsMissingMetric            = "LRPsMissing"
	LRPsExtraMetric              = "LRPsExtra"
)

//go:generate counterfeiter -o fakes/fake_lrp_stat_metron_notifier.go . LRPStatMetronNotifier
type LRPStatMetronNotifier interface {
	ifrit.Runner

	RecordLRPConvergenceDuration(duration time.Duration)
	RecordFreshDomains(domains []string)
	RecordStateOfLRPs(unclaimed, claimed, running, crashed, crashingDesired int)
	RecordDesiredLRPs(desired int)
	RecordMissingLRPs(missing int)
	RecordExtraLRPs(extras int)
}

type lrpStatMetronNotifier struct {
	clock        clock.Clock
	mutex        sync.Mutex
	metronClient logging.IngressClient

	metrics lrpMetrics
}

type lrpMetrics struct {
	convergenceLRPRuns     uint64
	convergenceLRPDuration time.Duration

	domainsMetric []string

	lrpsUnclaimed       int
	lrpsClaimed         int
	lrpsRunning         int
	crashedActualLRPs   int
	crashingDesiredLRPs int
	lrpsDesired         int
	lrpsMissing         int
	lrpsExtra           int
}

func NewLRPStatMetronNotifier(clock clock.Clock, metronClient logging.IngressClient) LRPStatMetronNotifier {
	return &lrpStatMetronNotifier{
		clock:        clock,
		metronClient: metronClient,
	}
}

func (t *lrpStatMetronNotifier) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	ticker := t.clock.NewTicker(DefaultEmitMetricsFrequency)
	close(ready)
	for {
		select {
		case <-ticker.C():
			t.emitMetrics()
		case <-signals:
			return nil
		}
	}
}

func (lrp *lrpStatMetronNotifier) RecordLRPConvergenceDuration(duration time.Duration) {
	lrp.mutex.Lock()
	defer lrp.mutex.Unlock()

	lrp.metrics.convergenceLRPRuns++
	lrp.metrics.convergenceLRPDuration = duration
}

func (lrp *lrpStatMetronNotifier) RecordFreshDomains(domains []string) {
	lrp.mutex.Lock()
	defer lrp.mutex.Unlock()

	lrp.metrics.domainsMetric = domains
}

func (lrp *lrpStatMetronNotifier) RecordStateOfLRPs(unclaimed, claimed, running, crashed, crashingDesired int) {
	lrp.mutex.Lock()
	defer lrp.mutex.Unlock()

	lrp.metrics.lrpsUnclaimed = unclaimed
	lrp.metrics.lrpsClaimed = claimed
	lrp.metrics.lrpsRunning = running
	lrp.metrics.crashedActualLRPs = crashed
	lrp.metrics.crashingDesiredLRPs = crashingDesired
}

func (lrp *lrpStatMetronNotifier) RecordDesiredLRPs(desired int) {
	lrp.mutex.Lock()
	defer lrp.mutex.Unlock()

	lrp.metrics.lrpsDesired = desired
}

func (lrp *lrpStatMetronNotifier) RecordMissingLRPs(missing int) {
	lrp.mutex.Lock()
	defer lrp.mutex.Unlock()

	lrp.metrics.lrpsMissing = missing
}

func (lrp *lrpStatMetronNotifier) RecordExtraLRPs(extras int) {
	lrp.mutex.Lock()
	defer lrp.mutex.Unlock()

	lrp.metrics.lrpsExtra = extras
}

func (lrp *lrpStatMetronNotifier) emitMetrics() {
	lrp.mutex.Lock()
	defer lrp.mutex.Unlock()

	if lrp.metrics.convergenceLRPRuns > 0 {
		lrp.metronClient.IncrementCounterWithDelta(ConvergenceLRPRunsMetric, lrp.metrics.convergenceLRPRuns)
		lrp.metrics.convergenceLRPRuns = 0
	}

	lrp.metronClient.SendDuration(ConvergenceLRPDurationMetric, lrp.metrics.convergenceLRPDuration)

	for _, domain := range lrp.metrics.domainsMetric {
		lrp.metronClient.SendMetric(domainMetricPrefix+domain, 1)
	}

	lrp.metronClient.SendMetric(LRPsUnclaimedMetric, lrp.metrics.lrpsUnclaimed)
	lrp.metronClient.SendMetric(LRPsClaimedMetric, lrp.metrics.lrpsClaimed)
	lrp.metronClient.SendMetric(LRPsRunningMetric, lrp.metrics.lrpsRunning)
	lrp.metronClient.SendMetric(CrashedActualLRPsMetric, lrp.metrics.crashedActualLRPs)
	lrp.metronClient.SendMetric(CrashingDesiredLRPsMetric, lrp.metrics.crashingDesiredLRPs)
	lrp.metronClient.SendMetric(LRPsDesiredMetric, lrp.metrics.lrpsDesired)
	lrp.metronClient.SendMetric(LRPsMissingMetric, lrp.metrics.lrpsMissing)
	lrp.metronClient.SendMetric(LRPsExtraMetric, lrp.metrics.lrpsExtra)
}
