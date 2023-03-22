package encryptor_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestEncryptor(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Encryptor Suite")
}
