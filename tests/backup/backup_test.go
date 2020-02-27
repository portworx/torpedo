package tests

import (
	"testing"
	"fmt"
	"os"

	api "github.com/portworx/px-backup-api/pkg/apis/v1"
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
	fmt.Printf("calling init instance\n")
	InitInstance()
})

// This test performs basic test of starting an application and destroying it (along with storage)
var _ = Describe("{BackupSetup}", func() {
	It("has to connect and check the backup setup", func() {
		fmt.Printf("Create an org\n")
		CreateOrg("sample-org")
	})
})

func CreateOrg(orgName string){
	fmt.Printf("creating sample org\n")
	backupDriver := Inst().Backup
	fmt.Printf("backupDriver = %v\n", backupDriver)
	req := &api.OrganizationCreateRequest{
		CreateMetadata: &api.CreateMetadata{
			Name: orgName,
		},
	}
	_, err := backupDriver.CreateOrganization(req)
	Expect(err).NotTo(HaveOccurred())
}

var _ = AfterSuite(func() {
	//PerformSystemCheck()
	//ValidateCleanup()
})

func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags
	ParseFlags()
	os.Exit(m.Run())
}
