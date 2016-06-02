package sqldb

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/pivotal-golang/lager"
)

const MySQL = "mysql"
const Postgres = "postgres"

type SQLUpdate map[string]interface{}

const (
	CreateConfigurationsTableQuery = "CreateConfigurationsTableQuery"

	InsertActualLRPQuery        = "InsertActualLRPQuery"
	UnclaimActualLRPQuery       = "UnclaimActualLRPQuery"
	ClaimActualLRPQuery         = "ClaimActualLRPQuery"
	StartActualLRPQuery         = "StartActualLRPQuery"
	CrashActualLRPQuery         = "CrashActualLRPQuery"
	FailActualLRPQuery          = "FailActualLRPQuery"
	DeleteActualLRPQuery        = "DeleteActualLRPQuery"
	CreateRunningActualLRPQuery = "CreateRunningActualLRPQuery"
	SelectActualLRPQuery        = "SelectActualLRPQuery"
	UpsertActualLRPQuery        = "UpsertActualLRPQuery"

	SetConfigurationValueQuery = "SetConfigurationValueQuery"
	GetConfigurationValueQuery = "GetConfigurationValueQuery"

	DesireLRPQuery                = "DesireLRPQuery"
	DesiredLRPsQuery              = "DesiredLRPsQuery"
	DesiredLRPSchedulingInfoQuery = "DesiredLRPSchedulingInfoQuery"
	SelectDesiredLRPByGuidQuery   = "SelectDesiredLRPByGuidQuery"
	LockDesiredLRPByGuidQuery     = "LockDesiredLRPByGuidQuery"
	UpdateDesiredLRPQuery         = "UpdateDesiredLRPQuery"
	DeleteDesiredLRPQuery         = "DeleteDesiredLRPQuery"
	CountDesiredInstancesQuery    = "CountDesiredInstancesQuery"

	DomainsQuery      = "DomainsQuery"
	UpsertDomainQuery = "UpsertDomainQuery"

	ReEncryptSelectQuery = "ReEncryptSelectQuery"
	ReEncryptUpdateQuery = "ReEncryptUpdateQuery"

	EvacuateActualLRPQuery = "EvacuateActualLRPQuery"

	SelectStaleLRPsQuery          = "SelectStaleLRPsQuery"
	SelectLRPsByStateQuery        = "SelectLRPsByStateQuery"
	SelectOrphanedActualLRPsQuery = "SelectOrphanedActualLRPsQuery"
	SelectLRPsQuery               = "SelectLRPsQuery"
	SelectLRPInstanceCountsQuery  = "SelectLRPInstanceCountsQuery"

	PruneDomainsQuery     = "PruneDomainsQuery"
	PruneActualLRPsQuery  = "PruneActualLRPsQuery"
	SelectLRPMetricsQuery = "SelectLRPMetricsQuery"

	SelectTasksBaseQuery                = "SelectTasksBaseQuery"
	SelectTaskByGuidQuery               = "SelectTaskByGuidQuery"
	UpdateTaskByGuidQuery               = "UpdateTaskByGuidQuery"
	UpdateTasksQuery                    = "UpdateTasksQuery"
	SelectTasksQuery                    = "SelectTasksQuery"
	UpdateTasksByStateQuery             = "UpdateTasksByStateQuery"
	UpdateTaskStateQuery                = "UpdateTaskStateQuery"
	DeleteTasksQuery                    = "DeleteTasksQuery"
	DeleteTaskQuery                     = "DeleteTaskQuery"
	SelectTasksByStateAndUpdatedAtQuery = "SelectTasksByStateAndUpdatedAtQuery"
	CountTasksQuery                     = "CountTasksQuery"
	InsertTaskQuery                     = "InsertTaskQuery"
	CompleteTaskByGuidQuery             = "CompleteTaskByGuidQuery"
)

var schedulingInfoColumns = `
	desired_lrps.process_guid,
	desired_lrps.domain,
	desired_lrps.log_guid,
	desired_lrps.annotation,
	desired_lrps.instances,
	desired_lrps.memory_mb,
	desired_lrps.disk_mb,
	desired_lrps.rootfs,
	desired_lrps.routes,
	desired_lrps.volume_placement,
	desired_lrps.modification_tag_epoch,
	desired_lrps.modification_tag_index`

