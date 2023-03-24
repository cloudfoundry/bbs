package main_test

import (
	"fmt"
	"net"

	"code.cloudfoundry.org/bbs/cmd/bbs/testrunner"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	ginkgomon "github.com/tedsuo/ifrit/ginkgomon_v2"
)

var _ = Describe("Debug Address", func() {
	var debugAddress string

	BeforeEach(func() {
		port, err := portAllocator.ClaimPorts(1)
		Expect(err).NotTo(HaveOccurred())
		debugAddress = fmt.Sprintf("127.0.0.1:%d", port)
		bbsConfig.DebugAddress = debugAddress
	})

	JustBeforeEach(func() {
		bbsRunner = testrunner.New(bbsBinPath, bbsConfig)
		bbsProcess = ginkgomon.Invoke(bbsRunner)
	})

	It("listens on the debug address specified", func() {
		_, err := net.Dial("tcp", debugAddress)
		Expect(err).NotTo(HaveOccurred())
	})
})
