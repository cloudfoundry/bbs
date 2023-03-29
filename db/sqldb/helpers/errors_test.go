package helpers_test

import (
	"code.cloudfoundry.org/bbs/db/sqldb/helpers"
	"github.com/go-sql-driver/mysql"
	"github.com/jackc/pgx"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("SQL Helpers Errors", func() {
	var (
		helper helpers.SQLHelper
	)

	BeforeEach(func() {
		helper = helpers.NewSQLHelper(dbFlavor)
	})

	Describe("ConvertSQLError", func() {
		It("returns a descriptive error for unknown postgres SQL errors", func() {
			err := helper.ConvertSQLError(pgx.PgError{Code: "foo"})
			Expect(err).To(MatchError("sql-unknown, error code: foo, flavor: postgres"))
		})

		It("returns a descriptive error for unknown MySQL SQL errors", func() {
			err := helper.ConvertSQLError(&mysql.MySQLError{Number: 9999})
			Expect(err).To(MatchError("sql-unknown, error code: 9999, flavor: mysql"))
		})
	})
})
