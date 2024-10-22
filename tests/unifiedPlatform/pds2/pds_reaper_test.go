package tests

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	"github.com/pure-px/torpedo/pkg/log"
	. "github.com/pure-px/torpedo/tests"
	. "github.com/pure-px/torpedo/tests/unifiedPlatform"
	"strings"
)

var _ = Describe("{DeleteAllUnhealthyTargetClusters}", func() {

	JustBeforeEach(func() {
		StartTorpedoTest("DeleteAllUnhealthyTargetClusters", "Deletes All Unhealthy Target Cluster's", nil, 0)
	})

	It("Delete Unhealthy Target Cluster", func() {
		steplog := "Delete Unhealthy TargetClusters"
		Step(steplog, func() {
			log.InfoD(steplog)
			for _, tcPrefix := range NewPdsParams.CleanUpParams.TargetClusterPrefix {
				allErrors := WorkflowRepear.TargetCluster.DeleteAllUnhealthyTargetClustersWithPrefix(WorkflowPlatform.TenantId, tcPrefix, true)
				if len(allErrors) > 0 {
					var allErrorStrings []string
					for _, err := range allErrors {
						allErrorStrings = append(allErrorStrings, err.Error())
					}
					log.FailOnError(fmt.Errorf("[%s]", strings.Join(allErrorStrings, "\n\n")), "errors occurred while cleanup of target clusters")
				}
			}
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})

var _ = Describe("{DeleteBackUpLocationAndCloudCredentials}", func() {

	JustBeforeEach(func() {
		StartTorpedoTest("DeleteBackUpLocationAndCloudCredentials", "Delete's backup location and cloud credentials with the given prefix", nil, 0)
	})

	It("Delete Backup Location", func() {
		var allErrorStrings []string
		steplog := "Delete Backup Location with the given prefix"
		Step(steplog, func() {
			log.InfoD(steplog)
			for _, blPrefix := range NewPdsParams.CleanUpParams.BackupLocationPrefix {
				allErrors := WorkflowRepear.BackUpLocation.DeleteAllBackUpLocationWithPrefix(WorkflowPlatform.TenantId, blPrefix)
				if len(allErrors) > 0 {

					for _, err := range allErrors {
						allErrorStrings = append(allErrorStrings, err.Error())
					}
				}
			}
		})
		steplog = "Delete Cloud Credential with the given prefix"
		Step(steplog, func() {
			log.InfoD(steplog)
			for _, ccPrefix := range NewPdsParams.CleanUpParams.CloudCredentialPrefix {
				allErrors := WorkflowRepear.CloudCredential.DeleteAllCloudCredentialWithPrefix(WorkflowPlatform.TenantId, ccPrefix)
				if len(allErrors) > 0 {
					for _, err := range allErrors {
						allErrorStrings = append(allErrorStrings, err.Error())
					}
				}
			}
		})
		log.FailOnError(fmt.Errorf("[%s]", strings.Join(allErrorStrings, "\n\n")), "errors occurred while cleanup of BackupLocations and CloudCredentials")
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})
