package tests

import (
	//"context"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/portworx/sched-ops/k8s/apps"
	"github.com/portworx/sched-ops/k8s/core"
	"github.com/portworx/sched-ops/task"
	storkapi "github.com/pure-px/stork/pkg/apis/stork/v1alpha1"
	"github.com/pure-px/stork/pkg/crud/stork"
	storkops "github.com/pure-px/stork/pkg/crud/stork"
	"github.com/pure-px/stork/pkg/k8sutils"
	v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/pure-px/torpedo/drivers/node"
	"github.com/pure-px/torpedo/drivers/scheduler"
	"github.com/pure-px/torpedo/pkg/applicationbackup"
	"github.com/pure-px/torpedo/pkg/asyncdr"
	"github.com/pure-px/torpedo/pkg/log"
	"github.com/pure-px/torpedo/pkg/osutils"
	"github.com/pure-px/torpedo/pkg/storkctlcli"

	//"github.com/pure-px/torpedo/driver	"github.com/pure-px/torpedo/drivers/scheduler"
	//"github.com/pure-px/torpedo/drivers/scheduler/spec"
	"github.com/pure-px/torpedo/pkg/testrailuttils"
	. "github.com/pure-px/torpedo/tests"
)

const (
	migrationRetryTimeout     = 30 * time.Minute
	migrationRetryInterval    = 10 * time.Second
	domainCheckRetryTimeout   = 1 * time.Minute
	defaultClusterPairDir     = "cluster-pair"
	defaultClusterPairDirNew  = "cluster-pair-new"
	defaultClusterPairName    = "remoteclusterpair"
	defaultClusterPairNameNew = "remoteclusterpairnew"
	azureSecret               = "azuresecret"
	azureBackupLocation       = "azure"
	googleSecret              = "googlesecret"
	googleBackupLocation      = "google"
	defaultMigSchedName       = "automation-migration-schedule-"
	migrationKey              = "async-dr-"
	migrationSchedKey         = "mig-sched-"
	metromigrationKey         = "metro-dr-"
	clusterwideNs             = "openshift-operators"
)

var (
	kubeConfigWritten     bool
	defaultBackupLocation = "s3"
	nfsBackupLocation     = "nfs"
	defaultSecret         = "s3secret"
	nfsSecret             = "nfssecret"
)

type failoverFailbackParam struct {
	action                    string
	failoverOrFailbackNs      string
	migrationSchedName        string
	configPath                string
	single                    bool
	skipSourceOp              bool
	includeNs                 bool
	excludeNs                 bool
	extraArgsFailoverFailback map[string]string
	contexts                  []*scheduler.Context
}

