package tests

import (
	"fmt"
	"strings"
	"time"

	"github.com/libopenstorage/openstorage/api"
	pxapi "github.com/pure-px/px-operator/api/px"
	opcorev1 "github.com/pure-px/px-operator/pkg/apis/core/v1"
	"github.com/pure-px/sched-ops/k8s/operator"
	"github.com/pure-px/torpedo/drivers/node"
	"github.com/pure-px/torpedo/drivers/scheduler"
	"github.com/pure-px/torpedo/pkg/ipv6util"
	"github.com/pure-px/torpedo/pkg/log"
	"github.com/pure-px/torpedo/pkg/pureutils"
	"github.com/pure-px/torpedo/pkg/testrailuttils"
	"golang.org/x/net/context"

	// v1 "k8s.io/api/core/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	apmv1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/pure-px/torpedo/tests"
)

const (
	defaultReadynessTimeout = 2 * time.Minute

	PureSecretNamespace = "kube-system"
	pureSecretDataField = "pure.json"
	expiredLicString    = "License is expired"

	essentialsFaFbSKU = "Portworx CSI for FA/FB"
	pxEssentials      = "PX-Essential"
)

var (
	essentialLicense = map[LabLabel]interface{}{
		LabNodes:                 &pxapi.LicensedFeature_Count{Count: 5},
		LabVolumes:               &pxapi.LicensedFeature_Count{Count: 200},
		LabVolumeSize:            &pxapi.LicensedFeature_CapacityTb{CapacityTb: 1},
		LabNodeCapacity:          &pxapi.LicensedFeature_CapacityTb{CapacityTb: 1},
		LabNodeCapacityExtend:    &pxapi.LicensedFeature_Enabled{Enabled: false}, //feature upgrade needed
		LabSnapshots:             &pxapi.LicensedFeature_Count{Count: 5},
		LabLocalAttaches:         &pxapi.LicensedFeature_Count{Count: 30},
		LabAggregatedVol:         &pxapi.LicensedFeature_Enabled{Enabled: false}, //feature upgrade needed
		LabSharedVol:             &pxapi.LicensedFeature_Enabled{Enabled: true},
		LabScaledVol:             &pxapi.LicensedFeature_Enabled{Enabled: true},
		LabEncryptedVol:          &pxapi.LicensedFeature_Enabled{Enabled: true},
		LabGlobalSecretsOnly:     &pxapi.LicensedFeature_Enabled{Enabled: true},
		LabResizeVolume:          &pxapi.LicensedFeature_Enabled{Enabled: true},
		LabFastPath:              &pxapi.LicensedFeature_Enabled{Enabled: false}, //feature upgrade needed
		LabCloudSnap:             &pxapi.LicensedFeature_Enabled{Enabled: true},
		LabCloudSnapDaily:        &pxapi.LicensedFeature_Count{Count: 1},
		LabCloudMigration:        &pxapi.LicensedFeature_Enabled{Enabled: false}, //feature upgrade needed
		LabDisasterRecovery:      &pxapi.LicensedFeature_Enabled{Enabled: false}, //feature upgrade needed
		LabAUTCapacityMgmt:       &pxapi.LicensedFeature_Enabled{Enabled: false}, //feature upgrade needed
		LabOIDCSecurity:          &pxapi.LicensedFeature_Enabled{Enabled: false}, //feature upgrade needed
		LabPlatformBare:          &pxapi.LicensedFeature_Enabled{Enabled: true},
		LabPlatformVM:            &pxapi.LicensedFeature_Enabled{Enabled: true},
		LabMultiTenantFlashArray: &pxapi.LicensedFeature_Enabled{Enabled: false}, //feature upgrade needed
	}
)

var (
	faLicense = map[LabLabel]interface{}{
		LabNodes:              &pxapi.LicensedFeature_Count{Count: 1000},
		LabVolumeSize:         &pxapi.LicensedFeature_CapacityTb{CapacityTb: 40},
		LabVolumes:            &pxapi.LicensedFeature_Count{Count: 200},
		LabHaLevel:            &pxapi.LicensedFeature_Count{Count: MaxHaLevel},
		LabSnapshots:          &pxapi.LicensedFeature_Count{Count: 5},
		LabAggregatedVol:      &pxapi.LicensedFeature_Enabled{Enabled: false},
		LabSharedVol:          &pxapi.LicensedFeature_Enabled{Enabled: true},
		LabEncryptedVol:       &pxapi.LicensedFeature_Enabled{Enabled: true},
		LabGlobalSecretsOnly:  &pxapi.LicensedFeature_Enabled{Enabled: true},
		LabScaledVol:          &pxapi.LicensedFeature_Enabled{Enabled: true},
		LabResizeVolume:       &pxapi.LicensedFeature_Enabled{Enabled: true},
		LabCloudSnap:          &pxapi.LicensedFeature_Enabled{Enabled: true},
		LabCloudSnapDaily:     &pxapi.LicensedFeature_Count{Count: 1},
		LabCloudMigration:     &pxapi.LicensedFeature_Enabled{Enabled: false},
		LabDisasterRecovery:   &pxapi.LicensedFeature_Enabled{Enabled: false},
		LabPlatformBare:       &pxapi.LicensedFeature_Enabled{Enabled: true},
		LabPlatformVM:         &pxapi.LicensedFeature_Enabled{Enabled: true},
		LabNodeCapacity:       &pxapi.LicensedFeature_CapacityTb{CapacityTb: MaxNodeCapacity},
		LabNodeCapacityExtend: &pxapi.LicensedFeature_Enabled{Enabled: true},
		LabLocalAttaches:      &pxapi.LicensedFeature_Count{Count: 128},
		LabOIDCSecurity:       &pxapi.LicensedFeature_Enabled{Enabled: false},
		LabAUTCapacityMgmt:    &pxapi.LicensedFeature_Enabled{Enabled: false},
	}
)

