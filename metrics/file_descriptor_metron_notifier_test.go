package metrics_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"code.cloudfoundry.org/clock/fakeclock"
	mfakes "code.cloudfoundry.org/diego-logging-client/testhelpers"

	"code.cloudfoundry.org/bbs/metrics"
	"code.cloudfoundry.org/lager/lagertest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tedsuo/ifrit"
)

var _ = FDescribe("FileDescriptorMetronNotifier", func() {
	var (
		fdNotifier             ifrit.Process
		fakeMetronClient       *mfakes.FakeIngressClient
		fakeProcFileSystemPath string
		fakeClock              *fakeclock.FakeClock
		reportInterval         time.Duration
		logger                 *lagertest.TestLogger
	)

	BeforeEach(func() {
		fakeMetronClient = new(mfakes.FakeIngressClient)
		fakeProcFileSystemPath = createTestPath("", 10)
		reportInterval = 100 * time.Millisecond
		fakeClock = fakeclock.NewFakeClock(time.Unix(123, 456))
		logger = lagertest.NewTestLogger("test")
	})

	JustBeforeEach(func() {
		ticker := fakeClock.NewTicker(reportInterval)

		fdNotifier = ifrit.Invoke(
			metrics.NewFileDescriptorMetronNotifier(
				logger,
				fakeMetronClient,
				ticker,
				fakeProcFileSystemPath,
			),
		)
	})

	AfterEach(func() {
		fdNotifier.Signal(os.Interrupt)
		Eventually(fdNotifier.Wait(), 2*time.Second).Should(Receive())

		err := os.RemoveAll(fakeProcFileSystemPath)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("when the file descriptor metron notifier is running", func() {
		It("periodically emits the number of open file descriptors as a metric", func() {
			fakeClock.WaitForWatcherAndIncrement(reportInterval)

			Eventually(fakeMetronClient.SendMetricCallCount).Should(Equal(1))
			name, value := fakeMetronClient.SendMetricArgsForCall(0)
			Expect(name).To(Equal("OpenFileDescriptors"))
			Expect(value).To(BeEquivalentTo(10))

			fakeClock.WaitForWatcherAndIncrement(reportInterval)

			Eventually(fakeMetronClient.SendMetricCallCount).Should(Equal(2))
			name, value = fakeMetronClient.SendMetricArgsForCall(1)
			Expect(name).To(Equal("OpenFileDescriptors"))
			Expect(value).To(BeEquivalentTo(10))

			createTestPath(fakeProcFileSystemPath, 11)

			fakeClock.WaitForWatcherAndIncrement(reportInterval)

			Eventually(fakeMetronClient.SendMetricCallCount).Should(Equal(3))
			name, value = fakeMetronClient.SendMetricArgsForCall(2)
			Expect(name).To(Equal("OpenFileDescriptors"))
			Expect(value).To(BeEquivalentTo(11))
		})
	})

	Context("when the notifier fails to read the proc filesystem", func() {
		BeforeEach(func() {
			fakeProcFileSystemPath = "/proc/quack/moo"
		})

		It("doesn't send a metric", func() {
			fakeClock.WaitForWatcherAndIncrement(reportInterval)
			Consistently(fakeMetronClient.SendMetricCallCount).Should(Equal(0))
		})
	})
})

// TODO: check with routing about consolidating helpers
func createTestPath(path string, symlink int) string {
	// Create symlink structure similar to /proc/pid/fd in linux file system
	createSymlink := func(dir string, n int) {
		fd, err := ioutil.TempFile(dir, "socket")
		Expect(err).NotTo(HaveOccurred())
		for i := 0; i < n; i++ {
			fdId := strconv.Itoa(i)
			symlink := filepath.Join(dir, fdId)

			err := os.Symlink(fd.Name()+fdId, symlink)
			Expect(err).NotTo(HaveOccurred())
		}
	}
	if path != "" {
		createSymlink(path, symlink)
		return path
	}
	procPath, err := ioutil.TempDir("", "proc")
	Expect(err).NotTo(HaveOccurred())
	createSymlink(procPath, symlink)
	return procPath
}
