package middleware_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestMiddleware(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Middleware Suite")
}
