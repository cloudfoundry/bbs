package sqldb

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/pivotal-golang/lager"
)

const (
	MySQL    = "mysql"
	Postgres = "postgres"
)

type RowLock bool

const (
	LockRow   RowLock = true
	NoLockRow RowLock = false
)

type SQLAttributes map[string]interface{}

type ColumnList []string

var (
	schedulingInfoColumns = ColumnList{
		"desired_lrps.process_guid",
		"desired_lrps.domain",
		"desired_lrps.log_guid",
		"desired_lrps.annotation",
		"desired_lrps.instances",
		"desired_lrps.memory_mb",
		"desired_lrps.disk_mb",
		"desired_lrps.rootfs",
		"desired_lrps.routes",
		"desired_lrps.volume_placement",
		"desired_lrps.modification_tag_epoch",
		"desired_lrps.modification_tag_index",
	}

	desiredLRPColumns = append(schedulingInfoColumns,
		"desired_lrps.run_info",
	)

	taskColumns = ColumnList{
		"tasks.guid",
		"tasks.domain",
		"tasks.updated_at",
		"tasks.created_at",
		"tasks.first_completed_at",
		"tasks.state",
		"tasks.cell_id",
		"tasks.result",
		"tasks.failed",
		"tasks.failure_reason",
		"tasks.task_definition",
	}

	actualLRPColumns = ColumnList{
		"actual_lrps.process_guid",
		"actual_lrps.instance_index",
		"actual_lrps.evacuating",
		"actual_lrps.domain",
		"actual_lrps.state",
		"actual_lrps.instance_guid",
		"actual_lrps.cell_id",
		"actual_lrps.placement_error",
		"actual_lrps.since",
		"actual_lrps.net_info",
		"actual_lrps.modification_tag_epoch",
		"actual_lrps.modification_tag_index",
		"actual_lrps.crash_count",
		"actual_lrps.crash_reason",
	}

	domainColumns = ColumnList{
		"domain",
	}
)

func (db *SQLDB) CreateConfigurationsTable(logger lager.Logger) error {
	_, err := db.db.Exec(`
		CREATE TABLE IF NOT EXISTS configurations(
			id VARCHAR(255) PRIMARY KEY,
			value VARCHAR(255)
		)
	`)
	if err != nil {
		return err
	}

	return nil
}

func (db *SQLDB) selectLRPInstanceCounts(logger lager.Logger, q Queryable) (*sql.Rows, error) {
	var query string
	columns := schedulingInfoColumns
	columns = append(columns, "COUNT(actual_lrps.instance_index) AS actual_instances")

	switch db.flavor {
	case Postgres:
		columns = append(columns, "STRING_AGG(actual_lrps.instance_index::text, ',') AS existing_indices")
	case MySQL:
	default:
		// totally shouldn't happen
		panic("database flavor not implemented: " + db.flavor)
	}

	query = fmt.Sprintf(`
		SELECT %s
			FROM desired_lrps
			LEFT OUTER JOIN actual_lrps ON desired_lrps.process_guid = actual_lrps.process_guid AND actual_lrps.evacuating = false
			GROUP BY desired_lrps.process_guid
			HAVING COUNT(actual_lrps.instance_index) <> desired_lrps.instances
		`,
		strings.Join(columns, ", "),
	)

	return q.Query(query)
}
func (db *SQLDB) selectOrphanedActualLRPs(logger lager.Logger, q Queryable) (*sql.Rows, error) {
	query := `
		SELECT actual_lrps.process_guid, actual_lrps.instance_index, actual_lrps.domain
			FROM actual_lrps
			JOIN domains ON actual_lrps.domain = domains.domain
			WHERE actual_lrps.evacuating = false
			AND actual_lrps.process_guid NOT IN (SELECT process_guid FROM desired_lrps)
		`

	return q.Query(query)
}

func (db *SQLDB) selectLRPsWithMissingCells(logger lager.Logger, q Queryable, cellSet models.CellSet) (*sql.Rows, error) {
	wheres := []string{"actual_lrps.evacuating = false"}
	bindings := make([]interface{}, 0, len(cellSet))

	if len(cellSet) > 0 {
		wheres = append(wheres, fmt.Sprintf("actual_lrps.cell_id NOT IN (%s)", questionMarks(len(cellSet))))
		wheres = append(wheres, "actual_lrps.cell_id <> ''")
		for cellID := range cellSet {
			bindings = append(bindings, cellID)
		}
	}

	query := fmt.Sprintf(`
		SELECT %s
			FROM desired_lrps
			JOIN actual_lrps ON desired_lrps.process_guid = actual_lrps.process_guid
			WHERE %s
		`,
		strings.Join(append(schedulingInfoColumns, "actual_lrps.instance_index"), ", "),
		strings.Join(wheres, " AND "),
	)

	return q.Query(db.rebind(query), bindings...)
}

