package sqldb_test

import (
	"code.cloudfoundry.org/bbs/db/sqldb"
	"code.cloudfoundry.org/bbs/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Lock", func() {
	Context("Lock", func() {
		Context("when the lock does not exist", func() {
			It("inserts the lock and grabs it", func() {
				newLock := models.Lock{
					Key:   "quack",
					Owner: "iamthelizardking",
					Value: "i can do anything",
				}

				err := sqlDB.Lock(logger, newLock)
				Expect(err).NotTo(HaveOccurred())

				lockQuery := sqldb.RebindForFlavor(
					"SELECT * FROM Locks WHERE key = ?",
					dbFlavor,
				)

				var key, owner, value string
				row := db.QueryRow(lockQuery, newLock.Key)
				Expect(row.Scan(&key, &owner, &value)).To(Succeed())
				Expect(key).To(Equal(newLock.Key))
				Expect(owner).To(Equal(newLock.Owner))
				Expect(value).To(Equal(newLock.Value))
			})
		})

		Context("when the lock does exist", func() {
			It("returns an error without grabbing the lock", func() {
				oldLock := models.Lock{
					Key:   "quack",
					Owner: "iamthelizardking",
					Value: "i can do anything",
				}

				err := sqlDB.Lock(logger, oldLock)
				Expect(err).NotTo(HaveOccurred())

				newLock := models.Lock{
					Key:   "quack",
					Owner: "jim",
					Value: "i have never seen the princess bride and never will",
				}

				err = sqlDB.Lock(logger, newLock)
				Expect(err).To(Equal(models.ErrLockCollision))
			})
		})
	})

	Context("ReleaseLock", func() {
		var lock models.Lock

		BeforeEach(func() {
			lock = models.Lock{
				Key:   "test",
				Owner: "jim",
				Value: "locks stuff for days",
			}
		})

		Context("when the lock exists", func() {
			BeforeEach(func() {
				query := sqldb.RebindForFlavor(
					`INSERT INTO locks (key, owner, value) VALUES (?, ?, ?);`,
					dbFlavor,
				)
				result, err := db.Exec(query, lock.Key, lock.Owner, lock.Value)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.RowsAffected()).To(BeEquivalentTo(1))
			})

			It("removes the lock from the lock table", func() {
				err := sqlDB.ReleaseLock(logger, lock)
				Expect(err).NotTo(HaveOccurred())

				rows, err := db.Query(`SELECT key FROM locks;`)
				Expect(err).NotTo(HaveOccurred())
				Expect(rows.Next()).To(BeFalse())
			})

			Context("when the lock is owned by another owner", func() {
				It("returns an error", func() {
					err := sqlDB.ReleaseLock(logger, models.Lock{
						Key:   "test",
						Owner: "not jim",
						Value: "beep boop",
					})
					Expect(err).To(HaveOccurred())
				})
			})
		})

		Context("when the lock does not exist", func() {
			It("returns an error", func() {
				err := sqlDB.ReleaseLock(logger, lock)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
