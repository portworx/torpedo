package unifiedPlatform

import (
	"encoding/json"
	"fmt"
	"github.com/portworx/sched-ops/k8s/apps"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/portworx/sched-ops/task"
	"github.com/portworx/torpedo/drivers/volume"
	"github.com/portworx/torpedo/pkg/units"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/portworx/torpedo/drivers/node"

	"github.com/portworx/sched-ops/k8s/core"
	"github.com/portworx/torpedo/drivers/pds/parameters"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	dslibs "github.com/portworx/torpedo/drivers/unifiedPlatform/pdsLibs"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/stworkflows/pds"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/stworkflows/platform"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
	corev1 "k8s.io/api/core/v1"
)

var (
	WorkflowDataService     pds.WorkflowDataService
	WorkflowPDSBackupConfig pds.WorkflowPDSBackupConfig
	WorkflowPDSBackup       pds.WorkflowPDSBackup
	WorkflowPDSRestore      pds.WorkflowPDSRestore
	WorkflowPDSTemplate     pds.WorkflowPDSTemplates
)

var (
	k8sApps = apps.Instance()
	k8sCore = core.Instance()
)

const (
	EnvControlPlaneUrl     = "CONTROL_PLANE_URL"
	DefaultTestAccount     = "pds-qa"
	DefaultProject         = "PDS-Project"
	DefaultTenant          = "px-system-tenant"
	EnvPlatformAccountName = "PLATFORM_ACCOUNT_NAME"
	EnvAccountDisplayName  = "PLATFORM_ACCOUNT_DISPLAY_NAME"
	EnvUserMailId          = "USER_MAIL_ID"
	defaultParams          = "../drivers/pds/parameters/pds_default_parameters.json"
	pdsParamsConfigmap     = "pds-params"
	configmapNamespace     = "default"
)

const (
	defaultWaitRebootRetry       = 10 * time.Second
	defaultCommandRetry          = 5 * time.Second
	defaultCommandTimeout        = 1 * time.Minute
	defaultTestConnectionTimeout = 15 * time.Minute
)

var (
	AccID         string
	PDSBucketName string
	Namespace     string
	ProjectId     string
)

var (
	WorkflowPlatform                 platform.WorkflowPlatform
	WorkflowTargetCluster            platform.WorkflowTargetCluster
	WorkflowTargetClusterDestination platform.WorkflowTargetCluster
	WorkflowProject                  platform.WorkflowProject
	WorkflowNamespace                platform.WorkflowNamespace
	WorkflowNamespaceDestination     platform.WorkflowNamespace
	WorkflowCc                       platform.WorkflowCloudCredentials
	WorkflowbkpLoc                   platform.WorkflowBackupLocation
	WorkflowRepear                   platform.WorkflowRepear
	WorkflowSchedule                 platform.WorkflowSchedule
	NewPdsParams                     *parameters.NewPDSParams
	PdsLabels                        = make(map[string]string)
	PDS_DEFAULT_NAMESPACE            string
	DsNameAndAppTempId               map[string]string
	StTemplateId                     string
	ResourceTemplateId               string
	TemplateIds                      []string
)

// ReadParams reads the params from given or default json
func ReadNewParams(filename string) (*parameters.NewPDSParams, error) {
	var jsonPara parameters.NewPDSParams
	var err error

	if filename == "" {
		filename, err = filepath.Abs(defaultParams)
		log.Infof("filename %v", filename)
		if err != nil {
			return nil, err
		}
		log.Infof("Parameter json file is not used, use initial parameters value.")
		log.InfoD("Reading params from %v ", filename)
		file, err := ioutil.ReadFile(filename)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(file, &jsonPara)
		if err != nil {
			return nil, err
		}
	} else {
		cm, err := core.Instance().GetConfigMap(pdsParamsConfigmap, configmapNamespace)
		if err != nil {
			return nil, err
		}
		if len(cm.Data) > 0 {
			configmap := &cm.Data
			for key, data := range *configmap {
				log.InfoD("key %v \n value %v", key, data)
				jsonData := []byte(data)
				err = json.Unmarshal(jsonData, &jsonPara)
				if err != nil {
					return &jsonPara, err
				}
			}
		}
	}
	return &jsonPara, nil
}

