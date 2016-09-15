package etcd

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"code.cloudfoundry.org/cfhttp"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/runtimeschema/metric"
	"code.cloudfoundry.org/urljoiner"
)

var errRedirected = errors.New("redirected to leader")

const (
	etcdLeader                = metric.Metric("ETCDLeader")
	etcdReceivedBandwidthRate = metric.BytesPerSecond("ETCDReceivedBandwidthRate")
	etcdSentBandwidthRate     = metric.BytesPerSecond("ETCDSentBandwidthRate")
	etcdReceivedRequestRate   = metric.RequestsPerSecond("ETCDReceivedRequestRate")
	etcdSentRequestRate       = metric.RequestsPerSecond("ETCDSentRequestRate")
	etcdRaftTerm              = metric.Metric("ETCDRaftTerm")
	etcdWatchers              = metric.Metric("ETCDWatchers")
)

type ETCDMetrics struct {
	logger lager.Logger

	etcdCluster []string

	client *http.Client
}

func NewETCDMetrics(logger lager.Logger, etcdOptions *ETCDOptions) (*ETCDMetrics, error) {
	var tlsConfig *tls.Config
	if etcdOptions.CertFile != "" && etcdOptions.KeyFile != "" {
		var err error
		tlsConfig, err = cfhttp.NewTLSConfig(etcdOptions.CertFile, etcdOptions.KeyFile, etcdOptions.CAFile)
		if err != nil {
			return nil, err
		}
		tlsConfig.ClientSessionCache = tls.NewLRUClientSessionCache(etcdOptions.ClientSessionCacheSize)
	}

	client := cfhttp.NewClient()
	client.CheckRedirect = func(*http.Request, []*http.Request) error {
		return errRedirected
	}

	if tr, ok := client.Transport.(*http.Transport); ok {
		tr.TLSClientConfig = tlsConfig
	} else {
		return nil, errors.New("Invalid transport")
	}

	return &ETCDMetrics{
		logger: logger,

		etcdCluster: etcdOptions.ClusterUrls,

		client: client,
	}, nil
}

func (t *ETCDMetrics) Send() {
	for i, etcdAddr := range t.etcdCluster {
		t.sendLeaderStats(etcdAddr, i)
	}

	t.sendSelfStats()
}

func (t *ETCDMetrics) sendLeaderStats(etcdAddr string, index int) {
	resp, err := t.client.Get(t.leaderStatsEndpoint(etcdAddr))
	if err != nil {
		t.logger.Error("failed-to-collect-stats", err)
		return
	}

	defer resp.Body.Close()

	var stats etcdLeaderStats

	err = json.NewDecoder(resp.Body).Decode(&stats)
	if err != nil {
		t.logger.Error("failed-to-unmarshal-stats", err)
		return
	}

	err = etcdLeader.Send(index)
	if err != nil {
		t.logger.Error("failed-to-send-etcd-leader-metric", err)
	}

	var storeStats etcdStoreStats

	resp, err = t.client.Get(t.storeStatsEndpoint(etcdAddr))
	if err != nil {
		t.logger.Error("failed-to-collect-stats", err)
		return
	}

	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&storeStats)
	if err != nil {
		t.logger.Error("failed-to-unmarshal-stats", err)
		return
	}

	resp, err = t.client.Get(t.keysEndpoint((etcdAddr)))
	if err != nil {
		t.logger.Error("failed-to-get-keys", err)
		return
	}

	resp.Body.Close()

	raftTermHeader := resp.Header.Get("X-Raft-Term")

	raftTerm, err := strconv.ParseInt(raftTermHeader, 10, 0)
	if err != nil {
		t.logger.Error("failed-to-parse-raft-term", err, lager.Data{
			"term": raftTermHeader,
		})
		return
	}

	err = etcdRaftTerm.Send(int(raftTerm))
	if err != nil {
		t.logger.Error("failed-to-send-etcd-raft-term-metric", err)
	}

	err = etcdWatchers.Send(int(storeStats.Watchers))
	if err != nil {
		t.logger.Error("failed-to-send-etcd-watchers-metric", err)
	}

	return
}

