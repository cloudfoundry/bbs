package migrations_test

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	etcddb "github.com/cloudfoundry-incubator/bbs/db/etcd"
	"github.com/cloudfoundry-incubator/bbs/db/migrations"
	"github.com/cloudfoundry-incubator/bbs/format"
	"github.com/cloudfoundry-incubator/bbs/migration"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/bbs/models/test/model_helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ETCD to SQL Migration", func() {
	var (
		migration    migration.Migration
		serializer   format.Serializer
		migrationErr error
	)

	BeforeEach(func() {
		if !useSQL {
			Skip("SQL backing is not available")
		}

		migration = migrations.NewETCDToSQL()
		serializer = format.NewSerializer(cryptor)

		rawSQLDB.Exec("DROP TABLE domains;")
		rawSQLDB.Exec("DROP TABLE tasks;")
		rawSQLDB.Exec("DROP TABLE desired_lrps;")
		rawSQLDB.Exec("DROP TABLE actual_lrps;")
	})

	It("appends itself to the migration list", func() {
		Expect(migrations.Migrations).To(ContainElement(migration))
	})

	Describe("Version", func() {
		It("returns the timestamp from which it was created", func() {
			Expect(migration.Version()).To(BeEquivalentTo(1461790966))
		})
	})

	Describe("Up", func() {
		JustBeforeEach(func() {
			migration.SetStoreClient(storeClient)
			migration.SetRawSQLDB(rawSQLDB)
			migration.SetCryptor(cryptor)
			migration.SetClock(fakeClock)
			migrationErr = migration.Up(logger)
		})

		Context("when there is existing data in the database", func() {
			BeforeEach(func() {
				var err error

				_, err = rawSQLDB.Exec(`CREATE TABLE domains( domain VARCHAR(255) PRIMARY KEY);`)
				Expect(err).NotTo(HaveOccurred())

				_, err = rawSQLDB.Exec(`CREATE TABLE desired_lrps( process_guid VARCHAR(255) PRIMARY KEY);`)
				Expect(err).NotTo(HaveOccurred())

				_, err = rawSQLDB.Exec(`CREATE TABLE actual_lrps( process_guid VARCHAR(255) PRIMARY KEY);`)
				Expect(err).NotTo(HaveOccurred())

				_, err = rawSQLDB.Exec(`CREATE TABLE tasks( guid VARCHAR(255) PRIMARY KEY);`)
				Expect(err).NotTo(HaveOccurred())

				_, err = rawSQLDB.Exec(`INSERT INTO domains VALUES ('test-domain')`)
				Expect(err).NotTo(HaveOccurred())

				_, err = rawSQLDB.Exec(`INSERT INTO desired_lrps VALUES ('test-guid')`)
				Expect(err).NotTo(HaveOccurred())

				_, err = rawSQLDB.Exec(`INSERT INTO actual_lrps VALUES ('test-guid')`)
				Expect(err).NotTo(HaveOccurred())

				_, err = rawSQLDB.Exec(`INSERT INTO tasks VALUES ('test-guid')`)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should truncate the tables and start migration", func() {
				var value string
				err := rawSQLDB.QueryRow(`SELECT domain FROM domains WHERE domain='test-domain'`).Scan(&value)
				Expect(err).To(MatchError(sql.ErrNoRows))
			})

			It("should truncate desired_lrps table", func() {
				var value string
				err := rawSQLDB.QueryRow(`SELECT process_guid FROM desired_lrps WHERE process_guid='test-guid'`).Scan(&value)
				Expect(err).To(MatchError(sql.ErrNoRows))
			})

			It("should truncate actual_lrps table", func() {
				var value string
				err := rawSQLDB.QueryRow(`SELECT process_guid FROM actual_lrps WHERE process_guid='test-guid'`).Scan(&value)
				Expect(err).To(MatchError(sql.ErrNoRows))
			})

			It("should truncate tasks table", func() {
				var value string
				err := rawSQLDB.QueryRow(`SELECT guid FROM tasks WHERE guid='test-guid'`).Scan(&value)
				Expect(err).To(MatchError(sql.ErrNoRows))
			})
		})

		Context("when etcd is not configured", func() {
			BeforeEach(func() {
				storeClient = nil
			})

			It("creates the sql schema and returns", func() {
				Expect(migrationErr).NotTo(HaveOccurred())
				rows, err := rawSQLDB.Query(`SHOW TABLES`)
				Expect(err).NotTo(HaveOccurred())

				tables := []string{}
				for rows.Next() {
					var tableName string
					err := rows.Scan(&tableName)
					Expect(err).NotTo(HaveOccurred())
					tables = append(tables, tableName)
				}
				Expect(tables).To(ConsistOf("domains", "desired_lrps", "actual_lrps", "tasks"))
			})
		})

		Describe("Domains", func() {
			BeforeEach(func() {
				_, err := storeClient.Set(etcddb.DomainSchemaPath("domain-1"), []byte(""), 100)
				Expect(err).NotTo(HaveOccurred())
				_, err = storeClient.Set(etcddb.DomainSchemaPath("domain-2"), []byte(""), 100)
				Expect(err).NotTo(HaveOccurred())
			})

			It("creates a domain entry in sql for each domain in etcd", func() {
				Expect(migrationErr).NotTo(HaveOccurred())

				rows, err := rawSQLDB.Query(`SELECT domain, expire_time FROM domains`)
				Expect(err).NotTo(HaveOccurred())
				domains := map[string]int64{}
				for rows.Next() {
					var domain string
					var expireTime int64
					err := rows.Scan(&domain, &expireTime)
					Expect(err).NotTo(HaveOccurred())
					domains[domain] = expireTime
				}
				Expect(domains).To(HaveLen(2))
				Expect(domains).To(HaveKey("domain-1"))
				Expect(domains).To(HaveKeyWithValue("domain-1",
					BeNumerically(">", fakeClock.Now().Add(time.Second*95).UnixNano())))
				Expect(domains).To(HaveKeyWithValue("domain-2",
					BeNumerically(">", fakeClock.Now().Add(time.Second*95).UnixNano())))
			})
		})

		Describe("Desired LRPs", func() {
			var (
				existingDesiredLRPs []migrations.ETCDToSQLDesiredLRP
				desiredLRPsToCreate int
			)

			BeforeEach(func() {
				encoder := format.NewEncoder(cryptor)

				desiredLRPsToCreate = 3
				for i := 0; i < desiredLRPsToCreate; i++ {
					processGuid := fmt.Sprintf("process-guid-%d", i)
					var desiredLRP *models.DesiredLRP
					desiredLRP = model_helpers.NewValidDesiredLRP(processGuid)

					schedulingInfo, runInfo := desiredLRP.CreateComponents(fakeClock.Now())

					var (
						encryptedVolumePlacement []byte
						err                      error
					)
					if i == 0 { // test for nil and full VolumePlacements
						schedulingInfo.VolumePlacement = nil
						encryptedVolumePlacement, err = serializer.Marshal(logger, format.ENCRYPTED_PROTO, &models.VolumePlacement{})
					} else {
						encryptedVolumePlacement, err = serializer.Marshal(logger, format.ENCRYPTED_PROTO, schedulingInfo.VolumePlacement)
					}
					Expect(err).NotTo(HaveOccurred())

					volumePlacementData, err := encoder.Decode(encryptedVolumePlacement)
					Expect(err).NotTo(HaveOccurred())

					routesData, err := json.Marshal(desiredLRP.Routes)
					Expect(err).NotTo(HaveOccurred())

					schedInfoData, err := serializer.Marshal(logger, format.ENCRYPTED_PROTO, &schedulingInfo)
					Expect(err).NotTo(HaveOccurred())
					_, err = storeClient.Set(etcddb.DesiredLRPSchedulingInfoSchemaPath(processGuid), schedInfoData, 0)
					Expect(err).NotTo(HaveOccurred())

					runInfoData, err := serializer.Marshal(logger, format.ENCRYPTED_PROTO, &runInfo)
					Expect(err).NotTo(HaveOccurred())
					_, err = storeClient.Set(etcddb.DesiredLRPRunInfoSchemaPath(processGuid), runInfoData, 0)
					Expect(err).NotTo(HaveOccurred())

					existingDesiredLRPs = append(existingDesiredLRPs, migrations.ETCDToSQLDesiredLRP{
						ProcessGuid: desiredLRP.ProcessGuid, Domain: desiredLRP.Domain,
						LogGuid: desiredLRP.LogGuid, Annotation: desiredLRP.Annotation,
						Instances: desiredLRP.Instances, RootFS: desiredLRP.RootFs,
						DiskMB: desiredLRP.DiskMb, MemoryMB: desiredLRP.MemoryMb,
						Routes: routesData, ModificationTagEpoch: desiredLRP.ModificationTag.Epoch,
						ModificationTagIndex: desiredLRP.ModificationTag.Index, RunInfo: runInfoData,
						VolumePlacement: volumePlacementData,
					})
				}
			})

			It("creates a desired lrp in sqldb for each desired lrp in etcd", func() {
				Expect(migrationErr).NotTo(HaveOccurred())

				rows, err := rawSQLDB.Query(`
					SELECT
						process_guid, domain, log_guid, annotation, instances, memory_mb,
						disk_mb, rootfs, routes, modification_tag_epoch,
						modification_tag_index, run_info, volume_placement
					FROM desired_lrps
				`)
				Expect(err).NotTo(HaveOccurred())

				desiredLRPs := []migrations.ETCDToSQLDesiredLRP{}

				for rows.Next() {
					var desiredLRPTest migrations.ETCDToSQLDesiredLRP

					var encodedVolumePlacement []byte
					err := rows.Scan(&desiredLRPTest.ProcessGuid, &desiredLRPTest.Domain,
						&desiredLRPTest.LogGuid, &desiredLRPTest.Annotation,
						&desiredLRPTest.Instances, &desiredLRPTest.MemoryMB,
						&desiredLRPTest.DiskMB, &desiredLRPTest.RootFS,
						&desiredLRPTest.Routes, &desiredLRPTest.ModificationTagEpoch,
						&desiredLRPTest.ModificationTagIndex, &desiredLRPTest.RunInfo, &encodedVolumePlacement)
					Expect(err).NotTo(HaveOccurred())

					encoder := format.NewEncoder(cryptor)
					desiredLRPTest.VolumePlacement, err = encoder.Decode(encodedVolumePlacement)
					desiredLRPs = append(desiredLRPs, desiredLRPTest)
				}

				Expect(desiredLRPs).To(HaveLen(desiredLRPsToCreate))
				Expect(desiredLRPs).To(BeEquivalentTo(existingDesiredLRPs))
			})
		})

		Describe("Actual LRPs", func() {
			var (
				existingActualLRPs []migrations.ETCDToSQLActualLRP
				actualLRPsToCreate int
			)

			BeforeEach(func() {
				actualLRPsToCreate = 3
				for i := 0; i < actualLRPsToCreate; i++ {
					processGuid := fmt.Sprintf("process-guid-%d", i)
					actualLRP := model_helpers.NewValidActualLRP(processGuid, int32(i))

					actualLRPData, err := serializer.Marshal(logger, format.ENCRYPTED_PROTO, actualLRP)
					Expect(err).NotTo(HaveOccurred())
					_, err = storeClient.Set(etcddb.ActualLRPSchemaPath(processGuid, int32(i)), actualLRPData, 0)
					Expect(err).NotTo(HaveOccurred())

					encoder := format.NewEncoder(cryptor)
					encryptedNetInfo, err := serializer.Marshal(logger, format.ENCRYPTED_PROTO, &actualLRP.ActualLRPNetInfo)
					Expect(err).NotTo(HaveOccurred())
					netInfoData, err := encoder.Decode(encryptedNetInfo)
					Expect(err).NotTo(HaveOccurred())

					existingActualLRPs = append(existingActualLRPs, migrations.ETCDToSQLActualLRP{
						ProcessGuid:          actualLRP.ProcessGuid,
						Index:                actualLRP.Index,
						Domain:               actualLRP.Domain,
						InstanceGuid:         actualLRP.InstanceGuid,
						CellId:               actualLRP.CellId,
						ActualLRPNetInfo:     netInfoData,
						CrashCount:           actualLRP.CrashCount,
						CrashReason:          actualLRP.CrashReason,
						State:                actualLRP.State,
						PlacementError:       actualLRP.PlacementError,
						Since:                actualLRP.Since,
						ModificationTagEpoch: actualLRP.ModificationTag.Epoch,
						ModificationTagIndex: actualLRP.ModificationTag.Index,
					})
				}
			})

			It("creates a actual lrp in sqldb for each actual lrp in etcd", func() {
				Expect(migrationErr).NotTo(HaveOccurred())

				rows, err := rawSQLDB.Query(`
					SELECT
						process_guid, instance_index, domain, instance_guid, cell_id, net_info,
						crash_count, crash_reason, state, placement_error, since,
						modification_tag_epoch, modification_tag_index
					FROM actual_lrps
				`)
				Expect(err).NotTo(HaveOccurred())

				actualLRPs := []migrations.ETCDToSQLActualLRP{}
				encoder := format.NewEncoder(cryptor)

				for rows.Next() {
					var actualLRPTest migrations.ETCDToSQLActualLRP
					var encryptedNetInfo []byte

					err := rows.Scan(&actualLRPTest.ProcessGuid, &actualLRPTest.Index,
						&actualLRPTest.Domain, &actualLRPTest.InstanceGuid,
						&actualLRPTest.CellId, &encryptedNetInfo,
						&actualLRPTest.CrashCount, &actualLRPTest.CrashReason,
						&actualLRPTest.State, &actualLRPTest.PlacementError,
						&actualLRPTest.Since, &actualLRPTest.ModificationTagEpoch,
						&actualLRPTest.ModificationTagIndex)
					Expect(err).NotTo(HaveOccurred())

					actualLRPTest.ActualLRPNetInfo, err = encoder.Decode(encryptedNetInfo)
					Expect(err).NotTo(HaveOccurred())

					actualLRPs = append(actualLRPs, actualLRPTest)
				}

				Expect(actualLRPs).To(HaveLen(actualLRPsToCreate))
				Expect(actualLRPs).To(BeEquivalentTo(existingActualLRPs))
			})
		})

		Describe("Tasks", func() {
			var (
				existingTasks []migrations.ETCDToSQLTask
				tasksToCreate int
			)

			BeforeEach(func() {
				tasksToCreate = 3
				for i := 0; i < tasksToCreate; i++ {
					taskGuid := fmt.Sprintf("task-guid-%d", i)
					task := model_helpers.NewValidTask(taskGuid)

					taskData, err := serializer.Marshal(logger, format.ENCRYPTED_PROTO, task)
					Expect(err).NotTo(HaveOccurred())
					_, err = storeClient.Set(etcddb.TaskSchemaPath(task), taskData, 0)
					Expect(err).NotTo(HaveOccurred())

					encoder := format.NewEncoder(cryptor)
					encryptedDefinition, err := serializer.Marshal(logger, format.ENCRYPTED_PROTO, task.TaskDefinition)
					Expect(err).NotTo(HaveOccurred())
					definitionData, err := encoder.Decode(encryptedDefinition)
					Expect(err).NotTo(HaveOccurred())

					existingTasks = append(existingTasks, migrations.ETCDToSQLTask{
						TaskGuid:         task.TaskGuid,
						Domain:           task.Domain,
						CellId:           task.CellId,
						CreatedAt:        task.CreatedAt,
						UpdatedAt:        task.UpdatedAt,
						FirstCompletedAt: task.FirstCompletedAt,
						State:            int32(task.State),
						Result:           task.Result,
						Failed:           task.Failed,
						FailureReason:    task.FailureReason,
						TaskDefinition:   definitionData,
					})
				}
			})

			It("creates a task in sqldb for each task in etcd", func() {
				Expect(migrationErr).NotTo(HaveOccurred())

				rows, err := rawSQLDB.Query(`
					SELECT
						guid, domain, updated_at, created_at, first_completed_at, state,
						cell_id, result, failed, failure_reason, task_definition
					FROM tasks
				`)
				Expect(err).NotTo(HaveOccurred())

				tasks := []migrations.ETCDToSQLTask{}
				encoder := format.NewEncoder(cryptor)

				for rows.Next() {
					var taskTest migrations.ETCDToSQLTask
					var encryptedDefinition []byte

					err := rows.Scan(&taskTest.TaskGuid, &taskTest.Domain,
						&taskTest.UpdatedAt, &taskTest.CreatedAt, &taskTest.FirstCompletedAt,
						&taskTest.State, &taskTest.CellId, &taskTest.Result,
						&taskTest.Failed, &taskTest.FailureReason, &encryptedDefinition)
					Expect(err).NotTo(HaveOccurred())

					taskTest.TaskDefinition, err = encoder.Decode(encryptedDefinition)
					Expect(err).NotTo(HaveOccurred())

					tasks = append(tasks, taskTest)
				}

				Expect(tasks).To(HaveLen(tasksToCreate))
				Expect(tasks).To(BeEquivalentTo(existingTasks))
			})
		})
	})

	Describe("Down", func() {
		It("returns a not implemented error", func() {
			Expect(migration.Down(logger)).To(HaveOccurred())
		})
	})
})
