package controllers_test

import (
	"context"

	"code.cloudfoundry.org/bbs/serviceclient/serviceclientfakes"
	"code.cloudfoundry.org/lager/v3"
	"code.cloudfoundry.org/lager/v3/lagertest"
	modelsfakes "code.cloudfoundry.org/bbs/models/fakes"
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
	fakeRepClient        *modelsfakes.FakeRepClient
	fakeRepClientFactory *modelsfakes.FakeRepClientFactory
	logger               lager.Logger
	ctx                  context.Context
)

var _ = BeforeEach(func() {
	logger = lagertest.NewTestLogger("test")
	ctx = context.Background()
	fakeServiceClient = new(serviceclientfakes.FakeServiceClient)
	fakeRepClientFactory = new(modelsfakes.FakeRepClientFactory)
	fakeRepClient = new(modelsfakes.FakeRepClient)
	fakeRepClientFactory.CreateClientReturns(fakeRepClient, nil)
})
