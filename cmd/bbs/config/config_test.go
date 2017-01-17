package config_test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"time"

	"code.cloudfoundry.org/bbs/cmd/bbs/config"
	"code.cloudfoundry.org/bbs/encryption"
	"code.cloudfoundry.org/debugserver"
	"code.cloudfoundry.org/durationjson"
	"code.cloudfoundry.org/lager/lagerflags"
	"code.cloudfoundry.org/locket"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("BBSConfig", func() {
	var configFilePath, configData string

	BeforeEach(func() {
		configData = `{
  "session_name": "bbs-session",
  "access_log_path": "/var/vcap/sys/log/bbs/access.log",
  "require_ssl": true,
  "ca_file": "/var/vcap/jobs/bbs/config/ca.crt",
  "cert_file": "/var/vcap/jobs/bbs/config/bbs.crt",
  "key_file": "/var/vcap/jobs/bbs/config/bbs.key",
  "listen_address": "0.0.0.0:8889",
  "health_address": "127.0.0.1:8890",
  "advertise_url": "bbs.service.cf.internal",
  "communication_timeout": "20s",
  "desired_lrp_creation_timeout": "1m0s",
  "expire_completed_task_duration": "2m0s",
  "expire_pending_task_duration": "30m0s",
  "converge_repeat_interval": "30s",
  "kick_task_duration": "30s",
  "lock_retry_interval": "5s",
  "lock_ttl": "15s",
  "report_interval": "1m0s",
  "convergence_workers": 20,
  "update_workers": 1000,
  "task_callback_workers": 1000,
  "consul_cluster": "",
  "dropsonde_port": 3457,
  "database_connection_string": "",
  "database_driver": "postgres",
  "max_database_connections": 500,
  "sql_ca_cert_file": "/var/vcap/jobs/bbs/config/sql.ca",
  "auctioneer_address": "https://auctioneer.service.cf.internal:9016",
  "auctioneer_ca_cert": "/var/vcap/jobs/bbs/config/auctioneer.ca",
  "auctioneer_client_cert": "/var/vcap/jobs/bbs/config/auctioneer.crt",
  "auctioneer_client_key": "/var/vcap/jobs/bbs/config/auctioneer.key",
  "auctioneer_require_tls": true,
  "rep_ca_cert": "/var/vcap/jobs/bbs/config/rep.ca",
  "rep_client_cert": "/var/vcap/jobs/bbs/config/rep.crt",
  "rep_client_key": "/var/vcap/jobs/bbs/config/rep.key",
  "rep_client_session_cache_size": 10,
  "rep_require_tls": true,
  "etcd_cluster_urls": [
    "http://127.0.0.1:8500"
  ],
  "etcd_cert_file": "/var/vcap/jobs/bbs/config/etcd.crt",
  "etcd_key_file": "/var/vcap/jobs/bbs/config/etcd.key",
  "etcd_ca_file": "/var/vcap/jobs/bbs/config/etcd.ca",
  "etcd_client_session_cache_size": 10,
  "etcd_max_idle_conns_per_host": 10,
  "active_key_label": "label",
  "encryption_keys": {
    "label": "key"
  },
  "debug_address": "127.0.0.1:17017",
  "log_level": "debug"
}`
	})

	JustBeforeEach(func() {
		configFile, err := ioutil.TempFile("", "config-file")
		Expect(err).NotTo(HaveOccurred())

		n, err := configFile.WriteString(configData)
		Expect(err).NotTo(HaveOccurred())
		Expect(n).To(Equal(len(configData)))

		configFilePath = configFile.Name()
	})

	AfterEach(func() {
		err := os.RemoveAll(configFilePath)
		Expect(err).NotTo(HaveOccurred())
	})

	It("correctly parses the config file", func() {
		bbsConfig, err := config.NewBBSConfig(configFilePath)
		Expect(err).NotTo(HaveOccurred())

		config := config.BBSConfig{
			SessionName:                 "bbs-session",
			AccessLogPath:               "/var/vcap/sys/log/bbs/access.log",
			RequireSSL:                  true,
			CaFile:                      "/var/vcap/jobs/bbs/config/ca.crt",
			CertFile:                    "/var/vcap/jobs/bbs/config/bbs.crt",
			KeyFile:                     "/var/vcap/jobs/bbs/config/bbs.key",
			ListenAddress:               "0.0.0.0:8889",
			HealthAddress:               "127.0.0.1:8890",
			AdvertiseURL:                "bbs.service.cf.internal",
			CommunicationTimeout:        durationjson.Duration(20 * time.Second),
			DesiredLRPCreationTimeout:   durationjson.Duration(1 * time.Minute),
			ExpireCompletedTaskDuration: durationjson.Duration(2 * time.Minute),
			ExpirePendingTaskDuration:   durationjson.Duration(30 * time.Minute),
			ConvergeRepeatInterval:      durationjson.Duration(30 * time.Second),
			KickTaskDuration:            durationjson.Duration(30 * time.Second),
			LockTTL:                     durationjson.Duration(locket.DefaultSessionTTL),
			LockRetryInterval:           durationjson.Duration(locket.RetryInterval),
			ReportInterval:              durationjson.Duration(1 * time.Minute),
			ConvergenceWorkers:          20,
			UpdateWorkers:               1000,
			TaskCallbackWorkers:         1000,
			DropsondePort:               3457,
			DatabaseDriver:              "postgres",
			MaxDatabaseConnections:      500,
			SQLCACertFile:               "/var/vcap/jobs/bbs/config/sql.ca",
			AuctioneerAddress:           "https://auctioneer.service.cf.internal:9016",
			AuctioneerCACert:            "/var/vcap/jobs/bbs/config/auctioneer.ca",
			AuctioneerClientCert:        "/var/vcap/jobs/bbs/config/auctioneer.crt",
			AuctioneerClientKey:         "/var/vcap/jobs/bbs/config/auctioneer.key",
			AuctioneerRequireTLS:        true,
			RepCACert:                   "/var/vcap/jobs/bbs/config/rep.ca",
			RepClientCert:               "/var/vcap/jobs/bbs/config/rep.crt",
			RepClientKey:                "/var/vcap/jobs/bbs/config/rep.key",
			RepClientSessionCacheSize:   10,
			RepRequireTLS:               true,
			EncryptionConfig: encryption.EncryptionConfig{
				ActiveKeyLabel: "label",
				EncryptionKeys: map[string]string{
					"label": "key",
				},
			},
			ETCDConfig: config.ETCDConfig{
				ClusterUrls:            []string{"http://127.0.0.1:8500"},
				CaFile:                 "/var/vcap/jobs/bbs/config/etcd.ca",
				CertFile:               "/var/vcap/jobs/bbs/config/etcd.crt",
				KeyFile:                "/var/vcap/jobs/bbs/config/etcd.key",
				ClientSessionCacheSize: 10,
				MaxIdleConnsPerHost:    10,
			},
			DebugServerConfig: debugserver.DebugServerConfig{
				DebugAddress: "127.0.0.1:17017",
			},
			LagerConfig: lagerflags.LagerConfig{
				LogLevel: "debug",
			},
		}

		Expect(bbsConfig).To(Equal(config))
	})

	Context("when the file does not exist", func() {
		It("returns an error", func() {
			_, err := config.NewBBSConfig("foobar")
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when the file does not contain valid json", func() {
		BeforeEach(func() {
			configData = "{{"
		})

		It("returns an error", func() {
			_, err := config.NewBBSConfig(configFilePath)
			Expect(err).To(HaveOccurred())
		})

	})

	Context("when the file contains invalid durations", func() {
		BeforeEach(func() {
			configData = `{"expire_completed_task_duration": "4234342342"}`
		})

		It("returns an error", func() {
			_, err := config.NewBBSConfig(configFilePath)
			Expect(err).To(MatchError(ContainSubstring("missing unit")))
		})
	})

	Context("default values", func() {
		BeforeEach(func() {
			configData = `{}`
		})

		It("uses default values when they are not specified", func() {
			bbsConfig, err := config.NewBBSConfig(configFilePath)
			Expect(err).NotTo(HaveOccurred())

			Expect(bbsConfig).To(Equal(config.DefaultConfig()))
		})

		Context("when serialized from BBSConfig", func() {
			BeforeEach(func() {
				bbsConfig := config.BBSConfig{}
				bytes, err := json.Marshal(bbsConfig)
				Expect(err).NotTo(HaveOccurred())
				configData = string(bytes)
			})

			It("uses default values when they are not specified", func() {
				bbsConfig, err := config.NewBBSConfig(configFilePath)
				Expect(err).NotTo(HaveOccurred())

				Expect(bbsConfig).To(Equal(config.DefaultConfig()))
			})
		})
	})
})
