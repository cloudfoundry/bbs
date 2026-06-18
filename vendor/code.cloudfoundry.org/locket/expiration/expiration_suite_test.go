package expiration_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestExpiration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Expiration Suite")
}
