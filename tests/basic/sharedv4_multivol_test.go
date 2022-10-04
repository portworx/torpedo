package tests

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/portworx/torpedo/drivers/node"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/drivers/volume"
	"github.com/portworx/torpedo/pkg/testrailuttils"
	. "github.com/portworx/torpedo/tests"
	"github.com/sirupsen/logrus"
)

const (
	nodeDeleteTimeoutMins = 7 * time.Minute
)

// This test performs multi volume mounts to a single deployment
var _ = Describe("{MultiVolumeMountsForSharedV4}", func() {
	var testrailID = 58846
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/58846
	var runID int
	JustBeforeEach(func() {
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})
	var contexts []*scheduler.Context

	It("has to create multiple sharedv4 volumes and mount to single pod", func() {
		// set frequency mins depending on the chaos level
		var frequency int
		var timeout time.Duration

		chaosLevel := Inst().ChaosLevel
		if chaosLevel != 0 {
			frequency = 10 * chaosLevel
			timeout = (15 * time.Duration(chaosLevel) * time.Minute) / 10
		} else {
			frequency = 10
			timeout = 1 * time.Minute
		}
		logrus.Infof("setting number of volumes=%v and app readiness timeout=%v for chaos level %v",
			frequency, timeout, chaosLevel)

		customAppConfig := scheduler.AppConfig{
			ClaimsCount: frequency,
		}

		provider := Inst().V.String()
		contexts = []*scheduler.Context{}
		// there should be only 1 app
		Expect(len(Inst().AppList)).To(Equal(1))
		appName := Inst().AppList[0]

		Inst().CustomAppConfig[appName] = customAppConfig
		err := Inst().S.RescanSpecs(Inst().SpecDir, provider)

		Expect(err).NotTo(HaveOccurred(),
			fmt.Sprintf("Failed to rescan specs from %s for storage provider %s. Error: [%v]",
				Inst().SpecDir, provider, err))

		Step("schedule application with multiple sharedv4 volumes attached", func() {
			logrus.Infof("Number of Volumes to be mounted: %v", frequency)

			taskName := "sharedv4-multivol"

			logrus.Infof("Task name %s\n", taskName)

			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				newContexts := ScheduleApplications(taskName)
				contexts = append(contexts, newContexts...)
			}

			for _, ctx := range contexts {
				ctx.ReadinessTimeout = timeout
				ctx.SkipVolumeValidation = false
				ValidateContext(ctx)
			}
		})

		Step("get nodes where volume is attached and restart volume driver", func() {
			for _, ctx := range contexts {
				appVolumes, err := Inst().S.GetVolumes(ctx)
				Expect(err).NotTo(HaveOccurred())
				for _, appVolume := range appVolumes {
					attachedNode, err := Inst().V.GetNodeForVolume(appVolume, defaultCommandTimeout, defaultCommandRetry)
					Expect(err).NotTo(HaveOccurred())
					Step(
						fmt.Sprintf("stop volume driver %s on app %s's node: %s",
							Inst().V.String(), ctx.App.Key, attachedNode.Name),
						func() {
							StopVolDriverAndWait([]node.Node{*attachedNode})
						})

					Step(
						fmt.Sprintf("starting volume %s driver on app %s's node %s",
							Inst().V.String(), ctx.App.Key, attachedNode.Name),
						func() {
							StartVolDriverAndWait([]node.Node{*attachedNode})
						})

					Step("Giving few seconds for volume driver to stabilize", func() {
						time.Sleep(20 * time.Second)
					})

					Step(fmt.Sprintf("validate app %s", attachedNode.Name), func() {
						ctx.ReadinessTimeout = timeout
						ctx.SkipVolumeValidation = true
						ValidateContext(ctx)
					})
				}
			}
		})
	})
	JustAfterEach(func() {
		AfterEachTest(contexts, testrailID, runID)
	})
})

// This test performs sharedv4 nfs server pod termination failover use case
var _ = Describe("{NFSServerNodeDelete}", func() {
	var contexts []*scheduler.Context

	It("has to validate that the new pods started successfully after nfs server node is terminated", func() {
		contexts = make([]*scheduler.Context, 0)
		var err error

		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("nodekill-%d", i))...)
		}

		ValidateApplications(contexts)
		for _, ctx := range contexts {
			var appVolumes []*volume.Volume
			Step(fmt.Sprintf("get volumes for %s app", ctx.App.Key), func() {
				appVolumes, err = Inst().S.GetVolumes(ctx)
				Expect(err).NotTo(HaveOccurred())
				Expect(appVolumes).NotTo(BeEmpty())
			})
			for _, v := range appVolumes {

				Step("get attached node and stop the instance", func() {
					currNodes := node.GetStorageDriverNodes()
					countOfCurrNodes := len(currNodes)

					attachedNode, err := Inst().V.GetNodeForVolume(v, defaultCommandTimeout, defaultCommandRetry)

					// Delete node and check Apps status
					Step(fmt.Sprintf("delete node : %v having volume: %v attached", attachedNode.Name, v.Name), func() {

						sv4KillANodeAndValidate(*attachedNode)

						Step(fmt.Sprintf("validate node: %v is deleted", attachedNode.Name), func() {
							currNodes = node.GetStorageDriverNodes()
							for _, currNode := range currNodes {
								if currNode.Name == attachedNode.Name {
									Fail(fmt.Sprintf("Node: %v still exists",
										attachedNode.Name))
									break
								}
							}
						})

						Step(fmt.Sprintf("validate applications after node [%v] deletion", attachedNode.Name), func() {
							for _, ctx := range contexts {
								ValidateContext(ctx)
							}
						})

						Step(fmt.Sprintf("wait to new instance to start scheduler: %s and volume driver: %s",
							Inst().S.String(), Inst().V.String()), func() {
							time.Sleep(2 * time.Minute)
							currNodes = node.GetStorageDriverNodes()
							Expect(countOfCurrNodes).To(Equal(len(currNodes)))
							for _, n := range currNodes {

								err = Inst().S.IsNodeReady(n)
								Expect(err).NotTo(HaveOccurred())

								err = Inst().V.WaitDriverUpOnNode(n, Inst().DriverStartTimeout)
								Expect(err).NotTo(HaveOccurred())
							}
						})

						Step("validate apps after new node is ready", func() {
							for _, ctx := range contexts {
								ValidateContext(ctx)
							}
						})

					})
				})
			}

		}

	})
	JustAfterEach(func() {
		AfterEachTest(contexts)
	})
})

func sv4KillANodeAndValidate(nodeToKill node.Node) {

	Step(fmt.Sprintf("Deleting node [%v]", nodeToKill.Name), func() {
		logrus.Infof("Instance is of %v ", Inst().N.String())
		err := Inst().N.DeleteNode(nodeToKill, nodeDeleteTimeoutMins)
		Expect(err).NotTo(HaveOccurred())
	})

	Step(fmt.Sprintf("Wait for node: %v to be deleted", nodeToKill.Name), func() {
		maxWait := 10
	OUTER:
		for maxWait > 0 {
			for _, currNode := range node.GetStorageDriverNodes() {
				if currNode.Name == nodeToKill.Name {
					logrus.Infof("Node %v still exists. Waiting for a minute to check again", nodeToKill.Name)
					maxWait--
					time.Sleep(1 * time.Minute)
					continue OUTER
				}
			}
			break
		}
	})

	err := Inst().S.RefreshNodeRegistry()
	Expect(err).NotTo(HaveOccurred())

	err = Inst().V.RefreshDriverEndpoints()
	Expect(err).NotTo(HaveOccurred())
}
