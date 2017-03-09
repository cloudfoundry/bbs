package sqldb_test

import (
	"database/sql"
	"strings"

	"code.cloudfoundry.org/bbs/db/sqldb"
	"code.cloudfoundry.org/bbs/db/sqldb/helpers"
	"code.cloudfoundry.org/bbs/format"
	"code.cloudfoundry.org/bbs/test_helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Isolation Level", func() {
	var (
		dbSession      *sql.DB
		sqlDBIsolation *sqldb.SQLDB
	)

	BeforeEach(func() {
		var err error
		// We need a different db session to prevent test pollution
		dbSession, err = sql.Open(dbDriverName, dbBaseConnectionString)
		Expect(err).NotTo(HaveOccurred())
		Expect(dbSession.Ping()).NotTo(HaveOccurred())

		sqlDBIsolation = sqldb.NewSQLDB(dbSession, 5, 5, format.ENCRYPTED_PROTO, cryptor, fakeGUIDProvider, fakeClock, dbFlavor)
	})

	It("sets the transaction isolation level", func() {
		levels := []string{
			helpers.IsolationLevelReadUncommitted,
			helpers.IsolationLevelReadCommitted,
			helpers.IsolationLevelSerializable,
			helpers.IsolationLevelRepeatableRead,
		}

		for _, level := range levels {
			err := sqlDBIsolation.SetIsolationLevel(logger, level)
			Expect(err).NotTo(HaveOccurred())

			var isolationLevel, isolationVariable string
			if test_helpers.UsePostgres() {
				expectedLevel := strings.ToLower(level)
				row := dbSession.QueryRow("SHOW TRANSACTION ISOLATION LEVEL")
				err := row.Scan(&isolationLevel)
				Expect(err).NotTo(HaveOccurred())
				Expect(isolationLevel).To(Equal(expectedLevel))
			} else if test_helpers.UseMSSQL() {
				row := dbSession.QueryRow(`SELECT CASE transaction_isolation_level
								WHEN 0 THEN 'Unspecified'
								WHEN 1 THEN 'READ UNCOMMITTED'
								WHEN 2 THEN 'READ COMMITTED'
								WHEN 3 THEN 'REPEATABLE READ'
								WHEN 4 THEN 'SERIALIZABLE'
								WHEN 5 THEN 'SNAPSHOT' END AS TRANSACTION_ISOLATION_LEVEL
							FROM sys.dm_exec_sessions
							where session_id = @@SPID`)
				err := row.Scan(&isolationLevel)
				Expect(err).NotTo(HaveOccurred())
				Expect(isolationLevel).To(Equal(level))
			} else {
				expectedLevel := strings.Replace(level, " ", "-", -1)
				row := dbSession.QueryRow("SHOW VARIABLES LIKE '%isolation%'")
				err := row.Scan(&isolationVariable, &isolationLevel)
				Expect(err).NotTo(HaveOccurred())
				Expect(isolationLevel).To(Equal(expectedLevel))
			}
		}
	})

	Context("when the isolation level is not valid", func() {
		It("returns an error", func() {
			err := sqlDBIsolation.SetIsolationLevel(logger, "Not a valid isolation level")
			Expect(err).To(HaveOccurred())
		})
	})
})
