package models_test

import (
	"encoding/json"
	"errors"

	. "code.cloudfoundry.org/bbs/models"

	ginkgo "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = ginkgo.Describe("Errors", func() {
	ginkgo.Describe("ConvertError", func() {
		ginkgo.It("maintains nils", func() {
			var err error = nil
			bbsError := ConvertError(err)
			Expect(bbsError).To(BeNil())
			Expect(bbsError == nil).To(BeTrue())
		})

		ginkgo.It("can convert a *Error back to *Error", func() {
			var err error = NewError(Error_ResourceConflict, "some message")
			bbsError := ConvertError(err)
			Expect(bbsError.Type).To(Equal(Error_ResourceConflict))
			Expect(bbsError.Message).To(Equal("some message"))
		})

		ginkgo.It("can convert a regular error to a *Error with unknown type", func() {
			var err error = errors.New("fail")
			bbsError := ConvertError(err)
			Expect(bbsError.Type).To(Equal(Error_UnknownError))
			Expect(bbsError.Message).To(Equal("fail"))
		})
	})

	ginkgo.Describe("Equal", func() {
		ginkgo.It("is true when the types are the same", func() {
			err1 := &Error{Type: 0, Message: "some-message"}
			err2 := &Error{Type: 0, Message: "some-other-message"}
			Expect(err1.Equal(err2)).To(BeTrue())
		})

		ginkgo.It("is false when the types are different", func() {
			err1 := &Error{Type: 0, Message: "some-message"}
			err2 := &Error{Type: 1, Message: "some-other-message"}
			Expect(err1.Equal(err2)).To(BeFalse())
		})

		ginkgo.It("is false when one is nil", func() {
			var err1 *Error = nil
			err2 := &Error{Type: 0, Message: "some-other-message"}
			Expect(err1.Equal(err2)).To(BeFalse())
		})

		ginkgo.It("is true when both errors are nil", func() {
			var err1 *Error = nil
			var err2 *Error = nil
			Expect(err1.Equal(err2)).To(BeTrue())
		})
	})

	ginkgo.Describe("Type", func() {
		ginkgo.Describe("serialization", func() {
			ginkgo.DescribeTable("marshals and unmarshals between the value and the expected JSON output",
				func(v Error_Type, expectedJSON string) {
					Expect(json.Marshal(v)).To(MatchJSON(expectedJSON))
					var testV Error_Type
					Expect(json.Unmarshal([]byte(expectedJSON), &testV)).To(Succeed())
					Expect(testV).To(Equal(v))
				},
				ginkgo.Entry("UnknownError", Error_UnknownError, `"UnknownError"`),
				ginkgo.Entry("InvalidRecord", Error_InvalidRecord, `"InvalidRecord"`),
				ginkgo.Entry("InvalidRequest", Error_InvalidRequest, `"InvalidRequest"`),
				ginkgo.Entry("InvalidResponse", Error_InvalidResponse, `"InvalidResponse"`),
				ginkgo.Entry("InvalidProtobufMessage", Error_InvalidProtobufMessage, `"InvalidProtobufMessage"`),
				ginkgo.Entry("InvalidJSON", Error_InvalidJSON, `"InvalidJSON"`),
				ginkgo.Entry("FailedToOpenEnvelope", Error_FailedToOpenEnvelope, `"FailedToOpenEnvelope"`),
				ginkgo.Entry("InvalidStateTransition", Error_InvalidStateTransition, `"InvalidStateTransition"`),
				ginkgo.Entry("ResourceConflict", Error_ResourceConflict, `"ResourceConflict"`),
				ginkgo.Entry("ResourceExists", Error_ResourceExists, `"ResourceExists"`),
				ginkgo.Entry("ResourceNotFound", Error_ResourceNotFound, `"ResourceNotFound"`),
				ginkgo.Entry("RouterError", Error_RouterError, `"RouterError"`),
				ginkgo.Entry("ActualLRPCannotBeClaimed", Error_ActualLRPCannotBeClaimed, `"ActualLRPCannotBeClaimed"`),
				ginkgo.Entry("ActualLRPCannotBeStarted", Error_ActualLRPCannotBeStarted, `"ActualLRPCannotBeStarted"`),
				ginkgo.Entry("ActualLRPCannotBeCrashed", Error_ActualLRPCannotBeCrashed, `"ActualLRPCannotBeCrashed"`),
				ginkgo.Entry("ActualLRPCannotBeFailed", Error_ActualLRPCannotBeFailed, `"ActualLRPCannotBeFailed"`),
				ginkgo.Entry("ActualLRPCannotBeRemoved", Error_ActualLRPCannotBeRemoved, `"ActualLRPCannotBeRemoved"`),
				ginkgo.Entry("ActualLRPCannotBeUnclaimed", Error_ActualLRPCannotBeUnclaimed, `"ActualLRPCannotBeUnclaimed"`),
				ginkgo.Entry("RunningOnDifferentCell", Error_RunningOnDifferentCell, `"RunningOnDifferentCell"`),
				ginkgo.Entry("GUIDGeneration", Error_GUIDGeneration, `"GUIDGeneration"`),
				ginkgo.Entry("Deserialize", Error_Deserialize, `"Deserialize"`),
				ginkgo.Entry("Deadlock", Error_Deadlock, `"Deadlock"`),
				ginkgo.Entry("Unrecoverable", Error_Unrecoverable, `"Unrecoverable"`),
				ginkgo.Entry("LockCollision", Error_LockCollision, `"LockCollision"`),
				ginkgo.Entry("Timeout", Error_Timeout, `"Timeout"`),
			)
		})
	})
})
