package migrations

import (
	"database/sql"
	"encoding/json"
	"errors"
	"path"
	"time"

	"github.com/cloudfoundry-incubator/bbs/db/etcd"
	"github.com/cloudfoundry-incubator/bbs/encryption"
	"github.com/cloudfoundry-incubator/bbs/format"
	"github.com/cloudfoundry-incubator/bbs/migration"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/pivotal-golang/clock"
	"github.com/pivotal-golang/lager"
)

func init() {
	AppendMigration(NewETCDToSQL())
}

type ETCDToSQL struct {
	serializer  format.Serializer
	storeClient etcd.StoreClient
	clock       clock.Clock
	rawSQLDB    *sql.DB
}

type ETCDToSQLDesiredLRP struct {
	// DesiredLRPKey
	ProcessGuid string
	Domain      string
	LogGuid     string

	Annotation string
	Instances  int32

	// DesiredLRPResource
	RootFS   string
	DiskMB   int32
	MemoryMB int32

	// Routes
	Routes []byte

	// ModificationTag
	ModificationTagEpoch string
	ModificationTagIndex uint32

	// DesiredLRPRunInfo
	RunInfo         []byte
	VolumePlacement []byte
}

type ETCDToSQLActualLRP struct {
	// ActualLRPKey
	ProcessGuid string
	Index       int32
	Domain      string

	// ActualLRPInstanceKey
	InstanceGuid string
	CellId       string

	ActualLRPNetInfo []byte

	CrashCount     int32
	CrashReason    string
	State          string
	PlacementError string
	Since          int64

	// ModificationTag
	ModificationTagEpoch string
	ModificationTagIndex uint32
}

type ETCDToSQLTask struct {
	TaskGuid         string
	Domain           string
	CreatedAt        int64
	UpdatedAt        int64
	FirstCompletedAt int64
	State            int32
	CellId           string
	Result           string
	Failed           bool
	FailureReason    string
	TaskDefinition   []byte
}

func NewETCDToSQL() migration.Migration {
	return &ETCDToSQL{}
}

func (e *ETCDToSQL) String() string {
	return "1461790966"
}

func (e *ETCDToSQL) Version() int64 {
	return 1461790966
}

func (e *ETCDToSQL) SetStoreClient(storeClient etcd.StoreClient) {
	e.storeClient = storeClient
}

func (e *ETCDToSQL) SetCryptor(cryptor encryption.Cryptor) {
	e.serializer = format.NewSerializer(cryptor)
}

func (e *ETCDToSQL) SetClock(c clock.Clock) {
	e.clock = c
}