// EndPDSTorpedoTest ends the logging for PDS torpedo test and updates results in testrail
func EndPDSTorpedoTest() {
	defer EndTorpedoTest()

	if NewPdsParams.BackUpAndRestore.RunBkpAndRestrTest {
		defer func() {
			err := SetSourceKubeConfig()
			log.FailOnError(err, "failed to switch context to source cluster")
		}()
	}

	Step("Purging all PDS related objects", func() {
		errors := PurgePDS()
		if len(errors) > 0 {
			var errorStrings []string
			for _, err := range errors {
				errorStrings = append(errorStrings, err.Error())
			}
			log.FailOnError(fmt.Errorf("[%s]", strings.Join(errorStrings, "\n\n")), "errors occurred while cleanup")
		}
	})

}

// StartPDSTorpedoTest starts the logging for PDS torpedo test
func StartPDSTorpedoTest(testName string, testDescription string, tags map[string]string, testRepoID int) {

	PDS_DEFAULT_NAMESPACE = "pds-namespace-" + RandomString(5)

	Step("Create a namespace for PDS", func() {
		WorkflowNamespace.TargetCluster = &WorkflowTargetCluster
		WorkflowNamespace.Namespaces = make(map[string]string)
		_, err := WorkflowNamespace.CreateNamespaces(PDS_DEFAULT_NAMESPACE)
		log.FailOnError(err, "Unable to create namespace")
		log.Infof("Namespaces created - [%s]", WorkflowNamespace.Namespaces)
	})

	Step("Associate namespace to Project", func() {
		err := WorkflowProject.Associate(
			[]string{},
			[]string{WorkflowNamespace.Namespaces[PDS_DEFAULT_NAMESPACE]},
			[]string{},
			[]string{},
			[]string{},
			[]string{},
		)
		log.FailOnError(err, "Unable to associate Cluster to Project")
		log.Infof("Associated Resources - [%+v]", WorkflowProject.AssociatedResources)
	})

	Step("Creating all PDS related structs", func() {

		WorkflowNamespaceDestination.TargetCluster = &WorkflowTargetClusterDestination
		WorkflowNamespaceDestination.Namespaces = make(map[string]string)

		log.Infof("Creating schedule struct")
		WorkflowSchedule.Platform = &WorkflowPlatform
		WorkflowSchedule.Schedules = make(map[string]*automationModels.V1BackupPolicy)

		log.Infof("Creating data service struct")

		WorkflowDataService.Namespace = &WorkflowNamespace
		WorkflowDataService.SkipValidatation = make(map[string]bool)
		WorkflowDataService.DataServiceDeployment = make(map[string]*dslibs.DataServiceDetails)
		WorkflowDataService.Dash = Inst().Dash
		WorkflowDataService.UpdateDeploymentTemplates = false
		WorkflowDataService.PDSTemplates = WorkflowPDSTemplate
		WorkflowDataService.PDSParams = NewPdsParams

		//Initializing the parameters required for workload generation
		WorkflowDataService.WorkloadGenParams = &dslibs.LoadGenParams{
			LoadGenDepName: NewPdsParams.LoadGen.LoadGenDepName,
			Namespace:      PDS_DEFAULT_NAMESPACE,
			NumOfRows:      NewPdsParams.LoadGen.NumOfRows,
			Timeout:        NewPdsParams.LoadGen.Timeout,
			Replicas:       NewPdsParams.LoadGen.Replicas,
			Iterations:     NewPdsParams.LoadGen.Iterations,
			FailOnError:    NewPdsParams.LoadGen.FailOnError,
		}

		log.Infof("Creating backup config struct")
		WorkflowPDSBackupConfig.WorkflowBackupLocation = WorkflowbkpLoc
		WorkflowPDSBackupConfig.WorkflowDataService = &WorkflowDataService
		WorkflowPDSBackupConfig.BackupConfigs = make(map[string]*pds.BackupConfigDetails)
		WorkflowPDSBackupConfig.SkipValidatation = make(map[string]bool)

		log.Infof("Creating Backup struct")
		WorkflowPDSBackup.WorkflowDataService = &WorkflowDataService
		WorkflowPDSBackup.Backups = make(map[string]*pds.BackupDetails)
		WorkflowPDSBackup.WorkflowBackupConfig = &WorkflowPDSBackupConfig

		log.Infof("Creating restore object for same cluster and same project")
		WorkflowPDSRestore.Source = &WorkflowDataService
		WorkflowPDSRestore.WorkflowBackup = &WorkflowPDSBackup
		WorkflowPDSRestore.Restores = make(map[string]automationModels.PDSRestore)
		WorkflowPDSRestore.Destination = &WorkflowNamespace
		WorkflowPDSRestore.Validatation = make(map[string]bool)
		WorkflowPDSRestore.SkipValidation = make(map[string]bool)
		WorkflowPDSRestore.RestoredDeployments = &pds.WorkflowDataService{
			PDSParams:    NewPdsParams,
			Namespace:    &WorkflowNamespace,
			Dash:         Inst().Dash,
			PDSTemplates: WorkflowPDSTemplate,
			WorkloadGenParams: &dslibs.LoadGenParams{
				LoadGenDepName: NewPdsParams.LoadGen.LoadGenDepName,
				Namespace:      PDS_DEFAULT_NAMESPACE,
				NumOfRows:      NewPdsParams.LoadGen.NumOfRows,
				Timeout:        NewPdsParams.LoadGen.Timeout,
				Replicas:       NewPdsParams.LoadGen.Replicas,
				Iterations:     NewPdsParams.LoadGen.Iterations,
				FailOnError:    NewPdsParams.LoadGen.FailOnError,
			},
		}
		WorkflowPDSRestore.RestoredDeployments.DataServiceDeployment = make(map[string]*dslibs.DataServiceDetails)

		log.Infof("Creating Platform object for Template Workflow")
		WorkflowPDSTemplate.Platform = WorkflowPlatform

	})

	instanceIDString := strconv.Itoa(testRepoID)
	timestamp := time.Now().Format("01-02-15h04m05s")
	Inst().InstanceID = fmt.Sprintf("%s-%s", instanceIDString, timestamp)
	StartTorpedoTest(testName, testDescription, tags, testRepoID)
}

