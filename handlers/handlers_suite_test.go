package handlers_test

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"github.com/gogo/protobuf/proto"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestHandlers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Handlers Suite")
}

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

type explodingReader struct {
	ReadError error
}

func newExplodingReader(readErr error) explodingReader {
	return explodingReader{readErr}
}

func (rc explodingReader) Read([]byte) (int, error) {
	return 0, rc.ReadError
}
