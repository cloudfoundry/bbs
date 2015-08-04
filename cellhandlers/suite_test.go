package cellhandlers_test

import (
	"time"

	"github.com/cloudfoundry-incubator/cf_http"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

var cfHttpTimeout time.Duration

var _ = BeforeSuite(func() {
	cfHttpTimeout = 1 * time.Second
	cf_http.Initialize(cfHttpTimeout)
})

func TestCellHandlers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cell Handlers Suite")
}
