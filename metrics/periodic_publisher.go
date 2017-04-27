package metrics

import (
	"os/exec"
	"strconv"
	"strings"
	"time"

	m "github.com/cloudfoundry/dropsonde/metrics"
)

type MetricSample struct {
	Name    string
	Value   float64
	Unit    string
	Counter bool
}

var (
	MetricCh chan MetricSample
)

func init() {
	MetricCh = make(chan MetricSample, 100)
	go send()
	go networkStats()
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func send() {
	metrics := make(map[string]MetricSample)
	ticker := time.NewTicker(5 * time.Second)

	for {
		select {
		case metric := <-MetricCh:
			metrics = addMetric(metrics, metric)
		case <-ticker.C:
			go flush(metrics)
			metrics = make(map[string]MetricSample)
		}
	}
}

func addMetric(metrics map[string]MetricSample, metric MetricSample) map[string]MetricSample {
	m, ok := metrics[metric.Name]
	if ok {
		if m.Counter {
			metric.Value += m.Value
		} else {
			metric.Value = max(m.Value, metric.Value)
		}
	}
	if metric.Unit == "nanos" {
		metrics = addMetric(metrics, MetricSample{
			Name:    metric.Name + "-counter",
			Value:   1,
			Counter: true,
		})
	}
	metrics[metric.Name] = metric
	return metrics
}

func flush(metrics map[string]MetricSample) {
	for _, metric := range metrics {
		if metric.Counter {
			m.AddToCounter(metric.Name, uint64(metric.Value))
			continue
		}
		m.SendValue(metric.Name, metric.Value, metric.Unit)
	}
}

func networkStats() {
	for {
		cmd := exec.Command("netstat", "-s")
		output, err := cmd.CombinedOutput()
		if err != nil {
			panic(err)
		}
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			for _, v := range []string{"segments retransmited", "segments received", "segments send out"} {
				if strings.HasSuffix(line, v) {
					columns := strings.Fields(line)
					value, err := strconv.Atoi(columns[0])
					if err != nil {
						panic("cannot parse: " + columns[0])
					}
					name := strings.Replace(v, " ", "-", -1)
					m.SendValue(name, float64(value), "Metric")
				}
			}
		}
		time.Sleep(time.Second)
	}
}
