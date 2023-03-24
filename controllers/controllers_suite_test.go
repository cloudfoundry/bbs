package controllers_test

import (
	"context"

	"code.cloudfoundry.org/bbs/serviceclient/serviceclientfakes"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/rep/repfakes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestControllers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Controllers Suite")
}

var (
	fakeServiceClient    *serviceclientfakes.FakeServiceClient
	fakeRepClient        *repfakes.FakeClient
	fakeRepClientFactory *repfakes.FakeClientFactory
	logger               lager.Logger
	ctx                  context.Context
)

var _ = BeforeEach(func() {
	logger = lagertest.NewTestLogger("test")
	ctx = context.Background()
	fakeServiceClient = new(serviceclientfakes.FakeServiceClient)
	fakeRepClientFactory = new(repfakes.FakeClientFactory)
	fakeRepClient = new(repfakes.FakeClient)
	fakeRepClientFactory.CreateClientReturns(fakeRepClient, nil)
})
