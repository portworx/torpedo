package pdslibs

import (
	"errors"
	"fmt"
	"github.com/portworx/torpedo/drivers/pds/parameters"
	k8utils "github.com/portworx/torpedo/drivers/utilities"
	"github.com/portworx/torpedo/pkg/log"
	"sync"
)

const (
	StopPXDuringStorageResize                         = "stop-px-during-storage-resize"
	DeletePdsDeploymentPodAndValidateDeploymentHealth = "delete-pdsDeploymentPods-validate-deployment-health"
)

const (
	PDS_DEPLOYMENT_AVAILABLE   = "AVAILABLE"
	PDS_DEPLOYMENT_UNAVAILABLE = "UNAVAILABLE"
)

var (
	wg                        sync.WaitGroup
	ResiFlag                  = false
	CapturedErrors            = make(chan error, 10)
	ResiliencyCondition       = make(chan bool)
	hasResiliencyConditionMet = false
	NewPdsParams              *parameters.NewPDSParams
)

// TypeOfFailure Struct Definition for kind of Failure the framework needs to trigger
type TypeOfFailure struct {
	Type   string
	Method func() error
}

// CloseResiliencyChannel Close all open Resiliency channels here
func CloseResiliencyChannel() {
	// Close the Channel if it's empty. Otherwise there is no need to close as per Golang official documentation,
	// as far as we are making sure no writes are happening to a closed channel. Make sure to call this method only
	// during Post Test Case execution to avoid any unknown panics
	if len(ResiliencyCondition) == 0 {
		close(ResiliencyCondition)
	}
}

// InduceFailure Function to wait for event to induce failure
func InduceFailure(failure TypeOfFailure, ns string) {
	isResiliencyConditionset := <-ResiliencyCondition
	if isResiliencyConditionset {
		err := failure.Method()
		if err != nil {
			return
		}
	} else {
		CapturedErrors <- errors.New("Resiliency Condition did not meet. Failing this test case.")
		return
	}
	return
}

// ExecuteInParallel Executes all methods in parallel
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

func GenerateFailureTypeAndMethod(failuretype string, FunctionMap map[string]error) TypeOfFailure {
	var method error
	// Call a function based on a string argument
	if fn, ok := FunctionMap[failuretype]; ok {
		method = fn
	} else {
		fmt.Println("Function not found")
	}
	failureScenario := TypeOfFailure{
		Type: failuretype,
		Method: func() error {
			return method
		},
	}
	return failureScenario
}

// DefineFailureType Wrapper to Define failure type from Test Case
func DefineFailureType(failureType string, funcMap map[string]error) TypeOfFailure {
	return GenerateFailureTypeAndMethod(failureType, funcMap)
}

// InduceFailureAfterWaitingForCondition executes failure injection method and validation method in goroutines
func InduceFailureAfterWaitingForCondition(ds PDSDataService, dsPodName, deploymentId, namespaceId, projectId, imageId, appConfigId, resConfigId, stConfigId, namespace, failureType string, resiFlag bool) error {
	ResiFlag = resiFlag
	switch failureType {
	case StopPXDuringStorageResize:
		FunctionMap := map[string]error{
			StopPXDuringStorageResize: k8utils.StopPxOnReplicaVolumeNode(),
		}
		FailureTypeAndMethod := DefineFailureType(failureType, FunctionMap)

		log.InfoD("Entering to resize of the Data service Volume, while PX on volume node is stopped")
		func1 := func() {
			UpdateDataService(ds, deploymentId, namespaceId, projectId, imageId, appConfigId, resConfigId, stConfigId)
		}
		func2 := func() {
			InduceFailure(FailureTypeAndMethod, namespace)
		}
		ExecuteInParallel(func1, func2)
	case DeletePdsDeploymentPodAndValidateDeploymentHealth:
		log.Debugf("dsPodName [%s]", dsPodName)
		log.Debugf("namespace [%s]", namespace)
		FunctionMap := map[string]error{
			DeletePdsDeploymentPodAndValidateDeploymentHealth: k8utils.DeleteK8sPod(dsPodName, namespace, true),
		}
		FailureTypeAndMethod := DefineFailureType(failureType, FunctionMap)

		func1 := func() {
			ValidateDataServiceDeploymentHealth(deploymentId, PDS_DEPLOYMENT_UNAVAILABLE)
		}
		func2 := func() {
			InduceFailure(FailureTypeAndMethod, namespace)
		}
		ExecuteInParallel(func1, func2)
		ResiFlag = false
		ValidateDataServiceDeploymentHealth(deploymentId, PDS_DEPLOYMENT_AVAILABLE)
	}

	var aggregatedError error
	for w := 1; w <= len(CapturedErrors); w++ {
		if err := <-CapturedErrors; err != nil {
			aggregatedError = fmt.Errorf("%v : %v", aggregatedError, err)
		}
	}
	if aggregatedError != nil {
		return aggregatedError
	}
	return err
}
