package lib

import (
	"errors"
	"sync"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	"github.com/portworx/torpedo/drivers/node"

	_ "github.com/portworx/torpedo/drivers/scheduler/dcos"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/portworx/torpedo/pkg/log"
	"github.com/portworx/torpedo/tests"
)

const (
	active_node_reboot_during_deployment = "active-node-reboot-during-deployment"
)

// PDS vars
var (
	wg                        sync.WaitGroup
	ResiliencyFlag            = false
	hasResiliencyConditionMet = false
	FailureType               TypeOfFailure
	testError                 error
	check_till_replica        int32
	ResiliencyCondition       = make(chan bool)
)

// Struct Definition for kind of Failure the framework needs to trigger
type TypeOfFailure struct {
	Type   string
	Method func() error
}

// Wrapper to Define failure type from Test Case
func DefineFailureType(failuretype TypeOfFailure) {
	FailureType = failuretype
}

// Executes all methods in parallel
func ExecuteInParallel(functions ...func()) {
	wg.Add(len(functions))
	defer wg.Wait()
	for _, fn := range functions {
		go func(FuncToRun func()) {
			defer wg.Done()
			FuncToRun()
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
	// for !hasResiliencyConditionMet {
	// 	continue
	// }
	isResiliencyConditionset := <-ResiliencyCondition
	if isResiliencyConditionset {
		FailureType.Method()
	} else {
		testError = errors.New("Resiliency Condition did not meet. Failing this test case.")
		return
	}
	// Triggering Resiliency Failure now
	// ResiliencyDriver(failure, ns)
}

// Resiliency Driver Module
func ResiliencyDriver(failure string, ns string) {
	if failure == active_node_reboot_during_deployment {
		FailureType.Method()
	}
}

// Close all open Resiliency channels here
func CloseResiliencyChannel() {
	close(ResiliencyCondition)
}

//
func InduceFailureAfterWaitingForCondition(deployment *pds.ModelsDeployment, namespace string) error {
	switch FailureType.Type {
	// Case when we want to reboot a node onto which a deployment pod is coming up
	case active_node_reboot_during_deployment:
		check_till_replica = 1
		log.InfoD("Entering to check if Data service has %v active pods. Once it does, we will reboot the node it is hosted upon.", check_till_replica)
		func1 := func() {
			GetPdsSs(deployment.GetClusterResourceName(), namespace, check_till_replica)
		}
		func2 := func() {
			InduceFailure(FailureType.Type, namespace)
		}
		ExecuteInParallel(func1, func2)
		if testError != nil {
			return testError
		}
	}
	err := ValidateDataServiceDeployment(deployment, namespace)
	return err
}

// Reboot the Active Node onto which the application pod is coming up
func RebootActiveNodeDuringDeployment(ns string) error {
	// Get StatefulSet Object
	var ss *v1.StatefulSet
	// ss, testError = k8sApps.GetStatefulSet(deployment.GetClusterResourceName(), ns)
	// if testError != nil {
	// 	return testError
	// }
	// Waiting till atleast first pod have a node assigned
	var pods []corev1.Pod
	err = wait.Poll(resiliencyInterval, timeOut, func() (bool, error) {
		ss, testError = k8sApps.GetStatefulSet(deployment.GetClusterResourceName(), ns)
		if testError != nil {
			return false, testError
		}
		// Get Pods of this StatefulSet
		pods, testError = k8sApps.GetStatefulSetPods(ss)
		if testError != nil {
			return false, testError
		}
		// Check if Pods have a node assigned or it's in a window where it's just coming up
		for _, pod := range pods {
			log.Infof("======= Nodename of pod %v is :%v:========== ", pod.Name, pod.Spec.NodeName)
			if pod.Spec.NodeName == "" || pod.Spec.NodeName == " " {
				log.Infof("Pod %v still does not have a node assigned. Retrying in 5 seconds", pod.Name)
				return false, nil
			} else {
				return true, nil
			}
		}
		return true, nil
	})
	// Get Pods of this StatefulSet
	// pods, testError := k8sApps.GetStatefulSetPods(ss)
	// if testError != nil {
	// 	return testError
	// }
	// Check which Pod is still not up. Try to reboot the node on which this Pod is hosted.
	for _, pod := range pods {
		log.Infof("Checking Pod %v running on Node: %v", pod.Name, pod.Spec.NodeName)
		if k8sCore.IsPodReady(pod) {
			log.InfoD("This Pod running on Node %v is Ready so skipping this pod......", pod.Spec.NodeName)
			continue
		} else {
			// nodes := node.GetWorkerNodes()
			var nodeToReboot node.Node
			// for _, n := range nodes {
			// 	if n.Name != pod.Spec.NodeName {
			// 		continue
			// 	}
			// 	nodeToReboot = n
			// }
			nodeToReboot, testError = node.GetNodeByName(pod.Spec.NodeName)
			if testError != nil {
				return testError
			}
			if nodeToReboot.Name == "" {
				testError = errors.New("Something happened and node is coming out to be empty from Node registry")
				return testError
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
