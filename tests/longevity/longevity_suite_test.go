package tests

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/portworx/torpedo/tests"
)

func TestLongevity(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Torpedo : Longevity")
}

func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags
	ParseFlags()
	os.Exit(m.Run())
}
