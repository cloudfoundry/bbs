package converger_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestConverger(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Converger Process Suite")
}
