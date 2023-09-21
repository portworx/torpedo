package tests

import (
	"github.com/portworx/torpedo/pkg/log"
	"os"
	"sync"

	. "github.com/onsi/ginkgo"
	"github.com/portworx/torpedo/drivers/scheduler"
	. "github.com/portworx/torpedo/tests"
)

var _ = Describe("{PxUpgradeTest}", func() {
	BeforeEach(func() {

	})
	triggerEventsChan := make(chan *EventRecord, 100)
	contexts := make([]*scheduler.Context, 0)
	var triggerLock sync.Mutex
	triggerFunctions = map[string]func(*[]*scheduler.Context, *chan *EventRecord){
		UpgradePortwox: TriggerUpgradeVolumeDriver,
	}
	Inst().UpgradeStorageDriverEndpointList = "https://edge-install.portworx.com/3.0.2/"
	It("upgrade portworx", func() {
		var wg sync.WaitGroup
		Step("Register test triggers", func() {
			for triggerType, triggerFunc := range triggerFunctions {
				log.InfoD("Registering trigger: [%v]", triggerType)
				go testTriggerUpgrade(&wg, &contexts, triggerType, triggerFunc, &triggerLock, &triggerEventsChan)
				wg.Add(1)
			}
		})
	})

	JustAfterEach(func() {

	})
})

func testTriggerUpgrade(wg *sync.WaitGroup,
	contexts *[]*scheduler.Context,
	triggerType string,
	triggerFunc func(*[]*scheduler.Context, *chan *EventRecord),
	triggerLoc *sync.Mutex,
	triggerEventsChan *chan *EventRecord) {
	defer wg.Done()
	log.Infof("Waiting for lock for trigger [%s]\n", triggerType)
	triggerLoc.Lock()
	log.Infof("Successfully taken lock for trigger [%s]\n", triggerType)
	triggerFunc(contexts, triggerEventsChan)
	log.Infof("Trigger Function completed for [%s]\n", triggerType)
	triggerLoc.Unlock()
	log.Infof("Successfully released lock for trigger [%s]\n", triggerType)
	os.Exit(0)
}
