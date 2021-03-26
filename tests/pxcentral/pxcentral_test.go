package tests

import (
	"encoding/base64"
	"fmt"
	"os"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	"github.com/portworx/sched-ops/k8s/batch"
	"github.com/portworx/sched-ops/k8s/core"
	"github.com/portworx/torpedo/drivers/node"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/drivers/scheduler/k8s"
	. "github.com/portworx/torpedo/tests"
	"github.com/sirupsen/logrus"
)

const (
	defaultTimeout       = 5 * time.Minute
	defaultRetryInterval = 20 * time.Second
	appReadinessTimeout  = 20 * time.Minute
)

// To run tests here, k8s ConfigMaps need to be created for each app before starting torpedo
// ConfigMap names are: px-central, px-license-server and px-monitor
// the Date field should contains helm chart url, helm values (new url, repo and chart name for upgrade tests)
// for simple installation, only install-url and values are needed, while for upgrade test here is an example:
// install-url: http://charts.portworx.io/
// upgrade-url: https://raw.githubusercontent.com/portworx/helm/master/repo/staging (staging chart to upgrade to)
// values: persistentStorage.storageClassName=central-sc,persistentStorage.enabled=true
// repo-name: portworx-staging (create a new staging repo for upgrade)
// chart-name: px-central (if the chart name for release installed is changed, e.g. px-backup to px-central)
func TestPxcentral(t *testing.T) {
	RegisterFailHandler(Fail)

	var specReporters []Reporter
	junitReporter := reporters.NewJUnitReporter("/testresults/junit_basic.xml")
	specReporters = append(specReporters, junitReporter)
	RunSpecsWithDefaultAndCustomReporters(t, "Torpedo : px-central", specReporters)
}

var _ = BeforeSuite(func() {
	logrus.Infof("Init instance")
	InitInstance()
})