const taskColumns = `tasks.guid, tasks.domain, tasks.updated_at, tasks.created_at,
	tasks.first_completed_at, tasks.state, tasks.cell_id, tasks.result,
	tasks.failed, tasks.failure_reason, tasks.task_definition`

var postgresQueries = map[string]string{
	CreateConfigurationsTableQuery: `
		CREATE TABLE IF NOT EXISTS configurations(
			id VARCHAR(255) PRIMARY KEY,
			value VARCHAR(255)
		)`,

	InsertActualLRPQuery: `
		INSERT INTO actual_lrps
				(process_guid, instance_index, domain, state, since, net_info, modification_tag_epoch, modification_tag_index)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
	UnclaimActualLRPQuery: `
		UPDATE actual_lrps
					SET state = $1, instance_guid = $2, cell_id = $3,
						modification_tag_index = $4, since = $5, net_info = $6
					WHERE process_guid = $7 AND instance_index = $8 AND evacuating = $9`,
	ClaimActualLRPQuery: `
		UPDATE actual_lrps
					SET state = $1, instance_guid = $2, cell_id = $3, placement_error = $4,
						modification_tag_index = $5, net_info = $6, since = $7
					WHERE process_guid = $8 AND instance_index = $9 AND evacuating = $10`,
	StartActualLRPQuery: `
		UPDATE actual_lrps SET instance_guid = $1, cell_id = $2, net_info = $3,
						state = $4, since = $5, modification_tag_index = $6, placement_error = $7
						WHERE process_guid = $8 AND instance_index = $9 AND evacuating = $10`,
	CrashActualLRPQuery: `
		UPDATE actual_lrps
					SET state = $1, instance_guid = $2, cell_id = $3,
						modification_tag_index = $4, since = $5, net_info = $6,
						crash_count = $7, crash_reason = $8
					WHERE process_guid = $9 AND instance_index = $10 AND evacuating = $11`,
	FailActualLRPQuery: `
		UPDATE actual_lrps SET since = $1, modification_tag_index = $2, placement_error = $3
						WHERE process_guid = $4 AND instance_index = $5 AND evacuating = $6 `,
	DeleteActualLRPQuery: `
		DELETE FROM actual_lrps
						WHERE process_guid = $1 AND instance_index = $2 AND evacuating = $3`,
	CreateRunningActualLRPQuery: `
		INSERT INTO actual_lrps
						(process_guid, instance_index, domain, instance_guid, cell_id, state, net_info, since, modification_tag_epoch, modification_tag_index)
						VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
	SelectActualLRPQuery: `
		SELECT process_guid, instance_index, evacuating, domain, state,
					instance_guid, cell_id, placement_error, since, net_info,
					modification_tag_epoch, modification_tag_index, crash_count,
					crash_reason
					FROM actual_lrps `,
	UpsertActualLRPQuery: `
		INSERT INTO actual_lrps
					(process_guid, instance_index, domain, instance_guid, cell_id, state,
					net_info, since, modification_tag_epoch, modification_tag_index,
					evacuating)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
				ON CONFLICT (process_guid, instance_index, evacuating)
				DO UPDATE SET
					expire_time = $12, domain = EXCLUDED.domain,
					instance_guid = EXCLUDED.instance_guid, cell_id = EXCLUDED.cell_id,
					state = EXCLUDED.state, net_info = EXCLUDED.net_info, since = EXCLUDED.since,
					modification_tag_epoch = EXCLUDED.modification_tag_epoch,
					modification_tag_index = EXCLUDED.modification_tag_index `,

	SetConfigurationValueQuery: `
		INSERT INTO configurations (id, value) VALUES ($1, $2)
											ON CONFLICT (id) DO UPDATE SET value = $3`,
	GetConfigurationValueQuery: `
		SELECT value FROM configurations WHERE id = $1`,

	DesireLRPQuery: `
		INSERT INTO desired_lrps
					(process_guid, domain, log_guid, annotation, instances, memory_mb,
					disk_mb, rootfs, volume_placement, modification_tag_epoch, modification_tag_index,
					routes, run_info)
					VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
	DesiredLRPsQuery: `
		SELECT process_guid, domain, log_guid, annotation, instances, memory_mb,
				disk_mb, rootfs, routes, modification_tag_epoch, modification_tag_index,
				run_info
			FROM desired_lrps`,
	DesiredLRPSchedulingInfoQuery: `
		SELECT process_guid, domain, log_guid, annotation, instances, memory_mb,
					disk_mb, rootfs, routes, volume_placement, modification_tag_epoch,
					modification_tag_index
				FROM desired_lrps`,
	SelectDesiredLRPByGuidQuery: `
		SELECT process_guid, domain, log_guid, annotation, instances, memory_mb,
				disk_mb, rootfs, routes, modification_tag_epoch, modification_tag_index,
				run_info
			FROM desired_lrps
			WHERE process_guid = $1 `,
	UpdateDesiredLRPQuery: `
		UPDATE desired_lrps
			SET modification_tag_index = $1, annotation = $2, instances = $3, routes = $4
			WHERE process_guid = $5`,
	DeleteDesiredLRPQuery: `
		DELETE FROM desired_lrps
					WHERE process_guid = $1`,
	CountDesiredInstancesQuery: `
		SELECT COALESCE(SUM(desired_lrps.instances), 0) AS desired_instances
		FROM desired_lrps`,
	LockDesiredLRPByGuidQuery: `
		SELECT 1 FROM desired_lrps WHERE process_guid = $1 FOR UPDATE`,

	DomainsQuery: `
		SELECT domain FROM domains WHERE expire_time > $1`,
	UpsertDomainQuery: `
		INSERT INTO domains (domain, expire_time) VALUES ($1, $2)
				ON CONFLICT (domain) DO UPDATE SET expire_time = $3`,

	ReEncryptSelectQuery: `
		SELECT %s FROM %s WHERE %s = $1 FOR UPDATE`,
	ReEncryptUpdateQuery: `
		UPDATE %s SET %s = $1 WHERE %s = $2`,

	EvacuateActualLRPQuery: `
		UPDATE actual_lrps SET domain = $1, instance_guid = $2, cell_id = $3, net_info = $4,
				state = $5, since = $6, modification_tag_index = $7
				WHERE process_guid = $8 AND instance_index = $9 AND evacuating = $10`,

	SelectStaleLRPsQuery: `
		SELECT ` + schedulingInfoColumns + `, actual_lrps.instance_index
			FROM desired_lrps
			JOIN actual_lrps ON desired_lrps.process_guid = actual_lrps.process_guid
			WHERE actual_lrps.state = $1 AND actual_lrps.since < $2 AND actual_lrps.evacuating = $3
	`,
	SelectLRPsByStateQuery: `
		SELECT ` + schedulingInfoColumns + `, actual_lrps.instance_index, actual_lrps.since, actual_lrps.crash_count
			FROM desired_lrps
			JOIN actual_lrps ON desired_lrps.process_guid = actual_lrps.process_guid
			WHERE actual_lrps.evacuating = $1
			AND actual_lrps.state = $2
	`,
	SelectOrphanedActualLRPsQuery: `
		SELECT actual_lrps.process_guid, actual_lrps.instance_index, actual_lrps.domain
			FROM actual_lrps
			JOIN domains ON actual_lrps.domain = domains.domain
			WHERE actual_lrps.evacuating = $1
			AND actual_lrps.process_guid NOT IN (SELECT process_guid FROM desired_lrps)
	`,
	SelectLRPsQuery: `
		SELECT ` + schedulingInfoColumns + `, actual_lrps.instance_index
			FROM desired_lrps
			JOIN actual_lrps ON desired_lrps.process_guid = actual_lrps.process_guid
			WHERE actual_lrps.evacuating = $1`,

	SelectLRPInstanceCountsQuery: `
		SELECT ` + schedulingInfoColumns + `,
			COUNT(actual_lrps.instance_index) AS actual_instances,
			STRING_AGG(actual_lrps.instance_index::text, ',') AS existing_indices
		FROM desired_lrps
		LEFT OUTER JOIN actual_lrps ON desired_lrps.process_guid = actual_lrps.process_guid AND actual_lrps.evacuating = $1
		GROUP BY desired_lrps.process_guid
		HAVING COUNT(actual_lrps.instance_index) <> desired_lrps.instances
	`,

	PruneDomainsQuery: `
		DELETE FROM domains
		WHERE expire_time <= $1
	`,
	PruneActualLRPsQuery: `
		DELETE FROM actual_lrps
		WHERE evacuating = $1 AND expire_time <= $2
	`,
	SelectLRPMetricsQuery: `
		SELECT
		  COUNT(*) FILTER (WHERE actual_lrps.state = $1) AS claimed_instances,
		  COUNT(*) FILTER (WHERE actual_lrps.state = $2) AS unclaimed_instances,
		  COUNT(*) FILTER (WHERE actual_lrps.state = $3) AS running_instances,
		  COUNT(*) FILTER (WHERE actual_lrps.state = $4) AS crashed_instances,
			COUNT(DISTINCT process_guid) FILTER (WHERE actual_lrps.state = $5) AS crashing_desireds
		FROM actual_lrps
		WHERE evacuating = $6
	`,

	SelectTasksBaseQuery: `
		SELECT ` + taskColumns + ` FROM tasks`,
	SelectTaskByGuidQuery: `
		SELECT ` + taskColumns + ` FROM tasks WHERE guid = $1`,
	SelectTasksQuery: `
		SELECT ` + taskColumns + ` FROM tasks
		  WHERE state = $1 AND updated_at < $2 AND created_at > $3`,
	DeleteTasksQuery: `
		DELETE FROM tasks
		  WHERE state = $1 AND first_completed_at < $2 `,
	DeleteTaskQuery: `
		DELETE FROM tasks WHERE guid = $1`,
	SelectTasksByStateAndUpdatedAtQuery: `
		SELECT ` + taskColumns + ` FROM tasks
		  WHERE state = $1 AND updated_at < $2`,
	CountTasksQuery: `
		SELECT
		  COUNT(*) FILTER (WHERE state = $1) AS pending_tasks,
		  COUNT(*) FILTER (WHERE state = $2) AS running_tasks,
		  COUNT(*) FILTER (WHERE state = $3) AS completed_tasks,
		  COUNT(*) FILTER (WHERE state = $4) AS resolving_tasks
		FROM tasks`,
	InsertTaskQuery: `
		INSERT INTO tasks (guid, domain, created_at, updated_at, first_completed_at, state, task_definition)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
	UpdateTaskByGuidQuery: `
		UPDATE tasks
			SET state = $1, updated_at = $2, cell_id = $3
			WHERE guid = $4`,
	UpdateTasksQuery: `
		UPDATE tasks
		  SET failed = $1, failure_reason = $2, result = $3, state = $4, first_completed_at = $5, updated_at = $6
		  WHERE state = $7 AND created_at < $8 `,
	CompleteTaskByGuidQuery: `
		UPDATE tasks
			SET state = $1, updated_at = $2, first_completed_at = $3,
			failed = $4, failure_reason = $5, result = $6, cell_id = $7
			WHERE guid = $8`,
}

