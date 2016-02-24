package sqldb_test

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"time"

	"github.com/cloudfoundry-incubator/bbs/db/sqldb"
	"github.com/cloudfoundry-incubator/bbs/encryption"
	"github.com/cloudfoundry-incubator/bbs/format"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/clock/fakeclock"
	"github.com/pivotal-golang/lager/lagertest"

	_ "github.com/go-sql-driver/mysql"

	"testing"
)

var (
	db         *sql.DB
	sqlDB      *sqldb.SQLDB
	fakeClock  *fakeclock.FakeClock
	logger     *lagertest.TestLogger
	cryptor    encryption.Cryptor
	serializer format.Serializer
)

func TestSql(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Sql Suite")
}

var _ = BeforeSuite(func() {
	var err error
	fakeClock = fakeclock.NewFakeClock(time.Now())
	logger = lagertest.NewTestLogger("sql-db")

	db, err = sql.Open("mysql", "root:password@/")
	Expect(err).NotTo(HaveOccurred())
	Expect(db.Ping()).NotTo(HaveOccurred())

	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE diego_%d", GinkgoParallelNode()))
	Expect(err).NotTo(HaveOccurred())

	db, err = sql.Open("mysql", fmt.Sprintf("root:password@/diego_%d?parseTime=true", GinkgoParallelNode()))
	Expect(err).NotTo(HaveOccurred())
	Expect(db.Ping()).NotTo(HaveOccurred())

	createTables(db)

	encryptionKey, err := encryption.NewKey("label", "passphrase")
	Expect(err).NotTo(HaveOccurred())
	keyManager, err := encryption.NewKeyManager(encryptionKey, nil)
	Expect(err).NotTo(HaveOccurred())
	cryptor = encryption.NewCryptor(keyManager, rand.Reader)
	serializer = format.NewSerializer(cryptor)

	sqlDB = sqldb.NewSQLDB(db, format.ENCRYPTED_PROTO, cryptor, fakeClock)
})

var _ = AfterEach(func() {
	truncateTables(db)
})

var _ = AfterSuite(func() {
	_, err := db.Exec(fmt.Sprintf("DROP DATABASE diego_%d", GinkgoParallelNode()))
	Expect(err).NotTo(HaveOccurred())

	Expect(db.Close()).NotTo(HaveOccurred())
})

func truncateTables(db *sql.DB) {
	for _, query := range truncateTablesSQL {
		result, err := db.Exec(query)
		Expect(err).NotTo(HaveOccurred())
		Expect(result.RowsAffected()).To(BeEquivalentTo(0))
	}
}

func createTables(db *sql.DB) {
	for _, query := range createTablesSQL {
		result, err := db.Exec(query)
		Expect(err).NotTo(HaveOccurred())
		Expect(result.RowsAffected()).To(BeEquivalentTo(0))
	}
}

var truncateTablesSQL = []string{
	"TRUNCATE TABLE domains;",
	"TRUNCATE TABLE configurations;",
	"TRUNCATE TABLE tasks;",
	"TRUNCATE TABLE desired_lrps;",
}

var createTablesSQL = []string{
	createDomainSQL,
	createConfigurationsSQL,
	createTasksSQL,
	createDesiredLRPsSQL,
}

const createDomainSQL = `CREATE TABLE domains(
	domain varchar(255) PRIMARY KEY,
	expire_time timestamp
);`

const createConfigurationsSQL = `CREATE TABLE configurations(
	id varchar(255) PRIMARY KEY,
	value varchar(255)
);`

const createTasksSQL = `CREATE TABLE tasks(
	guid varchar(255) PRIMARY KEY,
	domain varchar(255) NOT NULL,
	created_at timestamp(6),
	updated_at timestamp(6),
	first_completed_at timestamp(6),
	state int,
	cell_id varchar(255) NOT NULL DEFAULT "",
	result text,
	failed bool DEFAULT false,
	failure_reason varchar(255) NOT NULL DEFAULT "",
	task_definition blob NOT NULL
);`

const createDesiredLRPsSQL = `CREATE TABLE desired_lrps(
	process_guid varchar(255) PRIMARY KEY,
	domain varchar(255) NOT NULL,
	log_guid varchar(255) NOT NULL,
	annotation text,
	instances int NOT NULL,
	memory_mb int NOT NULL,
	disk_mb int NOT NULL,
	rootfs varchar(255) NOT NULL,
	routes blob NOT NULL,
	modification_tag_epoch varchar(255) NOT NULL,
	modification_tag_index int,
	run_info blob NOT NULL
);`

func randStr(strSize int) string {
	alphanum := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var bytes = make([]byte, strSize)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}
	return string(bytes)
}
