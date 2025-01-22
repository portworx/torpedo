package tests

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	"github.com/portworx/sched-ops/k8s/core"
	"github.com/pure-px/torpedo/drivers/scheduler"
	"github.com/pure-px/torpedo/pkg/log"
	. "github.com/pure-px/torpedo/tests"
	"helm.sh/helm/v3/pkg/release"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
			if err != nil {
				log.InfoD("Install Px-Backup error - %s", err.Error())
				dash.VerifyFatal(strings.Contains(err.Error(), "px-backup chart installation failed"), true, "Failed to install px-backup")
			} else {
				log.FailOnError(fmt.Errorf("installation was successful when it should not have been"),
					"Installation was successful when it should not have been")
			}
		})

		Step("Validate the pre-install job pods logs", func() {
			log.InfoD("Validate the pre-install job pods logs")
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

// Test case to check if there is a string in the helm notes
var _ = Describe("{InstallPxBackupHelmNotesCheck}", Label(TestCaseLabelsMap[InstallPxBackupHelmNotesCheck]...), func() {
	var (
		namespace string
	)
	rel := &release.Release{}
	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("InstallPxBackupHelmNotesCheck",
			"Installing Px Backup and checking if helm notes contains the desired string", nil, 304870, Mkoppal, Q4FY25)
		log.InfoD("Uninstalling Px-Backup...")
		err := UninstallPxBackup()
		log.FailOnError(err, "Failed to delete px-backup")
		namespace = "px-backup"
	})
	It("Should check if there is a string in the helm notes", func() {

		Step("Install Px-Backup and validate the helm notes", func() {
			log.InfoD("Install Px-Backup and validate the helm notes")
			valuesMap, err := ParseValuesFromFile("values1")
			log.FailOnError(err, "Failed to parse values file")
			rel, err = InstallPxBackup(LatestPxBackupVersion, DefaultPxBackupHelmBranch, namespace, valuesMap)
			log.FailOnError(err, "Failed to install px-backup")
			log.InfoD("Helm notes - \n%s", rel.Info.Notes)
		})

		Step("Validate the helm notes", func() {
			searchString := "For more information on network pre-requisites: " +
				"https://docs.portworx.com/portworx-backup-on-prem/install/install-prereq/nw-prereqs.html"
			dash.VerifyFatal(strings.Contains(rel.Info.Notes, searchString), true,
				"Found the expected string in the helm notes")
		})
	})
	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(make([]*scheduler.Context, 0))
		log.Infof("No cleanup required for this testcase")
	})
})