func truncateTables(db *sql.DB) error {
	tableNames := []string{
		"domains",
		"tasks",
		"desired_lrps",
		"actual_lrps",
	}
	for _, tableName := range tableNames {
		var value int
		// check whether the table exists before truncating
		err := db.QueryRow("SELECT 1 FROM ? LIMIT 1;", tableName).Scan(&value)
		if err == sql.ErrNoRows {
			continue
		}
		_, err = db.Exec("TRUNCATE TABLE " + tableName)
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *ETCDToSQL) Up(logger lager.Logger) error {
	truncateTables(e.rawSQLDB)
	var createTablesSQL = []string{
		createDomainSQL,
		createDesiredLRPsSQL,
		createActualLRPsSQL,
		createTasksSQL,
	}
	for _, query := range createTablesSQL {
		_, err := e.rawSQLDB.Exec(query)
		if err != nil {
			return err
		}
	}

	if e.storeClient == nil {
		return nil
	}

	response, err := e.storeClient.Get(etcd.DomainSchemaRoot, false, true)
	if err != nil {
		logger.Error("failed-fetching-domains", err)
	}

	if response != nil {
		for _, node := range response.Node.Nodes {
			domain := path.Base(node.Key)
			expireTime := e.clock.Now().UnixNano() + int64(time.Second)*node.TTL

			_, err := e.rawSQLDB.Exec(`
			INSERT INTO domains
			(domain, expire_time)
			VALUES (?, ?)
		`, domain, expireTime)
			if err != nil {
				logger.Error("failed-inserting-domain", err)
				continue
			}
		}
	}

	response, err = e.storeClient.Get(etcd.DesiredLRPSchedulingInfoSchemaRoot, false, true)
	if err != nil {
		logger.Error("failed-fetching-desired-lrp-scheduling-infos", err)
	}

	schedInfos := make(map[string]*models.DesiredLRPSchedulingInfo)

	if response != nil {
		for _, node := range response.Node.Nodes {
			model := new(models.DesiredLRPSchedulingInfo)
			err := e.serializer.Unmarshal(logger, []byte(node.Value), model)
			if err != nil {
				logger.Error("failed-to-deserialize-desired-lrp-scheduling-info", err)
				continue
			}
			schedInfos[path.Base(node.Key)] = model
		}
	}

	response, err = e.storeClient.Get(etcd.DesiredLRPRunInfoSchemaRoot, false, true)
	if err != nil {
		logger.Error("failed-fetching-desired-lrp-run-infos", err)
	}

	if response != nil {
		for _, node := range response.Node.Nodes {
			schedInfo := schedInfos[path.Base(node.Key)]
			routeData, err := json.Marshal(schedInfo.Routes)
			if err != nil {
				logger.Error("failed-to-marshal-routes", err)
				continue
			}

			volumePlacementData, err := e.serializer.Marshal(logger, format.ENCRYPTED_PROTO, schedInfo.VolumePlacement)
			if err != nil {
				logger.Error("failed-marshalling-volume-placements", err)
			}

			_, err = e.rawSQLDB.Exec(`
				INSERT INTO desired_lrps
					(process_guid, domain, log_guid, annotation, instances, memory_mb,
					disk_mb, rootfs, volume_placement, routes, modification_tag_epoch,
					modification_tag_index, run_info)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			`, schedInfo.ProcessGuid, schedInfo.Domain, schedInfo.LogGuid, schedInfo.Annotation,
				schedInfo.Instances, schedInfo.MemoryMb, schedInfo.DiskMb, schedInfo.RootFs, volumePlacementData,
				routeData, schedInfo.ModificationTag.Epoch, schedInfo.ModificationTag.Index, []byte(node.Value))
			if err != nil {
				logger.Error("failed-inserting-desired-lrp", err)
				continue
			}
		}
	}

	response, err = e.storeClient.Get(etcd.ActualLRPSchemaRoot, false, true)
	if err != nil {
		logger.Error("failed-fetching-actual-lrps", err)
	}

	if response != nil {
		for _, parent := range response.Node.Nodes {
			for _, indices := range parent.Nodes {
				for _, node := range indices.Nodes {
					if path.Base(node.Key) == "instance" {
						actualLRP := new(models.ActualLRP)
						err := e.serializer.Unmarshal(logger, []byte(node.Value), actualLRP)
						if err != nil {
							logger.Error("failed-to-deserialize-actual-lrp", err)
							continue
						}

						netInfoData, err := e.serializer.Marshal(logger, format.ENCRYPTED_PROTO, &actualLRP.ActualLRPNetInfo)
						if err != nil {
							logger.Error("failed-to-marshal-net-info", err)
						}

						_, err = e.rawSQLDB.Exec(`
							INSERT INTO actual_lrps
								(process_guid, instance_index, domain, instance_guid, cell_id,
								net_info, crash_count, crash_reason, state, placement_error, since,
								modification_tag_epoch, modification_tag_index)
							VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
						`, actualLRP.ProcessGuid, actualLRP.Index, actualLRP.Domain, actualLRP.InstanceGuid,
							actualLRP.CellId, netInfoData, actualLRP.CrashCount, actualLRP.CrashReason,
							actualLRP.State, actualLRP.PlacementError, actualLRP.Since,
							actualLRP.ModificationTag.Epoch, actualLRP.ModificationTag.Index)
						if err != nil {
							logger.Error("failed-inserting-actual-lrp", err)
							continue
						}
					}
				}
			}
		}
	}

	response, err = e.storeClient.Get(etcd.TaskSchemaRoot, false, true)
	if err != nil {
		logger.Error("failed-fetching-tasks", err)
	}

	if response != nil {
		for _, node := range response.Node.Nodes {
			task := new(models.Task)
			err := e.serializer.Unmarshal(logger, []byte(node.Value), task)
			if err != nil {
				logger.Error("failed-to-deserialize-task", err)
				continue
			}

			definitionData, err := e.serializer.Marshal(logger, format.ENCRYPTED_PROTO, task.TaskDefinition)

			_, err = e.rawSQLDB.Exec(`
							INSERT INTO tasks
								(guid, domain, updated_at, created_at, first_completed_at,
								state, cell_id, result, failed, failure_reason,
								task_definition)
							VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
						`,
				task.TaskGuid, task.Domain, task.UpdatedAt, task.CreatedAt,
				task.FirstCompletedAt, task.State, task.CellId, task.Result,
				task.Failed, task.FailureReason, definitionData)
			if err != nil {
				logger.Error("failed-inserting-task", err)
				continue
			}
		}
	}

	return nil
}

func (e *ETCDToSQL) Down(logger lager.Logger) error {
	return errors.New("not implemented")
}

func (e *ETCDToSQL) SetRawSQLDB(rawSQLDB *sql.DB) {
	e.rawSQLDB = rawSQLDB
}

func (e *ETCDToSQL) RequiresSQL() bool {
	return true
}

const createDomainSQL = `CREATE TABLE IF NOT EXISTS domains(
	domain VARCHAR(255) PRIMARY KEY,
	expire_time BIGINT DEFAULT 0,

	INDEX(expire_time)
);`

const createDesiredLRPsSQL = `CREATE TABLE IF NOT EXISTS desired_lrps(
	process_guid VARCHAR(255) PRIMARY KEY,
	domain VARCHAR(255) NOT NULL,
	log_guid VARCHAR(255) NOT NULL,
	annotation TEXT,
	instances INT NOT NULL,
	memory_mb INT NOT NULL,
	disk_mb INT NOT NULL,
	rootfs VARCHAR(255) NOT NULL,
	routes BLOB NOT NULL,
	volume_placement BLOB NOT NULL,
	modification_tag_epoch VARCHAR(255) NOT NULL,
	modification_tag_index INT,
	run_info BLOB NOT NULL,

	INDEX(domain)
);`

const createActualLRPsSQL = `CREATE TABLE IF NOT EXISTS actual_lrps(
	process_guid VARCHAR(255),
	instance_index INT,
	evacuating BOOL DEFAULT false,
	domain VARCHAR(255) NOT NULL,
	state VARCHAR(255) NOT NULL,
	instance_guid VARCHAR(255) NOT NULL DEFAULT "",
	cell_id VARCHAR(255) NOT NULL DEFAULT "",
	placement_error VARCHAR(255) NOT NULL DEFAULT "",
	since BIGINT DEFAULT 0,
	net_info BLOB NOT NULL,
	modification_tag_epoch VARCHAR(255) NOT NULL,
	modification_tag_index INT,
	crash_count INT NOT NULL DEFAULT 0,
	crash_reason VARCHAR(255) NOT NULL DEFAULT "",
	expire_time BIGINT DEFAULT 0,

	PRIMARY KEY(process_guid, instance_index, evacuating),
	INDEX(domain),
	INDEX(cell_id),
	INDEX(since),
	INDEX(state),
	INDEX(expire_time)
);`

const createTasksSQL = `CREATE TABLE IF NOT EXISTS tasks(
	guid VARCHAR(255) PRIMARY KEY,
	domain VARCHAR(255) NOT NULL,
	updated_at BIGINT DEFAULT 0,
	created_at BIGINT DEFAULT 0,
	first_completed_at BIGINT DEFAULT 0,
	state INT,
	cell_id VARCHAR(255) NOT NULL DEFAULT "",
	result TEXT,
	failed BOOL DEFAULT false,
	failure_reason VARCHAR(255) NOT NULL DEFAULT "",
	task_definition BLOB NOT NULL,

	INDEX(domain),
	INDEX(state),
	INDEX(cell_id),
	INDEX(updated_at),
	INDEX(created_at),
	INDEX(first_completed_at)
);`
