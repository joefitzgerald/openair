package generator

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestOpenairgenerator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "OpenAir Generator Suite")
}
