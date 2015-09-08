package main_test

import (
	"encoding/json"
	"io"
	"os"

	"github.com/cloudfoundry-incubator/bbs/cmd/bbs/testrunner"
	"github.com/cloudfoundry-incubator/bbs/db/etcd"
	"github.com/cloudfoundry-incubator/bbs/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/tedsuo/ifrit/ginkgomon"
)

var _ = Describe("Migration Version", func() {
	var migrationFixtureFilePath, migrationFilePath string
	BeforeEach(func() {
		migrationFixtureFilePath = "fixtures/9999999999_test_migration.go"
		migrationFilePath = "../../db/migrations/9999999999_test_migration.go"
		migrationFixtureFile, err := os.Open(migrationFixtureFilePath)
		Expect(err).NotTo(HaveOccurred())

		migrationFile, err := os.Create(migrationFilePath)
		Expect(err).NotTo(HaveOccurred())

		_, err = io.Copy(migrationFile, migrationFixtureFile)
		Expect(err).NotTo(HaveOccurred())

		err = migrationFixtureFile.Close()
		Expect(err).NotTo(HaveOccurred())

		err = migrationFile.Close()
		Expect(err).NotTo(HaveOccurred())

		bbsConfig, err := gexec.Build("github.com/cloudfoundry-incubator/bbs/cmd/bbs", "-race")
		Expect(err).NotTo(HaveOccurred())
		bbsBinPath = string(bbsConfig)

		bbsRunner = testrunner.New(bbsBinPath, bbsArgs)
		bbsProcess = ginkgomon.Invoke(bbsRunner)
	})

	AfterEach(func() {
		ginkgomon.Kill(bbsProcess)
		err := os.RemoveAll(migrationFilePath)
		Expect(err).NotTo(HaveOccurred())

		bbsConfig, err := gexec.Build("github.com/cloudfoundry-incubator/bbs/cmd/bbs", "-race")
		Expect(err).NotTo(HaveOccurred())
		bbsBinPath = string(bbsConfig)
	})

	Context("Running Migrations", func() {
		It("loads and runs the given migrations", func() {
			response, err := storeClient.Get(etcd.VersionKey, false, false)
			Expect(err).NotTo(HaveOccurred())

			var version models.Version
			err = json.Unmarshal([]byte(response.Node.Value), &version)
			Expect(err).NotTo(HaveOccurred())

			Expect(version.CurrentVersion).To(BeEquivalentTo(9999999999))
			Expect(version.TargetVersion).To(BeEquivalentTo(9999999999))

			response, err = storeClient.Get("/test/key", false, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(response.Node.Value).To(Equal("jim is awesome"))
		})
	})
})
