package tests

import (
	"fmt"
	"os"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	"github.com/portworx/torpedo/drivers/node"
	"github.com/portworx/torpedo/drivers/scheduler"
	. "github.com/portworx/torpedo/tests"
	"github.com/sirupsen/logrus"
)

func TestSharedV4SVC(t *testing.T) {
	RegisterFailHandler(Fail)

	var specReporters []Reporter
	junitReporter := reporters.NewJUnitReporter("/testresults/junit_Sharedv4_SVC.xml")
	specReporters = append(specReporters, junitReporter)
	RunSpecsWithDefaultAndCustomReporters(t, "Torpedo: Sharedv4_SVC", specReporters)
}

var _ = BeforeSuite(func() {
	InitInstance()
})

// This test performs mutli volume mounts to a single  deployment
var _ = Describe("{MultiVolumeMounts}", func() {

	It("has to create multiple sharedv4-svc volumes and mount to sing pod", func() {
		// set frequency mins depending on the chaos level
		var frequency int
		var timeout time.Duration
		switch Inst().ChaosLevel {
		case 10:
			frequency = 100
			timeout = 10 * time.Minute
		case 9:
			frequency = 90
			timeout = 9 * time.Minute
		case 8:
			frequency = 80
			timeout = 8 * time.Minute
		case 7:
			frequency = 70
			timeout = 7 * time.Minute
		case 6:
			frequency = 60
			timeout = 6 * time.Minute
		case 5:
			frequency = 50
			timeout = 5 * time.Minute
		case 4:
			frequency = 40
			timeout = 4 * time.Minute
		case 3:
			frequency = 30
			timeout = 3 * time.Minute
		case 2:
			frequency = 20
			timeout = 2 * time.Minute
		case 1:
			frequency = 10
			timeout = 1 * time.Minute
		default:
			frequency = 10
			timeout = 1 * time.Minute

		}

		customAppConfig := scheduler.AppConfig{
			ClaimsCount: frequency,
		}

		provider := Inst().V.String()
		contexts := []*scheduler.Context{}

		Inst().CustomAppConfig["vdbench-sv4-svc-multivol"] = customAppConfig
		err := Inst().S.RescanSpecs(Inst().SpecDir, provider)

		Expect(err).NotTo(HaveOccurred(),
			fmt.Sprintf("Failed to rescan specs from %s for storage provider %s. Error: [%v]",
				Inst().SpecDir, provider, err))

		Step("schedule application with multiple sharedv4-svc volumes attached", func() {
			logrus.Infof("Number of Volumes to be mounted: %v", frequency)

			taskName := "sharedv4-svc-multivol"

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

		Step("Scale up and down all app", func() {
			for _, ctx := range contexts {
				globalScaleFactor := Inst().GlobalScaleFactor
				Step(fmt.Sprintf("scale up app: %s", ctx.App.Key), func() {
					applicationScaleUpMap, err := Inst().S.GetScaleFactorMap(ctx)
					Expect(err).NotTo(HaveOccurred())
					for name, scale := range applicationScaleUpMap {
						if globalScaleFactor == 100 && scale < 25 {

							applicationScaleUpMap[name] = scale + 25
						} else {
							applicationScaleUpMap[name] = scale + int32(len(node.GetWorkerNodes()))
						}
						logrus.Infof("scaling app %v by %d", name, applicationScaleUpMap[name])
					}
					err = Inst().S.ScaleApplication(ctx, applicationScaleUpMap)
					Expect(err).NotTo(HaveOccurred())
				})

				Step("Giving few seconds for scaled up applications to stabilize", func() {
					time.Sleep(10 * time.Second)
				})

				ctx.SkipVolumeValidation = false
				ValidateContext(ctx)

				Step(fmt.Sprintf("scale up app: %s", ctx.App.Key), func() {
					applicationScaleUpMap, err := Inst().S.GetScaleFactorMap(ctx)
					Expect(err).NotTo(HaveOccurred())
					for name, scale := range applicationScaleUpMap {
						if globalScaleFactor == 100 && scale < 50 {

							applicationScaleUpMap[name] = scale + 25
						} else {
							applicationScaleUpMap[name] = scale + int32(len(node.GetWorkerNodes()))
						}
						logrus.Infof("scaling app %v by %d", name, applicationScaleUpMap[name])
					}
					err = Inst().S.ScaleApplication(ctx, applicationScaleUpMap)
					Expect(err).NotTo(HaveOccurred())
				})

				Step("Giving few seconds for scaled up applications to stabilize", func() {
					time.Sleep(10 * time.Second)
				})

				ctx.SkipVolumeValidation = false
				ValidateContext(ctx)

				Step(fmt.Sprintf("scale up app: %s", ctx.App.Key), func() {
					applicationScaleUpMap, err := Inst().S.GetScaleFactorMap(ctx)
					Expect(err).NotTo(HaveOccurred())
					for name, scale := range applicationScaleUpMap {
						if globalScaleFactor == 100 && scale < 75 {

							applicationScaleUpMap[name] = scale + 25
						} else {
							applicationScaleUpMap[name] = scale + int32(len(node.GetWorkerNodes()))
						}
						logrus.Infof("scaling app %v by %d", name, applicationScaleUpMap[name])
					}
					err = Inst().S.ScaleApplication(ctx, applicationScaleUpMap)
					Expect(err).NotTo(HaveOccurred())
				})

				Step("Giving few seconds for scaled up applications to stabilize", func() {
					time.Sleep(10 * time.Second)
				})

				ctx.SkipVolumeValidation = false
				ValidateContext(ctx)

				Step(fmt.Sprintf("scale up app: %s", ctx.App.Key), func() {
					applicationScaleUpMap, err := Inst().S.GetScaleFactorMap(ctx)
					Expect(err).NotTo(HaveOccurred())
					for name, scale := range applicationScaleUpMap {
						if globalScaleFactor == 100 && scale < 100 {

							applicationScaleUpMap[name] = scale + 25
						} else {
							applicationScaleUpMap[name] = scale + int32(len(node.GetWorkerNodes()))
						}
						logrus.Infof("scaling app %v by %d", name, applicationScaleUpMap[name])
					}
					err = Inst().S.ScaleApplication(ctx, applicationScaleUpMap)
					Expect(err).NotTo(HaveOccurred())
				})

				Step("Giving few seconds for scaled up applications to stabilize", func() {
					time.Sleep(10 * time.Second)
				})

				ctx.SkipVolumeValidation = false
				ValidateContext(ctx)

				Step(fmt.Sprintf("scale down app %s", ctx.App.Key), func() {
					applicationScaleDownMap, err := Inst().S.GetScaleFactorMap(ctx)
					Expect(err).NotTo(HaveOccurred())
					for name, scale := range applicationScaleDownMap {
						if globalScaleFactor == 100 && scale == 100 {

							applicationScaleDownMap[name] = scale - 25
						} else {
							applicationScaleDownMap[name] = scale - 1
						}
						logrus.Infof("scaling app %v by %d", name, applicationScaleDownMap[name])
					}
					err = Inst().S.ScaleApplication(ctx, applicationScaleDownMap)
					Expect(err).NotTo(HaveOccurred())
				})

				Step("Giving few seconds for scaled down applications to stabilize", func() {
					time.Sleep(10 * time.Second)
				})

				ctx.SkipVolumeValidation = false
				ValidateContext(ctx)

				Step(fmt.Sprintf("scale down app %s", ctx.App.Key), func() {
					applicationScaleDownMap, err := Inst().S.GetScaleFactorMap(ctx)
					Expect(err).NotTo(HaveOccurred())
					for name, scale := range applicationScaleDownMap {
						if globalScaleFactor == 100 && scale > 50 {

							applicationScaleDownMap[name] = scale - 25
						} else {
							applicationScaleDownMap[name] = scale - 1
						}
						logrus.Infof("scaling app %v by %d", name, applicationScaleDownMap[name])
					}
					err = Inst().S.ScaleApplication(ctx, applicationScaleDownMap)
					Expect(err).NotTo(HaveOccurred())
				})

				Step("Giving few seconds for scaled down applications to stabilize", func() {
					time.Sleep(10 * time.Second)
				})

				ctx.SkipVolumeValidation = false
				ValidateContext(ctx)

				Step(fmt.Sprintf("scale down app %s", ctx.App.Key), func() {
					applicationScaleDownMap, err := Inst().S.GetScaleFactorMap(ctx)
					Expect(err).NotTo(HaveOccurred())
					for name, scale := range applicationScaleDownMap {
						if globalScaleFactor == 100 && scale > 25 {

							applicationScaleDownMap[name] = scale - 15
						} else {
							applicationScaleDownMap[name] = scale - 1
						}
						logrus.Infof("scaling app %v by %d", name, applicationScaleDownMap[name])
					}
					err = Inst().S.ScaleApplication(ctx, applicationScaleDownMap)
					Expect(err).NotTo(HaveOccurred())
				})

				Step("Giving few seconds for scaled down applications to stabilize", func() {
					time.Sleep(10 * time.Second)
				})

				ctx.SkipVolumeValidation = false
				ValidateContext(ctx)
			}
		})

		Step("get nodes where app is running and restart volume driver", func() {
			for _, ctx := range contexts {
				appNodes, err := Inst().S.GetNodesForApp(ctx)
				Expect(err).NotTo(HaveOccurred())
				for _, appNode := range appNodes {
					Step(
						fmt.Sprintf("stop volume driver %s on app %s's node: %s",
							Inst().V.String(), ctx.App.Key, appNode.Name),
						func() {
							StopVolDriverAndWait([]node.Node{appNode})
						})

					Step(
						fmt.Sprintf("starting volume %s driver on app %s's node %s",
							Inst().V.String(), ctx.App.Key, appNode.Name),
						func() {
							StartVolDriverAndWait([]node.Node{appNode})
						})

					Step("Giving few seconds for volume driver to stabilize", func() {
						time.Sleep(20 * time.Second)
					})

					Step(fmt.Sprintf("validate app %s", appNode.Name), func() {
						ValidateContext(ctx)
					})
				}
			}
		})

	})

})

func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags
	ParseFlags()
	os.Exit(m.Run())
}
