package main_test

import (
	"encoding/base64"

	"github.com/cloudfoundry-incubator/bbs/cmd/bbs/testrunner"
	"github.com/cloudfoundry-incubator/bbs/db/etcd"
	"github.com/cloudfoundry-incubator/bbs/format"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/bbs/models/test/model_helpers"
	"github.com/tedsuo/ifrit/ginkgomon"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SerializationFormat", func() {
	var task *models.Task

	JustBeforeEach(func() {
		task = model_helpers.NewValidTask("task-1")

		bbsRunner = testrunner.New(bbsBinPath, bbsArgs)
		bbsProcess = ginkgomon.Invoke(bbsRunner)

		err := client.DesireTask(task.TaskGuid, task.Domain, task.TaskDefinition)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		ginkgomon.Kill(bbsProcess)
	})

	Context("when the format is set to legacy", func() {
		BeforeEach(func() {
			bbsArgs.SerializationFormat = "json_no_envelope"
		})

		It("writes the value as unencoded json with no metadata", func() {
			res, err := etcdClient.Get(etcd.TaskSchemaPathByGuid(task.TaskGuid), false, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(res.Node.Value[0]).To(BeEquivalentTo('{'))
		})
	})

	Context("when the format is set to unencoded_json", func() {
		BeforeEach(func() {
			bbsArgs.SerializationFormat = "json"
		})

		It("writes the value as unencoded json with metadata", func() {
			res, err := etcdClient.Get(etcd.TaskSchemaPathByGuid(task.TaskGuid), false, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(res.Node.Value[:2]).To(BeEquivalentTo(format.UNENCODED[:]))
			Expect(res.Node.Value[2]).To(BeEquivalentTo(format.JSON))
			Expect(res.Node.Value[3]).To(BeEquivalentTo(format.V0))
			Expect(res.Node.Value[4]).To(BeEquivalentTo('{'))
		})
	})

	Context("when the format is set to encoded_proto", func() {
		BeforeEach(func() {
			bbsArgs.SerializationFormat = "proto"
		})

		It("writes the value as base64 encoded protobufs with metadata", func() {
			res, err := etcdClient.Get(etcd.TaskSchemaPathByGuid(task.TaskGuid), false, false)
			Expect(err).NotTo(HaveOccurred())

			Expect(res.Node.Value[:2]).To(BeEquivalentTo(format.BASE64[:]))

			payload, err := base64.StdEncoding.DecodeString(string(res.Node.Value[2:]))
			Expect(payload[0]).To(BeEquivalentTo(format.PROTO))
			Expect(payload[1]).To(BeEquivalentTo(format.V0))
		})
	})
})
