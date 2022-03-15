package tests

import (
	"context"
	"net/url"
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	. "github.com/portworx/torpedo/pkg/pdsutils/api"
	. "github.com/portworx/torpedo/pkg/pdsutils/lib"
	. "github.com/portworx/torpedo/tests"
	"github.com/sirupsen/logrus"
)

func TestPDS(t *testing.T) {
	RegisterFailHandler(Fail)

	var specReporters []Reporter
	junitReporter := reporters.NewJUnitReporter("/testresults/junit_basic.xml")
	specReporters = append(specReporters, junitReporter)
	RunSpecsWithDefaultAndCustomReporters(t, "Torpedo : Basic", specReporters)
}

var (
	components *Components
)

var _ = BeforeSuite(func() {
	logrus.Info("Check for environmental variable.")
	env := MustHaveEnvVariables()
	logrus.Info("Get control plane.")
	controlPlane := NewControlPlane(env.ControlPlaneKubeconfig)

	logrus.Info("Test control plane url connectivity.")
	Expect(controlPlane.IsReachbale(env.ControlPlaneURL)).To(BeTrue())

	// Improvement: Remove hardcoded token, keeping it for a while for sample automation run.
	pdsClusterSaToken := "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsImp0aSI6IjRmNDg3MDBhYzgyNzFiMjVhZDJhZjBjYmE4ZmE4MzI0OWVhYjQwNTQ0OTcxNzVjNTQ1MzlhN2U3MzRmODA2NGI3NDliNDJiZjZkOTBmMThhIn0.eyJhdWQiOiIzIiwianRpIjoiNGY0ODcwMGFjODI3MWIyNWFkMmFmMGNiYThmYTgzMjQ5ZWFiNDA1NDQ5NzE3NWM1NDUzOWE3ZTczNGY4MDY0Yjc0OWI0MmJmNmQ5MGYxOGEiLCJpYXQiOjE2NDM5MTk2MTQsIm5iZiI6MTY0MzkxOTYxNCwiZXhwIjoxNjc1NDU1NjE0LCJzdWIiOiI0NTI2IiwiaXNzIjoiaHR0cHM6XC9cL3JlbGVhc2Utc3RhZ2luZy1hcGkucG9ydHdvcnguZGV2XC9hcGkiLCJuYW1lIjoiIiwiZW1haWwiOiJ6ZnhuaHFlY2ZzdGthb3ZjbGtAYWRmc2tqLmNvbSIsInNjb3BlcyI6W10sInJvbGVzIjpbInB4LWJhY2t1cC1pbmZyYS5hZG1pbiJdfQ.GcVyTtuXzbppP82tMDsX1ZLkGFpC05yf_oqwzYuQpC64i7aePvCDXUNU8dIPY3e73p8WOFN8P7ZqxtSYns4fdvw0NfqhqKgSi5kIzop4tHzjQbU2w5Nd0MY6egenrzpF8FBSqxfdaGMb8OqdLxA69g6I0OmIrCcF5zu2b0veXDn2a_jNNTcKV8lRb0PxJPhNBo54ajIBcTbBnnw9K7MYyq0ILDVc5AlTadtr5VtsMc8sAz8pnhx3LAvgnWEEi6yYHMdEB_lPc-4lvrs32aRrci6NrJhzZzIaA8fg3Mf2PYgtfjrahBE_jWrHDrNUzbMs6KBsDyxDjmLStyszFJWCNvH15whsD2y4fUQZzvwDyizIZhcyUo9qZr1A8YHGY_xwLU_WdnRpworosObFMRyy5r4jRYCH3kqOr8HDSnaA6ZPrCNFSBUjzHDL-TbPUYW0Qy44KVIVYuJ0e5IW92H6vKPkpPD43Iubw71iZtqiYVRVBaCo8aqonaVBz09MU8EYzZhboJDmQ9QqVbB7rUgM3JLpZH0sq3buuiPusubQnO2K0jVbX3Wfj2G7ZY97yrnKNO0moP82r4Ado5TnNOnKgbceTXiPFJTqQuH7e-vqLSVvrsQTTH8CNaiNWR7P9_bVo6Ep_qaaCT23IIY4Fvdp7KMJ5kpJrWkT069MvsqVFFTE"
	endpointURL, _ := url.Parse(env.ControlPlaneURL)
	apiConf := pds.NewConfiguration()
	apiConf.Host = endpointURL.Host
	apiConf.Scheme = endpointURL.Scheme

	// Use Configuration or context with WithValue (above)
	context := context.WithValue(context.Background(), pds.ContextAPIKeys, map[string]pds.APIKey{"ApiKeyAuth": {Key: pdsClusterSaToken, Prefix: "Bearer"}})
	apiClient := pds.NewAPIClient(apiConf)
	components = NewComponents(context, apiClient)
})

var _ = AfterSuite(func() {
	// PerformSystemCheck()
	// ValidateCleanup()
})

func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags
	ParseFlags()
	os.Exit(m.Run())
}

// This test performs basic PDS test and just list accounts/tenants/projects.
var _ = Describe("{PDSTestListAccountsTenantsProjects}", func() {
	It("has to list accounts, tenants and projects.", func() {
		Step("Get the list of accounts and verify ", func() {
			acc := components.Account
			accounts, _ := acc.GetAccountsList()
			accountID := accounts[0].GetId()
			accountName := accounts[0].GetName()
			logrus.Infof("Account Detail- Name: %s, UUID: %s ", accountName, accountID)
			Expect(accountName).To(Equal("Portworx"))
		})

		Step("Get the list of tenants to belong to the account Portworx.", func() {
			acc := components.Account
			accounts, _ := acc.GetAccountsList()
			accountID := accounts[0].GetId()
			tnts := components.Tenant
			tenants, _ := tnts.GetTenantsList(accountID)
			tenantID := tenants[0].GetId()
			tenantName := tenants[0].GetName()
			logrus.Infof("Tenant Details- Name: %s, UUID: %s ", tenantName, tenantID)
			Expect(tenantName).To(Equal("Default"))
		})

		Step("Get the list of projects to belong to the detault tenant.", func() {
			acc := components.Account
			accounts, _ := acc.GetAccountsList()
			accountID := accounts[0].GetId()
			tnts := components.Tenant
			tenants, _ := tnts.GetTenantsList(accountID)
			tenantID := tenants[0].GetId()
			pjts := components.Project
			projects, _ := pjts.GetprojectsList(tenantID)
			projectUUID := projects[0].GetId()
			projectName := projects[0].GetName()
			logrus.Infof("Project Details- Name: %s, UUID: %s ", projectName, projectUUID)
			Expect(projectName).To(Equal("Default"))
		})
	})
})
