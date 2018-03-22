package bumper_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestBumper(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Bumper Suite")
}
