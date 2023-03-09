package lib

import (
	"sync"
	"time"

	"github.com/portworx/torpedo/drivers/node"
	"github.com/portworx/torpedo/drivers/scheduler"

	// import scheduler drivers to invoke it's init
	_ "github.com/portworx/torpedo/drivers/scheduler/dcos"

	// import scheduler drivers to invoke it's init
	_ "github.com/portworx/torpedo/drivers/scheduler/openshift"
	_ "github.com/portworx/torpedo/drivers/scheduler/rke"

	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
	v1 "k8s.io/api/apps/v1"
)

const (
	defaultWaitRebootTimeout     = 5 * time.Minute
	defaultWaitRebootRetry       = 10 * time.Second
	defaultCommandRetry          = 5 * time.Second
	defaultCommandTimeout        = 1 * time.Minute
	defaultTestConnectionTimeout = 15 * time.Minute
	defaultRebootTimeRange       = 5 * time.Minute
)

// PDS vars
var (
	wg                        sync.WaitGroup
	ResiliencyFlag            = false
	hasResiliencyConditionMet = false
	FailureType               = "node-reboot"
)

// Executes all methods in parallel
func ExecuteInParallel(functions ...func()) {
	wg.Add(len(functions))
	defer wg.Wait()
	for _, fn := range functions {
		go func(copy func()) {
			defer wg.Done()
			copy()
		}(fn)
	}
}

// Function to enable Resiliency Test
func MarkResiliencyTC(resiliency bool) {
	ResiliencyFlag = resiliency
}

// Function to wait for event to induce failure
func InduceFailure(failure string, ns string) {
	for !hasResiliencyConditionMet {
		continue
	}
	log.InfoD("================ SS is now at 2 replica active pods ==============")
	ResiliencyDriver(failure, ns)
	log.InfoD("================ Came out of Resiliency Driver ===================")
}

// Resiliency Driver Module
func ResiliencyDriver(failure string, ns string) {
	if failure == "node-reboot" {
		RebootActiveNode(ns)
	}
}

// Reboot Any one Active Node
func RebootActiveNode(ns string) error {
	// Get StatefulSet Object
	var ss *v1.StatefulSet
	ss, err = k8sApps.GetStatefulSet(deployment.GetClusterResourceName(), ns)
	if err != nil {
		return err
	}
	// Get Pods of this StatefulSet
	pods, err := k8sApps.GetStatefulSetPods(ss)
	if err != nil {
		return err
	}
	// Check which Pod is still not up. Try to reboot the node on which this Pod is hosted.
	for _, pod := range pods {
		if k8sCore.IsPodReady(pod) {
			log.InfoD("Pod running on Node %v is Ready so skipping this pod......", pod.Spec.NodeName)
			continue
		} else {
			log.InfoD(" ================ Node Selected is : %v =================", pod.Spec.NodeName)
			err = Inst().S.Init(scheduler.InitOptions{
				NodeDriverName: Inst().N.String(),
			})

			log.FailOnError(err, "Error occured while Scheduler Driver Initialization")

			err = Inst().N.Init(node.InitOptions{
				SpecDir: Inst().SpecDir,
			})
			log.FailOnError(err, "Error occured while Node Driver Initialization")

			nodes := node.GetWorkerNodes()
			var nodeToReboot node.Node
			for _, n := range nodes {
				log.Info("============= Checking Node ============ : %s", n.Name)
				if n.Name != pod.Spec.NodeName {
					continue
				}
				nodeToReboot = n
			}

			log.InfoD(" ================ Rebooting the above node %v ===================", nodeToReboot.Name)
			err = Inst().N.RebootNode(nodeToReboot, node.RebootNodeOpts{
				Force: true,
				ConnectionOpts: node.ConnectionOpts{
					Timeout:         defaultCommandTimeout,
					TimeBeforeRetry: defaultCommandRetry,
				},
			})
		}
	}
	return nil
}
