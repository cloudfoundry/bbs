package migration_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"code.cloudfoundry.org/bbs/db/dbfakes"
	"code.cloudfoundry.org/bbs/db/sqldb/helpers"
	"code.cloudfoundry.org/bbs/encryption"
	"code.cloudfoundry.org/bbs/encryption/encryptionfakes"
	"code.cloudfoundry.org/bbs/migration"
	"code.cloudfoundry.org/bbs/migration/migrationfakes"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/bbs/test_helpers"
	"code.cloudfoundry.org/clock"
	mfakes "code.cloudfoundry.org/diego-logging-client/testhelpers"
	"code.cloudfoundry.org/lager/v3"
	"code.cloudfoundry.org/lager/v3/lagertest"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tedsuo/ifrit"
	ginkgomon "github.com/tedsuo/ifrit/ginkgomon_v2"
)

var _ = Describe("Migration Manager", func() {
	var (
		manager          ifrit.Runner
		migrationProcess ifrit.Process

		logger *lagertest.TestLogger

		fakeSQLDB *dbfakes.FakeDB
		rawSQLDB  *sql.DB

		migrations []migration.Migration

		migrationsDone chan struct{}

		fakeMigration *migrationfakes.FakeMigration

		cryptor encryption.Cryptor

		fakeMetronClient *mfakes.FakeIngressClient
	)

	BeforeEach(func() {
		migrationsDone = make(chan struct{})

		fakeMetronClient = new(mfakes.FakeIngressClient)

		logger = lagertest.NewTestLogger("test")

		fakeSQLDB = &dbfakes.FakeDB{}

		cryptor = &encryptionfakes.FakeCryptor{}

		fakeMigration = &migrationfakes.FakeMigration{}
		migrations = []migration.Migration{fakeMigration}
	})

	JustBeforeEach(func() {
		manager = migration.NewManager(logger, fakeSQLDB, rawSQLDB, cryptor, migrations, migrationsDone, clock.NewClock(), "db-driver", fakeMetronClient)
		migrationProcess = ifrit.Background(manager)
	})

	AfterEach(func() {
		ginkgomon.Kill(migrationProcess)
		Eventually(migrationProcess.Wait()).Should(Receive())
	})

	Context("when configured with a SQL database", func() {
		var sqlProcess ifrit.Process
		BeforeEach(func() {
			dbName := fmt.Sprintf("diego_%d", GinkgoParallelProcess())
			sqlRunner := test_helpers.NewSQLRunner(dbName)
			sqlProcess = ginkgomon.Invoke(sqlRunner)
			var err error
			dbParams := &helpers.BBSDBParam{
				DriverName:                    sqlRunner.DriverName(),
				DatabaseConnectionString:      sqlRunner.ConnectionString(),
				SqlCACertFile:                 "",
				SqlEnableIdentityVerification: false,
			}
			rawSQLDB, err = helpers.Connect(logger, dbParams)
			Expect(err).NotTo(HaveOccurred())
			Expect(rawSQLDB.Ping()).NotTo(HaveOccurred())

			fakeSQLDB.VersionReturns(&models.Version{}, nil)
		})

		AfterEach(func() {
			Expect(rawSQLDB.Close()).NotTo(HaveOccurred())
			ginkgomon.Kill(sqlProcess, 5*time.Second)
		})

		It("fetches the stored version from sql", func() {
			Eventually(fakeSQLDB.VersionCallCount).Should(Equal(1))
			Consistently(fakeSQLDB.VersionCallCount).Should(Equal(1))

			ginkgomon.Interrupt(migrationProcess)
			Eventually(migrationProcess.Wait()).Should(Receive(BeNil()))
		})

		Context("when there is no version", func() {
			var (
				fakeMigrationToSQL   *migrationfakes.FakeMigration
				fakeSQLOnlyMigration *migrationfakes.FakeMigration
			)

			BeforeEach(func() {
				fakeSQLDB.VersionReturns(nil, models.ErrResourceNotFound)

				fakeMigrationToSQL = &migrationfakes.FakeMigration{}
				fakeMigrationToSQL.VersionReturns(100)

				fakeSQLOnlyMigration = &migrationfakes.FakeMigration{}
				fakeSQLOnlyMigration.VersionReturns(101)

				migrations = []migration.Migration{fakeSQLOnlyMigration, fakeMigrationToSQL}
			})

			It("runs all the migrations in the correct order and sets the version to the latest migration version", func() {
				Eventually(fakeSQLDB.SetVersionCallCount).Should(Equal(3))

				_, _, _, version := fakeSQLDB.SetVersionArgsForCall(0)
				Expect(version.CurrentVersion).To(BeEquivalentTo(0))

				_, _, _, version = fakeSQLDB.SetVersionArgsForCall(1)
				Expect(version.CurrentVersion).To(BeEquivalentTo(100))

				_, _, _, version = fakeSQLDB.SetVersionArgsForCall(2)
				Expect(version.CurrentVersion).To(BeEquivalentTo(101))

				Expect(fakeMigrationToSQL.UpCallCount()).To(Equal(1))
				Expect(fakeSQLOnlyMigration.UpCallCount()).To(Equal(1))
			})
		})

		Context("when fetching the version fails", func() {
			BeforeEach(func() {
				fakeSQLDB.VersionReturns(nil, errors.New("kablamo"))
			})

			It("fails early", func() {
				var err error
				Eventually(migrationProcess.Wait()).Should(Receive(&err))
				Expect(err).To(MatchError("kablamo"))
				Expect(migrationProcess.Ready()).ToNot(BeClosed())
				Expect(migrationsDone).NotTo(BeClosed())
			})
		})

		Context("when the current version is newer than bbs migration version", func() {
			BeforeEach(func() {
				fakeSQLDB.VersionReturns(&models.Version{CurrentVersion: 100}, nil)
				fakeMigration.VersionReturns(99)
			})

			It("shuts down without signalling ready", func() {
				var err error
				Eventually(migrationProcess.Wait()).Should(Receive(&err))
				Expect(err).To(MatchError("Existing DB version (100) exceeds bbs version (99)"))
				Expect(migrationProcess.Ready()).ToNot(BeClosed())
				Expect(migrationsDone).NotTo(BeClosed())
			})
		})

		Context("when the current version is the same as the bbs migration version", func() {
			BeforeEach(func() {
				fakeSQLDB.VersionReturns(&models.Version{CurrentVersion: 100}, nil)
				fakeMigration.VersionReturns(100)
			})

			It("signals ready and does not change the version", func() {
				Eventually(migrationProcess.Ready()).Should(BeClosed())
				Expect(migrationsDone).To(BeClosed())
				Consistently(fakeSQLDB.SetVersionCallCount).Should(Equal(0))
			})
		})

		Context("when the current version is older than the maximum migration version", func() {
			var fakeMigration102 *migrationfakes.FakeMigration

			BeforeEach(func() {
				_, err := rawSQLDB.Exec("create table foo (bar int)")
				Expect(err).NotTo(HaveOccurred())
				fakeMigration102 = &migrationfakes.FakeMigration{}
				fakeMigration102.VersionReturns(102)
				fakeMigration102.UpStub = func(tx *sql.Tx, _ lager.Logger) error {
					_, err := tx.Exec("insert into foo values (123)")
					return err
				}

				fakeSQLDB.VersionReturns(&models.Version{CurrentVersion: 99}, nil)
				fakeMigration.VersionReturns(100)

				migrations = []migration.Migration{fakeMigration102, fakeMigration}
			})

			Context("when starting transaction fails", func() {
				BeforeEach(func() {
					Expect(rawSQLDB.Close()).NotTo(HaveOccurred())
				})

				It("fails early", func() {
					var err error
					Eventually(migrationProcess.Wait()).Should(Receive(&err))
					Expect(err).To(MatchError("sql: database is closed"))
					Expect(migrationProcess.Ready()).ToNot(BeClosed())
					Expect(migrationsDone).NotTo(BeClosed())
				})
			})

			Context("when starting transaction succeeds", func() {
				AfterEach(func() {
					_, err := rawSQLDB.Exec("drop table if exists foo")
					Expect(err).NotTo(HaveOccurred())
				})

				Describe("reporting", func() {
					It("reports the duration that it took to migrate", func() {
						Eventually(migrationProcess.Ready()).Should(BeClosed())
						Expect(migrationsDone).To(BeClosed())

						Expect(fakeMetronClient.SendDurationCallCount()).To(Equal(1))
						name, value, _ := fakeMetronClient.SendDurationArgsForCall(0)
						Expect(name).To(Equal("MigrationDuration"))
						Expect(value).NotTo(BeZero())
					})
				})

				It("sorts the migrations and runs them sequentially", func() {
					Eventually(migrationProcess.Ready()).Should(BeClosed())
					Expect(migrationsDone).To(BeClosed())
					Consistently(fakeSQLDB.SetVersionCallCount).Should(Equal(2))

					_, _, _, version := fakeSQLDB.SetVersionArgsForCall(0)
					Expect(version).To(Equal(&models.Version{CurrentVersion: 100}))

					_, _, _, version = fakeSQLDB.SetVersionArgsForCall(1)
					Expect(version).To(Equal(&models.Version{CurrentVersion: 102}))

					Expect(fakeMigration.UpCallCount()).To(Equal(1))
					Expect(fakeMigration102.UpCallCount()).To(Equal(1))
				})

				It("commits transaction", func() {
					Eventually(migrationProcess.Ready()).Should(BeClosed())
					Expect(migrationsDone).To(BeClosed())
					row := rawSQLDB.QueryRow("select count(*) from foo")
					var value int
					err := row.Scan(&value)
					Expect(err).NotTo(HaveOccurred())
					Expect(value).To(Equal(1))
				})

				Context("when migration fails", func() {
					BeforeEach(func() {
						fakeMigration102.UpStub = func(tx *sql.Tx, _ lager.Logger) error {
							_, err := tx.Exec("invalid sql")
							return err
						}
					})

					It("returns an error", func() {
						var err error
						Eventually(migrationProcess.Wait()).Should(Receive(&err))
						Expect(err.Error()).To(ContainSubstring("invalid"))
						Expect(migrationProcess.Ready()).ToNot(BeClosed())
						Expect(migrationsDone).NotTo(BeClosed())
					})
				})

				Context("when commiting transaction fails", func() {
					BeforeEach(func() {
						fakeMigration102.UpStub = func(tx *sql.Tx, _ lager.Logger) error {
							return tx.Commit()
						}
					})

					It("returns an error", func() {
						var err error
						Eventually(migrationProcess.Wait()).Should(Receive(&err))
						Expect(migrationProcess.Ready()).ToNot(BeClosed())
						Expect(migrationsDone).NotTo(BeClosed())
					})
				})

				Context("when saving the version fails", func() {
					BeforeEach(func() {
						fakeSQLDB.SetVersionStub = func(_ helpers.Tx, _ context.Context, _ lager.Logger, version *models.Version) error {
							if version.CurrentVersion == 102 {
								return errors.New("some-error")
							}
							return nil
						}
					})

					It("rolls back transaction", func() {
						var err error
						Eventually(migrationProcess.Wait()).Should(Receive(&err))
						Expect(migrationsDone).ToNot(BeClosed())
						row := rawSQLDB.QueryRow("select count(*) from foo")
						var value int
						err = row.Scan(&value)
						Expect(err).NotTo(HaveOccurred())
						Expect(value).To(Equal(0))
					})
				})

				Describe("and one of the migrations takes a long time", func() {
					var longMigrationExitChan chan struct{}

					BeforeEach(func() {
						longMigrationExitChan = make(chan struct{}, 1)
						longMigration := &migrationfakes.FakeMigration{}
						longMigration.UpStub = func(tx *sql.Tx, logger lager.Logger) error {
							<-longMigrationExitChan
							return nil
						}
						longMigration.VersionReturns(103)
						migrations = []migration.Migration{longMigration}
					})

					AfterEach(func() {
						longMigrationExitChan <- struct{}{}
					})

					It("should not close the channel until the migration finishes", func() {
						Consistently(migrationProcess.Ready()).ShouldNot(BeClosed())
					})

					Context("when the migration finishes", func() {
						JustBeforeEach(func() {
							Eventually(longMigrationExitChan).Should(BeSent(struct{}{}))
						})

						It("should close the ready channel", func() {
							Eventually(migrationProcess.Ready()).Should(BeClosed())
						})
					})

					Context("when interrupted", func() {
						JustBeforeEach(func() {
							ginkgomon.Interrupt(migrationProcess)
						})

						It("exits and does not wait for the migration to finish", func() {
							Eventually(migrationProcess.Wait()).Should(Receive())
						})
					})
				})

				It("sets the cryptor on the migration", func() {
					Eventually(migrationProcess.Ready()).Should(BeClosed())
					Expect(migrationsDone).To(BeClosed())
					Expect(fakeMigration.SetCryptorCallCount()).To(Equal(1))
					actualCryptor := fakeMigration.SetCryptorArgsForCall(0)
					Expect(actualCryptor).To(Equal(cryptor))
				})

			})
		})

		Context("when there are no migrations", func() {
			BeforeEach(func() {
				migrations = []migration.Migration{}
			})

			Context("and there is an existing version", func() {
				BeforeEach(func() {
					fakeSQLDB.VersionReturns(&models.Version{CurrentVersion: 100}, nil)
				})

				It("treats the bbs migration version as 0", func() {
					var err error
					Eventually(migrationProcess.Wait()).Should(Receive(&err))
					Expect(err).To(MatchError("Existing DB version (100) exceeds bbs version (0)"))
					Expect(migrationProcess.Ready()).ToNot(BeClosed())
				})
			})

			Context("and there is an existing version 0", func() {
				BeforeEach(func() {
					fakeSQLDB.VersionReturns(&models.Version{CurrentVersion: 0}, nil)
				})

				It("it skips writing the version into the db", func() {
					Consistently(fakeSQLDB.SetVersionCallCount).Should(Equal(0))
				})
			})

			Context("and there is no existing version", func() {
				BeforeEach(func() {
					fakeSQLDB.VersionReturns(nil, models.ErrResourceNotFound)
				})

				It("writes a zero version into the db", func() {
					Eventually(fakeSQLDB.SetVersionCallCount).Should(Equal(1))

					_, _, _, version := fakeSQLDB.SetVersionArgsForCall(0)
					Expect(version.CurrentVersion).To(BeEquivalentTo(0))
					Expect(version.CurrentVersion).To(BeEquivalentTo(0))
				})
			})
		})
	})

	Context("when not configured with a database", func() {
		BeforeEach(func() {
			rawSQLDB = nil
		})

		It("fails early", func() {
			var err error
			Eventually(migrationProcess.Wait()).Should(Receive(&err))
			Expect(err).To(MatchError("no database configured"))
			Expect(migrationProcess.Ready()).ToNot(BeClosed())
			Expect(migrationsDone).NotTo(BeClosed())
		})
	})
})
