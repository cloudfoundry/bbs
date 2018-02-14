package sqldb_test

import (
	"errors"

	. "code.cloudfoundry.org/bbs/db/sqldb"
	"code.cloudfoundry.org/bbs/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = FDescribe("Errors", func() {

	Context("E", func() {
		It("creates a new non-fatal error", func() {
			e := E("set-version", errors.New("BOOOOOM!"))
			Expect(e).To(MatchError("set-version: BOOOOOM!"))
			Expect(IsAvoidableE(e)).To(BeFalse())
		})

		Context("when the error is avoidable", func() {
			It("creates a new non-fatal avoidable error", func() {
				e := E("set-version", models.ErrResourceNotFound)
				Expect(e).To(MatchError("set-version: the requested resource could not be found"))
				Expect(IsAvoidableE(e)).To(BeTrue())
			})
		})

		Context("nested ops", func() {
			It("concatenate all ops", func() {
				e := E("set-version", errors.New("BOOOOOM!"))
				e = E("migration", e)
				Expect(e).To(MatchError("migration / set-version: BOOOOOM!"))
				Expect(IsAvoidableE(e)).To(BeFalse())
			})
		})

		Context("F", func() {
			It("create a fatal error", func() {
				e := F("get-version", models.ErrResourceNotFound)
				Expect(e).To(MatchError("get-version: the requested resource could not be found"))
				Expect(IsAvoidableE(e)).To(BeFalse())
			})

			It("convert non-fatal error to a fatal error", func() {
				e := E("set-version", models.ErrResourceExists)
				e = F("migration", e)
				Expect(e).To(MatchError("migration / set-version: the requested resource already exists"))
				Expect(IsAvoidableE(e)).To(BeFalse())
			})
		})
	})
})
