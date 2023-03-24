package fakesqldriver_test

import (
	"database/sql/driver"
	"strings"

	"code.cloudfoundry.org/bbs/db/sqldb/fakesqldriver/fakesqldriverfakes"
	"github.com/go-sql-driver/mysql"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Invalid Connections", func() {
	BeforeEach(func() {
		fakeConn.PrepareStub = func(query string) (driver.Stmt, error) {
			fakeStmt := &fakesqldriverfakes.FakeStmt{}
			fakeStmt.NumInputReturns(strings.Count(query, "?"))
			fakeStmt.QueryReturns(nil, mysql.ErrInvalidConn)
			return fakeStmt, nil
		}
	})

	Context("Domains", func() {
		It("retries on deadlocks", func() {
			_, err := sqlDB.FreshDomains(ctx, logger)
			Expect(err).To(HaveOccurred())
			Expect(fakeConn.BeginCallCount()).To(Equal(3))
		})
	})
})
