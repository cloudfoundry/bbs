package sqldb_test

import (
	"crypto/rand"
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/bbs/db/sqldb"
	"code.cloudfoundry.org/bbs/encryption"
	"code.cloudfoundry.org/bbs/format"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/bbs/test_helpers"
)

var _ = Describe("Encryption", func() {
	Describe("SetEncryptionKeyLabel", func() {
		It("sets the encryption key label into the database", func() {
			expectedLabel := "expectedLabel"
			err := sqlDB.SetEncryptionKeyLabel(ctx, logger, expectedLabel)
			Expect(err).NotTo(HaveOccurred())

			queryStr := `SELECT value FROM configurations WHERE id = ?`
			if test_helpers.UsePostgres() {
				queryStr = test_helpers.ReplaceQuestionMarks(queryStr)
			}

			rows, err := db.QueryContext(ctx, queryStr, sqldb.EncryptionKeyID)
			Expect(err).NotTo(HaveOccurred())
			defer rows.Close()
			Expect(rows.Next()).To(BeTrue())
			var label string
			err = rows.Scan(&label)
			Expect(err).NotTo(HaveOccurred())
			Expect(label).To(Equal(expectedLabel))
		})

		Context("when the label is too long", func() {
			It("returns an error trying to insert", func() {
				expectedLabel := randStr(256)
				err := sqlDB.SetEncryptionKeyLabel(ctx, logger, expectedLabel)
				Expect(err).To(Equal(models.ErrBadRequest))
			})
		})

		Context("When the encryption key is already set", func() {
			BeforeEach(func() {
				previouslySetLabel := "jim-likes-kittens-meow"
				err := sqlDB.SetEncryptionKeyLabel(ctx, logger, previouslySetLabel)
				Expect(err).NotTo(HaveOccurred())
			})

			It("replaces the encryption key label in the database", func() {
				expectedLabel := "expectedLabel"
				err := sqlDB.SetEncryptionKeyLabel(ctx, logger, expectedLabel)
				Expect(err).NotTo(HaveOccurred())

				queryStr := "SELECT value FROM configurations WHERE id = ?"
				if test_helpers.UsePostgres() {
					queryStr = test_helpers.ReplaceQuestionMarks(queryStr)
				}
				rows, err := db.QueryContext(ctx, queryStr, sqldb.EncryptionKeyID)
				Expect(err).NotTo(HaveOccurred())
				defer rows.Close()
				Expect(rows.Next()).To(BeTrue())
				var label string
				err = rows.Scan(&label)
				Expect(err).NotTo(HaveOccurred())
				Expect(label).To(Equal(expectedLabel))
			})

			Context("when the label is too long", func() {
				It("returns an error trying to insert", func() {
					expectedLabel := randStr(256)
					err := sqlDB.SetEncryptionKeyLabel(ctx, logger, expectedLabel)
					Expect(err).To(Equal(models.ErrBadRequest))
				})
			})
		})
	})

	Describe("EncryptionKeyLabel", func() {
		Context("when the encription key label key exists", func() {
			It("retrieves the encrption key label from the database", func() {
				label := "expectedLabel"
				queryStr := "INSERT INTO configurations VALUES (?, ?)"
				if test_helpers.UsePostgres() {
					queryStr = test_helpers.ReplaceQuestionMarks(queryStr)
				}
				_, err := db.ExecContext(ctx, queryStr, sqldb.EncryptionKeyID, label)
				Expect(err).NotTo(HaveOccurred())

				keyLabel, err := sqlDB.EncryptionKeyLabel(ctx, logger)
				Expect(err).NotTo(HaveOccurred())

				Expect(keyLabel).To(Equal(label))
			})
		})

		Context("when the encryption key label key does not exist", func() {
			It("returns a ErrResourceNotFound", func() {
				keyLabel, err := sqlDB.EncryptionKeyLabel(ctx, logger)
				Expect(err).To(MatchError(models.ErrResourceNotFound))
				Expect(keyLabel).To(Equal(""))
			})
		})
	})

	makeCryptor := func(activeLabel string, decryptionLabels ...string) encryption.Cryptor {
		activeKey, err := encryption.NewKey(activeLabel, fmt.Sprintf("%s-passphrase", activeLabel))
		Expect(err).NotTo(HaveOccurred())

		decryptionKeys := []encryption.Key{}
		for _, label := range decryptionLabels {
			key, err := encryption.NewKey(label, fmt.Sprintf("%s-passphrase", label))
			Expect(err).NotTo(HaveOccurred())
			decryptionKeys = append(decryptionKeys, key)
		}
		if len(decryptionKeys) == 0 {
			decryptionKeys = nil
		}

		keyManager, err := encryption.NewKeyManager(activeKey, decryptionKeys)
		Expect(err).NotTo(HaveOccurred())
		return encryption.NewCryptor(keyManager, rand.Reader)
	}

	Describe("PerformEncryption", func() {
		It("recursively re-encrypts tasks and desired lrps", func() {
			var cryptor encryption.Cryptor
			var encoder format.Encoder

			unencodedTaskDef := []byte("some text")
			unencodedRunInfo := []byte("another value")
			unencodedRoutes := []byte("some random routes")
			unencodedVolumePlacement := []byte("more value")
			unencodedMetricTags := []byte("some metric tags")
			taskGuid := "uniquetaskguid"
			processGuid := "uniqueprocessguid"

			cryptor = makeCryptor("old")
			encoder = format.NewEncoder(cryptor)

			encodedTaskDef, err := encoder.Encode(unencodedTaskDef)
			Expect(err).NotTo(HaveOccurred())

			encodedRunInfo, err := encoder.Encode(unencodedRunInfo)
			Expect(err).NotTo(HaveOccurred())

			encodedRoutes, err := encoder.Encode(unencodedRoutes)
			Expect(err).NotTo(HaveOccurred())

			encodedVolumePlacement, err := encoder.Encode(unencodedVolumePlacement)
			Expect(err).NotTo(HaveOccurred())

			encodedMetricTags, err := encoder.Encode(unencodedMetricTags)
			Expect(err).NotTo(HaveOccurred())

			queryStr := "INSERT INTO tasks (guid, domain, task_definition) VALUES (?, ?, ?)"
			if test_helpers.UsePostgres() {
				queryStr = test_helpers.ReplaceQuestionMarks(queryStr)
			}
			_, err = db.ExecContext(ctx, queryStr, taskGuid, "fake-domain", encodedTaskDef)
			Expect(err).NotTo(HaveOccurred())

			queryStr = `
				INSERT INTO desired_lrps
					(process_guid, domain, log_guid, instances, run_info, memory_mb,
					disk_mb, rootfs, routes, volume_placement, metric_tags, modification_tag_epoch)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
			if test_helpers.UsePostgres() {
				queryStr = test_helpers.ReplaceQuestionMarks(queryStr)
			}
			_, err = db.ExecContext(ctx, queryStr, processGuid, "fake-domain", "some-log-guid", 1, encodedRunInfo, 10, 10,
				"some-root-fs", encodedRoutes, encodedVolumePlacement, encodedMetricTags, "10")
			Expect(err).NotTo(HaveOccurred())
			cryptor = makeCryptor("new", "old")

			sqlDB := sqldb.NewSQLDB(db, 5, 5, cryptor, fakeGUIDProvider, fakeClock, dbFlavor, fakeMetronClient)
			err = sqlDB.PerformEncryption(ctx, logger)
			Expect(err).NotTo(HaveOccurred())

			cryptor = makeCryptor("new")
			encoder = format.NewEncoder(cryptor)

			var result []byte
			queryStr = "SELECT task_definition FROM tasks WHERE guid = ?"
			if test_helpers.UsePostgres() {
				queryStr = test_helpers.ReplaceQuestionMarks(queryStr)
			}
			row := db.QueryRowContext(ctx, queryStr, taskGuid)
			err = row.Scan(&result)
			Expect(err).NotTo(HaveOccurred())
			decryptedTaskDef, err := encoder.Decode(result)
			Expect(err).NotTo(HaveOccurred())
			Expect(decryptedTaskDef).To(Equal(unencodedTaskDef))

			var runInfo, routes, volumePlacement, metricTags []byte
			queryStr = "SELECT run_info, routes, volume_placement, metric_tags FROM desired_lrps WHERE process_guid = ?"
			if test_helpers.UsePostgres() {
				queryStr = test_helpers.ReplaceQuestionMarks(queryStr)
			}
			row = db.QueryRowContext(ctx, queryStr, processGuid)
			err = row.Scan(&runInfo, &routes, &volumePlacement, &metricTags)
			Expect(err).NotTo(HaveOccurred())

			decryptedRunInfo, err := encoder.Decode(runInfo)
			Expect(decryptedRunInfo).To(Equal(unencodedRunInfo))
			Expect(err).NotTo(HaveOccurred())

			decryptedRoutes, err := encoder.Decode(routes)
			Expect(err).NotTo(HaveOccurred())
			Expect(decryptedRoutes).To(Equal(unencodedRoutes))

			decryptedVolumePlacement, err := encoder.Decode(volumePlacement)
			Expect(err).NotTo(HaveOccurred())
			Expect(decryptedVolumePlacement).To(Equal(unencodedVolumePlacement))

			decryptedMetricTags, err := encoder.Decode(metricTags)
			Expect(err).NotTo(HaveOccurred())
			Expect(decryptedMetricTags).To(Equal(unencodedMetricTags))
		})

		Context("actual_lrps encryption", func() {
			var (
				processGuid    = "uniqueprocessguid"
				netInfo        string
				internalRoutes string
				metricTags     string
				cryptor        encryption.Cryptor
				encoder        format.Encoder
			)

			BeforeEach(func() {
				cryptor = makeCryptor("old")
				encoder = format.NewEncoder(cryptor)
			})

			JustBeforeEach(func() {
				cryptor = makeCryptor("new", "old")
				sqlDB := sqldb.NewSQLDB(db, 5, 5, cryptor, fakeGUIDProvider, fakeClock, dbFlavor, fakeMetronClient)
				err := sqlDB.PerformEncryption(ctx, logger)
				Expect(err).NotTo(HaveOccurred())
			})

			Context("when there is an actual lrp with non-empty net-info", func() {
				BeforeEach(func() {
					var err error
					info, err := encoder.Encode([]byte("actual value"))
					Expect(err).NotTo(HaveOccurred())
					netInfo = string(info)

					internalRoutesEncoded, err := encoder.Encode([]byte("{}"))
					Expect(err).NotTo(HaveOccurred())
					internalRoutes = string(internalRoutesEncoded)

					metricTagsEncoded, err := encoder.Encode([]byte("{}"))
					Expect(err).NotTo(HaveOccurred())
					metricTags = string(metricTagsEncoded)

					queryStr := `
						INSERT INTO actual_lrps
							(process_guid, domain, net_info, instance_index, modification_tag_epoch, state, internal_routes, metric_tags)
						VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
					if test_helpers.UsePostgres() {
						queryStr = test_helpers.ReplaceQuestionMarks(queryStr)
					}
					_, err = db.ExecContext(ctx, queryStr,
						processGuid, "fake-domain", netInfo, 0, "10", "yo", internalRoutes, metricTags)
					Expect(err).NotTo(HaveOccurred())
				})

				It("gets encrypted properly", func() {
					cryptor := makeCryptor("new")
					encoder := format.NewEncoder(cryptor)

					var dbNetInfo, dbInternalRoutes, dbMetricTags []byte
					queryStr := "SELECT net_info, internal_routes, metric_tags FROM actual_lrps WHERE process_guid = ?"
					if test_helpers.UsePostgres() {
						queryStr = test_helpers.ReplaceQuestionMarks(queryStr)
					}
					row := db.QueryRowContext(ctx, queryStr, processGuid)
					err := row.Scan(&dbNetInfo, &dbInternalRoutes, &dbMetricTags)
					Expect(err).NotTo(HaveOccurred())
					decryptedNetInfo, err := encoder.Decode(dbNetInfo)
					Expect(err).NotTo(HaveOccurred())
					Expect(string(decryptedNetInfo)).To(Equal("actual value"))
					decryptedInternalRoutes, err := encoder.Decode(dbInternalRoutes)
					Expect(err).NotTo(HaveOccurred())
					Expect(string(decryptedInternalRoutes)).To(Equal("{}"))
					decryptedMetricTags, err := encoder.Decode(dbMetricTags)
					Expect(err).NotTo(HaveOccurred())
					Expect(string(decryptedMetricTags)).To(Equal("{}"))
				})
			})

			Context("when there is an actual lrp with empty net_info", func() {
				BeforeEach(func() {
					netInfo = ""

					internalRoutesEncoded, err := encoder.Encode([]byte("{}"))
					Expect(err).NotTo(HaveOccurred())
					internalRoutes = string(internalRoutesEncoded)

					metricTagsEncoded, err := encoder.Encode([]byte("{}"))
					Expect(err).NotTo(HaveOccurred())
					metricTags = string(metricTagsEncoded)

					queryStr := `
						INSERT INTO actual_lrps
							(process_guid, domain, net_info, instance_index, modification_tag_epoch, state, internal_routes, metric_tags)
						VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
					if test_helpers.UsePostgres() {
						queryStr = test_helpers.ReplaceQuestionMarks(queryStr)
					}
					_, err = db.ExecContext(ctx, queryStr,
						processGuid, "fake-domain", netInfo, 0, "10", "yo", internalRoutes, metricTags)
					Expect(err).NotTo(HaveOccurred())
				})

				It("is left empty without getting encrypted", func() {
					var netInfo []byte
					queryStr := "SELECT net_info FROM actual_lrps WHERE process_guid = ?"
					if test_helpers.UsePostgres() {
						queryStr = test_helpers.ReplaceQuestionMarks(queryStr)
					}
					row := db.QueryRowContext(ctx, queryStr, processGuid)
					err := row.Scan(&netInfo)
					Expect(err).NotTo(HaveOccurred())
					Expect(netInfo).To(HaveLen(0))
				})
			})

			Context("when where are more than 1 lrp with the same process guid", func() {
				BeforeEach(func() {
					info, err := encoder.Encode([]byte("actual value 1"))
					Expect(err).NotTo(HaveOccurred())
					netInfo1 := string(info)

					internalRoutesEncoded, err := encoder.Encode([]byte("{}"))
					Expect(err).NotTo(HaveOccurred())
					internalRoutes = string(internalRoutesEncoded)

					metricTagsEncoded, err := encoder.Encode([]byte("{}"))
					Expect(err).NotTo(HaveOccurred())
					metricTags = string(metricTagsEncoded)

					queryStr := `
						INSERT INTO actual_lrps
							(process_guid, domain, net_info, instance_index, modification_tag_epoch, state, internal_routes, metric_tags)
						VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
					if test_helpers.UsePostgres() {
						queryStr = test_helpers.ReplaceQuestionMarks(queryStr)
					}
					_, err = db.ExecContext(ctx, queryStr,
						processGuid, "fake-domain", netInfo1, 0, "10", "yo", internalRoutes, metricTags)
					Expect(err).NotTo(HaveOccurred())

					info, err = encoder.Encode([]byte("actual value 2"))
					Expect(err).NotTo(HaveOccurred())
					netInfo2 := string(info)

					_, err = db.ExecContext(ctx, queryStr,
						processGuid, "fake-domain", netInfo2, 1, "10", "yo", internalRoutes, metricTags)
					Expect(err).NotTo(HaveOccurred())

					info, err = encoder.Encode([]byte("actual value 3"))
					Expect(err).NotTo(HaveOccurred())
					netInfo3 := string(info)

					_, err = db.ExecContext(ctx, queryStr,
						processGuid, "fake-domain", netInfo3, 2, "10", "yo", internalRoutes, metricTags)
					Expect(err).NotTo(HaveOccurred())
				})

				It("gets encrypted properly", func() {
					cryptor := makeCryptor("new")
					encoder := format.NewEncoder(cryptor)

					var netInfo []byte
					queryStr := "SELECT net_info FROM actual_lrps WHERE process_guid = ? and instance_index = ?"
					if test_helpers.UsePostgres() {
						queryStr = test_helpers.ReplaceQuestionMarks(queryStr)
					}
					row := db.QueryRowContext(ctx, queryStr, processGuid, 0)
					err := row.Scan(&netInfo)
					Expect(err).NotTo(HaveOccurred())
					decrypted, err := encoder.Decode(netInfo)
					Expect(err).NotTo(HaveOccurred())
					Expect(string(decrypted)).To(Equal("actual value 1"))

					row = db.QueryRowContext(ctx, queryStr, processGuid, 1)
					err = row.Scan(&netInfo)
					Expect(err).NotTo(HaveOccurred())
					decrypted, err = encoder.Decode(netInfo)
					Expect(err).NotTo(HaveOccurred())
					Expect(string(decrypted)).To(Equal("actual value 2"))
				})
			})
		})

		It("does not fail encryption if it can't read a record", func() {
			var cryptor encryption.Cryptor
			var encoder format.Encoder

			value1 := []byte("some text")
			taskGuid := "uniquetaskguid"

			cryptor = makeCryptor("unknown")
			encoder = format.NewEncoder(cryptor)

			encoded1, err := encoder.Encode(value1)
			Expect(err).NotTo(HaveOccurred())

			queryStr := "INSERT INTO tasks (guid, domain, task_definition) VALUES (?, ?, ?)"
			if test_helpers.UsePostgres() {
				queryStr = test_helpers.ReplaceQuestionMarks(queryStr)
			}
			_, err = db.ExecContext(ctx, queryStr, taskGuid, "fake-domain", encoded1)
			Expect(err).NotTo(HaveOccurred())

			cryptor = makeCryptor("new", "old")

			sqlDB := sqldb.NewSQLDB(db, 5, 5, cryptor, fakeGUIDProvider, fakeClock, dbFlavor, fakeMetronClient)
			err = sqlDB.PerformEncryption(ctx, logger)
			Expect(err).NotTo(HaveOccurred())
		})

		It("encrypts all text columns except for whitelisted", func() {
			cryptor := makeCryptor("old")
			encoder := format.NewEncoder(cryptor)

			tablesWithRequiredKeys := map[string]map[string]interface{}{
				"tasks": {"guid": "some-guid", "domain": "fake-domain"},
				"desired_lrps": {
					"process_guid":           "some-guid",
					"domain":                 "fake-domain",
					"log_guid":               "some-log-guid",
					"instances":              1,
					"memory_mb":              10,
					"disk_mb":                10,
					"rootfs":                 "some-root-fs",
					"modification_tag_epoch": "10",
				},
				"actual_lrps": {
					"process_guid":           "some-guid",
					"instance_index":         0,
					"domain":                 "fake-domain",
					"state":                  "running",
					"modification_tag_epoch": "10",
				},
			}
			dataTypesToEncrypt := map[string]bool{"text": true, "mediumtext": true, "longtext": true}
			whiteListedFields := map[string]map[string]bool{
				"tasks":        {"result": true},
				"desired_lrps": {"annotation": true, "placement_tags": true},
				"actual_lrps":  {},
			}
			var columnName, dataType string
			dataToStore, err := encoder.Encode([]byte("actual value"))
			Expect(err).NotTo(HaveOccurred())

			needsToBeEncrypted := map[string]map[string]bool{}

			for table, primaryKeys := range tablesWithRequiredKeys {
				queryStr := "SELECT column_name, data_type FROM information_schema.columns WHERE table_name = ?"
				if test_helpers.UsePostgres() {
					queryStr = test_helpers.ReplaceQuestionMarks(queryStr)
				}
				rows, err := db.QueryContext(ctx, queryStr, table)
				Expect(err).NotTo(HaveOccurred())
				defer rows.Close()

				for rows.Next() {
					err := rows.Scan(&columnName, &dataType)
					Expect(err).NotTo(HaveOccurred())
					if _, ok := dataTypesToEncrypt[dataType]; ok {
						if _, ok := whiteListedFields[table][columnName]; ok {
							continue
						}
						if _, ok := needsToBeEncrypted[table]; !ok {
							needsToBeEncrypted[table] = map[string]bool{}
						}
						needsToBeEncrypted[table][columnName] = true
					}
				}

				var columnNames, marks []string
				var values []interface{}
				for primaryKeyName, primaryKeyValue := range primaryKeys {
					columnNames = append(columnNames, primaryKeyName)
					values = append(values, primaryKeyValue)
					marks = append(marks, "?")
				}

				for columnName, encryptIt := range needsToBeEncrypted[table] {
					if encryptIt {
						columnNames = append(columnNames, columnName)
						values = append(values, dataToStore)
						marks = append(marks, "?")
					}
				}

				columnNamesStr := strings.Join(columnNames, ",")
				marksStr := strings.Join(marks, ",")

				queryStr = fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s)`, table, columnNamesStr, marksStr)
				if test_helpers.UsePostgres() {
					queryStr = test_helpers.ReplaceQuestionMarks(queryStr)
				}
				_, err = db.ExecContext(ctx, queryStr, values...)
				Expect(err).NotTo(HaveOccurred())
			}

			cryptor = makeCryptor("new", "old")
			sqlDB := sqldb.NewSQLDB(db, 5, 5, cryptor, fakeGUIDProvider, fakeClock, dbFlavor, fakeMetronClient)
			err = sqlDB.PerformEncryption(ctx, logger)
			Expect(err).NotTo(HaveOccurred())

			cryptor = makeCryptor("new")
			encoder = format.NewEncoder(cryptor)

			for table, columns := range needsToBeEncrypted {
				for column, encryptIt := range columns {
					if encryptIt {
						var reEncryptedData []byte
						queryStr := fmt.Sprintf("SELECT %s FROM %s", column, table)
						rowsEncrypted, err := db.QueryContext(ctx, queryStr)
						Expect(err).NotTo(HaveOccurred())
						defer rowsEncrypted.Close()

						for rowsEncrypted.Next() {
							err = rowsEncrypted.Scan(&reEncryptedData)
							Expect(err).NotTo(HaveOccurred())
							decryptedData, err := encoder.Decode(reEncryptedData)
							Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("Decrypting column %s on table %s", column, table))
							Expect(string(decryptedData)).To(Equal("actual value"), fmt.Sprintf("Decrypting column %s on table %s", column, table))
						}
					}
				}
			}
		})
	})
})
