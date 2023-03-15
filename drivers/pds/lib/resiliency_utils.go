package lib

import (
	"sync"
	"time"

	"github.com/portworx/torpedo/drivers/node"

	_ "github.com/portworx/torpedo/drivers/scheduler/dcos"
	v1 "k8s.io/api/apps/v1"

	"github.com/portworx/torpedo/pkg/log"
	"github.com/portworx/torpedo/tests"
)

const (
	defaultWaitRebootTimeout             = 5 * time.Minute
	defaultWaitRebootRetry               = 10 * time.Second
	defaultCommandRetry                  = 5 * time.Second
	defaultCommandTimeout                = 1 * time.Minute
	defaultTestConnectionTimeout         = 15 * time.Minute
	defaultRebootTimeRange               = 5 * time.Minute
	active_node_reboot_during_deployment = "active-node-reboot-during-deployment"
)

// PDS vars
var (
	wg                        sync.WaitGroup
	ResiliencyFlag            = false
	hasResiliencyConditionMet = false
	FailureType               ResiliencyFailure
	testError                 error
	check_till_replica        int32
)

// Struct Definition for kind of Failure the framework needs to trigger
type ResiliencyFailure struct {
	Type   string
	Method func() error
}

// Wrapper to Define failure type from Test Case
func DefineFailureType(failure ResiliencyFailure) {
	FailureType = failure
}

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
	if resiliency {
		tests.InitInstance()
	}
}

// Function to wait for event to induce failure
func InduceFailure(failure string, ns string) {
	for !hasResiliencyConditionMet {
		continue
	}
	// Triggering Resiliency Failure now
	ResiliencyDriver(failure, ns)
}

// Resiliency Driver Module
func ResiliencyDriver(failure string, ns string) {
	if failure == active_node_reboot_during_deployment {
		FailureType.Method()
	}
}

// Reboot the Active Node onto which the application pod is coming up
func RebootActiveNodeDuringDeployment(ns string) error {
	// Get StatefulSet Object
	var ss *v1.StatefulSet
	ss, testError = k8sApps.GetStatefulSet(deployment.GetClusterResourceName(), ns)
	if testError != nil {
		return testError
	}
	// Get Pods of this StatefulSet
	pods, testError := k8sApps.GetStatefulSetPods(ss)
	if testError != nil {
		return testError
	}
	// Check which Pod is still not up. Try to reboot the node on which this Pod is hosted.
	for _, pod := range pods {
		if k8sCore.IsPodReady(pod) {
			log.InfoD("Pod running on Node %v is Ready so skipping this pod......", pod.Spec.NodeName)
			continue
		} else {
			nodes := node.GetWorkerNodes()
			var nodeToReboot node.Node
			for _, n := range nodes {
				if n.Name != pod.Spec.NodeName {
					continue
				}
				nodeToReboot = n
			}
			log.Infof("Going ahead and rebooting the node %v as there is an application pod thats coming up on this node", pod.Spec.NodeName)
			testError = tests.Inst().N.RebootNode(nodeToReboot, node.RebootNodeOpts{
				Force: true,
				ConnectionOpts: node.ConnectionOpts{
					Timeout:         defaultCommandTimeout,
					TimeBeforeRetry: defaultCommandRetry,
				},
			})
			if testError != nil {
				return testError
			}
			log.Infof("Node %v rebooted successfully", pod.Spec.NodeName)
		}
	}
	return testError
}