// This test performs basic test of installing px-central with helm
// px-license-server and px-minotor will be installed after px-central is validated
var _ = Describe("{Installpxcentral}", func() {
	It("has to setup, validate and teardown apps", func() {
		var context *scheduler.Context

		centralApp := "px-central"
		centralOptions := scheduler.ScheduleOptions{
			AppKeys:            []string{centralApp},
			StorageProvisioner: Inst().Provisioner,
		}

		lsApp := "px-license-server"
		lsOptions := scheduler.ScheduleOptions{
			AppKeys:            []string{lsApp},
			StorageProvisioner: Inst().Provisioner,
		}

		monitorApp := "px-monitor"
		monitorOptions := scheduler.ScheduleOptions{
			AppKeys:            []string{monitorApp},
			StorageProvisioner: Inst().Provisioner,
		}

		Step("Install px-central using the px-backup helm chart then validate", func() {
			contexts, err := Inst().S.Schedule(Inst().InstanceID, centralOptions)
			Expect(err).NotTo(HaveOccurred())
			Expect(contexts).NotTo(BeEmpty())

			// Skipping volume validation until other volume providers are implemented.
			// Also change the app readinessTimeout to 20 mins
			context = contexts[0]
			context.SkipVolumeValidation = true
			context.ReadinessTimeout = appReadinessTimeout

			ValidateContext(context)
			logrus.Infof("Successfully validated specs for px-central")
		})

		Step("Install px-license-server then validate", func() {
			// label px/ls=true on 2 worker nodes
			for i, node := range node.GetWorkerNodes() {
				if i == 2 {
					break
				}
				err := Inst().S.AddLabelOnNode(node, "px/ls", "true")
				Expect(err).NotTo(HaveOccurred())
			}

			err := Inst().S.AddTasks(context, lsOptions)
			Expect(err).NotTo(HaveOccurred())

			ValidateContext(context)
			logrus.Infof("Successfully validated specs for px-license-server")
		})

		Step("Install px-monitor then validate", func() {
			var endpoint, oidcSecret string
			Step("Getting px-backup UI endpoint IP:PORT", func() {
				endpointIP := node.GetNodes()[0].GetMgmtIp()

				serviceObj, err := core.Instance().GetService("px-backup-ui", context.GetID())
				Expect(err).NotTo(HaveOccurred())
				endpointPort := serviceObj.Spec.Ports[0].NodePort

				endpoint = fmt.Sprintf("%s:%v", endpointIP, endpointPort)
				logrus.Infof("Got px-backup-ui endpoint: %s", endpoint)
			})

			Step("Getting OIDC client secret", func() {
				secretObj, err := core.Instance().GetSecret("pxc-backup-secret", context.GetID())
				Expect(err).NotTo(HaveOccurred())

				secretData, exist := secretObj.Data["OIDC_CLIENT_SECRET"]
				Expect(exist).To(Equal(true))
				oidcSecret = base64.StdEncoding.EncodeToString(secretData)
				logrus.Infof("Got OIDC client secret: %s", oidcSecret)
			})

			Step("Adding extra values to helm ConfigMap", func() {
				configMap, err := core.Instance().GetConfigMap(monitorApp, "default")
				Expect(err).NotTo(HaveOccurred())

				configMap.Data[k8s.HelmExtraValues] = fmt.Sprintf("pxmonitor.pxCentralEndpoint=%s,pxmonitor.oidcClientSecret=%s",
					                                              endpoint,
														 		  oidcSecret)
				configMap, err = core.Instance().UpdateConfigMap(configMap)
				Expect(err).NotTo(HaveOccurred())
				logrus.Infof("Updated helm values config map for px-monitor: %s", configMap.Data[monitorApp])
			})

			Step("Install px-monitor", func() {
				err := Inst().S.AddTasks(context, monitorOptions)
				Expect(err).NotTo(HaveOccurred())

				ValidateContext(context)
				logrus.Infof("Successfully validated specs for px-monitor")

				// remove extra values in config map
				configMap, err := core.Instance().GetConfigMap(monitorApp, "default")
				configMap.Data[k8s.HelmExtraValues] = ""
				configMap, err = core.Instance().UpdateConfigMap(configMap)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Step("Uninstall license server and monitoring", func() {
			err := Inst().S.ScheduleUninstall(context, monitorOptions)
			Expect(err).NotTo(HaveOccurred())
			err = Inst().S.ScheduleUninstall(context, lsOptions)
			Expect(err).NotTo(HaveOccurred())

			ValidateContext(context)
			logrus.Infof("Successfully uninstalled px-license-server and px-monitor")
		})

		Step("destroy apps", func() {
			opts := make(map[string]bool)
			opts[scheduler.OptionsWaitForResourceLeakCleanup] = true

			TearDownContext(context, opts)
			logrus.Infof("Successfully destroyed px-central")

			err := core.Instance().DeleteNamespace(context.GetID())
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

// This test installs px-central from release repo then upgrade using staging repo
// testing 1.2.3 -> 1.3.0 multi charts to single chart upgrade
// TODO: need to clarify what it the upgrade path and steps, once the new chart is ready
var _ = Describe("{Upgradepxcentral}", func() {
	It("has to setup, upgrade, validate and teardown apps", func() {
		var context *scheduler.Context

		centralApp := "px-central"
		centralOptions := scheduler.ScheduleOptions{
			AppKeys:            []string{centralApp},
			StorageProvisioner: Inst().Provisioner,
		}

		Step("Install px-central using 1.2.x helm repo then validate", func() {
			contexts, err := Inst().S.Schedule(Inst().InstanceID, centralOptions)
			Expect(err).NotTo(HaveOccurred())
			Expect(contexts).NotTo(BeEmpty())

			context = contexts[0]
			context.SkipVolumeValidation = true
			context.ReadinessTimeout = appReadinessTimeout

			ValidateContext(context)
			logrus.Infof("Successfully validated specs for px-central")
		})

		Step("Upgrade px-central using single chart helm repo then validate", func() {
			err := batch.Instance().DeleteJob("pxcentral-post-install-hook", context.GetID())
			Expect(err).NotTo(HaveOccurred())

			// set a different helm chart repo name for staging chart
			// and update the chart name from px-backup to px-central
			configMap, err := core.Instance().GetConfigMap(centralApp, "default")
			Expect(err).NotTo(HaveOccurred())

			configMap.Data[k8s.HelmRepoName] = "portworx-staging"
			configMap.Data[k8s.HelmChartName] = centralApp
			_, err = core.Instance().UpdateConfigMap(configMap)
			Expect(err).NotTo(HaveOccurred())

			centralOptions.Upgrade = true
			err = Inst().S.AddTasks(context, centralOptions)
			Expect(err).NotTo(HaveOccurred())

			ValidateContext(context)
			logrus.Infof("Successfully upgraded px-central to staging version")
		})

		Step("destroy apps", func() {
			opts := make(map[string]bool)
			opts[scheduler.OptionsWaitForResourceLeakCleanup] = true

			TearDownContext(context, opts)
			logrus.Infof("Successfully destroyed px-central")

			err := core.Instance().DeleteNamespace(context.GetID())
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

var _ = Describe("{InstallpxcentralWithoutBackup}", func() {
	It("has to setup, validate and teardown apps", func() {
		var context *scheduler.Context

		centralApp := "px-central"
		centralOptions := scheduler.ScheduleOptions{
			AppKeys:            []string{centralApp},
			StorageProvisioner: Inst().Provisioner,
		}

		Step("Install px-central with px-backup disabled then validate", func() {
			configMap, err := core.Instance().GetConfigMap(centralApp, "default")
			Expect(err).NotTo(HaveOccurred())

			configMap.Data[k8s.HelmExtraValues] = "pxbackup.enabled=false"
			_, err = core.Instance().UpdateConfigMap(configMap)
			Expect(err).NotTo(HaveOccurred())

			contexts, err := Inst().S.Schedule(Inst().InstanceID, centralOptions)
			Expect(err).NotTo(HaveOccurred())
			Expect(contexts).NotTo(BeEmpty())

			context = contexts[0]
			context.SkipVolumeValidation = true
			context.ReadinessTimeout = appReadinessTimeout

			ValidateContext(context)
			logrus.Infof("Successfully validated specs for px-central")
		})

		Step("destroy apps", func() {
			opts := make(map[string]bool)
			opts[scheduler.OptionsWaitForResourceLeakCleanup] = true

			TearDownContext(context, opts)
			logrus.Infof("Successfully destroyed px-central")

			// remove extra values in config map
			configMap, err := core.Instance().GetConfigMap(centralApp, "default")
			configMap.Data[k8s.HelmExtraValues] = ""
			_, err = core.Instance().UpdateConfigMap(configMap)
			Expect(err).NotTo(HaveOccurred())

			err = core.Instance().DeleteNamespace(context.GetID())
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

var _ = AfterSuite(func() {
	PerformSystemCheck()
	ValidateCleanup()
})

func TestMain(m *testing.M) {
	ParseFlags()
	os.Exit(m.Run())
}