// PurgePDS purges all default PDS related resources created during testcase run
func PurgePDS() []error {

	var allErrors []error

	if NewPdsParams.BackUpAndRestore.RunBkpAndRestrTest {

		log.InfoD("Purging all backup objects")
		err := WorkflowPDSBackup.Purge()
		if err != nil {
			log.Infof("Error while deleting backup objects - Error - [%s]", err.Error())
			allErrors = append(allErrors, err)
		}

		log.InfoD("Purging all backup config objects")
		err = WorkflowPDSBackupConfig.Purge(true)
		if err != nil {
			log.Infof("Error while deleting backup config objects - Error - [%s]", err.Error())
			allErrors = append(allErrors, err)
		}

		log.InfoD("Puring all schedules")
		err = WorkflowSchedule.Purge()
		if err != nil {
			log.Infof("Error while deleting schedules - Error - [%s]", err.Error())
			allErrors = append(allErrors, err)
		}

		if WorkflowPDSRestore.Source.Namespace.TargetCluster.ClusterUID != WorkflowPDSRestore.Destination.TargetCluster.ClusterUID {
			err := SetDestinationKubeConfig()
			allErrors = append(allErrors, err)
		} else {
			log.Infof("Source and target cluster are same. Switch is not required")
		}

		log.InfoD("Purging all restore objects")
		err = WorkflowPDSRestore.Purge()
		if err != nil {
			log.Errorf("error while purging restore - [%s]", err.Error())
			allErrors = append(allErrors, err)
		}

		if WorkflowPDSRestore.Source.Namespace.TargetCluster.ClusterUID != WorkflowPDSRestore.Destination.TargetCluster.ClusterUID {
			log.InfoD("Purging all destination namespaces")
			err = WorkflowPDSRestore.Destination.Purge(true)
			if err != nil {
				log.Errorf("error while purging destination namespaces - [%s]", err.Error())
				allErrors = append(allErrors, err)
			}
		}

		if WorkflowPDSRestore.Source.Namespace.TargetCluster.ClusterUID != WorkflowPDSRestore.Destination.TargetCluster.ClusterUID {
			err = SetSourceKubeConfig()
			allErrors = append(allErrors, err)
		} else {
			log.Infof("Source and target cluster are same. Switch is not required")
		}
	}

	log.InfoD("Purging all dataservice objects")
	err := WorkflowDataService.Purge(false)
	if err != nil {
		log.Errorf("error while purging dataservices - [%s]", err.Error())
		allErrors = append(allErrors, err)
	}

	log.InfoD("Purging all source namespace objects")
	err = WorkflowNamespace.Purge(true)
	if err != nil {
		log.Errorf("error while purging all namespaces - [%s]", err.Error())
		allErrors = append(allErrors, err)
	}

	if NewPdsParams.BackUpAndRestore.RunBkpAndRestrTest {
		log.InfoD("Purging all destination namespace objects")
		err = WorkflowNamespaceDestination.Purge(true)
		if err != nil {
			log.Errorf("error while purging all namespaces - [%s]", err.Error())
			allErrors = append(allErrors, err)
		}
	}

	return allErrors
}