func (db *SQLDB) selectCrashedLRPs(logger lager.Logger, q Queryable) (*sql.Rows, error) {
	query := fmt.Sprintf(`
		SELECT %s
			FROM desired_lrps
			JOIN actual_lrps ON desired_lrps.process_guid = actual_lrps.process_guid
			WHERE actual_lrps.state = ? AND actual_lrps.evacuating = ?
		`,
		strings.Join(
			append(schedulingInfoColumns, "actual_lrps.instance_index", "actual_lrps.since", "actual_lrps.crash_count"),
			", ",
		),
	)

	return q.Query(db.rebind(query), models.ActualLRPStateCrashed, false)
}

func (db *SQLDB) selectStaleUnclaimedLRPs(logger lager.Logger, q Queryable, now time.Time) (*sql.Rows, error) {
	query := fmt.Sprintf(`
		SELECT %s
			FROM desired_lrps
			JOIN actual_lrps ON desired_lrps.process_guid = actual_lrps.process_guid
			WHERE actual_lrps.state = ? AND actual_lrps.since < ? AND actual_lrps.evacuating = ?
		`,
		strings.Join(append(schedulingInfoColumns, "actual_lrps.instance_index"), ", "),
	)

	return q.Query(db.rebind(query),
		models.ActualLRPStateUnclaimed,
		now.Add(-models.StaleUnclaimedActualLRPDuration).UnixNano(),
		false,
	)
}

func (db *SQLDB) countDesiredInstances(logger lager.Logger, q Queryable) int {
	query := `
		SELECT COALESCE(SUM(desired_lrps.instances), 0) AS desired_instances
			FROM desired_lrps
	`

	var desiredInstances int
	row := q.QueryRow(db.rebind(query))
	err := row.Scan(&desiredInstances)
	if err != nil {
		logger.Error("failed-desired-instances-query", err)
	}
	return desiredInstances
}

func (db *SQLDB) countActualLRPsByState(logger lager.Logger, q Queryable) (claimedCount, unclaimedCount, runningCount, crashedCount, crashingDesiredCount int) {
	var query string
	switch db.flavor {
	case Postgres:
		query = `
			SELECT
				COUNT(*) FILTER (WHERE actual_lrps.state = $1) AS claimed_instances,
				COUNT(*) FILTER (WHERE actual_lrps.state = $2) AS unclaimed_instances,
				COUNT(*) FILTER (WHERE actual_lrps.state = $3) AS running_instances,
				COUNT(*) FILTER (WHERE actual_lrps.state = $4) AS crashed_instances,
				COUNT(DISTINCT process_guid) FILTER (WHERE actual_lrps.state = $5) AS crashing_desireds
			FROM actual_lrps
			WHERE evacuating = $6
		`
	case MySQL:
	default:
		// totally shouldn't happen
		panic("database flavor not implemented: " + db.flavor)
	}

	row := db.db.QueryRow(query, models.ActualLRPStateClaimed, models.ActualLRPStateUnclaimed, models.ActualLRPStateRunning, models.ActualLRPStateCrashed, models.ActualLRPStateCrashed, false)
	err := row.Scan(&claimedCount, &unclaimedCount, &runningCount, &crashedCount, &crashingDesiredCount)
	if err != nil {
		logger.Error("failed-counting-actual-lrps", err)
	}
	return
}

func (db *SQLDB) countTasksByState(logger lager.Logger, q Queryable) (pendingCount, runningCount, completedCount, resolvingCount int) {
	var query string
	switch db.flavor {
	case Postgres:
		query = `
			SELECT
				COUNT(*) FILTER (WHERE state = $1) AS pending_tasks,
				COUNT(*) FILTER (WHERE state = $2) AS running_tasks,
				COUNT(*) FILTER (WHERE state = $3) AS completed_tasks,
				COUNT(*) FILTER (WHERE state = $4) AS resolving_tasks
			FROM tasks
		`
	case MySQL:
	default:
		// totally shouldn't happen
		panic("database flavor not implemented: " + db.flavor)
	}

	row := db.db.QueryRow(query, models.Task_Pending, models.Task_Running, models.Task_Completed, models.Task_Resolving)
	err := row.Scan(&pendingCount, &runningCount, &completedCount, &resolvingCount)
	if err != nil {
		logger.Error("failed-counting-tasks", err)
	}
	return
}

// SELECT <columns> FROM <table> WHERE ... LIMIT 1 [FOR UPDATE]
func (db *SQLDB) one(logger lager.Logger, q Queryable, table string,
	columns ColumnList, lockRow RowLock,
	wheres string, whereBindings ...interface{},
) *sql.Row {
	query := fmt.Sprintf("SELECT %s FROM %s\n", strings.Join(columns, ", "), table)

	if len(wheres) > 0 {
		query += "WHERE " + wheres
	}

	query += "\nLIMIT 1"

	if lockRow {
		query += "\nFOR UPDATE"
	}

	return q.QueryRow(db.rebind(query), whereBindings...)
}

