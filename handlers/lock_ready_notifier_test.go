package handlers_test

import (
	"code.cloudfoundry.org/bbs/handlers"
	"github.com/tedsuo/ifrit"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("LockReadyNotifier", func() {
	It("closes the lock ready channel", func() {
		lockReady := make(chan struct{})
		lockReadyNotifier := handlers.NewLockReadyNotifier(lockReady)

		ifrit.Invoke(lockReadyNotifier)
		Eventually(lockReady).Should(BeClosed())
	})
})
