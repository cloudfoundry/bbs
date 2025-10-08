package sqldb_test

import (
	"database/sql"
	"encoding/json"
	"time"

	"code.cloudfoundry.org/bbs/db/sqldb"
	"code.cloudfoundry.org/bbs/db/sqldb/helpers"
	"code.cloudfoundry.org/bbs/db/sqldb/helpers/monitor"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/bbs/test_helpers"
	"code.cloudfoundry.org/lager/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Version", func() {
	var helper helpers.SQLHelper

	BeforeEach(func() {
		helper = helpers.NewSQLHelper(dbFlavor)
	})

	Describe("SetVersion", func() {
		Context("when the version is not set", func() {
			It("sets the version into the database", func() {
				expectedVersion := &models.Version{CurrentVersion: 99}
				err := helper.Transact(ctx, logger, db, func(l lager.Logger, tx helpers.Tx) error {
					return sqlDB.SetVersion(tx, ctx, logger, expectedVersion)
				})
				Expect(err).NotTo(HaveOccurred())

				queryStr := "SELECT value FROM configurations WHERE id = ?"
				if test_helpers.UsePostgres() {
					queryStr = test_helpers.ReplaceQuestionMarks(queryStr)
				}
				rows, err := db.QueryContext(ctx, queryStr, sqldb.VersionID)
				Expect(err).NotTo(HaveOccurred())
				defer rows.Close()

				Expect(rows.Next()).To(BeTrue())

				var versionData string
				err = rows.Scan(&versionData)
				Expect(err).NotTo(HaveOccurred())

				var actualVersion models.Version
				err = json.Unmarshal([]byte(versionData), &actualVersion)
				Expect(err).NotTo(HaveOccurred())

				Expect(actualVersion).To(Equal(*expectedVersion))
			})
		})

		Context("when a version is already set", func() {
			var existingVersion *models.Version
			BeforeEach(func() {
				existingVersion = &models.Version{CurrentVersion: 99}
				versionJSON, err := json.Marshal(existingVersion)
				Expect(err).NotTo(HaveOccurred())

				queryStr := "UPDATE configurations SET value = ? WHERE id = ?"
				_, err = db.ExecContext(ctx, helpers.RebindForFlavor(queryStr, dbDriverName), versionJSON, sqldb.VersionID)
				Expect(err).NotTo(HaveOccurred())
			})

			It("updates the version in the db", func() {
				version := &models.Version{CurrentVersion: 20}

				err := helper.Transact(ctx, logger, db, func(l lager.Logger, tx helpers.Tx) error {
					return sqlDB.SetVersion(tx, ctx, logger, version)
				})
				Expect(err).NotTo(HaveOccurred())

				queryStr := "SELECT value FROM configurations WHERE id = ?"
				if test_helpers.UsePostgres() {
					queryStr = test_helpers.ReplaceQuestionMarks(queryStr)
				}
				rows, err := db.QueryContext(ctx, queryStr, sqldb.VersionID)
				Expect(err).NotTo(HaveOccurred())
				defer rows.Close()

				Expect(rows.Next()).To(BeTrue())

				var versionData string
				err = rows.Scan(&versionData)
				Expect(err).NotTo(HaveOccurred())

				var actualVersion models.Version
				err = json.Unmarshal([]byte(versionData), &actualVersion)
				Expect(err).NotTo(HaveOccurred())

				Expect(actualVersion).To(Equal(*version))
			})
		})
	})

	Describe("Version", func() {
		Context("when the version exists", func() {
			It("retrieves the version from the database", func() {
				expectedVersion := &models.Version{CurrentVersion: 199}
				err := helper.Transact(ctx, logger, db, func(l lager.Logger, tx helpers.Tx) error {
					return sqlDB.SetVersion(tx, ctx, logger, expectedVersion)
				})
				Expect(err).NotTo(HaveOccurred())

				var version *models.Version
				err = helper.Transact(ctx, logger, db, func(l lager.Logger, tx helpers.Tx) error {
					version, err = sqlDB.Version(tx, ctx, logger)
					return err
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(*version).To(Equal(*expectedVersion))
			})
		})

		Context("when the database is down", func() {
			var (
				sqlDB    *sqldb.SQLDB
				db       *sql.DB
				helperDB helpers.QueryableDB
			)

			BeforeEach(func() {
				var err error
				dbParams := &helpers.BBSDBParam{
					DriverName:                    dbDriverName,
					DatabaseConnectionString:      dbBaseConnectionString,
					SqlCACertFile:                 "",
					SqlEnableIdentityVerification: false,
					ConnectionTimeout:             time.Duration(600),
					ReadTimeout:                   time.Duration(600),
					WriteTimeout:                  time.Duration(600),
				}
				db, err = helpers.Connect(logger, dbParams)
				Expect(err).NotTo(HaveOccurred())
				helperDB = helpers.NewMonitoredDB(db, monitor.New())
				sqlDB = sqldb.NewSQLDB(helperDB, 5, 5, cryptor, fakeGUIDProvider, fakeClock, dbFlavor, fakeMetronClient)
			})

			AfterEach(func() {
				Expect(db.Close()).To(Succeed())
			})

			It("does not return an ErrResourceNotFound", func() {
				err := helper.Transact(ctx, logger, helperDB, func(l lager.Logger, tx helpers.Tx) error {
					_, err := sqlDB.Version(tx, ctx, logger)
					return err
				})
				Expect(err).NotTo(MatchError(models.ErrResourceNotFound))
			})
		})

		Context("when the version key does not exist", func() {
			BeforeEach(func() {
				_, err := db.ExecContext(ctx, "DELETE FROM configurations")
				Expect(err).NotTo(HaveOccurred())
			})

			It("does not log an error", func() {
				Expect(logger.Errors).To(BeEmpty())
			})

			It("returns a ErrResourceNotFound", func() {
				var version *models.Version
				err := helper.Transact(ctx, logger, db, func(l lager.Logger, tx helpers.Tx) error {
					var err error
					version, err = sqlDB.Version(tx, ctx, logger)
					return err
				})
				Expect(err).To(MatchError(models.ErrResourceNotFound))
				Expect(version).To(BeNil())
			})
		})

		Context("when the version key is not valid json", func() {
			It("returns :a ErrDeserialize", func() {
				queryStr := "UPDATE configurations SET value = '{{' WHERE id = ?"
				_, err := db.ExecContext(ctx, helpers.RebindForFlavor(queryStr, dbDriverName), sqldb.VersionID)
				Expect(err).NotTo(HaveOccurred())

				err = helper.Transact(ctx, logger, db, func(l lager.Logger, tx helpers.Tx) error {
					_, err := sqlDB.Version(tx, ctx, logger)
					return err
				})
				Expect(err).To(MatchError(models.ErrDeserialize))
			})
		})
	})
})