// This test performs basic test of starting an application, creating cluster pair,
// and migrating application to the destination clsuter
var _ = Describe("{MigrateDeployment}", Label("p0", "positive", "AsyncDR"), func() {
	testrailID = 50803
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/50803
	BeforeEach(func() {
		if !kubeConfigWritten {
			// Write kubeconfig files after reading from the config maps created by torpedo deploy script
			WriteKubeconfigToFiles()
			kubeConfigWritten = true
		}
		wantAllAfterSuiteActions = false
	})
	JustBeforeEach(func() {
		StartTorpedoTest("MigrateDeployment", "Migration of application to destination cluster", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})
	var (
		contexts              []*scheduler.Context
		migrationNamespaces   []string
		taskNamePrefix        = "async-dr-mig"
		allMigrations         []*storkapi.Migration
		includeResourcesFlag  = true
		startApplicationsFlag = false
	)

	It("has to deploy app, create cluster pair, migrate app", func() {
		Step("Deploy applications", func() {

			err := SetSourceKubeConfig()
			log.FailOnError(err, "Switching context to source cluster failed")
			// Schedule applications
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				taskName := fmt.Sprintf("%s-%d", taskNamePrefix, i)
				log.Infof("Task name %s\n", taskName)
				appContexts := ScheduleApplications(taskName)
				contexts = append(contexts, appContexts...)
				ValidateApplications(contexts)
				for _, ctx := range appContexts {
					// Override default App readiness time out of 5 mins with 10 mins
					ctx.ReadinessTimeout = appReadinessTimeout
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

var _ = Describe("{MigrateDeploymentMetroAsync}", Label("p0", "positive", "MetroDR"), func() {
	testrailID = 297595
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/297595
	BeforeEach(func() {
		if !kubeConfigWritten {
			// Write kubeconfig files after reading from the config maps created by torpedo deploy script
			WriteKubeconfigToFiles()
			kubeConfigWritten = true
		}
		wantAllAfterSuiteActions = false
	})
	JustBeforeEach(func() {
		skipFlag := getClusterDomainsInfo()
		if skipFlag {
			Skip("Skip test because cluster domains are not set")
		}
		StartTorpedoTest("MigrateDeploymentMetroAsync", "Migration of application using metro+async combination", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})

	var (
		contexts                []*scheduler.Context
		migrationNamespaces     []string
		taskNamePrefix          = "metro-async-dr-mig"
		allMigrationsMetro      []*storkapi.Migration
		allMigrationsAsync      []*storkapi.Migration
		includeResourcesFlag    = true
		includeVolumesFlagMetro = false
		includeVolumesFlagAsync = true
		startApplicationsFlag   = false
	)

	It("has to deploy app, create cluster pair, migrate app", func() {
		Step("Deploy applications", func() {
			err = SetCustomKubeConfig(asyncdr.FirstCluster)
			log.FailOnError(err, "Switching context to first cluster failed")
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				taskName := fmt.Sprintf("%s-%d", taskNamePrefix, i)
				log.Infof("Task name %s\n", taskName)
				appContexts := ScheduleApplications(taskName)
				contexts = append(contexts, appContexts...)
				ValidateApplications(contexts)
				for _, ctx := range contexts {
					// Override default App readiness time out of 5 mins with 10 mins
					ctx.ReadinessTimeout = appReadinessTimeout
					namespace := GetAppNamespace(ctx, taskName)
					migrationNamespaces = append(migrationNamespaces, namespace)
					log.Infof("Creating clusterpair between first and second cluster")
					err = ScheduleBidirectionalClusterPair(defaultClusterPairName, namespace, "", "", "", "sync-dr", asyncdr.FirstCluster, asyncdr.SecondCluster, nil)
					log.FailOnError(err, "Failed creating bidirectional cluster pair")
				}
			}
			log.Infof("Migration Namespaces: %v", migrationNamespaces)
		})

		log.Infof("Start migration Metro")

		for i, currMigNamespace := range migrationNamespaces {
			migrationName := metromigrationKey + fmt.Sprintf("%d", i) + time.Now().Format("15h03m05s")
			currMig, err := asyncdr.CreateMigration(migrationName, currMigNamespace, defaultClusterPairName, currMigNamespace, &includeVolumesFlagMetro, &includeResourcesFlag, &startApplicationsFlag, nil)
			Expect(err).NotTo(HaveOccurred(),
				fmt.Sprintf("failed to create migration: %s in namespace %s. Error: [%v]",
					migrationKey, currMigNamespace, err))
			allMigrationsMetro = append(allMigrationsMetro, currMig)
		}

		// Validate all migrations
		for _, mig := range allMigrationsMetro {
			err := storkops.Instance().ValidateMigration(mig.Name, mig.Namespace, migrationRetryTimeout, migrationRetryInterval)
			Expect(err).NotTo(HaveOccurred(),
				fmt.Sprintf("failed to validate migration: %s in namespace %s. Error: [%v]",
					mig.Name, mig.Namespace, err))
		}

		err = SetCustomKubeConfig(asyncdr.SecondCluster)
		log.FailOnError(err, "Switching context to second cluster failed")

		log.Infof("Start Async migration from Second DR to Third DR")

		for i, currMigNamespace := range migrationNamespaces {
			log.Infof("Creating clusterpair between second and third cluster")
			ScheduleBidirectionalClusterPair(defaultClusterPairNameNew, currMigNamespace, "", storkapi.BackupLocationType(defaultBackupLocation), defaultSecret, "async-dr", asyncdr.SecondCluster, asyncdr.ThirdCluster, nil)
			migrationName := migrationKey + fmt.Sprintf("%d", i) + time.Now().Format("15h03m05s")
			currMig, err := asyncdr.CreateMigration(migrationName, currMigNamespace, defaultClusterPairNameNew, currMigNamespace, &includeVolumesFlagAsync, &includeResourcesFlag, &startApplicationsFlag, nil)
			Expect(err).NotTo(HaveOccurred(),
				fmt.Sprintf("failed to create migration: %s in namespace %s. Error: [%v]",
					migrationKey, currMigNamespace, err))
			allMigrationsAsync = append(allMigrationsAsync, currMig)
		}

		for _, mig := range allMigrationsAsync {
			err := storkops.Instance().ValidateMigration(mig.Name, mig.Namespace, migrationRetryTimeout, migrationRetryInterval)
			Expect(err).NotTo(HaveOccurred(),
				fmt.Sprintf("failed to validate migration: %s in namespace %s. Error: [%v]",
					mig.Name, mig.Namespace, err))
		}
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

var _ = Describe("{StorkctlPerformFailoverFailbackDefaultAsyncSingle}", Label("p0", "positive", "AsyncDR"), func() {
	testrailID = 296255
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/296255
	BeforeEach(func() {
		if !kubeConfigWritten {
			// Write kubeconfig files after reading from the config maps created by torpedo deploy script
			WriteKubeconfigToFiles()
			kubeConfigWritten = true
		}
		wantAllAfterSuiteActions = false
	})
	JustBeforeEach(func() {
		StartTorpedoTest("StorkctlPerformFailoverFailbackDefaultAsyncSingle", "Failover and Failback using storkctl on async cluster for single NS", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})

	It("has to deploy app, create cluster pair, migrate app and do failover/failback", func() {
		Step("Deploy app, Create cluster pair, Migrate app and Do failover/failback", func() {
			validateFailoverFailback("asyncdr", "asyncdr-failover-failback", true, false, false, false, false)
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

var _ = Describe("{StorkctlPerformFailoverFailbackDefaultAsyncJob}", Label("p0", "positive", "AsyncDR"), func() {
	testrailID = 302499
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/302499
	// run this test only with k8s job resource spec
	BeforeEach(func() {
		if !kubeConfigWritten {
			// Write kubeconfig files after reading from the config maps created by torpedo deploy script
			WriteKubeconfigToFiles()
			kubeConfigWritten = true
		}
		wantAllAfterSuiteActions = false
	})
	JustBeforeEach(func() {
		StartTorpedoTest("StorkctlPerformFailoverFailbackDefaultAsyncJob", "Failover and Failback using storkctl on async cluster for a job resource", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})

	It("has to deploy app, create cluster pair, migrate app and do failover/failback", func() {
		Step("Deploy app, Create cluster pair, Migrate app and Do failover/failback", func() {
			validateFailoverFailback("asyncdr", "asyncdr-failover-failback", true, false, false, false, true)
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

var _ = Describe("{StorkctlPerformFailoverFailbackDefaultAsyncSkipSourceOperations}", Label("p1", "positive", "AsyncDR"), func() {
	testrailID = 296256
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/296256
	BeforeEach(func() {
		if !kubeConfigWritten {
			// Write kubeconfig files after reading from the config maps created by torpedo deploy script
			WriteKubeconfigToFiles()
			kubeConfigWritten = true
		}
		wantAllAfterSuiteActions = false
	})
	JustBeforeEach(func() {
		StartTorpedoTest("StorkctlPerformFailoverFailbackDefaultAsyncSkipSourceOperations", "Failover and Failback using storkctl on async cluster with skip source operations", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})

	It("has to deploy app, create cluster pair, migrate app and do failover/failback", func() {
		Step("Deploy app, Create cluster pair, Migrate app and Do failover/failback", func() {
			validateFailoverFailback("asyncdr", "asyncdr-failover-failback", false, true, false, false, false)
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

var _ = Describe("{StorkctlPerformFailoverFailbackDefaultAsyncIncludeNs}", Label("p1", "positive", "AsyncDR"), func() {
	testrailID = 296368
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/296368
	BeforeEach(func() {
		if !kubeConfigWritten {
			// Write kubeconfig files after reading from the config maps created by torpedo deploy script
			WriteKubeconfigToFiles()
			kubeConfigWritten = true
		}
		wantAllAfterSuiteActions = false
	})
	JustBeforeEach(func() {
		StartTorpedoTest("StorkctlPerformFailoverFailbackDefaultAsyncIncludeNs", "Failover and Failback using storkctl on async cluster with Include Ns", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})

	It("has to deploy app, create cluster pair, migrate app and do failover/failback", func() {
		Step("Deploy app, Create cluster pair, Migrate app and Do failover/failback", func() {
			validateFailoverFailback("asyncdr", "asyncdr-failover-failback", false, false, true, false, false)
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

var _ = Describe("{StorkctlPerformFailoverFailbackDefaultAsyncExcludeNs}", Label("p0", "positive", "AsyncDR"), func() {
	testrailID = 296367
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/296367
	BeforeEach(func() {
		if !kubeConfigWritten {
			// Write kubeconfig files after reading from the config maps created by torpedo deploy script
			WriteKubeconfigToFiles()
			kubeConfigWritten = true
		}
		wantAllAfterSuiteActions = false
	})
	JustBeforeEach(func() {
		StartTorpedoTest("StorkctlPerformFailoverFailbackDefaultAsyncExcludeNs", "Failover and Failback using storkctl on async cluster with Exclude Ns", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})

	It("has to deploy app, create cluster pair, migrate app and do failover/failback", func() {
		Step("Deploy app, Create cluster pair, Migrate app and Do failover/failback", func() {
			validateFailoverFailback("asyncdr", "asyncdr-failover-failback", false, false, false, true, false)
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

var _ = Describe("{StorkctlPerformFailoverFailbackDefaultAsyncMultiple}", Label("p1", "positive", "AsyncDR", "MiniScale"), func() {
	testrailID = 296255
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/296255
	BeforeEach(func() {
		if !kubeConfigWritten {
			// Write kubeconfig files after reading from the config maps created by torpedo deploy script
			WriteKubeconfigToFiles()
			kubeConfigWritten = true
		}
		wantAllAfterSuiteActions = false
	})
	JustBeforeEach(func() {
		StartTorpedoTest("StorkctlPerformFailoverFailbackDefaultAsyncMultiple", "Failover and Failback using storkctl on async cluster for multiple NS", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})

	It("has to deploy app, create cluster pair, migrate app and do failover/failback", func() {
		Step("Deploy app, Create cluster pair, Migrate app and Do failover/failback", func() {
			validateFailoverFailback("asyncdr", "asyncdr-failover-failback", false, false, false, false, false)
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

var _ = Describe("{StorkctlPerformFailoverFailbackDefaultMetroSingle}", Label("p0", "positive", "MetroDR"), func() {
	testrailID = 296291
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/296291
	BeforeEach(func() {
		if !kubeConfigWritten {
			// Write kubeconfig files after reading from the config maps created by torpedo deploy script
			WriteKubeconfigToFiles()
			kubeConfigWritten = true
		}
		wantAllAfterSuiteActions = false
	})
	JustBeforeEach(func() {
		skipFlag := getClusterDomainsInfo()
		if skipFlag {
			Skip("Skip test because cluster domains are not set")
		}
		StartTorpedoTest("StorkctlPerformFailoverFailbackDefaultMetroSingle", "Failover and Failback using storkctl on metro cluster for single NS", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})

	It("has to deploy app, create cluster pair, migrate app and do failover/failback", func() {
		Step("Deploy app, Create cluster pair, Migrate app and Do failover/failback", func() {
			validateFailoverFailback("metrodr", "metrodr-failover-failback", true, false, false, false, false)
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

var _ = Describe("{StorkctlPerformFailoverFailbackDefaultMetroMultiple}", Label("p0", "positive", "MetroDR", "MiniScale"), func() {
	testrailID = 296291
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/296291
	BeforeEach(func() {
		if !kubeConfigWritten {
			// Write kubeconfig files after reading from the config maps created by torpedo deploy script
			WriteKubeconfigToFiles()
			kubeConfigWritten = true
		}
		wantAllAfterSuiteActions = false
	})
	JustBeforeEach(func() {
		skipFlag := getClusterDomainsInfo()
		if skipFlag {
			Skip("Skip test because cluster domains are not set")
		}
		StartTorpedoTest("StorkctlPerformFailoverFailbackDefaultMetroMultiple", "Failover and Failback using storkctl on metro cluster for multiple NS", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})

	It("has to deploy app, create cluster pair, migrate app and do failover/failback", func() {
		Step("Deploy app, Create cluster pair, Migrate app and Do failover/failback", func() {
			validateFailoverFailback("metrodr", "metrodr-failover-failback", false, false, false, false, false)
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

var _ = Describe("{StorkctlPerformFailoverFailbackPostgresql}", Label("p1", "positive", "AsyncDR"), func() {
	testrailID = 296287
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/296287
	BeforeEach(func() {
		if !kubeConfigWritten {
			WriteKubeconfigToFiles()
			kubeConfigWritten = true
		}
		wantAllAfterSuiteActions = false
	})

	JustBeforeEach(func() {
		StartTorpedoTest("StorkctlPerformFailoverFailbackPostgresql", "Failover and Failback using storkctl for postgresql namespaced operator", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})

	var (
		appPath = "/torpedo/deployments/customconfigs/pgo.yaml"
		opPath  = "/torpedo/deployments/customconfigs/pgo-operator.yaml"
		opName  = "pgo"
		crName  = "postgrescluster"
		ns      = "post"
	)

	It("has to deploy app, create cluster pair, migrate app and do failover/failback", func() {
		Step("Deploy app, Create cluster pair, Migrate app and Do failover/failback", func() {

			podList, err := createOperatorBasedApp(appPath, opPath, ns, false)
			log.FailOnError(err, "Failed to create operator based app")
			log.Infof("PodList is: %v", podList)
			podCount := len(podList.Items)
			validateOperatorMigFailover(ns, "asyncdr", opName, crName, podCount, false)
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

var _ = Describe("{StorkctlPerformFailoverFailbackElasticSearch}", Label("p1", "positive", "AsyncDR"), func() {
	testrailID = 296285
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/296285
	BeforeEach(func() {
		if !kubeConfigWritten {
			// Write kubeconfig files after reading from the config maps created by torpedo deploy script
			WriteKubeconfigToFiles()
			kubeConfigWritten = true
		}
		wantAllAfterSuiteActions = false
	})
	JustBeforeEach(func() {
		StartTorpedoTest("StorkctlPerformFailoverFailbackElasticSearch", "Failover and Failback using storkctl for elasticsearch namespaced operator", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})

	var (
		appPath = "/torpedo/deployments/customconfigs/elasticcr.yaml"
		opPath  = "/torpedo/deployments/customconfigs/elastic-op.yaml"
		opName  = "elastic-operator"
		crName  = "elasticsearch"
		ns      = "esop"
	)

	It("has to deploy app, create cluster pair, migrate app and do failover/failback", func() {
		Step("Deploy app, Create cluster pair, Migrate app and Do failover/failback", func() {
			podList, err := createOperatorBasedApp(appPath, opPath, ns, false)
			log.FailOnError(err, "Failed to create operator based app")
			log.Infof("PodList is: %v", podList)
			podCount := len(podList.Items)
			validateOperatorMigFailover(ns, "asyncdr", opName, crName, podCount, false)
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

var _ = Describe("{StorkctlPerformFailoverFailbackPostgresqlClusterwide}", Label("p1", "positive", "AsyncDR"), func() {
	testrailID = 297919
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/297919
	BeforeEach(func() {
		if !kubeConfigWritten {
			// Write kubeconfig files after reading from the config maps created by torpedo deploy script
			WriteKubeconfigToFiles()
			kubeConfigWritten = true
		}
		wantAllAfterSuiteActions = false
	})

	JustBeforeEach(func() {
		StartTorpedoTest("StorkctlPerformFailoverFailbackPostgresqlClusterwide", "Failover and Failback using storkctl for postgresql clusterwide operator", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})

	var (
		appPath = "/torpedo/deployments/customconfigs/pgo.yaml"
		opPath  = "/torpedo/deployments/customconfigs/pgo-operator-clusterwide.yaml"
		opName  = "pgo"
		crName  = "postgrescluster"
		ns      = "pgcw-" + time.Now().Format("15h03m05s")
	)

	It("has to deploy app, create cluster pair, migrate app and do failover/failback", func() {
		Step("Deploy app, Create cluster pair, Migrate app and Do failover/failback", func() {
			podList, err := createOperatorBasedApp(appPath, opPath, ns, true)
			log.FailOnError(err, "Failed to create operator based app")
			log.Infof("PodList is: %v", podList)
			podCount := len(podList.Items)
			validateOperatorMigFailover(ns, "asyncdr", opName, crName, podCount, true)
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

var _ = Describe("{StorkctlPerformFailoverFailbackeckEsClusterwide}", Label("p1", "positive", "AsyncDR"), func() {
	testrailID = 297921
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/297921
	BeforeEach(func() {
		if !kubeConfigWritten {
			// Write kubeconfig files after reading from the config maps created by torpedo deploy script
			WriteKubeconfigToFiles()
			kubeConfigWritten = true
		}
		wantAllAfterSuiteActions = false
	})

	JustBeforeEach(func() {
		StartTorpedoTest("StorkctlPerformFailoverFailbackeckEsClusterwide", "Failover and Failback using storkctl for elasticsearch clusterwide operator", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})

	var (
		appPath = "/torpedo/deployments/customconfigs/elasticcr.yaml"
		opPath  = "/torpedo/deployments/customconfigs/elastic-op-cw.yaml"
		opName  = "elastic-operator"
		crName  = "elasticsearch"
		ns      = "escw-" + time.Now().Format("15h03m05s")
	)

	It("has to deploy app, create cluster pair, migrate app and do failover/failback", func() {
		Step("Deploy app, Create cluster pair, Migrate app and Do failover/failback", func() {
			podList, err := createOperatorBasedApp(appPath, opPath, ns, true)
			log.FailOnError(err, "Failed to create operator based app")
			log.Infof("PodList is: %v", podList)
			podCount := len(podList.Items)
			validateOperatorMigFailover(ns, "asyncdr", opName, crName, podCount, true)
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

var _ = Describe("{FaFbPodEvictionTest}", Label("p1", "positive", "VolumeSnapshot"), func() {
	testrailID = 302503
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/302503
	BeforeEach(func() {
		if !kubeConfigWritten {
			// Write kubeconfig files after reading from the config maps created by torpedo deploy script
			WriteKubeconfigToFiles()
			kubeConfigWritten = true
		}
		wantAllAfterSuiteActions = false
	})
	JustBeforeEach(func() {
		StartTorpedoTest("FaFbPodEvictionTest", "Fada/Fbda pods should not get evicted when px is down on a node", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})

	It("has to deploy FADA/FBDA app and make sure pod should not get evicted when PX is offline on the scheduled node", func() {
		Step("has to deploy FADA/FBDA app and make sure pod should not get evicted when PX is offline on the scheduled node", func() {
			log.Infof("AppList is %v, it should be FA/FB spec for running this test", Inst().AppList)
			var (
				taskNamePrefix = "fafbtest"
			)
			appList := Inst().AppList
			Inst().AppList = []string{"nginx-fa-davol"}
			defer func() {
				Inst().AppList = appList
			}()
			appNs, contexts := initialSetupApps(taskNamePrefix, true, false)
			pods, err := core.Instance().GetPods(appNs[0], nil)
			log.FailOnError(err, fmt.Sprintf("Failed to get pods in [%v] namespace", appNs[0]))
			for _, pod := range pods.Items {
				log.Infof("Pod name: %s, Node name: %s", pod.Name, pod.Spec.NodeName)
				podNode := pod.Spec.NodeName
				nodeObj, err := node.GetNodeByName(podNode)
				log.FailOnError(err, "Failed to get node object")
				err = Inst().V.StopDriver([]node.Node{nodeObj}, false, nil)
				log.FailOnError(err, "Failed to stop px on node")
				log.Infof("Sleeping for 300 seconds to check if pod is evicted")
				time.Sleep(5 * time.Minute)
				podObj, err := core.Instance().GetPodByName(pod.Name, pod.Namespace)
				log.FailOnError(err, "Failed to get pod object")
				podNodeStop := podObj.Spec.NodeName
				dash.VerifyFatal(podNodeStop == podNode, true, "Pod should be running on the same node after stopping px")
				dash.VerifyFatal(podObj.Status.Phase == v1.PodRunning, true, "Pod should be in running state")
				err = Inst().V.StartDriver(nodeObj)
				log.FailOnError(err, "Failed to start px on node")
				err = Inst().V.WaitDriverUpOnNode(nodeObj, Inst().DriverStartTimeout)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Validate volume is driver up on node [%v]", nodeObj.Name))
				podObj, err = core.Instance().GetPodByName(pod.Name, pod.Namespace)
				log.FailOnError(err, "Failed to get pod object")
				podNodeStart := podObj.Spec.NodeName
				dash.VerifyFatal(podNodeStart == podNode, true, "Pod should be running on the same node after starting px")
				dash.VerifyFatal(podObj.Status.Phase == v1.PodRunning, true, "Pod should be in running state")
			}
			for _, ctx := range contexts {
				TearDownContext(ctx, nil)
				ctxNamespace := GetAppNamespace(ctx, "")
				err = asyncdr.WaitForNamespaceDeletion([]string{ctxNamespace})
				log.Errorf("Failed to delete namespaces: %v", err)
			}
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

var _ = Describe("{UpdateVolumeSnapshotSchedule}", Label("p1", "positive", "VolumeSnapshot"), func() {
	testrailID = 302510
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/302510
	BeforeEach(func() {
		if !kubeConfigWritten {
			// Write kubeconfig files after reading from the config maps created by torpedo deploy script
			WriteKubeconfigToFiles()
			kubeConfigWritten = true
		}
		wantAllAfterSuiteActions = false
	})
	JustBeforeEach(func() {
		StartTorpedoTest("UpdateVolumeSnapshotSchedule", "Update PVC name in VolumeSnapshotSchedule", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})

	It("has to create a volumesnapshotschedule and update it to use another PVC", func() {
		Step("has to create a volumesnapshotschedule and update it to use another PVC", func() {
			log.Infof("AppList is %v, it should be FA/FB spec for running this test", Inst().AppList)
			var (
				snapInterval                 = 1
				retain       storkapi.Retain = 1
				scpolName                    = "auto-schedule-policy"
				ns                           = "test-update-pvc"
				schdName                     = "test-schedule"
			)
			_, err := core.Instance().GetNamespace(ns)
			if err == nil {
				err = asyncdr.WaitForNamespaceDeletion([]string{ns})
				if err != nil {
					log.Infof("Failed to delete namespaces: %v", err)
				}
			}
			_, err = asyncdr.CreateSchedulePolicyWithRetain(scpolName, snapInterval, retain)
			log.FailOnError(err, "Failed to create schedule policy")
			for i := 0; i < 2; i++ {
				_, err := createPVC(fmt.Sprintf("pvc-%d", i), "px-csi-db", "10Gi", ns)
				log.FailOnError(err, "Failed to create PVC")
			}
			snapSched, err := asyncdr.CreateSnapshotSchedule(ns, "pvc-0", schdName, scpolName, "local")
			log.FailOnError(err, "Failed to create snapshot schedule")
			time.Sleep(10 * time.Second)
			err = asyncdr.WaitForRetainSnapshotsSuccessful(schdName, ns, int(retain), snapInterval)
			log.FailOnError(err, "Failed to wait for retain snapshots")
			err = storkctlcli.UpdateVolumeSnapshotSchedulePVC(schdName, ns, "pvc-1")
			log.FailOnError(err, "Failed to update snapshot schedule")
			time.Sleep(time.Duration(snapInterval) * time.Minute)
			snapSched, err = storkops.Instance().GetSnapshotSchedule(schdName, ns)
			log.FailOnError(err, "Failed to get snapshot schedule")
			snapSchedPVC := snapSched.Spec.Template.Spec.PersistentVolumeClaimName
			dash.VerifyFatal(snapSchedPVC == "pvc-1", true, "PVC name in snapshot schedule should be updated")
			err = asyncdr.WaitForRetainSnapshotsSuccessful(schdName, ns, int(retain), snapInterval)
			for i := 0; i < 2; i++ {
				err := core.Instance().DeletePersistentVolumeClaim(fmt.Sprintf("pvc-%d", i), ns)
				log.FailOnError(err, "Failed to delete PVC")
			}
			err = asyncdr.WaitForNamespaceDeletion([]string{ns})
			if err != nil {
				log.Errorf("Failed to delete namespaces: %v", err)
			}
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

var _ = Describe("{UpgradeVolumeDriverDuringAppBkpRestore}", Label("p2", "positive", "AsyncDR", "Upgrade"), func() {
	BeforeEach(func() {
		if !kubeConfigWritten {
			// Write kubeconfig files after reading from the config maps created by torpedo deploy script
			WriteKubeconfigToFiles()
			kubeConfigWritten = true
		}
		wantAllAfterSuiteActions = false
	})

	JustBeforeEach(func() {
		upgradeHopsList := make(map[string]string)
		upgradeHopsList["upgradeHops"] = Inst().UpgradeStorageDriverEndpointList
		upgradeHopsList["UpgradeVolumeDriverDuringAppBkpRestore"] = "true"
		StartTorpedoTest("UpgradeVolumeDriverDuringAppBkpRestore", "Validating volume driver upgrade during migration", upgradeHopsList, 0)
		log.InfoD("Volume driver upgrade hops list [%s]", upgradeHopsList)
	})

	It("upgrade volume driver during app backup and restore and ensure everything is running fine", func() {
		log.InfoD("upgrade volume driver during app backup and restore and ensure everything is running fine")

		var (
			nsRest             = "apprest-" + time.Now().Format("15h03m05s")
			restName           = "storkrestore-" + time.Now().Format("15h03m05s")
			s3SecretName       = "s3secret"
			backupLocationName = "storkbackuplocation-" + time.Now().Format("15h03m05s")
			backupName         = "storkbackup-" + time.Now().Format("15h03m05s")
			taskNamePrefix     = "appbkprest-upgradepx"
			defaultNs          = "kube-system"
			timeout            = 10 * time.Minute
		)
		bkpNs, contexts := initialSetupApps(taskNamePrefix, true, false)
		storageNodes := node.GetStorageNodes()

		// AddDrive is added to test to Vsphere Cloud drive upgrades when kvdb-device is part of storage in non-kvdb nodes
		isCloudDrive, err := IsCloudDriveInitialised(storageNodes[0])
		log.FailOnError(err, "Cloud drive installation failed")
		if !isCloudDrive {
			for _, storageNode := range storageNodes {
				err := Inst().V.AddBlockDrives(&storageNode, nil)
				if err != nil && strings.Contains(err.Error(), "no block drives available to add") {
					continue
				}
				log.FailOnError(err, "Adding block drive(s) failed.")
			}
		}

		Step("Create App Backup", func() {
			log.Infof("Contexts is: %v", contexts)
			log.FailOnError(err, "Failed to create app")
			backupLocation, err := applicationbackup.CreateBackupLocation(backupLocationName, defaultNs, s3SecretName)
			log.FailOnError(err, "Failed to create backup location")
			appBackup, bkp_create_err := applicationbackup.CreateApplicationBackupKs(backupName, defaultNs, backupLocation, bkpNs)
			log.FailOnError(bkp_create_err, "Failed to create backup")
			log.Infof("Backup is %v", appBackup)
			bkp_comp_err := applicationbackup.WaitForAppBackupCompletion(backupName, defaultNs, timeout)
			log.FailOnError(bkp_comp_err, "Backup completion failed")
			log.InfoD("backup successful, backup name - %v, backup location - %v", backupName, backupLocationName)
		})

		Step("start the upgrade of volume driver", func() {
			log.InfoD("start the upgrade of volume driver")

			if len(Inst().UpgradeStorageDriverEndpointList) == 0 {
				log.Fatalf("Unable to perform volume driver upgrade hops, none were given")
			}
			// Perform upgrade hops of volume driver based on a given list of upgradeEndpoints passed
			for _, upgradeHop := range strings.Split(Inst().UpgradeStorageDriverEndpointList, ",") {
				currPXVersion, err := Inst().V.GetDriverVersionOnNode(storageNodes[0])
				if err != nil {
					log.Warnf("error getting driver version, Err: %v", err)
				}
				upgradeStatus, updatedPXVersion, durationInMins := upgradePX(upgradeHop, storageNodes)
				log.InfoD("current PX version: %s, no of nodes: %d, upgrade status: %s, updated px version: %s, duration in mins: %d on source cluster",
					currPXVersion, len(storageNodes), upgradeStatus, updatedPXVersion, durationInMins)
				majorVersion := strings.Split(currPXVersion, "-")[0]
				statsData := make(map[string]string)
				statsData["numOfNodes"] = fmt.Sprintf("%d", len(storageNodes))
				statsData["fromVersion"] = currPXVersion
				statsData["toVersion"] = updatedPXVersion
				statsData["duration"] = fmt.Sprintf("%d mins", durationInMins)
				statsData["status"] = upgradeStatus
				dash.UpdateStats("px-upgrade-stats", "px-enterprise", "upgrade", majorVersion, statsData)
			}
		})

		Step("Perform Restore to new Ns", func() {
			nsSpec := &v1.Namespace{
				ObjectMeta: meta_v1.ObjectMeta{
					Name: nsRest,
				},
			}
			_, err = core.Instance().CreateNamespace(nsSpec)
			log.FailOnError(err, "Failed to create namespace")
			bl, err := storkops.Instance().GetBackupLocation(backupLocationName, defaultNs)
			log.FailOnError(err, "Failed to get backup location")
			log.Infof("BackupLocation is %v", bl)
			ab, err := storkops.Instance().GetApplicationBackup(backupName, defaultNs)
			log.FailOnError(err, "Failed to get backup location")
			namespaceMapping := make(map[string]string)
			namespaceMapping[bkpNs[0]] = nsRest
			appRestore, restore_err := applicationbackup.CreateApplicationRestore(restName, defaultNs, bl, ab.Name, namespaceMapping)
			log.FailOnError(restore_err, "Failed to create restore")
			log.Infof("Restore is %v", appRestore)
			restr_comp_err := applicationbackup.WaitForAppRestoreCompletion(restName, defaultNs, timeout)
			log.FailOnError(restr_comp_err, "Restore completion failed")
			log.InfoD("restore successful, restore name - %v", restName)
			restContext, err := CloneAppContextAndTransformWithMappings(contexts[0], namespaceMapping, nil, true)
			waitForPodsToBeRunning(restContext, false)
			log.FailOnError(err, "Failed to get mapped context")
			contexts = append(contexts, restContext)
		})

		Step("Destroy apps", func() {
			log.InfoD("Destroy apps")
			opts := make(map[string]bool)
			opts[scheduler.OptionsWaitForResourceLeakCleanup] = true
			for _, ctx := range contexts {
				TearDownContext(ctx, opts)
			}
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts)
	})
})

var _ = Describe("{UpgradeVolumeDriverDuringAsyncDrMigration}", Label("p2", "positive", "AsyncDR", "Upgrade"), func() {
	BeforeEach(func() {
		if !kubeConfigWritten {
			// Write kubeconfig files after reading from the config maps created by torpedo deploy script
			WriteKubeconfigToFiles()
			kubeConfigWritten = true
		}
		wantAllAfterSuiteActions = false
	})

	JustBeforeEach(func() {
		upgradeHopsList := make(map[string]string)
		upgradeHopsList["upgradeHops"] = Inst().UpgradeStorageDriverEndpointList
		upgradeHopsList["UpgradeVolumeDriverDuringAsyncDrMigration"] = "true"
		StartTorpedoTest("UpgradeVolumeDriverDuringAsyncDrMigration", "Validating volume driver upgrade during migration", upgradeHopsList, 0)
		log.InfoD("Volume driver upgrade hops list [%s]", upgradeHopsList)
	})
	var contexts []*scheduler.Context

	It("upgrade volume driver during migration and ensure everything is running fine", func() {
		log.InfoD("upgrade volume driver during migration and ensure everything is running fine")

		taskNamePrefix := "asyncdr-upgradepx"
		defaultNs := "kube-system"
		migrationNamespaces, contexts := initialSetupApps(taskNamePrefix, false, false)
		migNamespaces := strings.Join(migrationNamespaces, ",")
		kubeConfigPath := map[int]string{}
		for _, cluster := range []int{asyncdr.FirstCluster, asyncdr.SecondCluster} {
			kubeConfigPath[cluster], err = GetCustomClusterConfigPath(cluster)
			log.FailOnError(err, "Getting error while fetching path for %v cluster, error is %v", cluster, err)
		}

		var migrationSchedName string
		var schdPol *storkapi.SchedulePolicy
		cpName := defaultClusterPairName + time.Now().Format("15h03m05s")
		scpolName := "async-policy"
		migrationInterval := 5
		Step("Create Schedule Policy", func() {
			schdPol, err = asyncdr.CreateSchedulePolicy(scpolName, migrationInterval)
			log.FailOnError(err, "Failed to create schedule policy")
		})

		extraArgs := map[string]string{
			"namespaces":           migNamespaces,
			"kubeconfig":           kubeConfigPath[asyncdr.FirstCluster],
			"schedule-policy-name": schdPol.Name,
		}

		Step("create clusterpair and start migration", func() {
			log.InfoD("Creating clusterpair between first and second cluster")
			err = ScheduleBidirectionalClusterPair(cpName, defaultNs, "", storkapi.BackupLocationType(defaultBackupLocation), defaultSecret, "async-dr", asyncdr.FirstCluster, asyncdr.SecondCluster, nil)
			log.FailOnError(err, "Failed creating bidirectional cluster pair")

			log.InfoD("Start migration schedule and perform failover")
			migrationSchedName = migrationSchedKey + time.Now().Format("15h03m05s")
			createMigSchdAndValidateMigration(migrationSchedName, cpName, defaultNs, kubeConfigPath[asyncdr.FirstCluster], extraArgs)
		})

		stNodeClusterMap := make(map[int][]node.Node)

		for cluster, path := range kubeConfigPath {
			if cluster == asyncdr.SecondCluster {
				err = hardSetConfig(path)
				log.FailOnError(err, "Switching context to %v cluster failed", cluster)
				err = SetCustomKubeConfig(asyncdr.SecondCluster)
				log.FailOnError(err, "Switching context to %v cluster failed", cluster)
				err = Inst().S.RefreshNodeRegistry()
				log.FailOnError(err, "Node registry refresh failed")
				err = Inst().V.RefreshDriverEndpoints()
				log.FailOnError(err, "Refresh Driver end points failed")
				stNodeClusterMap[cluster] = node.GetStorageNodes()
				Step("Create Schedule Policy", func() {
					schdPol, err = asyncdr.CreateSchedulePolicy(scpolName, migrationInterval)
					log.FailOnError(err, "Failed to create schedule policy")
				})
			} else {
				stNodeClusterMap[cluster] = node.GetStorageNodes()
			}
			//AddDrive is added to test to Vsphere Cloud drive upgrades when kvdb-device is part of storage in non-kvdb nodes
			isCloudDrive, err := IsCloudDriveInitialised(stNodeClusterMap[cluster][0])
			log.FailOnError(err, "Cloud drive installation failed")
			if !isCloudDrive {
				for _, storageNode := range stNodeClusterMap[cluster] {
					err := Inst().V.AddBlockDrives(&storageNode, nil)
					if err != nil && strings.Contains(err.Error(), "no block drives available to add") {
						continue
					}
					log.FailOnError(err, "Adding block drive(s) failed.")
				}
			}
		}

		Step("start the upgrade of volume driver", func() {
			log.InfoD("start the upgrade of volume driver")

			if len(Inst().UpgradeStorageDriverEndpointList) == 0 {
				log.Fatalf("Unable to perform volume driver upgrade hops, none were given")
			}
			// Perform upgrade hops of volume driver based on a given list of upgradeEndpoints passed
			for _, upgradeHop := range strings.Split(Inst().UpgradeStorageDriverEndpointList, ",") {
				for _, cluster := range []int{asyncdr.FirstCluster, asyncdr.SecondCluster} {
					err = SetCustomKubeConfig(cluster)
					log.FailOnError(err, "Switching context to %v cluster failed", cluster)
					err = Inst().S.RefreshNodeRegistry()
					log.FailOnError(err, "Node registry refresh failed")
					err = Inst().V.RefreshDriverEndpoints()
					log.FailOnError(err, "Refresh Driver end points failed")
					currPXVersion, err := Inst().V.GetDriverVersionOnNode(stNodeClusterMap[cluster][0])
					if err != nil {
						log.Warnf("error getting driver version, Err: %v", err)
					}
					upgradeStatus, updatedPXVersion, durationInMins := upgradePX(upgradeHop, stNodeClusterMap[cluster])
					log.InfoD("current PX version: %s, no of nodes: %d, upgrade status: %s, updated px version: %s, duration in mins: %d on source cluster",
						currPXVersion, len(stNodeClusterMap[cluster]), upgradeStatus, updatedPXVersion, durationInMins)
					majorVersion := strings.Split(currPXVersion, "-")[0]
					statsData := make(map[string]string)
					statsData["numOfNodes"] = fmt.Sprintf("%d", len(stNodeClusterMap[cluster]))
					statsData["fromVersion"] = currPXVersion
					statsData["toVersion"] = updatedPXVersion
					statsData["duration"] = fmt.Sprintf("%d mins", durationInMins)
					statsData["status"] = upgradeStatus
					dash.UpdateStats("px-upgrade-stats", "px-enterprise", "upgrade", majorVersion, statsData)
				}
			}
		})

		Step("Perform failover and failback", func() {
			time.Sleep(2 * time.Duration(migrationInterval) * time.Minute)
			extraArgsFailoverFailback := map[string]string{
				"kubeconfig": kubeConfigPath[asyncdr.SecondCluster],
			}
			failoverParam := failoverFailbackParam{
				action:                    "failover",
				failoverOrFailbackNs:      defaultNs,
				migrationSchedName:        migrationSchedName,
				configPath:                kubeConfigPath[asyncdr.SecondCluster],
				single:                    false,
				skipSourceOp:              false,
				includeNs:                 false,
				excludeNs:                 false,
				extraArgsFailoverFailback: extraArgsFailoverFailback,
				contexts:                  contexts,
			}

			performFailoverFailback(failoverParam)

			err = hardSetConfig(kubeConfigPath[asyncdr.SecondCluster])
			log.FailOnError(err, "Error setting destination config: %v", err)
			extraArgs["kubeconfig"] = kubeConfigPath[asyncdr.SecondCluster]
			newMigSched := migrationSchedName + "-rev"
			createMigSchdAndValidateMigration(newMigSched, cpName, defaultNs, kubeConfigPath[asyncdr.SecondCluster], extraArgs)
			failback := failoverFailbackParam{
				action:                    "failback",
				failoverOrFailbackNs:      defaultNs,
				migrationSchedName:        newMigSched,
				configPath:                kubeConfigPath[asyncdr.SecondCluster],
				single:                    false,
				skipSourceOp:              false,
				includeNs:                 false,
				excludeNs:                 false,
				extraArgsFailoverFailback: extraArgsFailoverFailback,
				contexts:                  contexts,
			}
			performFailoverFailback(failback)
		})

		err = SetCustomKubeConfig(asyncdr.FirstCluster)
		log.FailOnError(err, "Switching context to source cluster failed")
		err = Inst().S.RefreshNodeRegistry()
		log.FailOnError(err, "Node registry refresh failed")
		err = Inst().V.RefreshDriverEndpoints()
		log.FailOnError(err, "Refresh Driver end points failed")

		Step("Destroy apps", func() {
			log.InfoD("Destroy apps")
			opts := make(map[string]bool)
			opts[scheduler.OptionsWaitForResourceLeakCleanup] = true
			for _, ctx := range contexts {
				TearDownContext(ctx, opts)
			}
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts)
	})
})

var _ = Describe("{UpgradeStorkDuringAsyncDrMigration}", Label("p2", "positive", "AsyncDR", "Upgrade"), func() {
	BeforeEach(func() {
		if !kubeConfigWritten {
			// Write kubeconfig files after reading from the config maps created by torpedo deploy script
			WriteKubeconfigToFiles()
			kubeConfigWritten = true
		}
		wantAllAfterSuiteActions = false
	})

	JustBeforeEach(func() {
		upgradeHopsList := make(map[string]string)
		upgradeHopsList["upgradeHops"] = Inst().UpgradeStorkVersionList
		upgradeHopsList["UpgradeStorkDuringAsyncDrMigration"] = "true"
		StartTorpedoTest("UpgradeStorkDuringAsyncDrMigration", "Validating asyncdr migration with stork upgrade", upgradeHopsList, 0)
		log.InfoD("Stork upgrade hops list [%s]", upgradeHopsList)
	})
	var contexts []*scheduler.Context

	It("upgrade stork during migration and ensure everything is running fine", func() {
		log.InfoD("upgrade stork during migration and ensure everything is running fine")

		taskNamePrefix := "asyncdr-upgradestork"
		defaultNs := "kube-system"
		migrationNamespaces, contexts := initialSetupApps(taskNamePrefix, false, false)
		migNamespaces := strings.Join(migrationNamespaces, ",")
		kubeConfigPath := map[int]string{}
		for _, cluster := range []int{asyncdr.FirstCluster, asyncdr.SecondCluster} {
			kubeConfigPath[cluster], err = GetCustomClusterConfigPath(cluster)
			log.FailOnError(err, "Getting error while fetching path for %v cluster, error is %v", cluster, err)
		}

		var migrationSchedName string
		var schdPol *storkapi.SchedulePolicy
		cpName := defaultClusterPairName + time.Now().Format("15h03m05s")
		scpolName := "async-policy"
		migrationInterval := 5
		Step("Create Schedule Policy", func() {
			schdPol, err = asyncdr.CreateSchedulePolicy(scpolName, migrationInterval)
			log.FailOnError(err, "Failed to create schedule policy")
		})

		extraArgs := map[string]string{
			"namespaces":           migNamespaces,
			"kubeconfig":           kubeConfigPath[asyncdr.FirstCluster],
			"schedule-policy-name": schdPol.Name,
		}

		Step("create clusterpair and start migration", func() {
			log.InfoD("Creating clusterpair between first and second cluster")
			err = ScheduleBidirectionalClusterPair(cpName, defaultNs, "", storkapi.BackupLocationType(defaultBackupLocation), defaultSecret, "async-dr", asyncdr.FirstCluster, asyncdr.SecondCluster, nil)
			log.FailOnError(err, "Failed creating bidirectional cluster pair")

			log.InfoD("Start migration schedule and perform failover")
			migrationSchedName = migrationSchedKey + time.Now().Format("15h03m05s")
			createMigSchdAndValidateMigration(migrationSchedName, cpName, defaultNs, kubeConfigPath[asyncdr.FirstCluster], extraArgs)
		})

		Step("start the upgrade of stork", func() {
			log.InfoD("start the upgrade of stork")

			if len(Inst().UpgradeStorkVersionList) == 0 {
				log.Fatalf("Unable to perform stork upgrade hops, none were given")
			}
			stc, err := Inst().V.GetDriver()
			log.FailOnError(err, "Failed to get storage cluster")
			// Perform upgrade hops of stork based on a given list of stork versions passed
			for _, upgradeHop := range strings.Split(Inst().UpgradeStorkVersionList, ",") {
				for _, cluster := range []int{asyncdr.FirstCluster, asyncdr.SecondCluster} {
					err = SetCustomKubeConfig(cluster)
					log.FailOnError(err, "Switching context to %v cluster failed", cluster)
					err = Inst().S.RefreshNodeRegistry()
					log.FailOnError(err, "Node registry refresh failed")
					err = Inst().V.RefreshDriverEndpoints()
					log.FailOnError(err, "Refresh Driver end points failed")
					err = asyncdr.PatchStorageClusterStorkImage(stc.Name, stc.Namespace, kubeConfigPath[cluster], upgradeHop)
					log.FailOnError(err, "Failed to upgrade stork on [%v] cluster with [%v] image", cluster, upgradeHop)
					if cluster == asyncdr.SecondCluster {
						schdPol, err = asyncdr.CreateSchedulePolicy(scpolName, migrationInterval)
						log.FailOnError(err, "Failed to create schedule policy")
					}
				}
			}
		})

		Step("Perform failover and failback", func() {
			time.Sleep(2 * time.Duration(migrationInterval) * time.Minute)
			extraArgsFailoverFailback := map[string]string{
				"kubeconfig": kubeConfigPath[asyncdr.SecondCluster],
			}
			failoverParam := failoverFailbackParam{
				action:                    "failover",
				failoverOrFailbackNs:      defaultNs,
				migrationSchedName:        migrationSchedName,
				configPath:                kubeConfigPath[asyncdr.SecondCluster],
				single:                    false,
				skipSourceOp:              false,
				includeNs:                 false,
				excludeNs:                 false,
				extraArgsFailoverFailback: extraArgsFailoverFailback,
				contexts:                  contexts,
			}

			performFailoverFailback(failoverParam)

			err = hardSetConfig(kubeConfigPath[asyncdr.SecondCluster])
			log.FailOnError(err, "Error setting destination config: %v", err)
			extraArgs["kubeconfig"] = kubeConfigPath[asyncdr.SecondCluster]
			newMigSched := migrationSchedName + "-rev"
			createMigSchdAndValidateMigration(newMigSched, cpName, defaultNs, kubeConfigPath[asyncdr.SecondCluster], extraArgs)
			failback := failoverFailbackParam{
				action:                    "failback",
				failoverOrFailbackNs:      defaultNs,
				migrationSchedName:        newMigSched,
				configPath:                kubeConfigPath[asyncdr.SecondCluster],
				single:                    false,
				skipSourceOp:              false,
				includeNs:                 false,
				excludeNs:                 false,
				extraArgsFailoverFailback: extraArgsFailoverFailback,
				contexts:                  contexts,
			}
			performFailoverFailback(failback)
		})

		err = SetCustomKubeConfig(asyncdr.FirstCluster)
		log.FailOnError(err, "Switching context to source cluster failed")
		err = Inst().S.RefreshNodeRegistry()
		log.FailOnError(err, "Node registry refresh failed")
		err = Inst().V.RefreshDriverEndpoints()
		log.FailOnError(err, "Refresh Driver end points failed")

		Step("Destroy apps", func() {
			log.InfoD("Destroy apps")
			opts := make(map[string]bool)
			opts[scheduler.OptionsWaitForResourceLeakCleanup] = true
			for _, ctx := range contexts {
				TearDownContext(ctx, opts)
			}
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts)
	})
})

var _ = Describe("{AutoVolumeSnapshot}", Label("p1", "positive", "VolumeSnapshot"), func() {
	testrailID = 302510
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/302510
	BeforeEach(func() {
		if !kubeConfigWritten {
			// Write kubeconfig files after reading from the config maps created by torpedo deploy script
			WriteKubeconfigToFiles()
			kubeConfigWritten = true
		}
		wantAllAfterSuiteActions = false
	})
	JustBeforeEach(func() {
		StartTorpedoTest("AutoVolumeSnaphpt", "AutoVolumeSnapshot test", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})

	It("has to deploy app, see if volumesnapshot started automatically", func() {
		appList := Inst().AppList
		stc, err := Inst().V.GetDriver()
		log.FailOnError(err, "Failed to get storage cluster")
		if stc.Spec.Security != nil && stc.Spec.Security.Enabled {
			log.InfoD("Security is enabled, changing the app to use auth app")
			Inst().AppList = []string{"auto-localsnap-auth"}
		}
		defer func() {
			Inst().AppList = appList
		}()
		Step("has to deploy app, see if volumesnapshot started automatically", func() {
			log.Infof("AppList is %v", Inst().AppList)
			var (
				taskNamePrefix                 = "auto-volumesnapshot"
				snapInterval                   = 2
				retain         storkapi.Retain = 2
				scpolName                      = "auto-schedule-policy"
			)
			snapNs, _ := initialSetupApps(taskNamePrefix, true, false)
			_, err := asyncdr.CreateSchedulePolicyWithRetain(scpolName, snapInterval, retain)
			log.FailOnError(err, "Failed to create schedule policy")
			pvcs, err := GetPVCListForNamespace(snapNs[0])
			log.FailOnError(err, "Failed to get PVC list")
			for _, pvcName := range pvcs {
				scheduleName := string(pvcName) + "-test-schedule"
				log.InfoD("SnapSchedule is %v", scheduleName)
				err := asyncdr.ValidateSnapshotScheduleRetainCount(scheduleName, snapNs[0], snapInterval, retain)
				log.FailOnError(err, "Failed to validate snapshot schedule")
			}
			for _, ctx := range contexts {
				TearDownContext(ctx, nil)
			}
			err = asyncdr.WaitForNamespaceDeletion(snapNs)
			if err != nil {
				log.Errorf("Failed to delete namespaces: %v", err)
			}
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

var _ = Describe("{AutoVolumeSnapshotCloud}", Label("p1", "positive", "VolumeSnapshot"), func() {
	testrailID = 302510
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/302510
	BeforeEach(func() {
		if !kubeConfigWritten {
			// Write kubeconfig files after reading from the config maps created by torpedo deploy script
			WriteKubeconfigToFiles()
			kubeConfigWritten = true
		}
		wantAllAfterSuiteActions = false
	})
	JustBeforeEach(func() {
		StartTorpedoTest("AutoVolumeSnaphptCloud", "AutoVolumeSnapshotCloud test", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})

	It("has to deploy app, see if volumesnapshot started automatically", func() {
		appList := Inst().AppList
		stc, err := Inst().V.GetDriver()
		log.FailOnError(err, "Failed to get storage cluster")
		if stc.Spec.Security != nil && stc.Spec.Security.Enabled {
			log.InfoD("Security is enabled, changing the app to use auth app")
			Inst().AppList = []string{"auto-cloudsnap-auth"}
		}
		defer func() {
			Inst().AppList = appList
		}()
		Step("has to deploy app, see if volumesnapshot started automatically", func() {
			log.Infof("AppList is %v", Inst().AppList)
			var (
				taskNamePrefix                 = "auto-volumesnapshot"
				snapInterval                   = 2
				retain         storkapi.Retain = 2
				scpolName                      = "auto-schedule-policy"
			)
			snapNs, _ := initialSetupApps(taskNamePrefix, true, false)
			_, err := asyncdr.CreateSchedulePolicyWithRetain(scpolName, snapInterval, retain)
			log.FailOnError(err, "Failed to create schedule policy")
			pvcs, err := GetPVCListForNamespace(snapNs[0])
			log.FailOnError(err, "Failed to get PVC list")
			for _, pvcName := range pvcs {
				scheduleName := string(pvcName) + "-test-schedule"
				log.InfoD("SnapSchedule is %v", scheduleName)
				err := asyncdr.ValidateSnapshotScheduleRetainCount(scheduleName, snapNs[0], snapInterval, retain)
				log.FailOnError(err, "Failed to validate snapshot schedule")
			}
			for _, ctx := range contexts {
				TearDownContext(ctx, nil)
			}
			err = asyncdr.WaitForNamespaceDeletion(snapNs)
			if err != nil {
				log.Errorf("Failed to delete namespaces: %v", err)
			}
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

func upgradePX(upgradeHop string, storageNodes []node.Node) (string, string, int) {
	var timeBeforeUpgrade time.Time
	var timeAfterUpgrade time.Time
	timeBeforeUpgrade = time.Now()
	isDmthinBeforeUpgrade, errDmthinCheck := IsDMthin()
	dash.VerifyFatal(errDmthinCheck, nil, "verified is setup dmthin before upgrade? ")
	err := Inst().V.UpgradeDriver(upgradeHop)
	timeAfterUpgrade = time.Now()
	dash.VerifyFatal(err, nil, "Volume driver upgrade successful?")
	durationInMins := int(timeAfterUpgrade.Sub(timeBeforeUpgrade).Minutes())
	expectedUpgradeTime := 9 * len(node.GetStorageDriverNodes())
	dash.VerifySafely(durationInMins <= expectedUpgradeTime, true, "Verify volume drive upgrade within expected time")
	upgradeStatus := "PASS"
	if durationInMins <= expectedUpgradeTime {
		log.InfoD("Upgrade successfully completed in %d minutes which is within %d minutes", durationInMins, expectedUpgradeTime)
	} else {
		log.Errorf("Upgrade took %d minutes to completed which is greater than expected time %d minutes", durationInMins, expectedUpgradeTime)
		dash.VerifySafely(durationInMins <= expectedUpgradeTime, true, "Upgrade took more than expected time to complete")
		upgradeStatus = "FAIL"
	}
	updatedPXVersion, err := Inst().V.GetDriverVersionOnNode(storageNodes[0])
	if err != nil {
		log.Warnf("error getting driver version, Err: %v", err)
	}
	isDmthinAfterUpgrade, errDmthinCheck := IsDMthin()
	dash.VerifyFatal(errDmthinCheck, nil, "verified is setup dmthin after upgrade? ")
	dash.VerifyFatal(isDmthinBeforeUpgrade, isDmthinAfterUpgrade, "setup type remained same pre and post upgrade")
	return upgradeStatus, updatedPXVersion, durationInMins
}

func validateFailoverFailback(clusterType, taskNamePrefix string, single, skipSourceOp, includeNs, excludeNs bool, volumeSkip bool) {
	isNfs := os.Getenv("IS_STORK_NFS_LOCATION") == "true"
	log.Infof(" User has opted for IS_STORK_NFS_LOCATION to  %s ", os.Getenv("IS_STORK_NFS_LOCATION"))

	defaultNs := "kube-system"
	migrationNamespaces, contexts := initialSetupApps(taskNamePrefix, single, volumeSkip)
	migNamespaces := strings.Join(migrationNamespaces, ",")
	kubeConfigPathSrc, err := GetCustomClusterConfigPath(asyncdr.FirstCluster)
	log.FailOnError(err, "Failed to get source configPath: %v", err)
	kubeConfigPathDest, err := GetCustomClusterConfigPath(asyncdr.SecondCluster)
	log.FailOnError(err, "Failed to get destination configPath: %v", err)
	if single {
		defaultNs = migrationNamespaces[0]
		migNamespaces = defaultNs
	}
	extraArgs := map[string]string{
		"namespaces":   migNamespaces,
		"kubeconfig":   kubeConfigPathSrc,
		"include-jobs": "",
	}
	stc, err := Inst().V.GetDriver()
	log.FailOnError(err, "Failed to get driver")

	isCloud, cloudName := asyncdr.IsCloud(stc)
	var srcEp, destEp string
	extraArgsCp := map[string]string{}
	const defaultPort = "9001"

	if isCloud {
		storageDriverName := Inst().V.String()
		if err = asyncdr.ChangePxServiceToLoadBalancer(false, storageDriverName, stc); err != nil {
			log.FailOnError(err, "failed to change PX service to LoadBalancer on source cluster")
		}
		if cloudName == "eks" {
			pxService, err := core.Instance().GetService("portworx-service", "kube-system")
			log.FailOnError(err, "failed to get px service")
			srcEp = pxService.Status.LoadBalancer.Ingress[0].Hostname
		}
		err = SetDestinationKubeConfig()
		log.FailOnError(err, "Failed to set destination kubeconfig")
		if err = asyncdr.ChangePxServiceToLoadBalancer(false, storageDriverName, stc); err != nil {
			log.FailOnError(err, "failed to change PX service to LoadBalancer on destination cluster")
		}
		if cloudName == "eks" {
			pxService, err := core.Instance().GetService("portworx-service", "kube-system")
			log.FailOnError(err, "failed to get px service")
			destEp = pxService.Status.LoadBalancer.Ingress[0].Hostname
		}
		err = SetSourceKubeConfig()
		log.FailOnError(err, "Failed to set source kubeconfig")
	}

	if cloudName == "aks" {
		defaultSecret = azureSecret
		defaultBackupLocation = azureBackupLocation
	} else if cloudName == "gke" {
		defaultSecret = googleSecret
		defaultBackupLocation = googleBackupLocation
	} else if cloudName == "eks" {
		extraArgsCp["src-ep"] = srcEp + ":" + defaultPort
		extraArgsCp["dest-ep"] = destEp + ":" + defaultPort
	}

	log.Infof("Creating clusterpair between first and second cluster")
	cpName := defaultClusterPairName + time.Now().Format("15h03m05s")

	if clusterType == "asyncdr" {
		if isNfs {
			log.Infof("Cluster pair will be created wrt nfs object location ")
			err = ScheduleBidirectionalClusterPair(cpName, defaultNs, "", storkapi.BackupLocationType(nfsBackupLocation), nfsSecret, "async-dr", asyncdr.FirstCluster, asyncdr.SecondCluster, extraArgsCp)
		} else {
			err = ScheduleBidirectionalClusterPair(cpName, defaultNs, "", storkapi.BackupLocationType(defaultBackupLocation), defaultSecret, "async-dr", asyncdr.FirstCluster, asyncdr.SecondCluster, extraArgsCp)
		}

	} else {
		err = ScheduleBidirectionalClusterPair(cpName, defaultNs, "", "", "", "sync-dr", asyncdr.FirstCluster, asyncdr.SecondCluster, extraArgsCp)
	}

	log.FailOnError(err, "Failed creating bidirectional cluster pair")

	if isCloud {
		err := patchClusterPair(cpName, defaultNs, kubeConfigPathSrc)
		log.FailOnError(err, "Failed patching cluster pair")
		err = patchClusterPair(cpName, defaultNs, kubeConfigPathDest)
		log.FailOnError(err, "Failed patching cluster pair")
	}

	log.Infof("Start migration schedule and perform failover")
	migrationSchedName := migrationSchedKey + time.Now().Format("15h03m05s")
	createMigSchdAndValidateMigration(migrationSchedName, cpName, defaultNs, kubeConfigPathSrc, extraArgs)
	err = SetCustomKubeConfig(asyncdr.SecondCluster)
	log.FailOnError(err, "Switching context to second cluster failed")
	extraArgsFailoverFailback := map[string]string{
		"kubeconfig": kubeConfigPathDest,
	}
	if includeNs {
		extraArgsFailoverFailback["include-namespaces"] = migrationNamespaces[0]
	}
	if excludeNs {
		extraArgsFailoverFailback["exclude-namespaces"] = migrationNamespaces[0]
	}
	failoverParam := failoverFailbackParam{
		action:                    "failover",
		failoverOrFailbackNs:      defaultNs,
		migrationSchedName:        migrationSchedName,
		configPath:                kubeConfigPathDest,
		single:                    single,
		skipSourceOp:              skipSourceOp,
		includeNs:                 includeNs,
		excludeNs:                 excludeNs,
		extraArgsFailoverFailback: extraArgsFailoverFailback,
		contexts:                  contexts,
	}
	performFailoverFailback(failoverParam)
	time.Sleep(1 * time.Minute)
	if clusterType != "asyncdr" {
		validateDomains("failover", kubeConfigPathDest)
	}
	if skipSourceOp {
		err = hardSetConfig(kubeConfigPathSrc)
		log.FailOnError(err, "Error setting source config: %v", err)
		for _, ctx := range contexts {
			waitForPodsToBeRunning(ctx, false)
		}
	} else {
		err = hardSetConfig(kubeConfigPathDest)
		log.FailOnError(err, "Error setting destination config: %v", err)
		extraArgs["kubeconfig"] = kubeConfigPathDest
		newMigSched := migrationSchedName + "-rev"
		if includeNs {
			extraArgs["namespaces"] = migrationNamespaces[0]
		}
		if excludeNs {
			extraArgs["namespaces"] = strings.Join(migrationNamespaces[1:], ",")
			extraArgsFailoverFailback["exclude-namespaces"] = migrationNamespaces[1]
		}
		createMigSchdAndValidateMigration(newMigSched, cpName, defaultNs, kubeConfigPathDest, extraArgs)
		failoverback := failoverFailbackParam{
			action:                    "failback",
			failoverOrFailbackNs:      defaultNs,
			migrationSchedName:        newMigSched,
			configPath:                kubeConfigPathDest,
			single:                    single,
			skipSourceOp:              false,
			includeNs:                 includeNs,
			excludeNs:                 excludeNs,
			extraArgsFailoverFailback: extraArgsFailoverFailback,
			contexts:                  contexts,
		}
		performFailoverFailback(failoverback)
		if clusterType != "asyncdr" {
			time.Sleep(1 * time.Minute)
			validateDomains("failback", kubeConfigPathSrc)
		}
	}
	err = asyncdr.WaitForNamespaceDeletion(migrationNamespaces)
	if err != nil {
		log.Infof("Failed to delete namespaces: %v", err)
	}
	err = hardSetConfig(kubeConfigPathDest)
	if err != nil {
		log.Infof("Failed to se dest kubeconfig for NS deletion on dest: %v", err)
	}
	err = asyncdr.WaitForNamespaceDeletion(migrationNamespaces)
	if err != nil {
		log.Infof("Failed to delete namespaces: %v", err)
	}
}

func getClusterDomainsInfo() bool {
	skipFlag := false
	listCdsTask := func() (interface{}, bool, error) {
		// Fetch the cluster domains
		cdses, err := storkops.Instance().ListClusterDomainStatuses()
		if err != nil || len(cdses.Items) == 0 {
			log.Infof("Failed to list cluster domains statuses. Error: %v. List of cluster domains: %v", err, len(cdses.Items))
			return "", true, fmt.Errorf("failed to list cluster domains statuses")
		}
		cds := cdses.Items[0]
		if len(cds.Status.ClusterDomainInfos) == 0 {
			log.Infof("Found 0 cluster domain info objects in cluster domain status.")
			return "", true, fmt.Errorf("failed to list cluster domains statuses")
		}
		return "", false, nil
	}
	_, err := task.DoRetryWithTimeout(listCdsTask, domainCheckRetryTimeout, migrationRetryInterval)
	if err != nil {
		skipFlag = true
	}
	return skipFlag
}

func WriteKubeconfigToFiles() {
	kubeconfigs := os.Getenv("KUBECONFIGS")
	Expect(kubeconfigs).NotTo(Equal(""),
		"KUBECONFIGS Environment variable should not be empty")

	kubeconfigList := strings.Split(kubeconfigs, ",")
	// Validate user has provided at least 1 kubeconfig for cluster
	Expect(len(kubeconfigList)).Should(BeNumerically(">=", 2), "At least minimum two kubeconfigs required")

	DumpKubeconfigs(kubeconfigList)

}

func CreateMigration(
	name string,
	namespace string,
	clusterPair string,
	migrationNamespace string,
	includeResources *bool,
	startApplications *bool,
) (*storkapi.Migration, error) {

	migration := &storkapi.Migration{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: storkapi.MigrationSpec{
			ClusterPair:       clusterPair,
			IncludeResources:  includeResources,
			StartApplications: startApplications,
			Namespaces:        []string{migrationNamespace},
		},
	}
	// TODO figure out a way to check if it's an auth-enabled and add security annotations
	//if authTokenConfigMap != "" {
	//	err := addSecurityAnnotation(migration)
	//	if err != nil {
	//		return nil, err
	//	}
	//}

	mig, err := storkops.Instance().CreateMigration(migration)
	return mig, err
}

func deleteMigrations(migrations []*storkapi.Migration) error {
	for _, mig := range migrations {
		err := storkops.Instance().DeleteMigration(mig.Name, mig.Namespace)
		if err != nil {
			return fmt.Errorf("Failed to delete migration %s in namespace %s. Error: %v", mig.Name, mig.Namespace, err)
		}
	}
	return nil
}

func WaitForMigration(migrationList []*storkapi.Migration) error {
	checkMigrations := func() (interface{}, bool, error) {
		isComplete := true
		for _, m := range migrationList {
			mig, err := storkops.Instance().GetMigration(m.Name, m.Namespace)
			if err != nil {
				return "", false, err
			}
			if mig.Status.Status != storkapi.MigrationStatusSuccessful && mig.Status.Status != storkapi.MigrationStatusPartialSuccess {
				log.Infof("Migration %s in namespace %s is pending", m.Name, m.Namespace)
				isComplete = false
			}
		}
		if isComplete {
			return "", false, nil
		}
		return "", true, fmt.Errorf("some migrations are still pending")
	}
	_, err := task.DoRetryWithTimeout(checkMigrations, migrationRetryTimeout, migrationRetryInterval)
	return err
}

func DeleteAndWaitForMigrationDeletion(name, namespace string) error {
	log.Infof("Deleting migration: %s in namespace: %s", name, namespace)
	err := storkops.Instance().DeleteMigration(name, namespace)
	if err != nil {
		return fmt.Errorf("Failed to delete migration: %s in namespace: %s", name, namespace)
	}
	getMigration := func() (interface{}, bool, error) {
		migration, err := storkops.Instance().GetMigration(name, namespace)
		if err == nil {
			return "", true, fmt.Errorf("Migration %s in %s has not completed yet.Status: %s. Retrying ", name, namespace, migration.Status.Status)
		}
		return "", false, nil
	}
	_, err = task.DoRetryWithTimeout(getMigration, migrationRetryTimeout, migrationRetryInterval)
	return err
}

func initialSetupApps(taskNamePrefix string, single bool, volumeSkip bool) ([]string, []*scheduler.Context) {
	var contexts []*scheduler.Context
	var migrationNamespaces []string

	err = SetCustomKubeConfig(asyncdr.FirstCluster)
	log.FailOnError(err, "Switching context to first cluster failed")
	if single {
		taskName := fmt.Sprintf("%s", taskNamePrefix)
		log.Infof("Task name %s\n", taskName)
		contexts = append(contexts, ScheduleApplications(taskName)...)
	} else {
		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			taskName := fmt.Sprintf("%s-%d", taskNamePrefix, i)
			log.Infof("Task name %s\n", taskName)
			contexts = append(contexts, ScheduleApplications(taskName)...)
		}
	}
	for _, ctx := range contexts {
		// Override default App readiness time out of 5 mins with 10 mins
		ctx.ReadinessTimeout = appReadinessTimeout
		namespace := GetAppNamespace(ctx, "")
		migrationNamespaces = append(migrationNamespaces, namespace)
	}
	log.Infof("Migration Namespaces are : [%v]", migrationNamespaces)
	if volumeSkip {
		for _, ctx := range contexts {
			ctx.SkipVolumeValidation = volumeSkip
		}
	}
	ValidateApplications(contexts)
	return migrationNamespaces, contexts
}

func createMigSchdAndValidateMigration(migSchedName, cpName, migNs, resetConfigPath string, extraArgs map[string]string) {
	var migration *storkapi.Migration
	err = storkctlcli.ScheduleStorkctlMigrationSched(migSchedName, cpName, migNs, extraArgs)
	log.FailOnError(err, "Error creating migrationschedule: %v", err)
	err = hardSetConfig(resetConfigPath)
	log.FailOnError(err, "Error setting destination config: %v", err)
	time.Sleep(time.Second * 30)
	migSchedule, err := storkops.Instance().GetMigrationSchedule(migSchedName, migNs)
	log.FailOnError(err, "failed to get migrationschedule %v, err: %v", migSchedName, err)
	migrations := migSchedule.Status.Items["Interval"]
	for _, mig := range migrations {
		migration, err = storkops.Instance().GetMigration(mig.Name, migNs)
		log.FailOnError(err, "failed to get migration for migrationschedule %v, err: %v", migSchedName, err)
		err = WaitForMigration([]*storkapi.Migration{migration})
		log.FailOnError(err, "Migration failed with error: %v", err)
	}
}

func performFailoverFailback(foFbParams failoverFailbackParam) {
	err, output := storkctlcli.PerformFailoverOrFailback(foFbParams.action, foFbParams.failoverOrFailbackNs, foFbParams.migrationSchedName, foFbParams.skipSourceOp, foFbParams.extraArgsFailoverFailback)
	log.FailOnError(err, "Error running perform %v: %v", foFbParams.action, err)
	splitOutput := strings.Split(output, "\n")
	prefix := fmt.Sprintf("To check %s status use the command : `", foFbParams.action)
	getStatusCommand := strings.TrimSpace(strings.TrimPrefix(splitOutput[1], prefix))
	getStatusCommand = strings.TrimSuffix(getStatusCommand, "`")
	getStatusCmdArgs := strings.Split(getStatusCommand, " ")
	// Extract the action Name from the command args
	actionName := getStatusCmdArgs[3]
	err = storkctlcli.WaitForActionSuccessful(actionName, foFbParams.failoverOrFailbackNs, Inst().GlobalScaleFactor)
	log.FailOnError(err, "Error in performing %v: %v", foFbParams.action, err)
	validatePodsRunning(foFbParams.action, foFbParams.single, foFbParams.includeNs, foFbParams.excludeNs, foFbParams.contexts)
}

func validatePodsRunning(action string, single, includeNs, excludeNs bool, contexts []*scheduler.Context) {
	switch action {
	case "failover":
		if includeNs {
			waitForPodsToBeRunning(contexts[0], false)
			for i := 1; i < len(contexts); i++ {
				ctx := contexts[i]
				waitForPodsToBeRunning(ctx, true)
			}
		} else if excludeNs {
			waitForPodsToBeRunning(contexts[0], true)
			for i := 1; i < len(contexts); i++ {
				ctx := contexts[i]
				waitForPodsToBeRunning(ctx, false)
			}
		} else if single {
			waitForPodsToBeRunning(contexts[0], false)
		} else {
			for _, ctx := range contexts {
				waitForPodsToBeRunning(ctx, false)
			}
		}
	case "failback":
		kubeConfigPathSrc, err := GetCustomClusterConfigPath(asyncdr.FirstCluster)
		log.FailOnError(err, "Failed to get source configPath: %v", err)
		err = hardSetConfig(kubeConfigPathSrc)
		log.FailOnError(err, "Error setting source config")
		if includeNs {
			for _, ctx := range contexts {
				waitForPodsToBeRunning(ctx, false)
			}
		} else if excludeNs {
			for i := 1; i < len(contexts); i++ {
				ctx := contexts[i]
				if i == 1 {
					waitForPodsToBeRunning(ctx, true)
				} else {
					waitForPodsToBeRunning(ctx, false)
				}
			}
		} else if single {
			waitForPodsToBeRunning(contexts[0], false)
		} else {
			for _, ctx := range contexts {
				waitForPodsToBeRunning(ctx, false)
			}
		}
	}
}

func validateDomains(failoverfailback, configPath string) {
	src, dest, witness, err := storkctlcli.GetActualClusterDomainStatus(failoverfailback, configPath)
	log.FailOnError(err, "Failed to get actual cluster domain status")
	if failoverfailback == "failover" {
		dash.VerifyFatal(src == "Inactive", true, "source cluster domain should not active")
	} else {
		dash.VerifyFatal(src == "Active", true, "source cluster domain should active")
	}
	dash.VerifyFatal(dest == "Active", true, "destination cluster domain should active")
	dash.VerifyFatal(witness == "Active", true, "witness cluster domain should active")
}

func hardSetConfig(configPath string) error {
	var config *rest.Config
	config, err = clientcmd.BuildConfigFromFlags("", configPath)
	if err != nil {
		return err
	}
	core.Instance().SetConfig(config)
	apps.Instance().SetConfig(config)
	stork.Instance().SetConfig(config)
	return nil
}

func waitForPodsToBeRunning(context *scheduler.Context, expectedFail bool) {
	log.Infof("Verifying Context [%v]", context.App.Key)
	err := Inst().S.WaitForRunning(context, 5*time.Minute, 10*time.Second)
	if expectedFail {
		log.FailOnNoError(err, "Pods are up on destination, they shouldn't be up")
	} else {
		log.FailOnError(err, "Error waiting for pods to be up")
	}
}

func validateOperatorMigFailover(namespace, clusterType, opName, crName string, podCount int, clusterwide bool) {
	cpName := defaultClusterPairName + time.Now().Format("15h03m05s")
	kubeConfigPathSrc, err := GetCustomClusterConfigPath(asyncdr.FirstCluster)
	log.FailOnError(err, "Failed to get source configPath: %v", err)
	kubeConfigPathDest, err := GetCustomClusterConfigPath(asyncdr.SecondCluster)
	log.FailOnError(err, "Failed to get destination configPath: %v", err)
	if clusterType == "asyncdr" {
		err = ScheduleBidirectionalClusterPair(cpName, namespace, "", storkapi.BackupLocationType(defaultBackupLocation), defaultSecret, "async-dr", asyncdr.FirstCluster, asyncdr.SecondCluster, nil)
	} else {
		err = ScheduleBidirectionalClusterPair(cpName, namespace, "", "", "", "sync-dr", asyncdr.FirstCluster, asyncdr.SecondCluster, nil)
	}
	log.FailOnError(err, "Failed creating bidirectional cluster pair")
	log.Infof("Start migration schedule and perform failover")
	migrationSchedName := migrationSchedKey + time.Now().Format("15h03m05s")
	extraArgs := map[string]string{
		"namespaces":             namespace,
		"kubeconfig":             kubeConfigPathSrc,
		"exclude-resource-types": "ClusterServiceVersion,operatorconditions,OperatorGroup,InstallPlan,Subscription",
		"exclude-selectors":      "olm.managed=true",
	}
	extraArgsFailoverFailback := map[string]string{
		"kubeconfig": kubeConfigPathDest,
	}
	err = patchStashStrategy(crName)
	log.FailOnError(err, "Failed to patch stash strategy")
	createMigSchdAndValidateMigration(migrationSchedName, cpName, namespace, kubeConfigPathSrc, extraArgs)
	err = hardSetConfig(kubeConfigPathSrc)
	log.FailOnError(err, "Switching context to source cluster failed")
	scaleCrApp(namespace, opName, true, clusterwide)
	err = SetCustomKubeConfig(asyncdr.SecondCluster)
	log.FailOnError(err, "Switching context to second cluster failed")
	err, output := storkctlcli.PerformFailoverOrFailback("failover", namespace, migrationSchedName, false, extraArgsFailoverFailback)
	log.FailOnError(err, "Error running perform %v: %v", "failover", err)
	actionName := extractActionName(output, "failover")
	err = storkctlcli.WaitForActionSuccessful(actionName, namespace, Inst().GlobalScaleFactor)
	log.FailOnError(err, "Error in performing %v: %v", "failover", err)
	time.Sleep(5 * time.Minute)
	podListDest, err := core.Instance().GetPods(namespace, nil)
	if err != nil {
		log.Errorf("Error getting podlist: %v", err)
	}
	err = asyncdr.WaitForPodToBeRunning(podListDest)
	if err != nil {
		log.Errorf("Error waiting for pod to be running: %v", err)
	}
	podCountDest := len(podListDest.Items)
	dash.VerifyFatal(podCountDest == podCount, true, "Pod count is not equal to expected")
	err = patchStashStrategy(crName)
	log.FailOnError(err, "Failed to patch stash strategy")
	err = hardSetConfig(kubeConfigPathDest)
	log.FailOnError(err, "Switching context to dest cluster failed")
	migrationSchedNameRev := migrationSchedKey + time.Now().Format("15h03m05s") + "-rev"
	extraArgs["kubeconfig"] = kubeConfigPathDest
	createMigSchdAndValidateMigration(migrationSchedNameRev, cpName, namespace, kubeConfigPathDest, extraArgs)
	err = hardSetConfig(kubeConfigPathDest)
	log.FailOnError(err, "Failed to set destination kubeconfig")
	scaleCrApp(namespace, opName, true, clusterwide)
	err = SetDestinationKubeConfig()
	log.FailOnError(err, "Failed to set destination kubeconfig")
	err, output = storkctlcli.PerformFailoverOrFailback("failback", namespace, migrationSchedNameRev, false, extraArgsFailoverFailback)
	log.FailOnError(err, "Error running perform %v: %v", "failback", err)
	actionName = extractActionName(output, "failback")
	err = storkctlcli.WaitForActionSuccessful(actionName, namespace, Inst().GlobalScaleFactor)
	log.FailOnError(err, "Error in performing %v: %v", "failback", err)
	err = SetSourceKubeConfig()
	log.FailOnError(err, "Failed to set source kubeconfig")
	scaleCrApp(namespace, opName, false, clusterwide)
	time.Sleep(5 * time.Minute)
	err = hardSetConfig(kubeConfigPathSrc)
	log.FailOnError(err, "Failed to set source kubeconfig")
	podListFailback, err := core.Instance().GetPods(namespace, nil)
	if err != nil {
		log.Errorf("Error getting podlist: %v", err)
	}
	err = asyncdr.WaitForPodToBeRunning(podListFailback)
	if err != nil {
		log.Errorf("Error waiting for pod to be running: %v", err)
	}
	podCountFailback := len(podListFailback.Items)
	dash.VerifyFatal(podCountFailback == podCount, true, "Pod count is not equal to expected")
}

func scaleCrApp(namespace, opName string, down, cw bool) {
	opNs := namespace
	if cw {
		opNs = clusterwideNs
	}
	desRepl := int32(1)
	if down {
		desRepl = int32(0)
	}
	deplop, err := apps.Instance().GetDeployment(opName, opNs)
	log.FailOnError(err, "Failed to get deployment")
	deplop.Spec.Replicas = &desRepl
	_, err = apps.Instance().UpdateDeployment(deplop)
	log.FailOnError(err, "Failed to scale deployment")

	if down {
		depl, err := apps.Instance().ListDeployments(namespace, meta_v1.ListOptions{})
		log.FailOnError(err, "Failed to list deployments")
		for _, d := range depl.Items {
			if d.Name != opName {
				d.Spec.Replicas = &desRepl
				_, err = apps.Instance().UpdateDeployment(&d)
				log.FailOnError(err, "Failed to scale down deployment")
			}
		}
		sts, err := apps.Instance().ListStatefulSets(namespace, meta_v1.ListOptions{})
		log.FailOnError(err, "Failed to list deployments")
		for _, s := range sts.Items {
			s.Spec.Replicas = &desRepl
			_, err = apps.Instance().UpdateStatefulSet(&s)
			log.FailOnError(err, "Failed to scale down statefulset")
		}
	}
}

func createOperatorBasedApp(appPath, opPath, ns string, clusterwide bool) (*v1.PodList, error) {
	nsSpec := &v1.Namespace{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: ns,
		},
	}
	_, err := core.Instance().CreateNamespace(nsSpec)
	if err != nil {
		return nil, err
	}
	if clusterwide {
		deployAndWaitForRunning(opPath, clusterwideNs, "src")
	} else {
		deployAndWaitForRunning(opPath, ns, "src")
	}
	deployAndWaitForRunning(appPath, ns, "src")
	podList, err := core.Instance().GetPods(ns, nil)
	if err != nil {
		return nil, err
	}
	err = SetDestinationKubeConfig()
	if err != nil {
		return nil, err
	}
	_, err = core.Instance().CreateNamespace(nsSpec)
	if err != nil {
		return nil, err
	}
	// Deploy operator
	if clusterwide {
		deployAndWaitForRunning(opPath, clusterwideNs, "dest")
	} else {
		deployAndWaitForRunning(opPath, ns, "dest")
	}
	return podList, nil
}

func deployAndWaitForRunning(path, namespace, cluster string) error {
	cmd := fmt.Sprintf("kubectl apply -f %v -n %v", path, namespace)
	if cluster == "dest" {
		kubeConfigPathDest, err := GetCustomClusterConfigPath(asyncdr.SecondCluster)
		if err != nil {
			return err
		}
		cmd = fmt.Sprintf("kubectl apply -f %v -n %v --kubeconfig %v", path, namespace, kubeConfigPathDest)
		err = hardSetConfig(kubeConfigPathDest)
		if err != nil {
			return err
		}
	}
	log.InfoD("Running command: %v", cmd)
	_, _, err = osutils.ExecShell(cmd)
	if err != nil {
		return err
	}
	// Sleeping here, as apps deploys one by one, which takes time to collect all pods
	time.Sleep(1 * time.Minute)
	podList, err := core.Instance().GetPods(namespace, nil)
	if err != nil {
		return err
	}
	err = asyncdr.WaitForPodToBeRunning(podList)
	if err != nil {
		return err
	}
	err = SetSourceKubeConfig()
	if err != nil {
		return err
	}
	return nil
}

type PatchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}

type ApplicationRegistration struct {
	Resources []struct {
	} `json:"resources"`
}

func getNumResources(crName string) (int, error) {
	cmd := exec.Command("kubectl", "get", "applicationregistration", crName, "-o", "json")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Infof("Error getting applicationregistration: %v", err)
		return 0, err
	}
	var appReg ApplicationRegistration
	err = json.Unmarshal(out.Bytes(), &appReg)
	if err != nil {
		log.Infof("Error unmarshalling JSON: %v", err)
		return 0, err
	}
	return len(appReg.Resources), nil
}

func patchStashStrategy(crName string) error {
	var patches []PatchOperation
	numResources, err := getNumResources(crName)
	if err != nil {
		return err
	}
	for i := 0; i < numResources; i++ {
		patches = append(patches, PatchOperation{
			Op:    "replace",
			Path:  fmt.Sprintf("/resources/%d/stashStrategy/stashCR", i),
			Value: true,
		})
	}
	patchBytes, err := json.Marshal(patches)
	if err != nil {
		log.Infof("Error marshalling patches: %v", err)
		return err
	}
	cmd := fmt.Sprintf(`kubectl patch applicationregistration %v --type='json' -p='%s'`, crName, string(patchBytes))
	log.Infof("Running command: %v", cmd)
	_, err = exec.Command("sh", "-c", cmd).CombinedOutput()
	if err != nil {
		log.Infof("Error running command: %v and err is: %v", cmd, err)
		return err
	}
	return nil
}

func patchClusterPair(cpName, cpNs, configPath string) error {
	patch := []byte(`[{"op": "remove", "path": "/spec/options/mode"}]`)
	cmd := fmt.Sprintf(`kubectl --kubeconfig %v patch clusterpair %v -n %v --type='json' -p='%s'`, configPath, cpName, cpNs, string(patch))
	log.Infof("Running command: %v", cmd)
	_, err = exec.Command("sh", "-c", cmd).CombinedOutput()
	if err != nil {
		log.Errorf("Error running command: %v and err is: %v", cmd, err)
		return err
	}
	return nil
}

func extractActionName(output, action string) string {
	splitOutput := strings.Split(output, "\n")
	prefix := fmt.Sprintf("To check %s status use the command : `", action)
	getStatusCommand := strings.TrimSpace(strings.TrimPrefix(splitOutput[1], prefix))
	getStatusCommand = strings.TrimSuffix(getStatusCommand, "`")
	getStatusCmdArgs := strings.Split(getStatusCommand, " ")
	return getStatusCmdArgs[3]
}

var _ = Describe("{RestartPXAndDeleteStorkLeaderDuringAsyncMigration}", Label("staging", "p1", "negative", "AsyncDR"), func() {
	testrailID := 0
	BeforeEach(func() {
		if !kubeConfigWritten {
			// Write kubeconfig files after reading from the config maps created by torpedo deploy script
			WriteKubeconfigToFiles()
			kubeConfigWritten = true
		}
		wantAllAfterSuiteActions = false
	})

	JustBeforeEach(func() {
		StartTorpedoTest("RestartPXAndDeleteStorkLeaderDuringAsyncMigration", "Migration of application to destination cluster with driver down", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})
	var contexts []*scheduler.Context

	stepLog := "Restart PX driver during migration and ensure migration completed successfully"
	It(stepLog, func() {
		log.InfoD(stepLog)

		taskNamePrefix := "asyncdr-restartpx"
		defaultNs := "kube-system"
		migrationNamespaces, contexts := initialSetupApps(taskNamePrefix, false, false)
		// Get the replica nodes from the context
		replicaNodeMap := make(map[string]node.Node)
		for _, ctx := range contexts {
			log.Infof("Get replicas nodes from the context")
			vols, err := Inst().S.GetVolumes(ctx)
			log.FailOnError(err, "Failed to get volumes for the app %v", ctx.App.Key)
			for _, appVol := range vols {
				apiVol, err := Inst().V.InspectVolume(appVol.ID)
				log.FailOnError(err, "Failed to inspect volume details")

				log.Infof("Get replicas from the volume")
				replicaSets := apiVol.ReplicaSets
				log.Infof("Replica details for the volume: %v, %v", appVol.Name, replicaSets)

				for _, rel := range replicaSets {
					for _, nodeName := range rel.Nodes {
						if _, ok := replicaNodeMap[nodeName]; !ok {
							replicaNodeMap[nodeName], err = node.GetNodeDetailsByNodeID(nodeName)
							log.FailOnError(err, "Failed to get node details for node: %v", nodeName)
						}
					}
				}
			}
		}
		migNamespaces := strings.Join(migrationNamespaces, ",")
		kubeConfigPath := map[int]string{}
		for _, cluster := range []int{asyncdr.FirstCluster, asyncdr.SecondCluster} {
			kubeConfigPath[cluster], err = GetCustomClusterConfigPath(cluster)
			log.FailOnError(err, "Getting error while fetching path for %v cluster, error is %v", cluster, err)
		}

		var migrationSchedName string
		var schdPol *storkapi.SchedulePolicy
		cpName := defaultClusterPairName + time.Now().Format("15h03m05s")
		scpolName := "async-policy"
		migrationInterval := 5
		Step("Create Schedule Policy", func() {
			schdPol, err = asyncdr.CreateSchedulePolicy(scpolName, migrationInterval)
			log.FailOnError(err, "Failed to create schedule policy")
		})

		extraArgs := map[string]string{
			"namespaces":           migNamespaces,
			"kubeconfig":           kubeConfigPath[asyncdr.FirstCluster],
			"schedule-policy-name": schdPol.Name,
		}

		stepLog = "create clusterpair and start migration"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			err = ScheduleBidirectionalClusterPair(cpName, defaultNs, "", storkapi.BackupLocationType(defaultBackupLocation), defaultSecret, "async-dr", asyncdr.FirstCluster, asyncdr.SecondCluster, nil)
			log.FailOnError(err, "Failed creating bidirectional cluster pair")

			log.InfoD("Start migration schedule and perform failover")
			migrationSchedName = migrationSchedKey + time.Now().Format("15h03m05s")

			// create migration schedule
			// once the migration created restart driver on above collected nodes
			err = storkctlcli.ScheduleStorkctlMigrationSched(migrationSchedName, cpName, defaultNs, extraArgs)
			log.FailOnError(err, "Error creating migrationschedule: %v", err)

			stepLog = "Restart the volume driver while migration is in progress"
			Step(stepLog, func() {
				log.InfoD(stepLog)

				restartPX := func() (interface{}, bool, error) {
					migSchedule, err := storkops.Instance().GetMigrationSchedule(migrationSchedName, defaultNs)
					if err != nil {
						return nil, false, err
					}
					migrations := migSchedule.Status.Items["Interval"]
					isRestarted := false
					for _, mig := range migrations {
						log.Infof("Validating migration status for: %v", mig.Name)
						migration, err := storkops.Instance().GetMigration(mig.Name, defaultNs)
						if err != nil {
							return nil, false, err
						}
						if migration.Status.Status == storkapi.MigrationStatusInProgress {
							log.Infof("Migration %s is %v. Restarting PX on replica nodes.", mig.Name, migration.Status.Status)

							for _, selectedNode := range replicaNodeMap {
								log.Infof("Restarting Px on the node: %s", selectedNode)
								err = Inst().V.RestartDriver(selectedNode, nil)
								if err != nil {
									return nil, false, err
								}
								log.Infof("PX restarted successfully on node %v", selectedNode)
							}
							isRestarted = true
							break
						}
					}

					if !isRestarted {
						return nil, true, fmt.Errorf("Migration not in progress. Retrying..")
					}
					return nil, false, nil
				}
				_, err = task.DoRetryWithTimeout(restartPX, migrationRetryTimeout, migrationRetryInterval)
				log.FailOnError(err, "Error occured when restarting px while migration in progress")
			})

			for _, selectedNode := range replicaNodeMap {
				err = Inst().V.WaitDriverUpOnNode(selectedNode, Inst().DriverStartTimeout)
				log.FailOnError(err, "failed to wait for px up on node: %v", selectedNode.Name)
			}

			_, err = storkops.Instance().ValidateMigrationSchedule(migrationSchedName, defaultNs, migrationRetryTimeout, migrationRetryInterval)
			log.FailOnError(err, "Error occured while validating migration schedule %v in the namespace %v", migrationSchedName, defaultNs)
		})

		stNodeClusterMap := make(map[int][]node.Node)

		for cluster, path := range kubeConfigPath {
			if cluster == asyncdr.SecondCluster {
				err = hardSetConfig(path)
				log.FailOnError(err, "Switching context to %v cluster failed", cluster)
				err = SetCustomKubeConfig(asyncdr.SecondCluster)
				log.FailOnError(err, "Switching context to %v cluster failed", cluster)
				err = Inst().S.RefreshNodeRegistry()
				log.FailOnError(err, "Node registry refresh failed")
				err = Inst().V.RefreshDriverEndpoints()
				log.FailOnError(err, "Refresh Driver end points failed")
				stNodeClusterMap[cluster] = node.GetStorageNodes()
				Step("Create Schedule Policy", func() {
					schdPol, err = asyncdr.CreateSchedulePolicy(scpolName, migrationInterval)
					log.FailOnError(err, "Failed to create schedule policy")
				})
			} else {
				stNodeClusterMap[cluster] = node.GetStorageNodes()
			}
		}

		stepLog = "Get stork leader pod and delete"
		Step(stepLog, func() {
			log.InfoD(stepLog)

			err = hardSetConfig(kubeConfigPath[asyncdr.FirstCluster])
			log.FailOnError(err, "Error setting source config: %v", err)
			time.Sleep(time.Second * 30)

			pxNamespace, err := Inst().V.GetVolumeDriverNamespace()
			log.FailOnError(err, "Error occurred while retrieving portworx namespace")
			holderIdentity, err := GetStorkLeaderPodName()
			log.FailOnError(err, "Error occurred while retrieving stork leader pod for configmap")

			stepLog = "Delete stork leader pod"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				err := k8sCore.DeletePod(holderIdentity, pxNamespace, false)
				log.FailOnError(err, "Error occurred while deleting the stork leader pod: %v", holderIdentity)
			})
			log.Infof("Stork leader pod: %v deleted successfully", holderIdentity)

			podList, err := core.Instance().GetPods(pxNamespace, nil)
			log.FailOnError(err, "Error occurred while getting pod list for the namespace: %v", pxNamespace)
			var podListNew = &v1.PodList{}
			for _, pod := range podList.Items {
				if strings.Contains(pod.Name, "stork") {
					podListNew.Items = append(podListNew.Items, pod)
				}
			}
			err = asyncdr.WaitForPodToBeRunning(podListNew)
			log.FailOnError(err, "Error occurred while waiting for pod to be running on the namespace: %v", pxNamespace)
		})

		Step("Perform failover", func() {
			time.Sleep(2 * time.Duration(migrationInterval) * time.Minute)

			log.Infof("Validate migration before failover")
			_, err = storkops.Instance().ValidateMigrationSchedule(migrationSchedName, defaultNs, migrationRetryTimeout, migrationRetryInterval)
			log.FailOnError(err, "Error occured while validating migration schedule")
			log.Infof("Validate migration before failover completed for the migration schedule: %v", migrationSchedName)

			extraArgsFailoverFailback := map[string]string{
				"kubeconfig": kubeConfigPath[asyncdr.SecondCluster],
			}
			failoverParam := failoverFailbackParam{
				action:                    "failover",
				failoverOrFailbackNs:      defaultNs,
				migrationSchedName:        migrationSchedName,
				configPath:                kubeConfigPath[asyncdr.SecondCluster],
				single:                    false,
				skipSourceOp:              false,
				includeNs:                 false,
				excludeNs:                 false,
				extraArgsFailoverFailback: extraArgsFailoverFailback,
				contexts:                  contexts,
			}

			performFailoverFailback(failoverParam)

			err = hardSetConfig(kubeConfigPath[asyncdr.SecondCluster])
			log.FailOnError(err, "Error setting destination config: %v", err)
			extraArgs["kubeconfig"] = kubeConfigPath[asyncdr.SecondCluster]
			newMigSched := migrationSchedName + "-rev"
			createMigSchdAndValidateMigration(newMigSched, cpName, defaultNs, kubeConfigPath[asyncdr.SecondCluster], extraArgs)
		})

		err = SetCustomKubeConfig(asyncdr.FirstCluster)
		log.FailOnError(err, "Switching context to source cluster failed")
		err = Inst().S.RefreshNodeRegistry()
		log.FailOnError(err, "Node registry refresh failed")
		err = Inst().V.RefreshDriverEndpoints()
		log.FailOnError(err, "Refresh Driver end points failed")

		Step("Destroy apps", func() {
			log.InfoD("Destroy apps")
			opts := make(map[string]bool)
			opts[scheduler.OptionsWaitForResourceLeakCleanup] = true
			for _, ctx := range contexts {
				TearDownContext(ctx, opts)
			}
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts)
	})
})

func GetStorkLeaderPodName() (string, error) {
	log.Infof("Get stork leader pod name")
	pxNamespace, err := Inst().V.GetVolumeDriverNamespace()
	if err != nil {
		return "", err
	}
	configMap, err := k8sCore.GetConfigMap("stork", pxNamespace)
	if err != nil {
		return "", err
	}
	log.FailOnError(err, "Error occurred while retrieving ConfigMap for stork")
	storkLeaderData, exists := configMap.Annotations["control-plane.alpha.kubernetes.io/leader"]
	if !exists {
		return "", fmt.Errorf("leader annotation not found in config map")
	}
	log.Infof("Stork leader data: %v", storkLeaderData)

	var leaderInfo map[string]interface{}
	if err := json.Unmarshal([]byte(storkLeaderData), &leaderInfo); err != nil {
		return "", fmt.Errorf("failed to unmarshal leader data: %v", err)
	}
	log.Infof("Stork leader info: %v", leaderInfo)

	holderIdentity, ok := leaderInfo["holderIdentity"].(string)
	if !ok {
		return "", fmt.Errorf("holderIdentity not found or is not a string in leader info")
	}
	log.Infof("Stork leader pod name: %v", holderIdentity)
	return holderIdentity, nil
}

var _ = Describe("{StorkPodsDownOnSourceDuringMigrationInProgress}", Label("staging", "p1", "negative", "AsyncDR"), func() {
	/*
	   https://purestorage.atlassian.net/browse/HAZEL-1050
	   1. Deploy application on source cluster
	   2. Create cluster pair between source and destination cluster
	   3. Create migration schedule and start migration on all namespace
	   4. while migration in progress delete all stork pods on source cluster
	   5. Validate migration
	*/
	var (
		testrailID          = 0
		runID               int
		contexts            []*scheduler.Context
		taskNamePrefix      = "dr-killstork"
		defaultNs           = "kube-system"
		migrationNamespaces []string
		kubeConfigPath      = map[int]string{}
		migrationSchedName  string
		schedulePolicy      *storkapi.SchedulePolicy
		clusterPairName     string
		migNamespaces       string
		schedulePolicyName  = "async-policy"
		migrationInterval   = 5
	)
	BeforeEach(func() {
		if !kubeConfigWritten {
			WriteKubeconfigToFiles()
			kubeConfigWritten = true
		}
	})

	JustBeforeEach(func() {
		StartTorpedoTest("StorkPodsDownOnSourceDuringMigrationInProgress", "Migration of application to destination cluster with stork pods down on source cluster", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})

	stepLog := "Kill stork pods during migration on source cluster and ensure migration completed successfully"
	It(stepLog, func() {
		log.InfoD(stepLog)

		cleanup := func() {
			log.Infof("Perform cleanup task")
			opts := make(map[string]bool)
			opts[SkipClusterScopedObjects] = true
			opts[scheduler.OptionsWaitForResourceLeakCleanup] = true
			opts[scheduler.OptionsWaitForDestroy] = true
			if len(contexts) > 0 {
				for _, ctx := range contexts {
					ctx.SkipVolumeValidation = true
					TearDownContext(ctx, opts)
					ctxNamespace := GetAppNamespace(ctx, "")
					err = Inst().S.DeletePvcsFromNamespace(ctx, ctxNamespace)
					log.FailOnError(err, "Failed to delete the pvcs from namespace  %v", ctxNamespace)
				}
			}
			log.Infof("Remove migration schedule from namespace [%v]", defaultNs)
			migrationSchedules, err := storkops.Instance().ListMigrationSchedules(defaultNs)
			log.FailOnError(err, "Failed to get migration schedule list from the namespace %v", defaultNs)
			for _, migrSched := range migrationSchedules.Items {
				err := asyncdr.DeleteAndWaitForMigrationSchedDeletion(migrSched.Name, defaultNs)
				log.FailOnError(err, "Failed to deleting migration schedule on destination cluster")
			}
			log.Infof("Remove volumes from namespace [%v]", defaultNs)
			volList, err := Inst().V.ListAllVolumes()
			log.FailOnError(err, "Failed to get volume list")
			if len(volList) > 0 {
				for _, volName := range volList {
					err = Inst().V.DeleteVolume(volName)
					log.FailOnError(err, "Failed to delete volume %v list %v", volName, err)
				}
			}
		}
		defer cleanup()

		stepLog = "Scheduling applications and creating a schedule policy for migration"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			migrationNamespaces, contexts = initialSetupApps(taskNamePrefix, false, false)
			migNamespaces = strings.Join(migrationNamespaces, ",")

			for _, cluster := range []int{asyncdr.FirstCluster, asyncdr.SecondCluster} {
				kubeConfigPath[cluster], err = GetCustomClusterConfigPath(cluster)
				log.FailOnError(err, "Getting error while fetching path for %v cluster", cluster)
			}

			schedulePolicy, err = asyncdr.CreateSchedulePolicy(schedulePolicyName, migrationInterval)
			log.FailOnError(err, "Failed to create schedule policy")
		})

		extraArgs := map[string]string{
			"namespaces":           migNamespaces,
			"kubeconfig":           kubeConfigPath[asyncdr.FirstCluster],
			"schedule-policy-name": schedulePolicy.Name,
		}

		stepLog = "Creating cluster pair and starting migration schedule"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			clusterPairName = defaultClusterPairName + time.Now().Format("15h03m05s")
			err = ScheduleBidirectionalClusterPair(clusterPairName, defaultNs, "", storkapi.BackupLocationType(defaultBackupLocation), defaultSecret, "async-dr", asyncdr.FirstCluster, asyncdr.SecondCluster, nil)
			log.FailOnError(err, "Failed creating bidirectional cluster pair")

			log.InfoD("Start migration schedule")
			migrationSchedName = migrationSchedKey + time.Now().Format("15h03m05s")
			err = storkctlcli.ScheduleStorkctlMigrationSched(migrationSchedName, clusterPairName, defaultNs, extraArgs)
			log.FailOnError(err, "Error creating migrationschedule: %v", err)
		})

		stepLog = "Bring down PX Pods on source cluster while migration is in progress"
		Step(stepLog, func() {
			log.InfoD(stepLog)

			deleteStorkPods := func() (interface{}, bool, error) {
				migSchedule, err := storkops.Instance().GetMigrationSchedule(migrationSchedName, defaultNs)
				if err != nil {
					return nil, true, err
				}
				migrations := migSchedule.Status.Items["Interval"]
				for _, mig := range migrations {
					log.Infof("Validating migration status for: %v", mig.Name)
					migration, err := storkops.Instance().GetMigration(mig.Name, defaultNs)
					if err != nil {
						return nil, true, err
					}
					if migration.Status.Status == storkapi.MigrationStatusInProgress {
						log.Infof("Migration [%s] is [%v]. Deleting Stork pods.", mig.Name, migration.Status.Status)
						storkNamespace, err := k8sutils.GetStorkPodNamespace()
						if err != nil {
							return nil, true, err
						}
						err = DeletePodWithWithoutLabelInNamespace(storkNamespace, StorkLabel, false)
						if err != nil {
							return nil, true, err
						}
						err = ValidatePodByLabel(StorkLabel, storkNamespace, 5*time.Minute, 30*time.Second)
						if err != nil {
							return nil, true, err
						}
						log.Infof("Successfully deleted and validated Stork pods in namespace [%s].", storkNamespace)
						return nil, false, nil
					}
				}
				return nil, true, fmt.Errorf("migration not in progress. Retrying...")
			}
			_, err := task.DoRetryWithTimeout(deleteStorkPods, migrationRetryTimeout, migrationRetryInterval)
			log.FailOnError(err, "Error occurred when deleting Stork pods while migration in progress")

			_, err = storkops.Instance().ValidateMigrationSchedule(migrationSchedName, defaultNs, migrationRetryTimeout, migrationRetryInterval)
			log.FailOnError(err, "Error validating migrationschedule [%v] in the namespace [%v] on the source cluster, %v", migrationSchedName, defaultNs, err)
		})

		stepLog = "Performing failover on the destination cluster"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			err := SwitchCluster(kubeConfigPath[asyncdr.SecondCluster], asyncdr.SecondCluster)
			log.FailOnError(err, "Failed to switch cluster")

			extraArgsFailoverFailback := map[string]string{
				"kubeconfig": kubeConfigPath[asyncdr.SecondCluster],
			}
			failoverParam := failoverFailbackParam{
				action:                    "failover",
				failoverOrFailbackNs:      defaultNs,
				migrationSchedName:        migrationSchedName,
				configPath:                kubeConfigPath[asyncdr.SecondCluster],
				single:                    false,
				skipSourceOp:              false,
				includeNs:                 false,
				excludeNs:                 false,
				extraArgsFailoverFailback: extraArgsFailoverFailback,
				contexts:                  contexts,
			}

			performFailoverFailback(failoverParam)
		})

		stepLog = "Destroy applications on the destination cluster"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			cleanup()
		})

		stepLog = "Destroy applications on the source cluster"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			err := SwitchCluster(kubeConfigPath[asyncdr.FirstCluster], asyncdr.FirstCluster)
			log.FailOnError(err, "Failed to switch cluster")
			cleanup()
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

func DeleteStorkPodsDuringMigration(migrationSchedName, namespace string) error {
	log.Infof("Checking migration schedule [%s] for in-progress migrations.", migrationSchedName)

	migSchedule, err := storkops.Instance().GetMigrationSchedule(migrationSchedName, namespace)
	if err != nil {
		return fmt.Errorf("failed to get migration schedule [%s]: %v", migrationSchedName, err)
	}
	migrations := migSchedule.Status.Items["Interval"]
	for _, mig := range migrations {
		log.Infof("Validating migration status for: %v", mig.Name)

		migration, err := storkops.Instance().GetMigration(mig.Name, namespace)
		if err != nil {
			return fmt.Errorf("failed to get migration [%s]: %v", mig.Name, err)
		}
		if migration.Status.Status == storkapi.MigrationStatusInProgress {
			log.Infof("Migration [%s] is in progress. Deleting Stork pods.", mig.Name)
			storkNamespace, err := k8sutils.GetStorkPodNamespace()
			if err != nil {
				return fmt.Errorf("failed to get Stork pod namespace: %v", err)
			}
			err = DeletePodWithWithoutLabelInNamespace(storkNamespace, StorkLabel, false)
			if err != nil {
				return fmt.Errorf("failed to delete Stork pods in namespace [%s]: %v", storkNamespace, err)
			}
			err = ValidatePodByLabel(StorkLabel, storkNamespace, 5*time.Minute, 30*time.Second)
			if err != nil {
				return fmt.Errorf("failed to validate Stork pods in namespace [%s]: %v", storkNamespace, err)
			}
			log.Infof("Successfully deleted and validated Stork pods in namespace [%s].", storkNamespace)
			return nil
		}
	}
	return fmt.Errorf("no migrations in progress for schedule [%s]. Retrying...", migrationSchedName)
}

var _ = Describe("{StorkPodsDownOnDestinationDuringMigrationInProgress}", Label("staging", "p1", "negative", "AsyncDR"), func() {
	/*
	   https://purestorage.atlassian.net/browse/HAZEL-1049
	   1. Deploy application on source cluster
	   2. Create cluster pair between source and destination cluster
	   3. Create migration schedule and start migration on all namespace
	   4. Switch to destination cluster
	   5. while migration in progress delete all stork pods on destination cluster
	   6. Validate migration
	*/
	var (
		testrailID          = 0
		runID               int
		contexts            []*scheduler.Context
		taskNamePrefix      = "dr-killstork"
		defaultNs           = "kube-system"
		migrationNamespaces []string
		kubeConfigPath      = map[int]string{}
		migrationSchedName  string
		schedulePolicy      *storkapi.SchedulePolicy
		clusterPairName     string
		migNamespaces       string
		schedulePolicyName  = "async-policy"
		migrationInterval   = 5
	)
	BeforeEach(func() {
		if !kubeConfigWritten {
			WriteKubeconfigToFiles()
			kubeConfigWritten = true
		}
	})

	JustBeforeEach(func() {
		StartTorpedoTest("StorkPodsDownOnDestinationDuringMigrationInProgress", "Migration of application to destination cluster with stork pods down on source cluster", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})

	stepLog := "Kill stork pods during migration on destination cluster and ensure migration completed successfully"
	It(stepLog, func() {
		log.InfoD(stepLog)

		cleanup := func() {
			log.Infof("Perform cleanup task")
			opts := make(map[string]bool)
			opts[SkipClusterScopedObjects] = true
			opts[scheduler.OptionsWaitForResourceLeakCleanup] = true
			opts[scheduler.OptionsWaitForDestroy] = true
			if len(contexts) > 0 {
				for _, ctx := range contexts {
					ctx.SkipVolumeValidation = true
					TearDownContext(ctx, opts)
					ctxNamespace := GetAppNamespace(ctx, "")
					err = Inst().S.DeletePvcsFromNamespace(ctx, ctxNamespace)
					log.FailOnError(err, "Failed to delete the pvcs from namespace  %v", ctxNamespace)
				}
			}
			log.Infof("Remove migrations from namespace [%v]", defaultNs)
			migrationSchedules, err := storkops.Instance().ListMigrationSchedules(defaultNs)
			log.FailOnError(err, "Failed to get migration schedule list from the namespace %v", defaultNs)
			for _, migrSched := range migrationSchedules.Items {
				err := asyncdr.DeleteAndWaitForMigrationSchedDeletion(migrSched.Name, defaultNs)
				log.FailOnError(err, "Failed to deleting migration schedule on destination cluster")
			}
			volList, err := Inst().V.ListAllVolumes()
			log.FailOnError(err, "Failed to get volume list")
			if len(volList) > 0 {
				for _, volName := range volList {
					err = Inst().V.DeleteVolume(volName)
					log.FailOnError(err, "Failed to delete volume %v list", volName)
				}
			}
		}
		defer cleanup()

		stepLog = "Scheduling applications and creating a schedule policy for migration"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			migrationNamespaces, contexts = initialSetupApps(taskNamePrefix, false, false)
			migNamespaces = strings.Join(migrationNamespaces, ",")

			for _, cluster := range []int{asyncdr.FirstCluster, asyncdr.SecondCluster} {
				kubeConfigPath[cluster], err = GetCustomClusterConfigPath(cluster)
				log.FailOnError(err, "Getting error while fetching path for %v cluster", cluster)
			}

			schedulePolicy, err = asyncdr.CreateSchedulePolicy(schedulePolicyName, migrationInterval)
			log.FailOnError(err, "Failed to create schedule policy")
		})

		extraArgs := map[string]string{
			"namespaces":           migNamespaces,
			"kubeconfig":           kubeConfigPath[asyncdr.FirstCluster],
			"schedule-policy-name": schedulePolicy.Name,
		}

		stepLog = "Creating cluster pair and starting migration schedule"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			clusterPairName = defaultClusterPairName + time.Now().Format("15h03m05s")
			err = ScheduleBidirectionalClusterPair(clusterPairName, defaultNs, "", storkapi.BackupLocationType(defaultBackupLocation), defaultSecret, "async-dr", asyncdr.FirstCluster, asyncdr.SecondCluster, nil)
			log.FailOnError(err, "Failed creating bidirectional cluster pair")

			log.InfoD("Start migration schedule")
			migrationSchedName = migrationSchedKey + time.Now().Format("15h03m05s")
			createMigSchdAndValidateMigration(migrationSchedName, clusterPairName, defaultNs, kubeConfigPath[asyncdr.FirstCluster], extraArgs)
		})

		stepLog = "Bring down stork Pods on destination cluster while migration is in progress"
		Step(stepLog, func() {
			log.InfoD(stepLog)

			deleteStorkPods := func() (interface{}, bool, error) {
				migSchedule, err := storkops.Instance().GetMigrationSchedule(migrationSchedName, defaultNs)
				if err != nil {
					return nil, true, err
				}
				migrations := migSchedule.Status.Items["Interval"]
				for _, mig := range migrations {
					log.Infof("Validating migration status for: %v", mig.Name)
					migration, err := storkops.Instance().GetMigration(mig.Name, defaultNs)
					if err != nil {
						return nil, true, err
					}
					if migration.Status.Status == storkapi.MigrationStatusInProgress {
						log.Infof("Migration [%s] is [%v]. Deleting Stork pods.", mig.Name, migration.Status.Status)
						err := SwitchCluster(kubeConfigPath[asyncdr.SecondCluster], asyncdr.SecondCluster)
						if err != nil {
							return nil, true, err
						}
						storkNamespace, err := k8sutils.GetStorkPodNamespace()
						if err != nil {
							return nil, true, err
						}
						err = DeletePodWithWithoutLabelInNamespace(storkNamespace, StorkLabel, false)
						if err != nil {
							return nil, true, err
						}
						err = ValidatePodByLabel(StorkLabel, storkNamespace, 5*time.Minute, 30*time.Second)
						if err != nil {
							return nil, true, err
						}
						log.Infof("Successfully deleted and validated Stork pods in namespace [%s].", storkNamespace)
						return nil, false, nil
					}
				}
				return nil, true, fmt.Errorf("migration not in progress. Retrying...")
			}
			_, err := task.DoRetryWithTimeout(deleteStorkPods, migrationRetryTimeout, migrationRetryInterval)
			log.FailOnError(err, "Error occurred when deleting Stork pods while migration in progress")

			err = SwitchCluster(kubeConfigPath[asyncdr.FirstCluster], asyncdr.FirstCluster)
			log.FailOnError(err, "Failed to switch destination to source cluster")
			_, err = storkops.Instance().ValidateMigrationSchedule(migrationSchedName, defaultNs, migrationRetryTimeout, migrationRetryInterval)
			log.FailOnError(err, "Failed to validate migration schedule [%v] in the namespace [%v] on the source cluster", migrationSchedName, defaultNs)
		})
		stepLog = "Performing failover on the destination cluster"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			err := SwitchCluster(kubeConfigPath[asyncdr.SecondCluster], asyncdr.SecondCluster)
			log.FailOnError(err, "Failed to switch source to destination cluster")

			extraArgsFailoverFailback := map[string]string{
				"kubeconfig": kubeConfigPath[asyncdr.SecondCluster],
			}
			failoverParam := failoverFailbackParam{
				action:                    "failover",
				failoverOrFailbackNs:      defaultNs,
				migrationSchedName:        migrationSchedName,
				configPath:                kubeConfigPath[asyncdr.SecondCluster],
				single:                    false,
				skipSourceOp:              false,
				includeNs:                 false,
				excludeNs:                 false,
				extraArgsFailoverFailback: extraArgsFailoverFailback,
				contexts:                  contexts,
			}

			performFailoverFailback(failoverParam)
		})

		stepLog = "Destroy applications on the destination cluster"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			cleanup()
		})

		stepLog = "Destroy applications on the source cluster"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			err := SwitchCluster(kubeConfigPath[asyncdr.FirstCluster], asyncdr.FirstCluster)
			log.FailOnError(err, "Failed to switch destination to source cluster")
			cleanup()
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

var _ = Describe("{PXPodsDownOnSourceAndDestinationDuringMigrationInProgress}", Label("staging", "p1", "negative", "AsyncDR"), func() {
	/*
	   https://purestorage.atlassian.net/browse/HAZEL-1052
	   1. Deploy application on source cluster
	   2. Create cluster pair between source and destination cluster
	   3. Create migration schedule and start migration on all namespace
	   4. while migration is in progress delete px pods on source cluster
	   5. wait for px pods to be up and validate migration schedule.
	   6. Switch to destination cluster and do 4 and 5.
	*/
	/*
			**Do not execute this case in the production pipeline.**
		    This issue is currently being investigated. For more details, please refer to the ongoing issue:
		    [PWX-40878](https://purestorage.atlassian.net/browse/PWX-40878).

	*/
	var (
		testrailID             = 0
		runID                  int
		contexts               []*scheduler.Context
		taskNamePrefix         = "dr-pxdown"
		defaultNs              = "kube-system"
		migrationNamespaces    []string
		kubeConfigPath         = map[int]string{}
		migrationSchedName     string
		destMigrationSchedName string
		schedulePolicy         *storkapi.SchedulePolicy
		clusterPairName        string
		migNamespaces          string
		schedulePolicyName     = "async-policy"
		migrationInterval      = 5
	)
	BeforeEach(func() {
		if !kubeConfigWritten {
			WriteKubeconfigToFiles()
			kubeConfigWritten = true
		}
	})

	JustBeforeEach(func() {
		StartTorpedoTest("PXPodsDownOnSourceAndDestinationDuringMigrationInProgress", "Migration of application to destination cluster with px pods down on source and destination cluster", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})

	stepLog := "Bring down PX pods during migration on source and destination cluster and ensure migration completed successfully"
	It(stepLog, func() {
		log.InfoD(stepLog)

		cleanup := func() {
			log.Infof("Performing cleanup task")
			opts := make(map[string]bool)
			opts[SkipClusterScopedObjects] = true
			opts[scheduler.OptionsWaitForResourceLeakCleanup] = true
			opts[scheduler.OptionsWaitForDestroy] = true
			if len(contexts) > 0 {
				for _, ctx := range contexts {
					ctx.SkipVolumeValidation = true
					TearDownContext(ctx, opts)
					ctxNamespace := GetAppNamespace(ctx, "")
					err = Inst().S.DeletePvcsFromNamespace(ctx, ctxNamespace)
					log.FailOnError(err, "Failed to delete the pvcs from namespace  %v", ctxNamespace)
				}
			}
			volList, err := Inst().V.ListAllVolumes()
			log.FailOnError(err, "Failed to get volume list")
			if len(volList) > 0 {
				for _, volName := range volList {
					err = Inst().V.DeleteVolume(volName)
					log.FailOnError(err, "Failed to delete volume %v list", volName)
				}
			}
		}
		defer cleanup()

		stepLog = "Scheduling applications and creating a schedule policy for migration"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			migrationNamespaces, contexts = initialSetupApps(taskNamePrefix, false, false)
			migNamespaces = strings.Join(migrationNamespaces, ",")

			for _, cluster := range []int{asyncdr.FirstCluster, asyncdr.SecondCluster} {
				kubeConfigPath[cluster], err = GetCustomClusterConfigPath(cluster)
				log.FailOnError(err, "Getting error while fetching path for %v cluster", cluster)
			}

			schedulePolicy, err = asyncdr.CreateSchedulePolicy(schedulePolicyName, migrationInterval)
			log.FailOnError(err, "Failed to create schedule policy")
		})

		extraArgs := map[string]string{
			"namespaces":           migNamespaces,
			"kubeconfig":           kubeConfigPath[asyncdr.FirstCluster],
			"schedule-policy-name": schedulePolicy.Name,
		}

		stepLog = "Creating cluster pair and starting migration schedule"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			clusterPairName = defaultClusterPairName + time.Now().Format("15h03m05s")
			err = ScheduleBidirectionalClusterPair(clusterPairName, defaultNs, "", storkapi.BackupLocationType(defaultBackupLocation), defaultSecret, "async-dr", asyncdr.FirstCluster, asyncdr.SecondCluster, nil)
			log.FailOnError(err, "Failed creating bidirectional cluster pair")

			log.InfoD("Start migration schedule")
			migrationSchedName = migrationSchedKey + time.Now().Format("15h03m05s")
			err = storkctlcli.ScheduleStorkctlMigrationSched(migrationSchedName, clusterPairName, defaultNs, extraArgs)
			log.FailOnError(err, "Error creating migrationschedule: %v", err)
		})

		stepLog = "Bring down PX Pods on source and destination clusters while migration is in progress"
		Step(stepLog, func() {
			log.InfoD(stepLog)

			isMigrationInProgress := func() (interface{}, bool, error) {
				isInprogress, err := IsMigrationInProgress(migrationSchedName, defaultNs)
				if err != nil {
					return nil, true, err
				}
				if isInprogress {
					return nil, false, nil
				}
				return nil, true, fmt.Errorf("retrying: no migration is in progress")
			}
			_, err = task.DoRetryWithTimeout(isMigrationInProgress, migrationRetryTimeout, migrationRetryInterval)
			log.FailOnError(err, "Failed to deleting px pods while migration in progress on source cluster")

			err = DeletePXOnSourceAndDestination(kubeConfigPath[asyncdr.SecondCluster], asyncdr.SecondCluster)

			err = SwitchCluster(kubeConfigPath[asyncdr.FirstCluster], asyncdr.FirstCluster)
			log.FailOnError(err, "Failed to switch cluster")

			_, err = storkops.Instance().ValidateMigrationSchedule(migrationSchedName, defaultNs, migrationRetryTimeout, migrationRetryInterval)
			log.FailOnError(err, "Failed to validate migration schedule [%v] in the namespace [%v] on the source cluster", migrationSchedName, defaultNs)
		})

		stepLog = "Performing failover"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			err = SwitchCluster(kubeConfigPath[asyncdr.SecondCluster], asyncdr.SecondCluster)
			log.FailOnError(err, "Failed to switch cluster")
			extraArgsFailoverFailback := map[string]string{
				"kubeconfig": kubeConfigPath[asyncdr.SecondCluster],
			}
			failoverParam := failoverFailbackParam{
				action:                    "failover",
				failoverOrFailbackNs:      defaultNs,
				migrationSchedName:        migrationSchedName,
				configPath:                kubeConfigPath[asyncdr.SecondCluster],
				single:                    false,
				skipSourceOp:              true,
				includeNs:                 false,
				excludeNs:                 false,
				extraArgsFailoverFailback: extraArgsFailoverFailback,
				contexts:                  contexts,
			}

			performFailoverFailback(failoverParam)
			extraArgs["kubeconfig"] = kubeConfigPath[asyncdr.SecondCluster]
			destMigrationSchedName = migrationSchedName + "-rev"
			err = storkctlcli.ScheduleStorkctlMigrationSched(destMigrationSchedName, clusterPairName, defaultNs, extraArgs)
			log.FailOnError(err, "Error creating migrationschedule: [%v] on destination cluster", destMigrationSchedName)
		})

		stepLog = "Bring down PX Pods on source and destination cluster while migration is in progress"
		Step(stepLog, func() {
			log.InfoD(stepLog)

			isMigrationInProgress := func() (interface{}, bool, error) {
				isInprogress, err := IsMigrationInProgress(destMigrationSchedName, defaultNs)
				if err != nil {
					return nil, true, err
				}
				if isInprogress {
					return nil, false, nil
				}
				return nil, true, fmt.Errorf("retrying: no migration is in progress")
			}
			_, err = task.DoRetryWithTimeout(isMigrationInProgress, migrationRetryTimeout, migrationRetryInterval)
			log.FailOnError(err, "Failed to deleting px pods while migration in progress on source cluster")

			err = SwitchCluster(kubeConfigPath[asyncdr.FirstCluster], asyncdr.FirstCluster)
			log.FailOnError(err, "Failed to switch cluster")

			err = DeletePXOnSourceAndDestination(kubeConfigPath[asyncdr.SecondCluster], asyncdr.SecondCluster)

			_, err = storkops.Instance().ValidateMigrationSchedule(destMigrationSchedName, defaultNs, migrationRetryTimeout, migrationRetryInterval)
			log.FailOnError(err, "Failed to validate migration schedule [%v] in the namespace [%v] on the source cluster", destMigrationSchedName, defaultNs)
		})

		stepLog = "Performing failback"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			extraArgsFailback := map[string]string{
				"kubeconfig": kubeConfigPath[asyncdr.SecondCluster],
			}

			failback := failoverFailbackParam{
				action:                    "failback",
				failoverOrFailbackNs:      defaultNs,
				migrationSchedName:        destMigrationSchedName,
				configPath:                kubeConfigPath[asyncdr.SecondCluster],
				single:                    false,
				skipSourceOp:              false,
				includeNs:                 false,
				excludeNs:                 false,
				extraArgsFailoverFailback: extraArgsFailback,
				contexts:                  contexts,
			}
			performFailoverFailback(failback)
		})

		stepLog = "Destroy applications on the destination cluster"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			cleanup()
		})

		stepLog = "Destroy applications on the source cluster"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			err := SwitchCluster(kubeConfigPath[asyncdr.SecondCluster], asyncdr.SecondCluster)
			log.FailOnError(err, "Failed to switch cluster")
			cleanup()
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

func IsMigrationInProgress(migrationSchedName, namespace string) (bool, error) {
	migSchedule, err := storkops.Instance().GetMigrationSchedule(migrationSchedName, namespace)
	if err != nil {
		return false, err
	}
	migrations := migSchedule.Status.Items["Interval"]
	for _, mig := range migrations {
		log.Infof("Validating migration status for: %v", mig.Name)
		migration, err := storkops.Instance().GetMigration(mig.Name, namespace)
		if err != nil {
			return false, err
		}
		if migration.Status.Status == storkapi.MigrationStatusInProgress {
			log.Infof("Migration name: [%s], Migration current status: [%v]", mig.Name, migration.Status.Status)
			return true, nil
		}
	}
	return false, nil
}

func SwitchCluster(configPath string, cluster int) error {
	err := hardSetConfig(configPath)
	if err != nil {
		return fmt.Errorf("error setting cluster %v config: %w", cluster, err)
	}

	err = SetCustomKubeConfig(cluster)
	if err != nil {
		return fmt.Errorf("switching context to cluster %v failed: %w", cluster, err)
	}

	err = Inst().S.RefreshNodeRegistry()
	if err != nil {
		return fmt.Errorf("node registry refresh failed: %w", err)
	}

	err = Inst().V.RefreshDriverEndpoints()
	if err != nil {
		return fmt.Errorf("refreshing driver endpoints failed: %w", err)
	}

	log.Infof("Successfully switched to cluster %v", cluster)
	return nil
}

func DeletePXOnSourceAndDestination(configPath string, cluster int) error {
	var pxNamespace string
	pxNamespace, err = Inst().V.GetVolumeDriverNamespace()
	if err != nil {
		return err
	}
	err = DeletePXPods(pxNamespace)
	if err != nil {
		return err
	}

	for _, n := range node.GetStorageDriverNodes() {
		err := Inst().V.WaitForPxPodsToBeUp(n)
		if err != nil {
			return err
		}
	}

	err := SwitchCluster(configPath, cluster)
	if err != nil {
		return err
	}

	pxNamespace, err = Inst().V.GetVolumeDriverNamespace()
	if err != nil {
		return err
	}
	err = DeletePXPods(pxNamespace)
	if err != nil {
		return err
	}

	log.InfoD("Waiting for PX Nodes to be up")
	for _, n := range node.GetStorageDriverNodes() {
		err := Inst().V.WaitForPxPodsToBeUp(n)
		if err != nil {
			return err
		}
	}
	return nil
}

var _ = Describe("{AsyncDRFailoverWithSourceClusterDown}", Label("staging", "p1", "negative", "AsyncDR"), func() {
	/*
		https://purestorage.atlassian.net/browse/HAZEL-1047
		1. Create apps
		2. Bring down Source Cluster (Restart all worker nodes)
		3. Trigger Failover
		4. Validate failover
	*/
	var (
		testrailID          = 0
		runID               int
		contexts            []*scheduler.Context
		taskNamePrefix      = "dr-sourcedown"
		defaultNs           = "kube-system"
		migrationNamespaces []string
		kubeConfigPath      = map[int]string{}
		migrationSchedName  string
		schedulePolicy      *storkapi.SchedulePolicy
		clusterPairName     string
		migNamespaces       string
		schedulePolicyName  = "async-policy"
		migrationInterval   = 5
		wg                  sync.WaitGroup
	)
	BeforeEach(func() {
		if !kubeConfigWritten {
			WriteKubeconfigToFiles()
			kubeConfigWritten = true
		}
	})

	JustBeforeEach(func() {
		StartTorpedoTest("AsyncDRFailoverWithSourceClusterDown", "Perform failover with source cluster down", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})

	stepLog := "Restart all worker nodes in the source cluster, perform failover, and validate the operation"
	It(stepLog, func() {
		log.InfoD(stepLog)

		cleanup := func() {
			log.Infof("Perform cleanup task")
			if len(contexts) > 0 {
				for _, ctx := range contexts {
					ctx.SkipVolumeValidation = true
					TearDownContext(ctx, map[string]bool{
						SkipClusterScopedObjects:                    true,
						scheduler.OptionsWaitForResourceLeakCleanup: true,
						scheduler.OptionsWaitForDestroy:             true,
					})
				}
			}
			log.Infof("Remove migrations from namespace [%v]", defaultNs)
			migrationSchedules, err := storkops.Instance().ListMigrationSchedules(defaultNs)
			log.FailOnError(err, "Failed to get migration schedule list from the namespace %v", defaultNs)
			for _, migrSched := range migrationSchedules.Items {
				err := asyncdr.DeleteAndWaitForMigrationSchedDeletion(migrSched.Name, defaultNs)
				log.FailOnError(err, "Failed to deleting migration schedule on destination cluster")
			}
			log.Infof("Remove volumes")
			volList, err := Inst().V.ListAllVolumes()
			log.FailOnError(err, "Failed to get volume list")
			if len(volList) > 0 {
				for _, volName := range volList {
					_ = Inst().V.DeleteVolume(volName)
				}
			}
		}
		defer cleanup()

		stepLog = "Scheduling applications and creating a schedule policy for migration"
		Step(stepLog, func() {
			log.InfoD(stepLog)

			migrationNamespaces, contexts = initialSetupApps(taskNamePrefix, false, false)
			migNamespaces = strings.Join(migrationNamespaces, ",")

			for _, cluster := range []int{asyncdr.FirstCluster, asyncdr.SecondCluster} {
				kubeConfigPath[cluster], err = GetCustomClusterConfigPath(cluster)
				log.FailOnError(err, "Getting error while fetching path for %v cluster", cluster)
			}

			schedulePolicy, err = asyncdr.CreateSchedulePolicy(schedulePolicyName, migrationInterval)
			log.FailOnError(err, "Failed to create schedule policy")
		})

		extraArgs := map[string]string{
			"namespaces":           migNamespaces,
			"kubeconfig":           kubeConfigPath[asyncdr.FirstCluster],
			"schedule-policy-name": schedulePolicy.Name,
		}

		stepLog = "Creating cluster pair and starting migration schedule"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			clusterPairName = defaultClusterPairName + time.Now().Format("15h03m05s")
			err = ScheduleBidirectionalClusterPair(clusterPairName, defaultNs, "", storkapi.BackupLocationType(defaultBackupLocation), defaultSecret, "async-dr", asyncdr.FirstCluster, asyncdr.SecondCluster, nil)
			log.FailOnError(err, "Failed creating bidirectional cluster pair")

			log.InfoD("Start migration schedule")
			migrationSchedName = migrationSchedKey + time.Now().Format("15h03m05s")
			createMigSchdAndValidateMigration(migrationSchedName, clusterPairName, defaultNs, kubeConfigPath[asyncdr.FirstCluster], extraArgs)
		})

		stepLog = "Restart all worker nodes in the source cluster"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			nodesToRestartPx := node.GetStorageNodes()
			for _, eachNode := range nodesToRestartPx {
				wg.Add(1)
				go func(n node.Node) {
					defer wg.Done()
					log.Infof("Restarting node: %v", n.Name)

					err := Inst().N.RebootNode(n, node.RebootNodeOpts{
						Force: true,
						ConnectionOpts: node.ConnectionOpts{
							Timeout:         1 * time.Minute,
							TimeBeforeRetry: 5 * time.Second,
						},
					})
					log.FailOnError(err, "Failed to reboot node [%v]", n.Name)
				}(eachNode)
			}
			wg.Wait()
			log.InfoD("Node restart Initiated on all storage nodes")
		})

		stepLog = "Performing failover on the destination cluster"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			err := SwitchCluster(kubeConfigPath[asyncdr.SecondCluster], asyncdr.SecondCluster)
			log.FailOnError(err, "Failed to switch source to destination cluster")

			extraArgsFailoverFailback := map[string]string{
				"kubeconfig": kubeConfigPath[asyncdr.SecondCluster],
			}
			failoverParam := failoverFailbackParam{
				action:                    "failover",
				failoverOrFailbackNs:      defaultNs,
				migrationSchedName:        migrationSchedName,
				configPath:                kubeConfigPath[asyncdr.SecondCluster],
				single:                    false,
				skipSourceOp:              false,
				includeNs:                 false,
				excludeNs:                 false,
				extraArgsFailoverFailback: extraArgsFailoverFailback,
				contexts:                  contexts,
			}
			performFailoverFailback(failoverParam)
		})

		stepLog = "Destroy applications on the destination cluster"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			cleanup()
		})

		stepLog = "Destroy applications on the source cluster"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			err := SwitchCluster(kubeConfigPath[asyncdr.FirstCluster], asyncdr.FirstCluster)
			log.FailOnError(err, "Failed to switch destination to source cluster")
			stNodes := node.GetStorageNodes()
			for _, eachNode := range stNodes {
				err = Inst().V.WaitDriverUpOnNode(eachNode, Inst().DriverStartTimeout)
				log.FailOnError(err, "Failed to wait for px up on node %v", eachNode.Name)
			}
			cleanup()
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})