// CheckforClusterSwitch checks if restore needs to be created on source or dest
func CheckforClusterSwitch() error {
	if WorkflowPDSRestore.Source.Namespace.TargetCluster.ClusterUID != WorkflowPDSRestore.Destination.TargetCluster.ClusterUID {
		err := SetDestinationKubeConfig()
		return err
	} else {
		log.Infof("Source and target cluster are same. Switch is not required")
		return nil
	}
}

// Stops px service for the given nodes
func StopPxServiceOnNodes(nodeList []*corev1.Node) error {
	// Getting all worker nodes
	workerNodes := node.GetWorkerNodes()
	for _, nodeToStop := range nodeList {
		log.InfoD("Disabling PX on Node %v ", nodeToStop.Name)
		for _, workerNode := range workerNodes {
			if workerNode.Name == nodeToStop.Name {
				err := Inst().V.StopDriver([]node.Node{workerNode}, false, nil)
				if err != nil {
					return err
				}
				break
			}
		}
	}
	return nil
}

// GetVolumeNodesOnWhichPxIsRunning fetches the lit of Volnodes on which PX is running
func GetVolumeNodesOnWhichPxIsRunning() []node.Node {
	var (
		nodesToStopPx []node.Node
		stopPxNode    []node.Node
	)
	stopPxNode = node.GetStorageNodes()
	log.InfoD("PX the node with vol running found is-  %v ", stopPxNode)
	nodesToStopPx = append(nodesToStopPx, stopPxNode[0])
	return nodesToStopPx
}

// StopPxOnReplicaVolumeNode is used to STOP PX on the given list of nodes
func StopPxOnReplicaVolumeNode(nodesToStopPx []node.Node) error {
	err := Inst().V.StopDriver(nodesToStopPx, true, nil)
	return err
}

