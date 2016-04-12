package sqldb

import "github.com/pivotal-golang/lager"

func (db *SQLDB) CreateInitialSchema(logger lager.Logger) error {
	var createTablesSQL = []string{
		createDomainSQL,
		createConfigurationsSQL,
		createTasksSQL,
		createDesiredLRPsSQL,
		createActualLRPsSQL,
	}

	for _, query := range createTablesSQL {
		_, err := db.db.Exec(query)
		if err != nil {
			return err
		}
	}

	return nil
}

const createDomainSQL = `CREATE TABLE IF NOT EXISTS domains(
	domain VARCHAR(255) PRIMARY KEY,
	expire_time BIGINT DEFAULT 0,

	INDEX(expire_time)
);`

const createConfigurationsSQL = `CREATE TABLE IF NOT EXISTS configurations(
	id VARCHAR(255) PRIMARY KEY,
	value VARCHAR(255)
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
