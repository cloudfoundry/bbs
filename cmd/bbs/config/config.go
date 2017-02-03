package config

import (
	"encoding/json"
	"os"
	"time"

	"code.cloudfoundry.org/bbs/encryption"
	"code.cloudfoundry.org/debugserver"
	"code.cloudfoundry.org/durationjson"
	"code.cloudfoundry.org/lager/lagerflags"
	"code.cloudfoundry.org/locket"
)

type BBSConfig struct {
	SessionName                 string                `json:"session_name,omitempty"`
	AccessLogPath               string                `json:"access_log_path,omitempty"`
	RequireSSL                  bool                  `json:"require_ssl,omitempty"`
	CaFile                      string                `json:"ca_file,omitempty"`
	CertFile                    string                `json:"cert_file,omitempty"`
	KeyFile                     string                `json:"key_file,omitempty"`
	ListenAddress               string                `json:"listen_address,omitempty"`
	HealthAddress               string                `json:"health_address,omitempty"`
	AdvertiseURL                string                `json:"advertise_url,omitempty"`
	CommunicationTimeout        durationjson.Duration `json:"communication_timeout,omitempty"`
	DesiredLRPCreationTimeout   durationjson.Duration `json:"desired_lrp_creation_timeout,omitempty"`
	ExpireCompletedTaskDuration durationjson.Duration `json:"expire_completed_task_duration,omitempty"`
	ExpirePendingTaskDuration   durationjson.Duration `json:"expire_pending_task_duration,omitempty"`
	ConvergeRepeatInterval      durationjson.Duration `json:"converge_repeat_interval,omitempty"`
	KickTaskDuration            durationjson.Duration `json:"kick_task_duration,omitempty"`
	LockRetryInterval           durationjson.Duration `json:"lock_retry_interval,omitempty"`
	LockTTL                     durationjson.Duration `json:"lock_ttl,omitempty"`
	ReportInterval              durationjson.Duration `json:"report_interval,omitempty"`
	ConvergenceWorkers          int                   `json:"convergence_workers,omitempty"`
	UpdateWorkers               int                   `json:"update_workers,omitempty"`
	TaskCallbackWorkers         int                   `json:"task_callback_workers,omitempty"`
	ConsulCluster               string                `json:"consul_cluster,omitempty"`
	DropsondePort               int                   `json:"dropsonde_port,omitempty"`
	DatabaseConnectionString    string                `json:"database_connection_string"`
	DatabaseDriver              string                `json:"database_driver,omitempty"`
	MaxOpenDatabaseConnections  int                   `json:"max_open_database_connections,omitempty"`
	MaxIdleDatabaseConnections  int                   `json:"max_idle_database_connections,omitempty"`
	SQLCACertFile               string                `json:"sql_ca_cert_file,omitempty"`
	AuctioneerAddress           string                `json:"auctioneer_address,omitempty"`
	AuctioneerCACert            string                `json:"auctioneer_ca_cert,omitempty"`
	AuctioneerClientCert        string                `json:"auctioneer_client_cert,omitempty"`
	AuctioneerClientKey         string                `json:"auctioneer_client_key,omitempty"`
	AuctioneerRequireTLS        bool                  `json:"auctioneer_require_tls,omitempty"`
	RepCACert                   string                `json:"rep_ca_cert,omitempty"`
	RepClientCert               string                `json:"rep_client_cert,omitempty"`
	RepClientKey                string                `json:"rep_client_key,omitempty"`
	RepClientSessionCacheSize   int                   `json:"rep_client_session_cache_size,omitempty"`
	RepRequireTLS               bool                  `json:"rep_require_tls,omitempty"`
	LocketAddress               string                `json:"locket_address,omitempty"`
	SkipConsulLock              bool                  `json:"skip_consul_lock,omitempty"`
	ETCDConfig
	encryption.EncryptionConfig
	debugserver.DebugServerConfig
	lagerflags.LagerConfig
}

func DefaultConfig() BBSConfig {
	return BBSConfig{
		SessionName:                 "bbs",
		CommunicationTimeout:        durationjson.Duration(10 * time.Second),
		RequireSSL:                  false,
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
		DatabaseDriver:              "mysql",
		MaxOpenDatabaseConnections:  200,
		MaxIdleDatabaseConnections:  200,
		AuctioneerRequireTLS:        false,
		RepClientSessionCacheSize:   0,
		RepRequireTLS:               false,
		ETCDConfig:                  DefaultETCDConfig(),
		EncryptionConfig:            encryption.DefaultEncryptionConfig(),
		LagerConfig:                 lagerflags.DefaultLagerConfig(),
	}
}

func NewBBSConfig(configPath string) (BBSConfig, error) {
	bbsConfig := DefaultConfig()
	configFile, err := os.Open(configPath)
	if err != nil {
		return BBSConfig{}, err
	}
	decoder := json.NewDecoder(configFile)

	err = decoder.Decode(&bbsConfig)
	if err != nil {
		return BBSConfig{}, err
	}

	return bbsConfig, nil
}
