package di_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestDi(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Di Suite")
}