// RebootReplicaVolumeNode reboots the node on which the replica of the volume exists
func RebootReplicaVolumeNode(namespace string, deploymentName string) error {
	_, replicaNodes, err := GetReplicaNodes(deploymentName, namespace)
	var volnodes []node.Node
	var dsNodes []string
	ss, err := k8sApps.GetStatefulSet(deploymentName, namespace)
	if err != nil {
		return err
	}
	// Get Pods of this StatefulSet
	pods, err := k8sApps.GetStatefulSetPods(ss)
	if err != nil {
		return err
	}
	// Get the node name of the pods
	for _, pod := range pods {
		log.Infof("Nodename of pod %v is :%v:", pod.Name, pod.Spec.NodeName)
		if pod.Spec.NodeName == "" || pod.Spec.NodeName == " " {
			return fmt.Errorf("Pod %v still does not have a node assigned. Retrying in 5 seconds", pod.Name)
		}
		dsNodes = append(dsNodes, pod.Spec.NodeName)
	}

	for _, replicaNode := range replicaNodes {
		nodeToReboot, err := node.GetNodeDetailsByNodeID(replicaNode)
		if err != nil {
			return err
		}
		// if the Node is same as data service node, skip reboot
		isDSNode := false
		for _, dsNode := range dsNodes {
			if dsNode == nodeToReboot.Name {
				isDSNode = true
				break
			}
		}
		if !isDSNode {
			volnodes = append(volnodes, nodeToReboot)
		}
	}

	err = RebootNodes(volnodes)
	if err != nil {
		return err
	}
	log.InfoD("successfully rebooted  node %v", volnodes)
	return nil
}

// GetReplicaNodes return the volume replicated nodes and its pool id's
func GetReplicaNodes(deploymentName, namespace string) ([]string, []string, error) {

	labelSelector := make(map[string]string)
	labelSelector["name"] = deploymentName
	pvcList, _ := k8sCore.GetPersistentVolumeClaims(namespace, labelSelector)
	pvc := pvcList.Items[0]
	appVolume := &volume.Volume{
		ID: pvc.Spec.VolumeName,
	}
	replicaSets, err := Inst().V.GetReplicaSets(appVolume)
	if err != nil {
		return nil, nil, err
	}
	replicaNodes := replicaSets[0].Nodes
	for _, node := range replicaNodes {
		log.Debugf("replica node [%s] of volume [%s]", node, appVolume.Name)
	}
	replPools := replicaSets[0].PoolUuids

	return replPools, replicaNodes, nil
}

// StartPxOnReplicaVolumeNode is used to START PX on the given list of nodes
func StartPxOnReplicaVolumeNode(nodesToStartPx []node.Node) error {
	for _, nodeName := range nodesToStartPx {
		log.InfoD("Going ahead and re-starting PX the node %v as there is an ", nodeName)
		err := Inst().V.StartDriver(nodeName)
		if err != nil {
			return err
		}
		log.InfoD("PX ReStarted successfully on node %v", nodeName)
	}
	return nil
}

