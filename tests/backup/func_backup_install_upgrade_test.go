package tests

import (
	. "github.com/onsi/ginkgo/v2"
	"github.com/pure-px/torpedo/drivers/scheduler"
	"github.com/pure-px/torpedo/pkg/log"
	. "github.com/pure-px/torpedo/tests"
	"strings"
)

// Test case to check resource quota for namespace
var _ = Describe("{InstallPxBackupResourceQuotaCheck}", Label(TestCaseLabelsMap[InstallPxBackupResourceQuotaCheck]...), func() {
	var (
		namespace string
	)
	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("InstallPxBackupResourceQuotaCheck", "Installing Px Backup and checking if resource quota satisfies the min requirement", nil, 304868, Mkoppal, Q4FY25)
		log.InfoD("Uninstalling Px-Backup...")
		err := UninstallPxBackup()
		log.FailOnError(err, "Failed to delete px-backup")
		namespace = "px-backup"
	})
	It("Should check resource quota for namespace", func() {

		Step("Create namespace with insufficient resource quota", func() {
			log.InfoD("Create namespace with insufficient resource quota")
			err := CreateNamespaceAndResourceQuota(namespace, "300Gi")
			log.FailOnError(err, "Failed to create namespace and resource quota")
		})

		Step("Install Px-Backup and validate the failure message", func() {
			log.InfoD("Install Px-Backup and validate the failure message")
			valuesMap, err := ParseValuesFromFile("values1")
			log.FailOnError(err, "Failed to parse values file")
			_, err = InstallPxBackup(LatestPxBackupVersion, DefaultPxBackupHelmBranch, namespace, valuesMap)
			log.InfoD("Install Px-Backup error - %s", err.Error())
			dash.VerifyFatal(strings.Contains(err.Error(), "px-backup chart installation failed"), true, "Failed to install px-backup")
		})

		Step("Validate the pre-install job pods logs", func() {
			searchString := "Namespace px-backup does not meet the storage quota requirements for Px-Backup installation"
			present, err := SearchPodLogs(map[string]string{"job-name": "pre-install-check"}, namespace, searchString)
			log.FailOnError(err, "Failed to search pod logs")
			dash.VerifyFatal(present, true, "Found the expected string in the pod logs")
		})

		Step("Delete namespace and resource quota", func() {
			log.InfoD("Delete namespace and resource quota")
			err := DeleteAppNamespace(namespace)
			log.FailOnError(err, "Failed to delete namespace and resource quota")
		})

		Step("Create namespace with sufficient resource quota", func() {
			log.InfoD("Create namespace with sufficient resource quota")
			err := CreateNamespaceAndResourceQuota(namespace, "600Gi")
			log.FailOnError(err, "Failed to create namespace and resource quota")
		})

		Step("Install Px-Backup and validate the successful installation", func() {
			log.InfoD("Install Px-Backup and validate the successful installation")
			valuesMap, err := ParseValuesFromFile("values1")
			log.FailOnError(err, "Failed to parse values file")
			_, err = InstallPxBackup(LatestPxBackupVersion, DefaultPxBackupHelmBranch, namespace, valuesMap)
			log.FailOnError(err, "Failed to install px-backup")
		})
	})

	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(make([]*scheduler.Context, 0))
		log.Infof("No cleanup required for this testcase")
	})
})
