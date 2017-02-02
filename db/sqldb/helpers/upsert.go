package helpers

import (
	"database/sql"
	"fmt"
	"strings"

	"code.cloudfoundry.org/lager"
)

func (h *sqlHelper) Upsert(
	logger lager.Logger,
	q Queryable,
	table string,
	keyAttributes,
	updateAttributes SQLAttributes,
) (sql.Result, error) {
	columns := make([]string, 0, len(keyAttributes)+len(updateAttributes))
	keyNames := make([]string, 0, len(keyAttributes))
	updateBindings := make([]string, 0, len(updateAttributes))
	bindingValues := make([]interface{}, 0, len(keyAttributes)+2*len(updateAttributes))

	keyBindingValues := make([]interface{}, 0, len(keyAttributes))
	nonKeyBindingValues := make([]interface{}, 0, len(updateAttributes))

	for column, value := range keyAttributes {
		columns = append(columns, column)
		keyNames = append(keyNames, column)
		keyBindingValues = append(keyBindingValues, value)
	}

	for column, value := range updateAttributes {
		columns = append(columns, column)
		updateBindings = append(updateBindings, fmt.Sprintf("%s = ?", column))
		nonKeyBindingValues = append(nonKeyBindingValues, value)
	}

	insertBindings := QuestionMarks(len(keyAttributes) + len(updateAttributes))

	var query string
	switch h.flavor {
	case Postgres:
		bindingValues = append(bindingValues, nonKeyBindingValues...)
		bindingValues = append(bindingValues, keyBindingValues...)
		bindingValues = append(bindingValues, keyBindingValues...)
		bindingValues = append(bindingValues, nonKeyBindingValues...)

		insert := fmt.Sprintf(`
				INSERT INTO %s
					(%s)
				SELECT %s`,
			table,
			strings.Join(columns, ", "),
			insertBindings)

		// TODO: Add where clause with key values.
		// Alternatively upgrade to postgres 9.5 :D
		whereClause := []string{}
		for _, key := range keyNames {
			whereClause = append(whereClause, fmt.Sprintf("%s = ?", key))
		}

		upsert := fmt.Sprintf(`
				UPDATE %s SET
					%s
				WHERE %s
				`,
			table,
			strings.Join(updateBindings, ", "),
			strings.Join(whereClause, " AND "),
		)

		query = fmt.Sprintf(`
				WITH upsert AS (%s RETURNING *)
				%s WHERE NOT EXISTS
				(SELECT * FROM upsert)
				`,
			upsert,
			insert)

		result, err := q.Exec(fmt.Sprintf("LOCK TABLE %s IN SHARE ROW EXCLUSIVE MODE", table))
		if err != nil {
			return result, err
		}

	case MySQL:
		bindingValues = append(bindingValues, keyBindingValues...)
		bindingValues = append(bindingValues, nonKeyBindingValues...)
		bindingValues = append(bindingValues, nonKeyBindingValues...)

		query = fmt.Sprintf(`
				INSERT INTO %s
					(%s)
				VALUES (%s)
				ON DUPLICATE KEY UPDATE
					%s
			`,
			table,
			strings.Join(columns, ", "),
			insertBindings,
			strings.Join(updateBindings, ", "),
		)
	default:
		// totally shouldn't happen
		panic("database flavor not implemented: " + h.flavor)
	}
	return q.Exec(h.Rebind(query), bindingValues...)
}