// This test performs basic test of starting an application and destroying it (along with storage)
var _ = Describe("{BasicEssentialsFaFbTest}", Label("p2", "positive", "license"), func() {
	var testrailID = 56354
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/56354
	var runID int
	JustBeforeEach(func() {
		StartTorpedoTest("BasicEssentialsFaFbTest", "Validates `Portworx CSI for FA/FB` license SKU", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})
	var contexts []*scheduler.Context

	It("has to setup, validate and teardown apps", func() {
		contexts = make([]*scheduler.Context, 0)

		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("setupteardown-license-%d", i))...)
		}

		ValidateApplications(contexts)

		Step("Get SKU and compare with PX-Essentials FA/FB", func() {
			summary, err := Inst().V.GetLicenseSummary()
			Expect(err).NotTo(HaveOccurred(),
				fmt.Sprintf("Failed to get license SKU. Error: [%v]", err))

			Expect(summary.SKU).To(Equal(essentialsFaFbSKU),
				fmt.Sprintf("SKU did not match: [%v]", essentialsFaFbSKU))

			Step("Compare PX-Essentials FA/FB features vs activated license", func() {
				for _, feature := range summary.Features {
					// if the feature limit exists in the hardcoded license limits we test it.
					if _, ok := faLicense[LabLabel(feature.Name)]; ok {
						Expect(feature.Quantity).To(Equal(faLicense[LabLabel(feature.Name)]),
							fmt.Sprintf("%v did not match: [%v]", feature.Quantity, faLicense[LabLabel(feature.Name)]))
					}
				}
			})
		})
		ValidateAndDestroy(contexts, nil)
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

// This test performs basic reboot test of starting an application and destroying it (along with storage)
var _ = Describe("{BasicEssentialsRebootTest}", Label("p1", "negative", "license", "node_reboot"), func() {
	var testrailID = 56356
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/56356
	var runID int
	JustBeforeEach(func() {
		StartTorpedoTest("BasicEssentialsRebootTest", "Validates `Portworx CSI for FA/FB` remains active after reboot", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})
	var err error
	var contexts []*scheduler.Context

	It("has to setup, validate and teardown apps", func() {
		contexts = make([]*scheduler.Context, 0)

		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("setupteardown-license-reboot-%d", i))...)
		}

		ValidateApplications(contexts)

		Step("get all nodes and reboot one by one", func() {
			nodesToReboot := node.GetWorkerNodes()

			// Reboot node and check driver status
			Step(fmt.Sprintf("reboot node one at a time from the node(s): %v", nodesToReboot), func() {
				for _, n := range nodesToReboot {
					if n.IsStorageDriverInstalled {
						Step(fmt.Sprintf("reboot node: %s", n.Name), func() {
							err = Inst().N.RebootNode(n, node.RebootNodeOpts{
								Force: true,
								ConnectionOpts: node.ConnectionOpts{
									Timeout:         defaultCommandTimeout,
									TimeBeforeRetry: defaultCommandRetry,
								},
							})
							Expect(err).NotTo(HaveOccurred())
						})

						Step(fmt.Sprintf("wait for node: %s to be back up", n.Name), func() {
							err = Inst().N.TestConnection(n, node.ConnectionOpts{
								Timeout:         defaultTestConnectionTimeout,
								TimeBeforeRetry: defaultWaitRebootRetry,
							})
							Expect(err).NotTo(HaveOccurred())
						})

						Step(fmt.Sprintf("wait for volume driver to stop on node: %v", n.Name), func() {
							err := Inst().V.WaitDriverDownOnNode(n)
							Expect(err).NotTo(HaveOccurred())
						})

						Step(fmt.Sprintf("wait to scheduler: %s and volume driver: %s to start",
							Inst().S.String(), Inst().V.String()), func() {

							err = Inst().S.IsNodeReady(n)
							Expect(err).NotTo(HaveOccurred())

							err = Inst().V.WaitDriverUpOnNode(n, Inst().DriverStartTimeout)
							Expect(err).NotTo(HaveOccurred())
						})

						Step("validate apps", func() {
							for _, ctx := range contexts {
								ValidateContext(ctx)
							}
						})
					}
				}
			})
		})

		Step("Get SKU and compare with PX-Essentials FA/FB", func() {
			summary, err := Inst().V.GetLicenseSummary()
			Expect(err).NotTo(HaveOccurred(),
				fmt.Sprintf("Failed to get license SKU. Error: [%v]", err))

			Expect(summary.SKU).To(Equal(essentialsFaFbSKU),
				fmt.Sprintf("SKU did not match: [%v] with [%v]",
					summary.SKU, essentialsFaFbSKU))
		})
		ValidateAndDestroy(contexts, nil)
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

// This test performs basic limit test of starting an application and destroying it (along with storage)
var _ = Describe("{BasicEssentialsAggrSnapLimitTest}", Label("p1", "positive", "license"), func() {
	var testrailID = 56355
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/56355
	var runID int
	JustBeforeEach(func() {
		StartTorpedoTest("BasicEssentialsAggrSnapLimitTest", "Validates `Portworx CSI for FA/FB` lic's limits ", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})
	var contexts []*scheduler.Context

	It("has to setup, validate and teardown apps", func() {
		contexts = make([]*scheduler.Context, 0)

		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("license-aggrsnaplimit-%d", i))...)
		}
		appScaleFactor := time.Duration(Inst().GlobalScaleFactor)
		for _, ctx := range contexts {
			if strings.Contains(ctx.App.Key, "snap") || strings.Contains(ctx.App.Key, "aggr") {
				Step(fmt.Sprintf("Expect volume validation for %s app to fail", ctx.App.Key), func() {
					err := Inst().S.ValidateVolumes(ctx, appScaleFactor*defaultReadynessTimeout, defaultRetryInterval, &scheduler.VolumeOptions{ExpectError: false})
					Expect(err).To(HaveOccurred(),
						fmt.Sprintf("No error occurred while validating storage for app [%s]", ctx.App.Key))
				})
			} else {
				Step(fmt.Sprintf("Expect volume validation for %s app to pass", ctx.App.Key), func() {
					err := Inst().S.ValidateVolumes(ctx, appScaleFactor*defaultReadynessTimeout, defaultRetryInterval, &scheduler.VolumeOptions{ExpectError: false})
					Expect(err).ToNot(HaveOccurred(),
						fmt.Sprintf("Error occurred during validating storage for app [%s]. Error: %v", ctx.App.Key, err))
				})
			}
			// If we are running the mysql-aggr test execute next steps.
			if strings.Contains(ctx.App.Key, "snap") || strings.Contains(ctx.App.Key, "aggr") {
				Step(fmt.Sprintf("Expect %s app to fail to start", ctx.App.Key), func() {
					err := Inst().S.WaitForRunning(ctx, appScaleFactor*defaultReadynessTimeout, defaultRetryInterval)
					Expect(err).To(HaveOccurred(),
						"app with aggregated volumes got deployed successfully when lic does not allow aggregated volumes.")
				})
			} else {
				Step(fmt.Sprintf("Wait for %s app to start running", ctx.App.Key), func() {
					err := Inst().S.WaitForRunning(ctx, appScaleFactor*defaultReadynessTimeout, defaultRetryInterval)
					Expect(err).NotTo(HaveOccurred())
				})
			}
		}
	})

	for _, ctx := range contexts {
		TearDownContext(ctx, nil)
	}

	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

/*
	This test

1. Deletes px-pure-secret
2. Waits for lic_expiry_timeout
3. Verifies that Essentials lic expires
4. Re-creates px-pure-secret
5. Waits for next metering interval
6. Verifies that Essentials lic gets renewed again
*/
var _ = Describe("{DeleteSecretLicExpiryAndRenewal}", Label("p1", "negative", "license"), func() {
	var testrailID = 56357
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/56357
	var runID int
	JustBeforeEach(func() {
		StartTorpedoTest("DeleteSecretLicExpiryAndRenewal", "Validates lic expires if `px-pure-secret` is deleted and it gets renewed when secret is re-created", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})
	var contexts []*scheduler.Context
	var pureSecretJSON string

	It("has to setup, validate and teardown apps", func() {
		contexts = make([]*scheduler.Context, 0)

		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("delseclicexprenewal-%d", i))...)
		}

		ValidateApplications(contexts)

		Step("Get SKU and compare with PX-Essentials FA/FB", func() {
			summary, err := Inst().V.GetLicenseSummary()
			Expect(err).NotTo(HaveOccurred(),
				fmt.Sprintf("Failed to get license SKU. Error: [%v]", err))

			Expect(summary.SKU).To(Equal(essentialsFaFbSKU),
				fmt.Sprintf("SKU did not match: [%v]", essentialsFaFbSKU))

			Step("Compare PX-Essentials FA/FB features vs activated license", func() {
				for _, feature := range summary.Features {
					// if the feature limit exists in the hardcoded license limits we test it.
					if _, ok := faLicense[LabLabel(feature.Name)]; ok {
						Expect(feature.Quantity).To(Equal(faLicense[LabLabel(feature.Name)]),
							fmt.Sprintf("%v did not match: [%v]", feature.Quantity, faLicense[LabLabel(feature.Name)]))
					}
				}
			})
		})

		Step("Fetch and store Pure secret", func() {
			var err error
			pureSecretJSON, err = Inst().S.GetSecretData(PureSecretNamespace, PureSecretName, pureSecretDataField)
			Expect(err).NotTo(HaveOccurred(),
				fmt.Sprintf("Failed to fetch secret [%s] in [%s] namespace. Error: [%v]",
					PureSecretName, PureSecretNamespace, err))
		})

		Step("Delete Pure secret", func() {
			err := Inst().S.DeleteSecret(PureSecretNamespace, PureSecretName)
			Expect(err).NotTo(HaveOccurred(),
				fmt.Sprintf("Failed to delete secret [%s] in [%s] namespace. Error: [%v]",
					PureSecretName, PureSecretNamespace, err))
		})

		Step(fmt.Sprintf("Wait for license expiry timeout of [%v]",
			Inst().LicenseExpiryTimeoutHours), func() {
			SleepWithContext(context.Background(), Inst().LicenseExpiryTimeoutHours)
			// Additional sleep to wait for lic to get expired on all nodes
			SleepWithContext(context.Background(), 10*time.Minute)
		})

		Step("Verify license is expired", func() {
			summary, err := Inst().V.GetLicenseSummary()
			Expect(err).NotTo(HaveOccurred(),
				fmt.Sprintf("Failed to get license SKU. Error: [%v]", err))
			Expect(summary.SKU).To(Equal(essentialsFaFbSKU),
				fmt.Sprintf("SKU did not match: [%v]", essentialsFaFbSKU))
			Expect(summary.LicenesConditionMsg).To(ContainSubstring(expiredLicString),
				fmt.Sprintf("License did not expire after deleting [%s] secret", PureSecretName))
		})

		Step("Re-create Pure secret", func() {
			err := Inst().S.CreateSecret(PureSecretNamespace, PureSecretName, pureSecretDataField, pureSecretJSON)
			Expect(err).NotTo(HaveOccurred(),
				fmt.Sprintf("Failed to create secret [%s] in [%s] namespace. Error: [%v]",
					PureSecretName, PureSecretNamespace, err))
		})

		Step(fmt.Sprintf("Wait for next metering interval which is going to happen in [%v]",
			Inst().MeteringIntervalMins), func() {
			SleepWithContext(context.Background(), Inst().MeteringIntervalMins)
			// Additional sleep to wait for lic to get renewed on all nodes
			SleepWithContext(context.Background(), 5*time.Minute)
		})

		Step("Verify correct license got re-activated for PX-Essentials FA/FB", func() {
			summary, err := Inst().V.GetLicenseSummary()
			Expect(err).NotTo(HaveOccurred(),
				fmt.Sprintf("Failed to get license SKU. Error: [%v]", err))

			Expect(summary.SKU).To(Equal(essentialsFaFbSKU),
				fmt.Sprintf("SKU did not match: [%v]", essentialsFaFbSKU))

			Expect(summary.LicenesConditionMsg).To(BeEmpty(),
				fmt.Sprintf("License did not got re-activated after recreating [%s] secret", PureSecretName))

			Step("Compare PX-Essentials FA/FB features vs activated license", func() {
				for _, feature := range summary.Features {
					// if the feature limit exists in the hardcoded license limits we test it.
					if _, ok := faLicense[LabLabel(feature.Name)]; ok {
						Expect(feature.Quantity).To(Equal(faLicense[LabLabel(feature.Name)]),
							fmt.Sprintf("%v did not match: [%v]", feature.Quantity, faLicense[LabLabel(feature.Name)]))
					}
				}
			})
		})

		ValidateAndDestroy(contexts, nil)
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

/*
This Test:

1. Deletes px-pure-secret
2. Restarts PX on all nodes
3. Expects PX-Essentials FA/FB lic does not falls back to PX-Essentials license
*/
var _ = Describe("{DeleteSecretRebootAllNodes}", Label("p0", "negative", "license", "node_reboot"), func() {
	var testrailID = 84245
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/84245
	var runID int
	JustBeforeEach(func() {
		StartTorpedoTest("DeleteSecretRebootAllNodes", "Validates `Portworx CSI for FA/FB` does not fall back to `PX-Essentials` after deleting `PX-Pure-Secret`", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})
	var contexts []*scheduler.Context
	var err error
	var pureSecretJSON string
	var nodesToReboot []node.Node

	It("has to setup, validate and teardown apps", func() {
		contexts = make([]*scheduler.Context, 0)

		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("delseclicexprenewal-%d", i))...)
		}

		ValidateApplications(contexts)

		Step("Get SKU and compare with PX-Essentials FA/FB", func() {
			summary, err := Inst().V.GetLicenseSummary()
			Expect(err).NotTo(HaveOccurred(),
				fmt.Sprintf("Failed to get license SKU. Error: [%v]", err))

			Expect(summary.SKU).To(Equal(essentialsFaFbSKU),
				fmt.Sprintf("SKU did not match: [%v]", essentialsFaFbSKU))

			Step("Compare PX-Essentials FA/FB features vs activated license", func() {
				for _, feature := range summary.Features {
					// if the feature limit exists in the hardcoded license limits we test it.
					if _, ok := faLicense[LabLabel(feature.Name)]; ok {
						Expect(feature.Quantity).To(Equal(faLicense[LabLabel(feature.Name)]),
							fmt.Sprintf("%v did not match: [%v]", feature.Quantity, faLicense[LabLabel(feature.Name)]))
					}
				}
			})
		})

		Step("Fetch and store Pure secret", func() {
			var err error
			pureSecretJSON, err = Inst().S.GetSecretData(PureSecretNamespace, PureSecretName, pureSecretDataField)
			Expect(err).NotTo(HaveOccurred(),
				fmt.Sprintf("Failed to fetch secret [%s] in [%s] namespace. Error: [%v]",
					PureSecretName, PureSecretNamespace, err))
		})

		Step("Delete Pure secret", func() {
			err := Inst().S.DeleteSecret(PureSecretNamespace, PureSecretName)
			Expect(err).NotTo(HaveOccurred(),
				fmt.Sprintf("Failed to delete secret [%s] in [%s] namespace. Error: [%v]",
					PureSecretName, PureSecretNamespace, err))
		})

		Step("get all nodes and reboot one by one", func() {
			nodesToReboot = node.GetWorkerNodes()

			// Reboot node and check driver status
			Step(fmt.Sprintf("reboot node one at a time from the node(s): %v", nodesToReboot), func() {
				for _, n := range nodesToReboot {
					if n.IsStorageDriverInstalled {
						Step(fmt.Sprintf("reboot node: %s", n.Name), func() {
							err = Inst().N.RebootNode(n, node.RebootNodeOpts{
								Force: true,
								ConnectionOpts: node.ConnectionOpts{
									Timeout:         defaultCommandTimeout,
									TimeBeforeRetry: defaultCommandRetry,
								},
							})
							Expect(err).NotTo(HaveOccurred())
						})

						Step(fmt.Sprintf("wait for node: %s to be back up", n.Name), func() {
							err = Inst().N.TestConnection(n, node.ConnectionOpts{
								Timeout:         defaultTestConnectionTimeout,
								TimeBeforeRetry: defaultWaitRebootRetry,
							})
							Expect(err).NotTo(HaveOccurred())
						})

						Step(fmt.Sprintf("wait for volume driver to stop on node: %v", n.Name), func() {
							err := Inst().V.WaitDriverDownOnNode(n)
							Expect(err).NotTo(HaveOccurred())
						})

						Step(fmt.Sprintf("wait to scheduler: %s and volume driver: %s to start",
							Inst().S.String(), Inst().V.String()), func() {

							err = Inst().S.IsNodeReady(n)
							Expect(err).NotTo(HaveOccurred())

							err = Inst().V.WaitDriverUpOnNode(n, Inst().DriverStartTimeout)
							Expect(err).NotTo(HaveOccurred())
						})

						Step("validate apps", func() {
							for _, ctx := range contexts {
								ValidateContext(ctx)
							}
						})
					}
				}
			})
		})

		Step("Get SKU and compare with PX-Essentials FA/FB", func() {
			summary, err := Inst().V.GetLicenseSummary()
			Expect(err).NotTo(HaveOccurred(),
				fmt.Sprintf("Failed to get license SKU. Error: [%v]", err))

			Expect(summary.SKU).To(Equal(essentialsFaFbSKU),
				fmt.Sprintf("SKU changed after deleting [%s] secret and reboot to [%v]", PureSecretName, summary.SKU))

			Step("Compare PX-Essentials FA/FB features vs activated license", func() {
				for _, feature := range summary.Features {
					// if the feature limit exists in the hardcoded license limits we test it.
					if _, ok := faLicense[LabLabel(feature.Name)]; ok {
						Expect(feature.Quantity).To(Equal(faLicense[LabLabel(feature.Name)]),
							fmt.Sprintf("%v did not match: [%v]", feature.Quantity, faLicense[LabLabel(feature.Name)]))
					}
				}
			})
		})

		// Perform below steps to recover setup for other tests to continue
		Step("Re-create Pure secret", func() {
			err := Inst().S.CreateSecret(PureSecretNamespace, PureSecretName, pureSecretDataField, pureSecretJSON)
			Expect(err).NotTo(HaveOccurred(),
				fmt.Sprintf("Failed to create secret [%s] in [%s] namespace. Error: [%v]",
					PureSecretName, PureSecretNamespace, err))
		})

		Step("Recover Portworx", func() {
			for _, node := range nodesToReboot {
				err := Inst().V.RestartDriver(node, nil)
				Expect(err).NotTo(HaveOccurred(), "failed to restart service on node: [%v]. Error: [%v]", node, err)
			}
		})

		ValidateAndDestroy(contexts, nil)
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

// This test performs basic test disabling callhome and checking if the licnse stays valid
var _ = Describe("{DisableCallHomeTest}", Label("p1", "negative", "license"), func() {
	var testrailID = 84245
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/84245
	var runID int
	JustBeforeEach(func() {
		StartTorpedoTest("DisableCallHomeTest", "Validates disabling callhome does not expires lic", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})
	var contexts []*scheduler.Context
	It("has to setup, validate and teardown apps, then disable callhome and wait 65 minutes to verify the license is still valid.", func() {
		contexts = make([]*scheduler.Context, 0)
		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("setupteardown-license-callhome-%d", i))...)
		}
		ValidateApplications(contexts)

		opts := make(map[string]bool)
		opts[scheduler.OptionsWaitForResourceLeakCleanup] = true

		currNode := node.GetWorkerNodes()[0]
		Step(fmt.Sprintf("Set License expiry timeout to 1 hour"), func() {
			err := Inst().V.SetClusterRunTimeOpts(currNode, map[string]string{
				"metering_interval_mins":       "10",
				"license_expiry_timeout_hours": "1",
			})
			Expect(err).NotTo(HaveOccurred())
		})

		Step(fmt.Sprintf("Disable call-home"), func() {
			err := Inst().V.ToggleCallHome(currNode, false)
			Expect(err).NotTo(HaveOccurred())
		})

		Step("get all nodes and reboot one by one", func() {
			nodesToReboot := node.GetWorkerNodes()

			// Reboot node and check driver status
			Step(fmt.Sprintf("reboot node one at a time from the node(s): %v", nodesToReboot), func() {
				for _, n := range nodesToReboot {
					if n.IsStorageDriverInstalled {
						Step(fmt.Sprintf("reboot node: %s", n.Name), func() {
							err := Inst().N.RebootNode(n, node.RebootNodeOpts{
								Force: true,
								ConnectionOpts: node.ConnectionOpts{
									Timeout:         defaultCommandTimeout,
									TimeBeforeRetry: defaultCommandRetry,
								},
							})
							Expect(err).NotTo(HaveOccurred())
						})

						Step(fmt.Sprintf("wait for node: %s to be back up", n.Name), func() {
							err := Inst().N.TestConnection(n, node.ConnectionOpts{
								Timeout:         defaultTestConnectionTimeout,
								TimeBeforeRetry: defaultWaitRebootRetry,
							})
							Expect(err).NotTo(HaveOccurred())
						})

						Step(fmt.Sprintf("wait for volume driver to stop on node: %v", n.Name), func() {
							err := Inst().V.WaitDriverDownOnNode(n)
							Expect(err).NotTo(HaveOccurred())
						})

						Step(fmt.Sprintf("wait to scheduler: %s and volume driver: %s to start",
							Inst().S.String(), Inst().V.String()), func() {

							err := Inst().S.IsNodeReady(n)
							Expect(err).NotTo(HaveOccurred())

							err = Inst().V.WaitDriverUpOnNode(n, Inst().DriverStartTimeout)
							Expect(err).NotTo(HaveOccurred())
						})

						Step("validate apps", func() {
							for _, ctx := range contexts {
								ValidateContext(ctx)
							}
						})
					}
				}
			})
		})

		Step("Wait 65 Minutes to make sure we passed the 1 hour mark and test if our license is still valid", func() {
			time.Sleep(65 * time.Minute)

			Step("Get SKU and compare with PX-Essentials FA/FB", func() {
				summary, err := Inst().V.GetLicenseSummary()
				Expect(err).NotTo(HaveOccurred(),
					fmt.Sprintf("Failed to get license SKU. Error: [%v]", err))

				Expect(summary.SKU).To(Equal(essentialsFaFbSKU),
					fmt.Sprintf("SKU did not match: [%v]", essentialsFaFbSKU))

				Step("Compare PX-Essentials FA/FB features vs activated license", func() {
					for _, feature := range summary.Features {
						// if the feature limit exists in the hardcoded license limits we test it.
						if _, ok := faLicense[LabLabel(feature.Name)]; ok {
							Expect(feature.Quantity).To(Equal(faLicense[LabLabel(feature.Name)]),
								fmt.Sprintf("%v: %v did not match: [%v]", feature.Name, feature.Quantity, faLicense[LabLabel(feature.Name)]))
						}
					}
				})
			})
		})

		for _, ctx := range contexts {
			TearDownContext(ctx, opts)
		}
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

// LicenseValidation validates license summary against expected SKU and features
var _ = Describe("{LicenseValidation}", Label("p1", "positive", "license", "Upgrade", "UpgradeVolDriver"), func() {
	JustBeforeEach(func() {
		StartTorpedoTest("LicenseValidation", "Validates license summary against expected SKU and features", nil, 0)
	})

	It("Validate license summary against expected SKU and features", func() {
		log.Infof("Validating license summary against expected SKU and features")
		err = ValidatePxLicenseSummary()
		log.FailOnError(err, "failed to validate license summary against expected SKU and features")
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})

// This test performs basic test of starting an application
// Validating those applications
// Getting Volume License Summary
// Validating License Summary with pxEssentials License defined
// and destroying it (along with storage)
var _ = Describe("{BasicEssentialsTest}", Label("p2", "positive", "license"), func() {
	var testrailID = 53350
	var runID int
	JustBeforeEach(func() {
		StartTorpedoTest("BasicEssentialsTest", "Validates `Portworx for Essentials` license SKU", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})
	var contexts []*scheduler.Context

	It("has to setup, validate and teardown apps", func() {
		contexts = make([]*scheduler.Context, 0)

		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("setupteardown-license-%d", i))...)
		}

		ValidateApplications(contexts)

		Step("Get SKU and compare with PX-Essentials", func() {
			summary, err := Inst().V.GetLicenseSummary()
			Expect(err).NotTo(HaveOccurred(),
				fmt.Sprintf("Failed to get license SKU. Error: [%v]", err))

			Expect(summary.SKU).To(Equal(pxEssentials),
				fmt.Sprintf("SKU did not match: [%v]", pxEssentials))
			log.InfoD("Comparing Px-Essentials List with Activated License List!!")
			Step("Compare PX-Essentials features vs activated license", func() {
				for _, feature := range summary.Features {
					// if the feature limit exists in the hardcoded license limits we test it.
					if labelFeatureName, ok := essentialLicense[LabLabel(feature.Name)]; ok {
						log.InfoD("Verifying the feature [%v] from the List Expected: [%v] == Actual: [%v]", feature.Name, labelFeatureName, feature.Quantity)
						Expect(feature.Quantity).To(Equal(labelFeatureName),
							fmt.Sprintf("%v did not match: [%v]", feature.Quantity, labelFeatureName))
					}
				}
			})
		})
		ValidateAndDestroy(contexts, nil)
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

func rebootNodes(nodesToReboot ...node.Node) error {
	for _, n := range nodesToReboot {
		if err := Inst().N.RebootNodeAndWait(n); err != nil {
			return err
		}

		if err := Inst().V.WaitDriverUpOnNode(n, Inst().DriverStartTimeout); err != nil {
			return err
		}
	}
	return nil
}

func countHealthyPxNodes(nodesToReboot ...node.Node) []node.Node {
	healthyNodes := make([]node.Node, 0)
	for _, n := range nodesToReboot {
		if n.StorageNode.Status == api.Status_STATUS_OK {
			healthyNodes = append(healthyNodes, n)
		}
	}
	return healthyNodes
}

func createPxPureSecret(pureJson, ns string) error {
	if err := Inst().S.CreateSecret(ns, pureutils.PureSecretName, pureutils.PureJSONKey, string(pureJson)); err != nil {
		return err
	}
	return nil
}

func getEssentialsSecretNS(stc *opcorev1.StorageCluster) string {
	const (
		defaultNS       = "kube-system"
		PxEssentialsEnv = "PXESSENTIAL_SECRET_NAMESPACE"
	)
	if stc == nil {
		return defaultNS
	}
	for _, env := range stc.Spec.Env {
		if env.Name == PxEssentialsEnv && env.Value != "" {
			return env.Value
		}
	}
	return defaultNS
}

func InstallStc(stc *opcorev1.StorageCluster, pureJson string) error {
	stc.Annotations["portworx.io/misc-args"] = "--oem esse"
	stc.Spec.DeleteStrategy = &opcorev1.StorageClusterDeleteStrategy{
		Type: opcorev1.UninstallAndWipeStorageClusterStrategyType,
	}
	if len(stc.Spec.RuntimeOpts) == 0 {
		stc.Spec.RuntimeOpts = make(map[string]string)
	}
	stc.Spec.RuntimeOpts["metering_interval_mins"] = "3"
	stc.Spec.RuntimeOpts["essentials_license_expiry_timeout_hours"] = "0"
	stc.ObjectMeta = apmv1.ObjectMeta{
		Name:        stc.GetName(),
		Namespace:   stc.GetNamespace(),
		Labels:      stc.GetLabels(),
		Annotations: stc.GetAnnotations(),
	}
	stc.Status = opcorev1.StorageClusterStatus{}
	stc.Spec.Autopilot = &opcorev1.AutopilotSpec{Enabled: false}
	stc.Spec.CSI = &opcorev1.CSISpec{Enabled: false}
	stc.Spec.Stork = &opcorev1.StorkSpec{Enabled: false}
	stc.Spec.Monitoring = &opcorev1.MonitoringSpec{
		Prometheus: &opcorev1.PrometheusSpec{Enabled: false},
		Telemetry:  &opcorev1.TelemetrySpec{Enabled: false},
	}

	err = Inst().S.DeleteSecret(getEssentialsSecretNS(stc), "px-essential")
	if !k8serror.IsNotFound(err) && err != nil {
		return err
	}
	err = Inst().S.DeleteSecret(stc.Namespace, pureutils.PureSecretName)
	if !k8serror.IsNotFound(err) && err != nil {
		return err
	}
	if stc.Spec.DeleteStrategy == nil || stc.Spec.DeleteStrategy.Type != opcorev1.UninstallAndWipeStorageClusterStrategyType {
		stc.Spec.DeleteStrategy = &opcorev1.StorageClusterDeleteStrategy{
			Type: opcorev1.UninstallAndWipeStorageClusterStrategyType,
		}
		stc, err = operator.Instance().UpdateStorageCluster(stc)
		if err != nil {
			return err
		}
	}

	if err = UninstallAndValidateStorageCluster(stc); err != nil {
		return err
	}

	if pureJson != "" {
		log.FailOnError(createPxPureSecret(pureJson, stc.GetNamespace()), "unable to create pure secret")
		defer func() {
			time.Sleep(4 * time.Minute)
		}()
	}
	_, err = DeployAndValidateStorageCluster(stc)
	return err
}

/*
1. Install px in essentials mode
2. Create pure-px-secret after px is healthy
3. Wait for metering cycle
4. The license should switch to CSI
*/
var _ = Describe("{ConversionToCSIDelayed}", Label("p0", "positive", "license"), func() {
	var testrailID = 0
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/84245
	var runID int
	var opts map[string]bool
	JustBeforeEach(func() {
		opts := make(map[string]bool)
		opts[scheduler.OptionsWaitForResourceLeakCleanup] = true
		StartTorpedoTest("ConversionToCSIDelayed", "Validate conversion to CSI license when pure-px-secret is added later", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})

	var contexts []*scheduler.Context
	It("has to setup, validate px license", func() {
		stc, err := Inst().V.GetDriver()
		log.FailOnError(err, "unable to get stc via volume driver")
		pureJson, err := Inst().S.GetSecretData(stc.Namespace, pureutils.PureSecretName, pureutils.PureJSONKey)
		log.FailOnError(err, "")
		Expect(err).ShouldNot(HaveOccurred())
		Expect(pureJson).ShouldNot(BeEmpty())
		err = Inst().S.DeleteSecret(getEssentialsSecretNS(stc), "px-essential")
		if !k8serror.IsNotFound(err) {
			Expect(err).NotTo(HaveOccurred())
		}
		err = Inst().S.DeleteSecret(stc.Namespace, pureutils.PureSecretName)
		if !k8serror.IsNotFound(err) {
			Expect(err).NotTo(HaveOccurred())
		}

		log.FailOnError(InstallStc(stc, ""), "Unable to install px")

		Step(fmt.Sprintf("Add pure-px-secret"), func() {
			log.FailOnError(createPxPureSecret(pureJson, stc.GetNamespace()), "unable to create pure secret")
			log.FailOnError(rebootNodes(node.GetWorkerNodes()...), "unable to reboot nodes after adding pure secret")

		})

		Step("Wait 7 Minutes to make sure we passed the metering cycle interval mark and test if our license is mutated to CSI", func() {
			time.Sleep(7 * time.Minute)

			Step("Get SKU and compare with PX-Essentials FA/FB", func() {
				summary, err := Inst().V.GetLicenseSummary()
				Expect(err).NotTo(HaveOccurred(),
					fmt.Sprintf("Failed to get license SKU. Error: [%v]", err))

				Expect(summary.SKU).To(Equal(essentialsFaFbSKU),
					fmt.Sprintf("SKU did not match: [%v]", essentialsFaFbSKU))
			})
		})

		Step("Validate all nodes are healthy", func() {
			workerNodes := node.GetWorkerNodes()
			healthyNodes := countHealthyPxNodes(workerNodes...)
			Expect(len(healthyNodes)).Should(BeEquivalentTo(len(workerNodes)))
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		for _, ctx := range contexts {
			TearDownContext(ctx, opts)
		}
		AfterEachTest(contexts, testrailID, runID)
	})
})

/*
1. Uninstall current PX
2. Install PX with px-pure-secret and without px-essential secret
3. PX should be healthy but Essentials license should show expired
*/
var _ = Describe("{InstallWithPurePxSecret}", Label("p0", "positive", "license"), func() {
	var testrailID = 90837609
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/84245
	var runID int
	var opts map[string]bool

	JustBeforeEach(func() {
		opts = make(map[string]bool)
		opts[scheduler.OptionsWaitForResourceLeakCleanup] = true

		StartTorpedoTest("InstallWithoutSecrets", "Validtes conversion to CSI license when pure-px-secret is present at time of installation", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})
	var contexts []*scheduler.Context
	It("has to validate installation with pure-px-secret", func() {

		Step("has to install PX without px-essentials and with px-pure-secret secrets", func() {
			stc, err := Inst().V.GetDriver()
			Expect(err).NotTo(HaveOccurred())
			pureJson, err := Inst().S.GetSecretData(stc.Namespace, pureutils.PureSecretName, pureutils.PureJSONKey)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(pureJson).ShouldNot(BeEmpty())
			err = Inst().S.DeleteSecret(getEssentialsSecretNS(stc), "px-essential")
			if !k8serror.IsNotFound(err) {
				Expect(err).NotTo(HaveOccurred())
			}
			err = Inst().S.DeleteSecret(stc.Namespace, pureutils.PureSecretName)
			if !k8serror.IsNotFound(err) {
				Expect(err).NotTo(HaveOccurred())
			}
			log.FailOnError(InstallStc(stc, pureJson), "Unable to install px with pure secret")

		})
		Step("has to validate license", func() {

			Step("Wait 7 Minutes to make sure we passed the metering cycle interval mark and test if our license is mutated to CSI", func() {
				time.Sleep(7 * time.Minute)

				Step("Get SKU and compare with PX-Essentials FA/FB", func() {
					summary, err := Inst().V.GetLicenseSummary()
					Expect(err).NotTo(HaveOccurred(),
						fmt.Sprintf("Failed to get license SKU. Error: [%v]", err))

					Expect(summary.SKU).To(Equal(essentialsFaFbSKU),
						fmt.Sprintf("SKU did not match: [%v]", essentialsFaFbSKU))
				})
			})
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		for _, ctx := range contexts {
			TearDownContext(ctx, opts)
		}
		AfterEachTest(contexts, testrailID, runID)
	})
})

/*
1. Uninstall current PX
2. Install PX without px-essential secret, px-pure-secret
3. PX should be healthy but Essentials license should show expired
*/
var _ = Describe("{InstallWithoutSecrets}", Label("p0", "positive", "license"), func() {
	var testrailID = 0
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/84245
	var runID int
	var pureJson string
	var opts map[string]bool
	var contexts []*scheduler.Context

	JustBeforeEach(func() {
		opts = make(map[string]bool)
		opts[scheduler.OptionsWaitForResourceLeakCleanup] = true

		StartTorpedoTest("InstallWithoutSecrets", "Validtes conversion to CSI license when pure-px-secret is added later", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})
	BeforeEach(func() {
		stc, err := Inst().V.GetDriver()
		log.FailOnError(err, "Unable to get stc from volume driver")
		pureJson, err = Inst().S.GetSecretData(stc.Namespace, pureutils.PureSecretName, pureutils.PureJSONKey)
		log.FailOnError(err, "Unable to get secret data from pure secret")
		dash.VerifyFatal(pureJson != "", true, "pure json field is empty in pure secret")
		err = Inst().S.DeleteSecret(getEssentialsSecretNS(stc), "px-essential")
		if !k8serror.IsNotFound(err) {
			log.FailOnError(err, "Unable to delete px-essential secret")
		}
		err = Inst().S.DeleteSecret(stc.Namespace, pureutils.PureSecretName)
		if !k8serror.IsNotFound(err) {
			log.FailOnError(err, "Unable to delete pure secret")
		}
	})
	AfterEach(func() {
		dash.VerifyFatal(pureJson != "", true, "empty pure json found")
		stc, err := Inst().V.GetDriver()
		log.FailOnError(err, "Unable to get stc via vol driver")
		log.FailOnError(createPxPureSecret(pureJson, stc.GetNamespace()), "unable to create pure secret")
	})
	It("it validates that px can be installed without any secrets in essentials mode", func() {
		Step("has to install PX without px-essentials and px-pure-secret secrets", func() {
			stc, err := Inst().V.GetDriver()
			log.FailOnError(err, "Unable to get stc via vol driver")
			stc.Annotations["portworx.io/misc-args"] = "--oem esse"
			stc.Spec.DeleteStrategy = &opcorev1.StorageClusterDeleteStrategy{
				Type: opcorev1.UninstallAndWipeStorageClusterStrategyType,
			}
			if len(stc.Spec.RuntimeOpts) == 0 {
				stc.Spec.RuntimeOpts = make(map[string]string)
			}
			stc.Spec.RuntimeOpts["metering_interval_mins"] = "3"
			stc.Spec.RuntimeOpts["essentials_license_expiry_timeout_hours"] = "0"
			stc.ObjectMeta = apmv1.ObjectMeta{
				Name:        stc.GetName(),
				Namespace:   stc.GetNamespace(),
				Labels:      stc.GetLabels(),
				Annotations: stc.GetAnnotations(),
			}
			stc.Status = opcorev1.StorageClusterStatus{}
			stc.Spec.Autopilot = &opcorev1.AutopilotSpec{Enabled: false}
			stc.Spec.CSI = &opcorev1.CSISpec{Enabled: false}
			stc.Spec.Stork = &opcorev1.StorkSpec{Enabled: false}
			stc.Spec.Monitoring = &opcorev1.MonitoringSpec{
				Prometheus: &opcorev1.PrometheusSpec{Enabled: false},
				Telemetry:  &opcorev1.TelemetrySpec{Enabled: false},
			}

			err = UninstallAndValidateStorageCluster(stc)
			_, err = DeployAndValidateStorageCluster(stc)
			log.FailOnError(err, "")

		})
		Step("has to validate license.", func() {
			Step("Wait 7 Minutes to make sure we passed the metering cycle interval mark and test if our license is mutated to CSI", func() {
				time.Sleep(7 * time.Minute)

				Step("Get SKU and compare with PX-Essentials FA/FB", func() {
					summary, err := Inst().V.GetLicenseSummary()
					log.FailOnError(err, fmt.Sprintf("Failed to get license SKU. Error: [%v]", err))
					dash.VerifyFatal(summary.SKU == pxEssentials, true, fmt.Sprintf("SKU did not match: [%v]", pxEssentials))
				})
			})
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		for _, ctx := range contexts {
			TearDownContext(ctx, opts)
		}
		AfterEachTest(contexts, testrailID, runID)
	})
})

/*
Steps
1. Corrupt the metering/license key
2. Wait for metering cycle to start
3. Verify license is valid and CSI
4. Verify metering/license key points to CSI license
*/

// This test performs basic test disabling callhome and checking if the licnse stays valid
var _ = Describe("{CorruptMeteringLicenseKey}", Label("p1", "positive", "license"), func() {
	var testrailID = 90837613
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/84245
	var runID int
	var opts map[string]bool

	JustBeforeEach(func() {
		opts = make(map[string]bool)
		opts[scheduler.OptionsWaitForResourceLeakCleanup] = true
		StartTorpedoTest("CorruptMeteringLicenseKey", "validates metering license key is fixed automatically", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})
	var contexts []*scheduler.Context
	It("has to setup px with with essentials fa/fb license", func() {
		stc, err := Inst().V.GetDriver()
		log.FailOnError(err, "")
		pureJson, err := Inst().S.GetSecretData(stc.Namespace, pureutils.PureSecretName, pureutils.PureJSONKey)
		log.FailOnError(err, "")
		dash.VerifyFatal(pureJson != "", true, "pure json is empty")
		err = Inst().S.DeleteSecret(getEssentialsSecretNS(stc), "px-essential")
		if !k8serror.IsNotFound(err) {
			log.FailOnError(err, "")
		}
		err = Inst().S.DeleteSecret(stc.Namespace, pureutils.PureSecretName)
		if !k8serror.IsNotFound(err) {
			log.FailOnError(err, "")
		}
		log.FailOnError(InstallStc(stc, pureJson), "Unable to install px with pure secret")

		currNode := node.GetWorkerNodes()[0]
		Step("Validate license is CSI", func() {
			summary, err := Inst().V.GetLicenseSummary()
			log.FailOnError(err, fmt.Sprintf("Failed to get license SKU. Error: [%v]", err))
			dash.VerifyFatal(summary.SKU == essentialsFaFbSKU, true, fmt.Sprintf("SKU did not match: [%v]", essentialsFaFbSKU))
		})
		Step(fmt.Sprintf("Set License expiry timeout to 1 hour"), func() {
			err := Inst().V.SetClusterRunTimeOpts(currNode, map[string]string{
				"metering_interval_mins":       "3",
				"license_expiry_timeout_hours": "1",
			})
			log.FailOnError(err, "")
		})

		Step("get all nodes and reboot one by one", func() {
			nodesToReboot := node.GetWorkerNodes()

			// Reboot node and check driver status
			Step(fmt.Sprintf("reboot node one at a time from the node(s): %v", nodesToReboot), func() {
				err := rebootNodes(nodesToReboot...)
				log.FailOnError(err, "Unable to reboot nodes")
			})
			// allow atleast 1 metering cycle to kick in
			time.Sleep(4 * time.Minute)
		})

		Step(fmt.Sprintf("Corrupt metering metering/license key"), func() {
			pxctlCmd := ipv6util.PxctlServiceKvdbEndpoints
			output, err := Inst().V.GetPxctlCmdOutput(currNode, pxctlCmd)
			log.FailOnError(err, "")
			ips := ipv6util.ParseIPAddressInPxctlServiceKvdbEndpointsWithPort(output)
			csEndpoints := strings.Join(ips, ",")
			stc, err := Inst().V.GetDriver()
			log.FailOnError(err, "")
			value := `"{'Duration':86400000000000,'Error':'','LicenseType':8}"`
			corruptKeyCmd := fmt.Sprintf("/opt/pwx/bin/runc exec -t portworx etcdctl --endpoints=%s put pwx/%s/metering/license %v", csEndpoints, stc.GetName(), value)
			output, err = runCmd(corruptKeyCmd, currNode)
			log.FailOnError(err, output)
			dash.VerifyFatal(strings.Contains(output, "OK"), true, "etcdctl command failed")
		})

		Step("Wait 15 Minutes to make sure we passed the metering cycle interval mark and test if our license is still valid", func() {
			time.Sleep(15 * time.Minute)

			Step("Get SKU and compare with PX-Essentials FA/FB", func() {
				summary, err := Inst().V.GetLicenseSummary()
				log.FailOnError(err, fmt.Sprintf("Failed to get license SKU. Error: [%v]", err))
				dash.VerifyFatal(summary.SKU == essentialsFaFbSKU, true, fmt.Sprintf("SKU did not match: [%v]", essentialsFaFbSKU))
			})
		})

		Step("Validate the metering/license key was corrected", func() {
			pxctlCmd := ipv6util.PxctlServiceKvdbEndpoints
			output, err := Inst().V.GetPxctlCmdOutput(currNode, pxctlCmd)
			log.FailOnError(err, "")
			ips := ipv6util.ParseIPAddressInPxctlServiceKvdbEndpointsWithPort(output)
			csEndpoints := strings.Join(ips, ",")
			stc, err := Inst().V.GetDriver()
			log.FailOnError(err, "")
			corruptKeyCmd := fmt.Sprintf("runc exec -t portworx etcdctl --endpoints=%s get pwx/%s/metering/license", csEndpoints, stc.GetName())
			output, err = runCmd(corruptKeyCmd, currNode)
			log.FailOnError(err, "")
			expectedLicense := `"LicenseType":8`
			dash.VerifyFatal(strings.Contains(output, expectedLicense), true, "metering license key doesn't reflect CSI as license type")
		})

	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		for _, ctx := range contexts {
			TearDownContext(ctx, opts)
		}
		AfterEachTest(contexts, testrailID, runID)
	})
})

// SleepWithContext will wait for the timer duration to expire, or the context
// is canceled. Which ever happens first. If the context is canceled the Context's
// error will be returned.
//
// Expects Context to always return a non-nil error if the Done channel is closed.
func SleepWithContext(ctx context.Context, dur time.Duration) error {
	t := time.NewTimer(dur)
	defer t.Stop()

	select {
	case <-t.C:
		break
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}