// RebootNodes will reboot the nodes in the given list
func RebootNodes(nodeList []node.Node) error {
	for _, n := range nodeList {
		log.InfoD("reboot node: %s", n.Name)
		err := Inst().N.RebootNode(n, node.RebootNodeOpts{
			Force: true,
			ConnectionOpts: node.ConnectionOpts{
				Timeout:         defaultCommandTimeout,
				TimeBeforeRetry: defaultCommandRetry,
			},
		})
		if err != nil {
			return err
		}

		log.Infof("wait for node: %s to be back up", n.Name)
		err = Inst().N.TestConnection(n, node.ConnectionOpts{
			Timeout:         defaultTestConnectionTimeout,
			TimeBeforeRetry: defaultWaitRebootRetry,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// Stops px service for the given nodes
func StartPxServiceOnNodes(nodeList []*corev1.Node) error {
	// Getting all worker nodes
	workerNodes := node.GetWorkerNodes()
	for _, nodeToStart := range nodeList {
		log.InfoD("Enabling PX on Node %v ", nodeToStart.Name)
		for _, workerNode := range workerNodes {
			if workerNode.Name == nodeToStart.Name {
				err := Inst().V.StartDriver(workerNode)
				if err != nil {
					return err
				}

				err = Inst().V.WaitDriverUpOnNode(workerNode, Inst().DriverStartTimeout)
				if err != nil {
					return err
				}
				break
			}
		}
	}
	return nil
}

// Increase PVC by 1 gb
func IncreasePVCSize(namespace string, deploymentName string, sizeInGb uint64) (*volume.Volume, error) {
	log.Info("Resizing of the PVC begins")
	var vol *volume.Volume
	pvcList, _ := GetPvsAndPVCsfromDeployment(namespace, deploymentName)
	initialCapacity, err := GetVolumeCapacityInGB(namespace, deploymentName)
	log.InfoD("Initial volume storage size is : %vG", initialCapacity)
	if err != nil {
		return nil, err
	}
	for _, pvc := range pvcList.Items {
		k8sOps := k8sCore
		storageSize := pvc.Spec.Resources.Requests[corev1.ResourceStorage]
		extraAmount, _ := resource.ParseQuantity(fmt.Sprintf("%dGi", sizeInGb))
		storageSize.Add(extraAmount)
		pvc.Spec.Resources.Requests[corev1.ResourceStorage] = storageSize
		_, err := k8sOps.UpdatePersistentVolumeClaim(&pvc)
		if err != nil {
			return nil, err
		}
		sizeInt64, _ := storageSize.AsInt64()
		vol = &volume.Volume{
			Name:          pvc.Name,
			RequestedSize: uint64(sizeInt64),
		}
	}

	// wait for the resize to take effect
	time.Sleep(30 * time.Second)
	newcapacity, err := GetVolumeCapacityInGB(namespace, deploymentName)
	log.InfoD("Resized volume storage size is : %vG", newcapacity)
	if err != nil {
		return nil, err
	}

	if newcapacity > initialCapacity {
		log.InfoD("Successfully resized the pvc by 1gb")
		return vol, nil
	} else {
		return vol, err
	}
}

func GetPvsAndPVCsfromDeployment(namespace string, deploymentName string) (*corev1.PersistentVolumeClaimList, []*volume.Volume) {
	log.Infof("Get PVC List based on namespace and deployment")
	var vols []*volume.Volume
	labelSelector := make(map[string]string)
	labelSelector["name"] = deploymentName
	pvcList, _ := k8sCore.GetPersistentVolumeClaims(namespace, labelSelector)
	for _, pvc := range pvcList.Items {
		vols = append(vols, &volume.Volume{
			ID: pvc.Spec.VolumeName,
		})
	}
	return pvcList, vols
}

func GetVolumeCapacityInGB(namespace string, deploymentName string) (uint64, error) {
	var pvcCapacity uint64
	_, vols := GetPvsAndPVCsfromDeployment(namespace, deploymentName)
	for _, vol := range vols {
		appVol, err := Inst().V.InspectVolume(vol.ID)
		if err != nil {
			return 0, err
		}
		pvcCapacity = appVol.Spec.Size / units.GiB
	}
	return pvcCapacity, nil
}

func GetVolumeUsage(namespace string, deploymentName string) (float64, error) {
	var pvcUsage float64
	_, vols := GetPvsAndPVCsfromDeployment(namespace, deploymentName)
	for _, vol := range vols {
		appVol, err := Inst().V.InspectVolume(vol.ID)
		if err != nil {
			return 0, err
		}
		pvcUsage = float64(appVol.GetUsage())
	}
	log.InfoD("Amount of PVC consumed is- [%v]", pvcUsage)
	return pvcUsage, nil
}

// Check the DS related PV usage and resize in case of 90% full
func CheckStorageFullCondition(namespace, deploymentName string, thresholdPercentage float64) error {
	log.Infof("Check PVC Usage")
	f := func() (interface{}, bool, error) {
		initialCapacity, err := GetVolumeCapacityInGB(namespace, deploymentName)
		if err != nil {
			return nil, false, err
		}
		floatCapacity := float64(initialCapacity)
		consumedCapacity, err := GetVolumeUsage(namespace, deploymentName)
		if err != nil {
			return nil, false, err
		}
		threshold := thresholdPercentage * (floatCapacity / 100)
		log.InfoD("threshold value calculated is- [%v]", threshold)
		if consumedCapacity >= threshold {
			log.Infof("The PVC capacity was %v , the consumed PVC in floating point value is- %v", initialCapacity, consumedCapacity)
			return nil, false, nil
		}
		return nil, true, fmt.Errorf("threshold not achieved for the PVC, ")
	}
	_, err := task.DoRetryWithTimeout(f, 35*time.Minute, 20*time.Second)
	return err
}