func (t *ETCDMetrics) sendSelfStats() {
	var receivedRequestsPerSecond float64
	var sentRequestsPerSecond float64

	var receivedBandwidthRate float64
	var sentBandwidthRate float64

	for _, addr := range t.etcdCluster {
		var selfStats etcdServerStats

		resp, err := t.client.Get(t.selfStatsEndpoint(addr))
		if err != nil {
			t.logger.Error("failed-to-collect-stats", err)
			return
		}

		defer resp.Body.Close()

		err = json.NewDecoder(resp.Body).Decode(&selfStats)
		if err != nil {
			t.logger.Error("failed-to-unmarshal-stats", err)
			return
		}

		if selfStats.RecvingPkgRate != nil {
			receivedRequestsPerSecond += *selfStats.RecvingPkgRate
		}

		if selfStats.RecvingBandwidthRate != nil {
			receivedBandwidthRate += *selfStats.RecvingBandwidthRate
		}

		if selfStats.SendingPkgRate != nil {
			sentRequestsPerSecond += *selfStats.SendingPkgRate
		}

		if selfStats.SendingBandwidthRate != nil {
			sentBandwidthRate += *selfStats.SendingBandwidthRate
		}
	}

	err := etcdSentBandwidthRate.Send(sentBandwidthRate)
	if err != nil {
		t.logger.Error("failed-to-send-etcd-sent-bandwidth-rate-metric", err)
	}
	err = etcdSentRequestRate.Send(float64(sentRequestsPerSecond))
	if err != nil {
		t.logger.Error("failed-to-send-etcd-sent-request-rate-metric", err)
	}

	err = etcdReceivedBandwidthRate.Send(receivedBandwidthRate)
	if err != nil {
		t.logger.Error("failed-to-send-etcd-received-bandwidth-rate-metric", err)
	}
	err = etcdReceivedRequestRate.Send(float64(receivedRequestsPerSecond))
	if err != nil {
		t.logger.Error("failed-to-send-etcd-received-request-rate-metric", err)
	}
}

func (t *ETCDMetrics) leaderStatsEndpoint(etcdAddr string) string {
	return urljoiner.Join(etcdAddr, "v2", "stats", "leader")
}

func (t *ETCDMetrics) selfStatsEndpoint(etcdAddr string) string {
	return urljoiner.Join(etcdAddr, "v2", "stats", "self")
}

func (t *ETCDMetrics) storeStatsEndpoint(etcdAddr string) string {
	return urljoiner.Join(etcdAddr, "v2", "stats", "store")
}

func (t *ETCDMetrics) keysEndpoint(etcdAddr string) string {
	return urljoiner.Join(etcdAddr, "v2", "keys")
}

type etcdLeaderStats struct {
	Leader    string `json:"leader"`
	Followers map[string]struct {
		Latency struct {
			Current           float64 `json:"current"`
			Average           float64 `json:"average"`
			averageSquare     float64
			StandardDeviation float64 `json:"standardDeviation"`
			Minimum           float64 `json:"minimum"`
			Maximum           float64 `json:"maximum"`
		} `json:"latency"`

		Counts struct {
			Fail    uint64 `json:"fail"`
			Success uint64 `json:"success"`
		} `json:"counts"`
	} `json:"followers"`
}

type etcdServerStats struct {
	Name  string `json:"name"`
	State string `json:"state"`

	LeaderInfo struct {
		Name   string `json:"leader"`
		Uptime string `json:"uptime"`
	} `json:"leaderInfo"`

	RecvAppendRequestCnt uint64   `json:"recvAppendRequestCnt,"`
	RecvingPkgRate       *float64 `json:"recvPkgRate,omitempty"`
	RecvingBandwidthRate *float64 `json:"recvBandwidthRate,omitempty"`

	SendAppendRequestCnt uint64   `json:"sendAppendRequestCnt"`
	SendingPkgRate       *float64 `json:"sendPkgRate,omitempty"`
	SendingBandwidthRate *float64 `json:"sendBandwidthRate,omitempty"`
}

type etcdStoreStats struct {
	Watchers uint64 `json:"watchers"`
}