// SELECT <columns> FROM <table> WHERE ... [FOR UPDATE]
func (db *SQLDB) all(logger lager.Logger, q Queryable, table string,
	columns ColumnList, lockRow RowLock,
	wheres string, whereBindings ...interface{},
) (*sql.Rows, error) {
	query := fmt.Sprintf("SELECT %s FROM %s\n", strings.Join(columns, ", "), table)

	if len(wheres) > 0 {
		query += "WHERE " + wheres
	}

	if lockRow {
		query += "\nFOR UPDATE"
	}

	return q.Query(db.rebind(query), whereBindings...)
}

func (db *SQLDB) upsert(logger lager.Logger, q Queryable, table string, keyAttributes, updateAttributes SQLAttributes) (sql.Result, error) {
	columns := make([]string, 0, len(keyAttributes)+len(updateAttributes))
	keyNames := make([]string, 0, len(keyAttributes))
	updateBindings := make([]string, 0, len(updateAttributes))
	bindingValues := make([]interface{}, 0, len(keyAttributes)+2*len(updateAttributes))

	for column, value := range keyAttributes {
		columns = append(columns, column)
		keyNames = append(keyNames, column)
		bindingValues = append(bindingValues, value)
	}

	for column, value := range updateAttributes {
		columns = append(columns, column)
		updateBindings = append(updateBindings, fmt.Sprintf("%s = ?", column))
		bindingValues = append(bindingValues, value)
	}

	// We need to copy the update column bindings so they each appear a second time in the binding list.
	bindingValues = append(bindingValues, bindingValues[len(keyAttributes):len(bindingValues)]...)

	insertBindings := questionMarks(len(keyAttributes) + len(updateAttributes))

	var query string
	switch db.flavor {
	case Postgres:
		query = fmt.Sprintf(`
				INSERT INTO %s
					(%s)
				VALUES (%s)
				ON CONFLICT (%s)
				DO UPDATE SET
					%s
			`,
			table,
			strings.Join(columns, ", "),
			insertBindings,
			strings.Join(keyNames, ", "),
			strings.Join(updateBindings, ", "),
		)
	case MySQL:
	default:
		// totally shouldn't happen
		panic("database flavor not implemented: " + db.flavor)
	}

	return q.Exec(db.rebind(query), bindingValues...)
}

// INSERT INTO <table> (...) VALUES ...
func (db *SQLDB) insert(logger lager.Logger, q Queryable, table string, attributes SQLAttributes) (sql.Result, error) {
	attributeCount := len(attributes)
	if attributeCount == 0 {
		return nil, nil
	}

	query := fmt.Sprintf("INSERT INTO %s\n", table)
	attributeNames := make([]string, 0, attributeCount)
	attributeBindings := make([]string, 0, attributeCount)
	bindings := make([]interface{}, 0, attributeCount)

	for column, value := range attributes {
		attributeNames = append(attributeNames, column)
		attributeBindings = append(attributeBindings, "?")
		bindings = append(bindings, value)
	}
	query += fmt.Sprintf("(%s)", strings.Join(attributeNames, ", "))
	query += fmt.Sprintf("VALUES (%s)", strings.Join(attributeBindings, ", "))

	return q.Exec(db.rebind(query), bindings...)
}

// UPDATE <table> SET ... WHERE ...
func (db *SQLDB) update(logger lager.Logger, q Queryable, table string, updates SQLAttributes, wheres string, whereBindings ...interface{}) (sql.Result, error) {
	updateCount := len(updates)
	if updateCount == 0 {
		return nil, nil
	}

	query := fmt.Sprintf("UPDATE %s SET\n", table)
	updateQueries := make([]string, 0, updateCount)
	bindings := make([]interface{}, 0, updateCount+len(whereBindings))

	for column, value := range updates {
		updateQueries = append(updateQueries, fmt.Sprintf("%s = ?", column))
		bindings = append(bindings, value)
	}
	query += strings.Join(updateQueries, ", ") + "\n"
	if len(wheres) > 0 {
		query += "WHERE " + wheres
		bindings = append(bindings, whereBindings...)
	}

	return q.Exec(db.rebind(query), bindings...)
}

// DELETE FROM <table> WHERE ...
func (db *SQLDB) delete(logger lager.Logger, q Queryable, table string, wheres string, whereBindings ...interface{}) (sql.Result, error) {
	query := fmt.Sprintf("DELETE FROM %s\n", table)

	if len(wheres) > 0 {
		query += "WHERE " + wheres
	}

	return q.Exec(db.rebind(query), whereBindings...)
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

func questionMarks(count int) string {
	if count == 0 {
		return ""
	}
	return strings.Repeat("?, ", count-1) + "?"
}
