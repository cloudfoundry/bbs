package controllers_test

import (
	"code.cloudfoundry.org/bbs/fake_bbs"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/rep/repfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestControllers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Controllers Suite")
}

var (
	fakeServiceClient    *fake_bbs.FakeServiceClient
	fakeRepClient        *repfakes.FakeClient
	fakeRepClientFactory *repfakes.FakeClientFactory
	logger               lager.Logger
)

var _ = BeforeEach(func() {
	logger = lagertest.NewTestLogger("test")
	fakeServiceClient = new(fake_bbs.FakeServiceClient)
	fakeRepClientFactory = new(repfakes.FakeClientFactory)
	fakeRepClient = new(repfakes.FakeClient)
	fakeRepClientFactory.CreateClientReturns(fakeRepClient)
})
