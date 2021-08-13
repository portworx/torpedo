package tests

import (
	"fmt"
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
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
		switch Inst().ChaosLevel {
		case 10:
			frequency = 100
		case 9:
			frequency = 90
		case 8:
			frequency = 80
		case 7:
			frequency = 70
		case 6:
			frequency = 60
		case 5:
			frequency = 50
		case 4:
			frequency = 40
		case 3:
			frequency = 30
		case 2:
			frequency = 20
		case 1:
			frequency = 10
		default:
			frequency = 10

		}

		customAppConfig := scheduler.AppConfig{
			ClaimsCount: frequency,
		}

		provider := Inst().V.String()

		Inst().CustomAppConfig["vdbench-sv4-svc-multivol"] = customAppConfig
		err := Inst().S.RescanSpecs(Inst().SpecDir, provider)

		Expect(err).NotTo(HaveOccurred(),
			fmt.Sprintf("Failed to rescan specs from %s for storage provider %s. Error: [%v]",
				Inst().SpecDir, provider, err))

		Step("schedule application with multiple sharedv4-svc volumes attached", func() {
			logrus.Infof("Number of Volumes to be mounted: %v", frequency)
			contexts := []*scheduler.Context{}
			taskName := "sharedv4-svc-multivol"

			logrus.Infof("Task name %s\n", taskName)

			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				newContexts := ScheduleApplications(taskName)
				contexts = append(contexts, newContexts...)
			}

			for _, ctx := range contexts {
				ctx.SkipVolumeValidation = false
				ValidateContext(ctx)

			}
		})

	})

})

func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags
	ParseFlags()
	os.Exit(m.Run())
}
