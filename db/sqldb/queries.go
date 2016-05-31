package sqldb

const MySql = "mysql"
const Postgres = "postgres"

const (
	CreateUnclaimedActualLRPQuery = iota
	UnclaimActualLRPQuery
	ClaimActualLRPQuery
	StartActualLRPQuery
	CrashActualLRPQuery
	FailActualLRPQuery
	RemoveActualLRPQuery
	CreateRunningActualLRPQuery
	SelectActualLRPQuery

	SetConfigurationValueQuery
	GetConfigurationValueQuery

	DesireLRPQuery
	DesiredLRPsByDomainQuery
	DesiredLRPsQuery
	DesiredLRPSchedulingInfoByDomainQuery
	DesiredLRPSchedulingInfoQuery
	SelectDesiredLRPByGuidQuery
	DeleteDesiredLRPQuery

	DomainsQuery
	UpsertDomainQuery

	ReEncryptSelectQuery
	ReEncryptUpdateQuery
)

var postgresQueries = []string{
	// CreateUnclaimedActualLRPQuery
	`INSERT INTO actual_lrps
			(process_guid, instance_index, domain, state, since, net_info, modification_tag_epoch, modification_tag_index)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
	// UnclaimActualLRPQuery
	`UPDATE actual_lrps
				SET state = $1, instance_guid = $2, cell_id = $3,
					modification_tag_index = $4, since = $5, net_info = $6
				WHERE process_guid = $7 AND instance_index = $8 AND evacuating = $9`,
	// ClaimActualLRPQuery
	`UPDATE actual_lrps
				SET state = $1, instance_guid = $2, cell_id = $3, placement_error = $4,
					modification_tag_index = $5, net_info = $6, since = $7
				WHERE process_guid = $8 AND instance_index = $9 AND evacuating = $10`,
	// StartActualLRPQuery
	`UPDATE actual_lrps SET instance_guid = $1, cell_id = $2, net_info = $3,
					state = $4, since = $5, modification_tag_index = $6, placement_error = $7
					WHERE process_guid = $8 AND instance_index = $9 AND evacuating = $10`,
	// CrashActualLRPQuery
	`UPDATE actual_lrps
				SET state = $1, instance_guid = $2, cell_id = $3,
					modification_tag_index = $4, since = $5, net_info = $6,
					crash_count = $7, crash_reason = $8
				WHERE process_guid = $9 AND instance_index = $10 AND evacuating = $11`,
	// FailActualLRPQuery
	`UPDATE actual_lrps SET since = $1, modification_tag_index = $2, placement_error = $3
					WHERE process_guid = $4 AND instance_index = $5 AND evacuating = $6 `,
	// RemoveActualLRPQuery
	`DELETE FROM actual_lrps
					WHERE process_guid = $1 AND instance_index = $2 AND evacuating = $3`,
	// CreateRunningActualLRPQuery
	`INSERT INTO actual_lrps
					(process_guid, instance_index, domain, instance_guid, cell_id, state, net_info, since, modification_tag_epoch, modification_tag_index)
					VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
	// SelectActualLRPQuery
	`SELECT process_guid, instance_index, evacuating, domain, state,
				instance_guid, cell_id, placement_error, since, net_info,
				modification_tag_epoch, modification_tag_index, crash_count,
				crash_reason
				FROM actual_lrps `,
	// SetConfigurationValueQuery
	`INSERT INTO configurations (id, value) VALUES ($1, $2)
										ON CONFLICT (id) DO UPDATE SET value = $3`,

	// GetConfigurationValueQuery
	`SELECT value FROM configurations WHERE id = $1`,

	// DesireLRPQuery
	`INSERT INTO desired_lrps
				(process_guid, domain, log_guid, annotation, instances, memory_mb,
				disk_mb, rootfs, volume_placement, modification_tag_epoch, modification_tag_index,
				routes, run_info)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
	// DesiredLRPsByDomainQuery
	`SELECT process_guid, domain, log_guid, annotation, instances, memory_mb,
				disk_mb, rootfs, routes, modification_tag_epoch, modification_tag_index,
				run_info
			FROM desired_lrps
			WHERE domain = $1`,
	// DesiredLRPsQuery
	`SELECT process_guid, domain, log_guid, annotation, instances, memory_mb,
			disk_mb, rootfs, routes, modification_tag_epoch, modification_tag_index,
			run_info
		FROM desired_lrps`,
	// DesiredLRPSchedulingInfoByDomainQuery
	`SELECT process_guid, domain, log_guid, annotation, instances, memory_mb,
				disk_mb, rootfs, routes, volume_placement, modification_tag_epoch,
				modification_tag_index
			FROM desired_lrps
			WHERE domain = $1`,
	// DesiredLRPSchedulingInfoQuery
	`SELECT process_guid, domain, log_guid, annotation, instances, memory_mb,
				disk_mb, rootfs, routes, volume_placement, modification_tag_epoch,
				modification_tag_index
			FROM desired_lrps`,
	// SelectDesiredLRPByGuidQuery
	`SELECT process_guid, domain, log_guid, annotation, instances, memory_mb,
			disk_mb, rootfs, routes, modification_tag_epoch, modification_tag_index,
			run_info
		FROM desired_lrps
		WHERE process_guid = $1 `,
	// DeleteDesiredLRPQuery
	`DELETE FROM desired_lrps
				WHERE process_guid = $1
				`,

	// DomainsQuery
	`SELECT domain FROM domains WHERE expire_time > $1`,
	// UpsertDomainQuery
	`INSERT INTO domains (domain, expire_time) VALUES ($1, $2)
			ON CONFLICT (domain) DO UPDATE SET expire_time = $3`,

	// ReEncryptSelectQuery
	`SELECT %s FROM %s WHERE %s = $1 FOR UPDATE`,
	// ReEncryptUpdateQuery
	`UPDATE %s SET %s = $1 WHERE %s = $2`,
}

var mySqlQueries = []string{
	// CreateUnclaimedActualLRPQuery
	`INSERT INTO actual_lrps
			(process_guid, instance_index, domain, state, since, net_info, modification_tag_epoch, modification_tag_index)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
	// UnclaimActualLRPQuery
	`UPDATE actual_lrps
				SET state = ?, instance_guid = ?, cell_id = ?,
					modification_tag_index = ?, since = ?, net_info = ?
				WHERE process_guid = ? AND instance_index = ? AND evacuating = ?`,
	// ClaimActualLRPQuery
	`UPDATE actual_lrps
				SET state = ?, instance_guid = ?, cell_id = ?, placement_error = ?,
					modification_tag_index = ?, net_info = ?, since = ?
				WHERE process_guid = ? AND instance_index = ? AND evacuating = ?`,
	// StartActualLRPQuery
	`UPDATE actual_lrps SET instance_guid = ?, cell_id = ?, net_info = ?,
					state = ?, since = ?, modification_tag_index = ?, placement_error = ?
					WHERE process_guid = ? AND instance_index = ? AND evacuating = ?`,
	// CrashActualLRPQuery
	`UPDATE actual_lrps
				SET state = ?, instance_guid = ?, cell_id = ?,
					modification_tag_index = ?, since = ?, net_info = ?,
					crash_count = ?, crash_reason = ?
				WHERE process_guid = ? AND instance_index = ? AND evacuating = ?`,
	// FailActualLRPQuery
	`UPDATE actual_lrps SET since = ?, modification_tag_index = ?, placement_error = ?
					WHERE process_guid = ? AND instance_index = ? AND evacuating = ? `,
	// RemoveActualLRPQuery
	`DELETE FROM actual_lrps
					WHERE process_guid = ? AND instance_index = ? AND evacuating = ?`,
	// CreateRunningActualLRPQuery
	`INSERT INTO actual_lrps
					(process_guid, instance_index, domain, instance_guid, cell_id, state, net_info, since, modification_tag_epoch, modification_tag_index)
					VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
	// SelectActualLRPQuery
	`SELECT process_guid, instance_index, evacuating, domain, state,
				instance_guid, cell_id, placement_error, since, net_info,
				modification_tag_epoch, modification_tag_index, crash_count,
				crash_reason
				FROM actual_lrps `,
	// SetConfigurationValueQuery
	// GetConfigurationValueQuery
	// DesireLRPQuery
	// DesiredLRPsByDomainQuery
	// DesiredLRPsQuery
	// DesiredLRPSchedulingInfoByDomainQuery
	// DesiredLRPSchedulingInfoQuery
	// SelectDesiredLRPByGuidQuery
	// DeleteDesiredLRPQuery

	// DomainsQuery
	// UpsertDomainQuery

	// ReEncryptSelectQuery
	// ReEncryptUpdateQuery
}

func (db *SQLDB) getQuery(queryId int) string {
	if db.flavor == MySql {
		return mySqlQueries[queryId]
	} else {
		return postgresQueries[queryId]
	}
}
