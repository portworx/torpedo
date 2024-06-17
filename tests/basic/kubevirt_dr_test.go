package tests

import (
	//"context"
	"fmt"
	"time"

	storkapi "github.com/libopenstorage/stork/pkg/apis/stork/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	storkops "github.com/portworx/sched-ops/k8s/stork"

	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/pkg/log"

	//"github.com/portworx/torpedo/driver	"github.com/portworx/torpedo/drivers/scheduler"
	//"github.com/portworx/torpedo/drivers/scheduler/spec"
	"github.com/portworx/torpedo/pkg/testrailuttils"
	. "github.com/portworx/torpedo/tests"
)

// This test does the following:
// Deploy Kubevirt VM on source cluster, validate it is running
// Create cluster pair with destination cluster and create a migration schedule
// Validate few migrations run successfully
// Failover to destination cluster and validate Kubevirt VM on destination cluster and Live VM migrate it
// Failback to source cluster and validate Kubevirt VM on source cluster
var _ = Describe("{AsyncDRKubevirtVMs}", func() {
	testrailID = 79656
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/79656
	BeforeEach(func() {
		if !kubeConfigWritten {
			// Write kubeconfig files after reading from the config maps created by torpedo deploy script
			WriteKubeconfigToFiles()
			kubeConfigWritten = true
		}
		wantAllAfterSuiteActions = false
	})
	JustBeforeEach(func() {
		StartTorpedoTest("AsyncDRKubevirtVMs", "AsyncDR failover/failback of Kubevirt VMs to destination cluster", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})
	var (
		contexts              []*scheduler.Context
		migrationNamespaces   []string
		taskNamePrefix        = "async-dr-kubevirt"
		allMigrations         []*storkapi.Migration
		includeResourcesFlag  = true
		startApplicationsFlag = false
	)

	It("has to deploy kubevirt VM, create cluster pair, migrate kubevirt VM to destination cluster", func() {
		Step("Deploy applications", func() {

			err := SetSourceKubeConfig()
			log.FailOnError(err, "Switching context to source cluster failed")
			// Schedule applications
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				taskName := fmt.Sprintf("%s-%d", taskNamePrefix, i)
				appContexts := ScheduleApplications(taskName)
				contexts = append(contexts, appContexts...)
				ValidateApplications(contexts)
				for _, ctx := range appContexts {
					// Override default App readiness time out of 5 mins with 10 mins
					ctx.ReadinessTimeout = kubevirtReadinessTimeout
					namespace := GetAppNamespace(ctx, taskName)
					migrationNamespaces = append(migrationNamespaces, namespace)
				}
				Step("Create cluster pair between source and destination clusters", func() {
					// Set cluster context to cluster where torpedo is running
					ScheduleValidateClusterPair(appContexts[0], false, true, defaultClusterPairDir, false)
				})
			}

			log.Infof("Migration Namespaces: %v", migrationNamespaces)
		})

		time.Sleep(5 * time.Minute)
		log.Info("Start migration")

		for i, currMigNamespace := range migrationNamespaces {
			migrationName := migrationKey + fmt.Sprintf("%d", i)
			currMig, err := CreateMigration(migrationName, currMigNamespace, defaultClusterPairName, currMigNamespace, &includeResourcesFlag, &startApplicationsFlag)
			Expect(err).NotTo(HaveOccurred(),
				fmt.Sprintf("failed to create migration: %s in namespace %s. Error: [%v]",
					migrationKey, currMigNamespace, err))
			allMigrations = append(allMigrations, currMig)
		}

		for _, mig := range allMigrations {
			err := storkops.Instance().ValidateMigration(mig.Name, mig.Namespace, migrationRetryTimeout, migrationRetryInterval)
			Expect(err).NotTo(HaveOccurred(),
				fmt.Sprintf("failed to validate migration: %s in namespace %s. Error: [%v]",
					mig.Name, mig.Namespace, err))
		}

		log.InfoD("Start volume only migration")
		includeResourcesFlag = false
		for i, currMigNamespace := range migrationNamespaces {
			migrationName := migrationKey + "volumeonly-" + fmt.Sprintf("%d", i)
			currMig, createMigErr := CreateMigration(migrationName, currMigNamespace, defaultClusterPairName, currMigNamespace, &includeResourcesFlag, &startApplicationsFlag)
			allMigrations = append(allMigrations, currMig)
			log.FailOnError(createMigErr, "Failed to create %s migration in %s namespace", migrationName, currMigNamespace)
			err := storkops.Instance().ValidateMigration(currMig.Name, currMig.Namespace, migrationRetryTimeout, migrationRetryInterval)
			dash.VerifyFatal(err, nil, "Migration successful?")
			resp, getMigErr := storkops.Instance().GetMigration(currMig.Name, currMig.Namespace)
			dash.VerifyFatal(getMigErr, nil, "Received migration response?")
			dash.VerifyFatal(resp.Status.Summary.NumberOfMigratedResources == 0, true, "Validate no resources migrated")
		}

		Step("teardown all applications on source cluster before switching context to destination cluster", func() {
			for _, ctx := range contexts {
				TearDownContext(ctx, map[string]bool{
					SkipClusterScopedObjects:                    true,
					scheduler.OptionsWaitForResourceLeakCleanup: true,
					scheduler.OptionsWaitForDestroy:             true,
				})
			}
		})

		Step("teardown migrations", func() {
			for _, mig := range allMigrations {
				err := DeleteAndWaitForMigrationDeletion(mig.Name, mig.Namespace)
				Expect(err).NotTo(HaveOccurred(),
					fmt.Sprintf("failed to delete migration: %s in namespace %s. Error: [%v]",
						mig.Name, mig.Namespace, err))
			}
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})
