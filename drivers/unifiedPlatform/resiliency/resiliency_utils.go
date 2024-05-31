package resiliency

import (
	"errors"
	"fmt"
	dslibs "github.com/portworx/torpedo/drivers/unifiedPlatform/pdsLibs"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/platformLibs"
	"sync"
	"time"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	restoreBkp "github.com/portworx/torpedo/drivers/pds/pdsrestore"
	_ "github.com/portworx/torpedo/drivers/scheduler/dcos"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/pkg/log"
)

const (
	PdsDeploymentControllerManagerPod   = "pds-deployment-controller-manager"
	PdsAgentPod                         = "pds-agent"
	PdsTeleportPod                      = "pds-teleport"
	PdsBackupControllerPod              = "pds-backup-controller-manager"
	PdsTargetControllerPod              = "pds-operator-target-controller-manager"
	ActiveNodeRebootDuringDeployment    = "active-node-reboot-during-deployment"
	KillDeploymentControllerPod         = "kill-deployment-controller-pod-during-deployment"
	RestartPxDuringDSScaleUp            = "restart-portworx-during-ds-scaleup"
	RebootNodesDuringDeployment         = "reboot-multiple-nodes-during-deployment"
	KillAgentPodDuringDeployment        = "kill-agent-pod-during-deployment"
	RestartAppDuringResourceUpdate      = "restart-app-during-resource-update"
	RebootNodeDuringAppVersionUpdate    = "reboot-node-during-app-version-update"
	KillTeleportPodDuringDeployment     = "kill-teleport-pod-during-deployment"
	KillPdsAgentPodDuringAppScaleUp     = "kill-pds-agent-pod-during-app-scale-up"
	RestoreDSDuringPXPoolExpansion      = "restore-ds-during-px-pool-expansion"
	RestoreDSDuringKVDBFailOver         = "restore-ds-during-kvdb-fail-over"
	RestoreDuringAllNodesReboot         = "restore-ds-during-node-reboot"
	StopPXDuringStorageResize           = "stop-px-during-storage-resize"
	RebootNodeDuringAppResourceUpdate   = "reboot-node-during-app-resource-update"
	KillDbMasterNodeDuringStorageResize = "kill-db-master-node-during-storage-resize"
	poolResizeTimeout                   = time.Minute * 120
	retryTimeout                        = time.Minute * 2
)

// PDS vars
var (
	wg                        sync.WaitGroup
	ResiliencyFlag            = false
	hasResiliencyConditionMet = false
	FailureType               TypeOfFailure
	CapturedErrors            = make(chan error, 10)
	checkTillReplica          int32
	ResiliencyCondition       = make(chan bool)
	restoredDeployment        *pds.ModelsDeployment
	dsEntity                  restoreBkp.DSEntity
	DynamicDeployments        []*pds.ModelsDeployment
	RestoredDeployments       []*pds.ModelsDeployment
	UpdateTemplate            string
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
}

// Function to wait for event to induce failure
func InduceFailure(failure string, ns string) {
	isResiliencyConditionset := <-ResiliencyCondition
	if isResiliencyConditionset {
		FailureType.Method()
	} else {
		CapturedErrors <- errors.New("Resiliency Condition did not meet. Failing this test case.")
		return
	}
	return
}

func InduceFailureAfterWaitingForCondition(deployment *automationModels.V1Deployment, namespace string, CheckTillReplica int32, ds dslibs.PDSDataService) error {
	switch FailureType.Type {
	// Case when we want to reboot a node onto which a deployment pod is coming up
	case ActiveNodeRebootDuringDeployment:
		checkTillReplica = CheckTillReplica
		log.InfoD("Entering to check if Data service has %v active pods. Once it does, we will reboot the node it is hosted upon.", checkTillReplica)
		func1 := func() {
			GetPdsSs(*deployment.Status.CustomResourceName, namespace, checkTillReplica)
		}
		func2 := func() {
			InduceFailure(FailureType.Type, namespace)
		}
		ExecuteInParallel(func1, func2)

	case KillAgentPodDuringDeployment:
		checkTillReplica = CheckTillReplica
		log.InfoD("Entering to check if Data service has %v active pods. Once it does, we will reboot the node it is hosted upon.", checkTillReplica)
		func1 := func() {
			GetPdsSs(*deployment.Status.CustomResourceName, namespace, checkTillReplica)
		}
		func2 := func() {
			InduceFailure(FailureType.Type, namespace)
		}
		ExecuteInParallel(func1, func2)

	case RebootNodesDuringDeployment:
		checkTillReplica = CheckTillReplica
		log.InfoD("Entering to check if Data service has %v active pods. Once it does, we will reboot the node it is hosted upon.", checkTillReplica)
		func1 := func() {
			GetPdsSs(*deployment.Status.CustomResourceName, namespace, checkTillReplica)
		}
		func2 := func() {
			InduceFailure(FailureType.Type, namespace)
		}
		ExecuteInParallel(func1, func2)

	case KillPdsAgentPodDuringAppScaleUp:
		checkTillReplica = CheckTillReplica
		log.InfoD("Entering to check if Data service has %v active pods. Once it does, we will reboot the node it is hosted upon.", checkTillReplica)
		func1 := func() {
			GetPdsSs(*deployment.Status.CustomResourceName, namespace, checkTillReplica)
		}
		func2 := func() {
			InduceFailure(FailureType.Type, namespace)
		}
		ExecuteInParallel(func1, func2)

	case StopPXDuringStorageResize:
		log.InfoD("Entering to resize of the Data service Volume, while PX on volume node is stopped")
		tenantID, err := platformLibs.GetDefaultTenantId(AccountID)
		if err != nil {
			return err
		}

		nameSpace, err := platformLibs.GetNamespace(tenantID, namespace)
		if err != nil {
			return err
		}

		func1 := func() {
			ResizeDataServiceStorage(deployment, ds, *nameSpace.Meta.Uid, UpdateTemplate)
		}
		func2 := func() {
			InduceFailure(FailureType.Type, namespace)
		}
		ExecuteInParallel(func1, func2)

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

	return nil
}
