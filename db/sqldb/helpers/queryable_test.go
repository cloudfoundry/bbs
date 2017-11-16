package helpers_test

import (
	"database/sql"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/bbs/db/sqldb/helpers"
	"code.cloudfoundry.org/lager/lagertest"
)

var _ = Describe("SQL Helpers", func() {
	var (
		logger  *lagertest.TestLogger
		helper  helpers.SQLHelper
		monitor helpers.QueryMonitor
	)

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("query-metrics-test")
		helper = helpers.NewSQLHelper(dbFlavor)
		monitor = helpers.NewQueryMonitor()

		tableName = fmt.Sprintf("dummy_%d", GinkgoParallelNode())
		tableQuery := fmt.Sprintf("CREATE TABLE %s (field1 INT);", tableName)
		_, err := db.Exec(tableQuery)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		_, err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s;", tableName))
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Transactions", func() {
		It("returns a transaction and increments metrics", func() {
			q := helpers.NewMonitoredDB(db, monitor)

			tx, err := q.Begin()
			query := helper.Rebind(fmt.Sprintf("INSERT INTO %s (field1) VALUES (?);", tableName))
			res, err := tx.Exec(query, 3)
			Expect(err).NotTo(HaveOccurred())
			rows, err := res.RowsAffected()
			Expect(err).NotTo(HaveOccurred())
			Expect(rows).To(BeEquivalentTo(1))

			query = helper.Rebind(fmt.Sprintf("SELECT * FROM %s;", tableName))
			row := tx.QueryRow(query)

			var value int
			err = row.Scan(&value)
			Expect(err).NotTo(HaveOccurred())
			Expect(value).To(Equal(3))

			err = tx.Commit()
			Expect(err).NotTo(HaveOccurred())

			Expect(monitor.QueriesSucceeded()).To(BeEquivalentTo(4))
		})

		It("rollsback a transaction and increments metrics", func() {
			q := helpers.NewMonitoredDB(db, monitor)

			tx, err := q.Begin()
			query := helper.Rebind(fmt.Sprintf("INSERT INTO %s (field1) VALUES (?);", tableName))
			res, err := tx.Exec(query, 3)
			Expect(err).NotTo(HaveOccurred())
			rows, err := res.RowsAffected()
			Expect(err).NotTo(HaveOccurred())
			Expect(rows).To(BeEquivalentTo(1))

			query = helper.Rebind(fmt.Sprintf("SELECT * FROM %s;", tableName))
			row := tx.QueryRow(query)

			var value int
			err = row.Scan(&value)
			Expect(err).NotTo(HaveOccurred())
			Expect(value).To(Equal(3))

			err = tx.Rollback()
			Expect(err).NotTo(HaveOccurred())

			Expect(monitor.QueriesSucceeded()).To(BeEquivalentTo(4))

			query = helper.Rebind(fmt.Sprintf("SELECT count(*) FROM %s;", tableName))
			row = q.QueryRow(query)

			err = row.Scan(&value)
			Expect(err).NotTo(HaveOccurred())
			Expect(value).To(Equal(0))
		})
	})

	Describe("Begin", func() {
		It("returns a transaction and increments", func() {
			q := helpers.NewMonitoredDB(db, monitor)

			tx, err := q.Begin()
			Expect(err).NotTo(HaveOccurred())

			Expect(monitor.QueriesSucceeded()).To(BeEquivalentTo(1))
			tx.Commit()
		})
	})

	Describe("Exec", func() {
		It("executes queries", func() {
			q := helpers.NewMonitoredDB(db, monitor)

			query := helper.Rebind(fmt.Sprintf("INSERT INTO %s (field1) VALUES (?);", tableName))
			res, err := q.Exec(query, 3)
			Expect(err).NotTo(HaveOccurred())
			rows, err := res.RowsAffected()
			Expect(err).NotTo(HaveOccurred())
			Expect(rows).To(BeEquivalentTo(1))

			Expect(monitor.QueriesSucceeded()).To(BeEquivalentTo(1))
		})

		It("returns an error on a bad query", func() {
			q := helpers.NewMonitoredDB(db, monitor)

			query := helper.Rebind(fmt.Sprintf("INSERT INTO %s (field2) VALUES (?);", tableName))
			_, err := q.Exec(query, 3)

			Expect(err).To(HaveOccurred())
			Expect(monitor.QueriesFailed()).To(BeEquivalentTo(1))
		})
	})

	Describe("Query", func() {
		BeforeEach(func() {
			query := helper.Rebind(fmt.Sprintf("INSERT INTO %s (field1) VALUES (3);", tableName))
			_, err := db.Exec(query)
			Expect(err).NotTo(HaveOccurred())

			query = helper.Rebind(fmt.Sprintf("INSERT INTO %s (field1) VALUES (4);", tableName))
			_, err = db.Exec(query)
			Expect(err).NotTo(HaveOccurred())
		})

		It("executes queries", func() {
			q := helpers.NewMonitoredDB(db, monitor)

			query := helper.Rebind(fmt.Sprintf("SELECT * FROM %s;", tableName))
			rows, err := q.Query(query)
			defer rows.Close()
			Expect(err).NotTo(HaveOccurred())

			expectedValue := 3
			for rows.Next() {
				var value int
				err := rows.Scan(&value)
				Expect(err).NotTo(HaveOccurred())
				Expect(value).To(Equal(expectedValue))
				expectedValue++
			}

			Expect(monitor.QueriesSucceeded()).To(BeEquivalentTo(1))
		})

		It("returns an error on a bad query", func() {
			q := helpers.NewMonitoredDB(db, monitor)

			query := helper.Rebind(fmt.Sprintf("SELECT * FROM doesnotexist;"))
			_, err := q.Query(query)

			Expect(err).To(HaveOccurred())
			Expect(monitor.QueriesFailed()).To(BeEquivalentTo(1))
		})
	})

	Describe("QueryRow", func() {
		BeforeEach(func() {
			query := helper.Rebind(fmt.Sprintf("INSERT INTO %s (field1) VALUES (3);", tableName))
			_, err := db.Exec(query)
			Expect(err).NotTo(HaveOccurred())
		})

		It("executes a query", func() {
			q := helpers.NewMonitoredDB(db, monitor)

			query := helper.Rebind(fmt.Sprintf("SELECT * FROM %s;", tableName))
			row := q.QueryRow(query)

			var value int
			err := row.Scan(&value)
			Expect(err).NotTo(HaveOccurred())
			Expect(value).To(Equal(3))
			Expect(monitor.QueriesSucceeded()).To(BeEquivalentTo(1))
		})

		It("returns an error on a bad query", func() {
			q := helpers.NewMonitoredDB(db, monitor)

			query := helper.Rebind(fmt.Sprintf("SELECT * FROM doesnotexist;"))
			row := q.QueryRow(query)
			var value int
			err := row.Scan(&value)
			Expect(err).To(HaveOccurred())
			Expect(monitor.QueriesFailed()).To(BeEquivalentTo(1))
		})

		It("does not return an error if the row does not exist", func() {
			q := helpers.NewMonitoredDB(db, monitor)

			query := helper.Rebind(fmt.Sprintf("SELECT * FROM %s where field1 = 12345;", tableName))
			row := q.QueryRow(query)
			var value int
			err := row.Scan(&value)
			Expect(err).To(MatchError(sql.ErrNoRows))
			Expect(monitor.QueriesFailed()).To(BeZero())
		})
	})

	Describe("OpenConnections", func() {
		It("returns the number of open connections to the database", func() {
			q := helpers.NewMonitoredDB(db, monitor)
			Expect(q.OpenConnections()).To(BeNumerically(">=", 0))
		})
	})
})
