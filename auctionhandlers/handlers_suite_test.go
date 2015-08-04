package auctionhandlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

var ErrBadRead = errors.New("bad read!")

func TestAuctionHandlers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Auction Handlers Suite")
}

func newTestRequest(body interface{}) *http.Request {
	var reader io.Reader
	switch body := body.(type) {
	case string:
		reader = strings.NewReader(body)
	case []byte:
		reader = bytes.NewReader(body)
	default:
		jsonBytes, err := json.Marshal(body)
		Expect(err).NotTo(HaveOccurred())
		reader = bytes.NewReader(jsonBytes)
	}

	request, err := http.NewRequest("", "", reader)
	Expect(err).NotTo(HaveOccurred())
	return request
}

type badReader struct{}

func (_ badReader) Read(_ []byte) (int, error) {
	return 0, ErrBadRead
}

func (_ badReader) Close() error {
	return nil
}