var mySQLQueries = map[string]string{}

func (db *SQLDB) updateQuery(logger lager.Logger, q Queryable, table string, updates map[string]interface{}, suffix string, values ...interface{}) (sql.Result, error) {
	updateCount := len(updates)
	if updateCount == 0 {
		return nil, nil
	}

	query := fmt.Sprintf("UPDATE %s SET\n", table)
	updateQueries := make([]string, 0, updateCount)
	updateValues := make([]interface{}, 0, updateCount)

	for column, value := range updates {
		updateQueries = append(updateQueries, fmt.Sprintf("%s = ?", column))
		updateValues = append(updateValues, value)
	}
	query += strings.Join(updateQueries, ", ") + "\n"
	query += suffix

	result, err := q.Exec(db.rebind(query), append(updateValues, values...)...)
	if err != nil {
		logger.Error("failed-query", err)
	}

	return result, err
}

func (db *SQLDB) getQuery(queryID string) string {
	switch db.flavor {
	case MySQL:
		return mySQLQueries[queryID]
	case Postgres:
		return postgresQueries[queryID]
	default:
		// totally shouldn't happen
		panic("database flavor not implemented: " + db.flavor)
	}
}

func (db *SQLDB) rebind(query string) string {
	if db.flavor == MySQL {
		return query
	}

	strParts := strings.Split(query, "?")
	for i := 1; i < len(strParts); i++ {
		strParts[i-1] = fmt.Sprintf("%s$%d", strParts[i-1], i)
	}
	return strings.Join(strParts, "")
}
