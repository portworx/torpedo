package tests

import (
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	. "github.com/portworx/torpedo/tests"
)

func TestBackup(t *testing.T) {
	RegisterFailHandler(Fail)

	var specReporters []Reporter
	junitReporter := reporters.NewJUnitReporter("/testresults/junit_basic.xml")
	specReporters = append(specReporters, junitReporter)
	RunSpecsWithDefaultAndCustomReporters(t, "Torpedo : Backup", specReporters)
}

var _ = BeforeSuite(func() {
	fmt.Printf("In BeforeSuite")
	InitInstance()
})

// This test performs basic test of starting an application and destroying it (along with storage)
var _ = Describe("{BackupSetup}", func() {
	fmt.Printf("In TestBackup")
	It("has to connect and check the backup setup", func() {
	})
})

var _ = AfterSuite(func() {
	fmt.Printf("In AfterSuite")
	//PerformSystemCheck()
	//ValidateCleanup()
})

func init() {
	fmt.Printf("In test init")
	ParseFlags()
}
