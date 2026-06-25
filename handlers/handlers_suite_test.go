package handlers_test

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	modelsfakes "code.cloudfoundry.org/bbs/models/fakes"
	"code.cloudfoundry.org/bbs/serviceclient/serviceclientfakes"
	"github.com/gogo/protobuf/proto"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestHandlers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Handlers Suite")
}

var (
	fakeServiceClient    *serviceclientfakes.FakeServiceClient
	fakeRepClient        *modelsfakes.FakeRepClient
	fakeRepClientFactory *modelsfakes.FakeRepClientFactory
)

var _ = BeforeEach(func() {
	fakeServiceClient = new(serviceclientfakes.FakeServiceClient)
	fakeRepClientFactory = new(modelsfakes.FakeRepClientFactory)
	fakeRepClient = new(modelsfakes.FakeRepClient)
	fakeRepClientFactory.CreateClientReturns(fakeRepClient, nil)
})

func newTestRequest(body interface{}) *http.Request {
	var reader io.Reader
	switch body := body.(type) {
	case io.Reader:
		reader = body
	case string:
		reader = strings.NewReader(body)
	case []byte:
		reader = bytes.NewReader(body)
	case proto.Message:
		protoBytes, err := proto.Marshal(body)
		Expect(err).NotTo(HaveOccurred())
		reader = bytes.NewReader(protoBytes)
	default:
		panic("cannot create test request")
	}

	request, err := http.NewRequest("", "", reader)
	Expect(err).NotTo(HaveOccurred())
	return request
}
