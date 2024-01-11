package tests

import (
	"sync"
	"time"

	. "github.com/onsi/ginkgo"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
	. "github.com/portworx/torpedo/tests/backup/longevity/pxbackuplongevitytriggers"
)

// func main() {
// 	var wg sync.WaitGroup
// 	var startTime = time.Now()

// 	for {
// 		startTime = TriggerLongevityWorkflows(startTime, &wg)
// 		time.Sleep(2 * time.Second)
// 	}
// }

var _ = Describe("{BackupLongevityTest}", func() {

	JustBeforeEach(func() {
		log.Infof("Inside Just before each")
		StartPxBackupTorpedoTest("BackupLongevityTest",
			"Starting longevity test", nil, 00000, ATrivedi, Q1FY24)
	})

	It("Running backup longevity tests", func() {
		Step("Running Longevity", func() {
			var wg sync.WaitGroup
			var startTime = time.Now()

			for {
				startTime = TriggerLongevityWorkflows(startTime, &wg)
				time.Sleep(2 * time.Second)
			}
		})
	})

	JustAfterEach(func() {
		log.Infof("Just after Each")
	})

})
