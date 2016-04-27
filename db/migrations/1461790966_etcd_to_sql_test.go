package migrations_test

import (
	"time"

	"github.com/cloudfoundry-incubator/bbs/db/deprecations"
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
		migration = migrations.NewETCDToSQL()
		serializer = format.NewSerializer(cryptor)
	})

	JustBeforeEach(func() {
		migration.SetStoreClient(storeClient)
		migration.SetRawSQLDB(rawSQLDB)
		migration.SetCryptor(cryptor)
		migration.SetClock(fakeClock)
		migrationErr = migration.Up(logger)
	})

	AfterEach(func() {
		truncateTablesSQL := []string{
			"TRUNCATE TABLE domains",
			"TRUNCATE TABLE configurations",
			"TRUNCATE TABLE tasks",
			"TRUNCATE TABLE desired_lrps",
			"TRUNCATE TABLE actual_lrps",
		}

		for _, query := range truncateTablesSQL {
			result, err := rawSQLDB.Exec(query)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.RowsAffected()).To(BeEquivalentTo(0))
		}
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
				Expect(domains).To(HaveKeyWithValue("domain-1", fakeClock.Now().Add(time.Second*100)))
				Expect(domains).To(HaveKeyWithValue("domain-2", fakeClock.Now().Add(time.Second*100)))
			})
		})

		Describe("LRPs", func() {
			var (
				existingDesiredLRP *models.DesiredLRP
			)
			BeforeEach(func() {
				existingDesiredLRP = model_helpers.NewValidDesiredLRP("process-guid")
				payload, err := serializer.Marshal(logger, format.ENCRYPTED_PROTO, existingDesiredLRP)
				Expect(err).NotTo(HaveOccurred())
				_, err = storeClient.Set(deprecations.DesiredLRPSchemaPath(existingDesiredLRP), payload, 0)
				Expect(err).NotTo(HaveOccurred())
			})

		})
	})

	Describe("Down", func() {
		It("returns a not implemented error", func() {
			Expect(migration.Down(logger)).To(HaveOccurred())
		})
	})
})