// Test case to verify that the installation fails with a PVC in pre-existing in the namespace where the installation is being attempted
var _ = Describe("{InstallPxBackupPvcCheck}", Label(TestCaseLabelsMap[InstallPxBackupPvcCheck]...), func() {

	var (
		namespace string
	)
	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("InstallPxBackupPvcCheck",
			"Installing Px Backup and checking if installation fails when a PVC is pre-existing in the namespace", nil, 304869, Mkoppal, Q4FY25)
		log.InfoD("Uninstalling Px-Backup...")
		err := UninstallPxBackup()
		log.FailOnError(err, "Failed to delete px-backup")
		namespace = "px-backup"
	})
	It("Should check if installation fails when a PVC is pre-existing in the namespace", func() {

		Step("Create namespace and PVC", func() {
			log.InfoD("Create namespace and PVC")
			// Create the Namespace object
			ns := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: namespace,
				},
			}
			// Create Namespace
			_, err := core.Instance().CreateNamespace(ns)
			log.FailOnError(err, "Failed to create namespace")
			sc, err := GetDefaultStorageClass()
			log.Infof("Using storage class - %s", sc.Name)

			// Get a random px-backup associated PVC
			pxbPvc, err := GetRandomSubset(PxBackupPVCs, 1)
			log.FailOnError(err, "Failed to get PVC")
			if len(pxbPvc) == 0 {
				log.FailOnError(fmt.Errorf("no PVCs found"), "No PVCs found")
			}

			// Create PVC
			log.Infof("Creating PVC - %s", pxbPvc[0])
			pvc := &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      pxbPvc[0],
					Namespace: namespace,
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					StorageClassName: &sc.Name,
					AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceStorage: resource.MustParse("1Gi"),
						},
					},
				},
			}
			_, err = core.Instance().CreatePersistentVolumeClaim(pvc)
			log.FailOnError(err, "Failed to create PVC")
		})

		Step("Install Px-Backup and validate the failure message", func() {
			log.InfoD("Install Px-Backup and validate the failure message")
			valuesMap, err := ParseValuesFromFile("values1")
			log.FailOnError(err, "Failed to parse values file")
			_, err = InstallPxBackup(LatestPxBackupVersion, DefaultPxBackupHelmBranch, namespace, valuesMap)
			if err != nil {
				log.InfoD("Install Px-Backup error - %s", err.Error())
				dash.VerifyFatal(strings.Contains(err.Error(), "px-backup chart installation failed"),
					true, "Failed to install px-backup")
			} else {
				log.FailOnError(fmt.Errorf("installation was successful when it should not have been"),
					"Installation was successful when it should not have been")
			}
		})

		Step("Validate the pre-install job pods logs", func() {
			log.InfoD("Validate the pre-install job pods logs")
			searchString := fmt.Sprintf("Namespace contains %d stale PX-Backup PVC(s).", 1)
			present, err := SearchPodLogs(map[string]string{"job-name": "pre-install-check"}, namespace, searchString)
			log.FailOnError(err, "Failed to search pod logs")
			dash.VerifyFatal(present, true, "Found the expected string in the pod logs")
		})

	})
	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(make([]*scheduler.Context, 0))
		log.Infof("No cleanup required for this testcase")
	})
})

// Test case to check port info in the release notes for upgrade
var _ = Describe("{UpgradePxBackupPortInfoCheck}", Label(TestCaseLabelsMap[UpgradePxBackupPortInfoCheck]...), func() {
	var (
		namespace string
	)
	rel := &release.Release{}
	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("UpgradePxBackupPortInfoCheck", "Upgrading Px-Backup and checking if desired port information is displayed in the helm release notes", nil, 304871, SS, Q4FY25)
		log.InfoD("Uninstalling Px-Backup...")
		err := UninstallPxBackup()
		log.FailOnError(err, "Failed to delete px-backup")
		namespace = "px-backup"

		log.InfoD("Installing Px-Backup...")
		valuesMap, err := ParseValuesFromFile("values1")
		log.FailOnError(err, "Failed to parse values file")
		rel, err = InstallPxBackup("2.8.0", "master", namespace, valuesMap)
		log.FailOnError(err, "Failed to install px-backup")
	})
	It("Should check port info in the release notes for upgrade", func() {

		Step("Upgrade Px-Backup and validate the helm notes", func() {
			log.InfoD("Upgrade Px-Backup and validate the helm notes")
			valuesMap, err := ParseValuesFromFile("values1")
			log.FailOnError(err, "Failed to parse values file")
			rel, err = HelmUpgradePxBackup(LatestPxBackupVersion, DefaultPxBackupHelmBranch, namespace, valuesMap)
			log.FailOnError(err, "Failed to upgrade px-backup")
			log.InfoD("Helm notes - \n%s", rel.Info.Notes)
		})

		Step("Validate the helm notes", func() {
			searchString := "For more information on network pre-requisites: " +
				"https://docs.portworx.com/portworx-backup-on-prem/install/install-prereq/nw-prereqs.html"
			dash.VerifyFatal(strings.Contains(rel.Info.Notes, searchString), true,
				"Found the expected string in the helm notes")
		})
	})

	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(make([]*scheduler.Context, 0))
		log.Infof("No cleanup required for this testcase")
	})
})
