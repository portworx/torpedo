package utils

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/onsi/gomega"
	"github.com/pure-px/sched-ops/k8s/core"
	"github.com/pure-px/torpedo/drivers/backup"
	"github.com/pure-px/torpedo/drivers/node"
	"github.com/pure-px/torpedo/drivers/scheduler"
	"github.com/pure-px/torpedo/pkg/log"
	"github.com/pure-px/torpedo/tests"
	kubevirtv1 "kubevirt.io/api/core/v1"
	"math/rand"
	"net/http"
	"regexp"
	"sync"
	"time"
)

var (
	IsTorpedoInitDone bool                 // flag to check if Drivers init is done
	context           []*scheduler.Context // Application context will be maintained globally for now
)

type JobStatus struct {
	Status string      // "in-progress", "completed", "failed"
	Result interface{} // Could be error messages or success messages
}

type CompositeJobStatus struct {
	Status    string // Overall status: "in-progress", "completed", "failed"
	SubTasks  map[string]*JobStatus
	StartTime time.Time
	EndTime   time.Time
}

var compositeJobStatuses = make(map[string]*CompositeJobStatus)
var mutex = &sync.Mutex{}

// Generate a unique job ID
func generateJobID() string {
	return uuid.New().String()
}

func RunCompositeJob(c *gin.Context) {
	gomega.RegisterFailHandler(func(m string, _ ...int) { c.String(500, m) })
	jobID := generateJobID()

	mutex.Lock()
	compositeJobStatuses[jobID] = &CompositeJobStatus{
		Status:    "in-progress",
		SubTasks:  make(map[string]*JobStatus),
		StartTime: time.Now(),
	}
	mutex.Unlock()

	go func(jobID string) {
		defer func() {
			if r := recover(); r != nil {
				mutex.Lock()
				compositeJobStatuses[jobID].Status = "failed"
				mutex.Unlock()
			}
		}()

		//subTasks := []string{"ResizeVolumeAfterFull"}
		subTasks := []string{"ResizeDiskVolUpdate"}
		//subTasks := []string{"RunSetupTeardown", "AppScaleUpAndDown", "CreateLargeNumberOfVolumesTest",
		//	"EnableTrashCanDeleteVol", "ChainedLocalSnapAndValidateRestoreTest", "LocalSkinnySnap",
		//	"VerifyNoPxRestartDueToPxPodRestart", "MultiVolumeMountsForSharedV4", "StoragePoolExpandDiskAuto"
		//  "AutopilotPvcResize", "AutopilotPvcVolDetached", "AutToggleAutopilot"}
		//subTasks := []string{"VolHAIncreaseAllVolumes"}
		//subTasks := []string{"ValidateSvMotion"}

		mutex.Lock()
		for _, task := range subTasks {
			compositeJobStatuses[jobID].SubTasks[task] = &JobStatus{Status: "in-progress"}
		}
		mutex.Unlock()

		// Running sub-tasks concurrently
		var wg sync.WaitGroup
		for _, task := range subTasks {
			wg.Add(1)
			go func(taskName string) {
				defer wg.Done()
				runSubTask(jobID, taskName)
			}(task)
		}

		wg.Wait()

		mutex.Lock()
		allCompleted := true
		for _, status := range compositeJobStatuses[jobID].SubTasks {
			if status.Status != "completed" {
				allCompleted = false
				break
			}
		}
		if allCompleted {
			compositeJobStatuses[jobID].Status = "completed"
		} else {
			compositeJobStatuses[jobID].Status = "failed"
		}
		compositeJobStatuses[jobID].EndTime = time.Now()
		mutex.Unlock()
	}(jobID)

	c.JSON(http.StatusAccepted, gin.H{
		"job_id": jobID,
		"status": "in-progress",
	})
}

func runSubTask(jobID string, taskName string) {
	var err error
	var res string

	switch taskName {
	case "ResizeVolumeAfterFull":
		err, res = tests.ResizeVolumeAfterFull()
	case "ResizeDiskVolUpdate": // PASS
		err, res = tests.ResizeDiskVolUpdate()
	case "RunSetupTeardown": // PASS
		err, res = tests.RunSetupTeardownTest()
	case "AppScaleUpAndDown": // PASS
		err, res = tests.AppScaleUpAndDownTest()
	case "CreateLargeNumberOfVolumesTest": // PASS
		err, res = tests.CreateLargeNumberOfVolumesTest()
	case "EnableTrashCanDeleteVol": // PASS
		err, res = tests.EnableTrashCanDeleteVol()
	case "ChainedLocalSnapAndValidateRestoreTest": // PASS
		err, res = tests.ChainedLocalSnapAndValidateRestoreTest()
	case "LocalSkinnySnap": // PASS
		err, res = tests.LocalSkinnySnap()
	case "VerifyNoPxRestartDueToPxPodRestart": // PASS
		err, res = tests.VerifyNoPxRestartDueToPxPodRestart()
	case "AutoPilotPvcPoolExpand": // Re-Run
		err, res = tests.AutoPilotPvcPoolExpand()
	case "MultiVolumeMountsForSharedV4": // PASS
		err, res = tests.MultiVolumeMountsForSharedV4()
	case "StoragePoolExpandDiskAuto": // PASS
		err, res = tests.StoragePoolExpandDiskAuto()
	case "AutopilotPvcResize": // PASS
		err, res = tests.AutopilotPvcResize()
	case "AutopilotPvcVolDetached": // PASS
		err, res = tests.AutopilotPvcVolDetached()
	case "AutToggleAutopilot": // PASS
		err, res = tests.AutToggleAutopilot()
	case "VolHAIncreaseAllVolumes": // FAIL
		err, res = tests.VolHAIncreaseAllVolumes()
	case "ValidateSvMotion": // FAIL
		err, res = tests.ValidateSvMotion()

	default:
		err = fmt.Errorf("Unknown task: %s", taskName)
	}

	mutex.Lock()
	defer mutex.Unlock()
	if err != nil {
		compositeJobStatuses[jobID].SubTasks[taskName].Status = "failed"
		compositeJobStatuses[jobID].SubTasks[taskName].Result = err.Error()
	} else if res == "Test executed successfully" {
		compositeJobStatuses[jobID].SubTasks[taskName].Status = "completed"
		compositeJobStatuses[jobID].SubTasks[taskName].Result = res
	} else {
		compositeJobStatuses[jobID].SubTasks[taskName].Status = "failed"
		compositeJobStatuses[jobID].SubTasks[taskName].Result = "Failure"
	}
}

func GetCompositeJobStatus(c *gin.Context) {
	jobID := c.Param("job_id")
	mutex.Lock()
	jobStatus, exists := compositeJobStatuses[jobID]
	mutex.Unlock()
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Job ID not found",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"job_id":     jobID,
		"status":     jobStatus.Status,
		"sub_tasks":  jobStatus.SubTasks,
		"start_time": jobStatus.StartTime,
		"end_time":   jobStatus.EndTime,
	})
}

// This method checks if test has done InitInstance once or not. If not, we will try to do it.
func checkTorpedoInit(c *gin.Context) bool {
	if !IsTorpedoInitDone {
		InitializeDrivers(c)
		if !IsTorpedoInitDone {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Torpedo Init failed",
			})
			return false
		}
	}
	return true
}

// InitializeDrivers : This API Call will init all Torpedo Drivers. This needs to be run as ginkgo test
// as multiple ginkgo and gomega dependencies are being called in InitInstance()
func InitializeDrivers(c *gin.Context) {
	// TODO: Remove the Ginkgo dependency from functions outside the tests package.
	// Redefining tests.Step to avoid Ginkgo's "spec structure" error with `go run`, ensuring compatibility.
	tests.Step = func(text string, callback ...func()) {
		log.Infof("Step: [%s]", text)
		if len(callback) == 1 {
			callback[0]()
		} else if len(callback) > 1 {
			panic(fmt.Sprintf("Step: [%s] has more than one callback", text))
		}
	}
	tests.ParseFlags()
	tests.InitInstance()
	IsTorpedoInitDone = true
}

// GetNodes : This API will return list of all worker nodes in the Cluster
func GetNodes(c *gin.Context) {
	if !checkTorpedoInit(c) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Errorf("Error happened while doing InitInstance()"),
		})
		return
	}
	nodes := node.GetWorkerNodes()
	c.JSON(http.StatusOK, gin.H{
		"message": "Nodes are: ",
		"nodes":   nodes,
	})
}

// RebootNode - API Call to Reboot node :
// taas/rebootnode/all - Reboots all nodes in cluster
// taas/rebootnode/random - Reboots a randomly selected node
// taas/rebootnode/<nodename> - Reboots a node with specific node name
func RebootNode(c *gin.Context) {
	if !checkTorpedoInit(c) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Errorf("Error happened while doing InitInstance()"),
		})
		return
	}
	nodes := node.GetWorkerNodes()
	nodename := c.Param("nodename")
	if nodename == "all" {
		for _, n := range nodes {
			err := tests.Inst().N.RebootNode(n, node.RebootNodeOpts{
				Force: true,
				ConnectionOpts: node.ConnectionOpts{
					Timeout:         defaultCommandTimeout,
					TimeBeforeRetry: defaultCommandRetry,
				},
			})
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
		}
		c.JSON(http.StatusOK, gin.H{"message": "All Nodes successfully rebooted"})
		return
	} else if nodename == "random" {
		randomNode := nodes[rand.Intn(len(nodes))]
		err := tests.Inst().N.RebootNode(randomNode, node.RebootNodeOpts{
			Force: true,
			ConnectionOpts: node.ConnectionOpts{
				Timeout:         defaultCommandTimeout,
				TimeBeforeRetry: defaultCommandRetry,
			},
		})
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Randomly selected node %s successfully rebooted", randomNode.Name)})
		}
	} else {
		for _, n := range nodes {
			if n.Name == nodename {
				err := tests.Inst().N.RebootNode(n, node.RebootNodeOpts{
					Force: true,
					ConnectionOpts: node.ConnectionOpts{
						Timeout:         defaultCommandTimeout,
						TimeBeforeRetry: defaultCommandRetry,
					},
				})
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				} else {
					c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Node with name %s successfully rebooted", n.Name)})
				}
				return
			}
		}
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Node with name %s not found", nodename)})
	}
}

// GetStorageNodes : Returns all Storage Node objects in the cluster
func GetStorageNodes(c *gin.Context) {
	if !checkTorpedoInit(c) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Errorf("Error happened while doing InitInstance()"),
		})
		return
	}
	nodes, err := tests.GetStorageNodes()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"nodes": nodes,
		})
	}
}

// GetStorageLessNodes : Returns all Storage less node objects in the cluster
func GetStorageLessNodes(c *gin.Context) {
	if !checkTorpedoInit(c) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Errorf("Error happened while doing InitInstance()"),
		})
		return
	}
	nodes := node.GetStorageLessNodes()
	c.JSON(http.StatusOK, gin.H{"nodes": nodes})
}

// CollectSupport : This API collects the support bundle
func CollectSupport(c *gin.Context) {
	if !checkTorpedoInit(c) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Errorf("Error happened while doing InitInstance()"),
		})
		return
	}
	tests.CollectSupport()
	c.JSON(http.StatusOK, gin.H{"message": "Collection of support bundle done from Torpedo End"})
}

// ScheduleAppsAndValidate : This API schedules multiple applications on the cluster and validates them
// context is created as a global context to be accessed later in further tests
func ScheduleAppsAndValidate(c *gin.Context) {

	var requestBody struct {
		NamespaceSuffix string   `json:"nsSuffix"`
		AppList         []string `json:"appList"`
	}
	var errors []error
	errStrings := make([]string, 0)
	errChan := make(chan error, 100)
	if !checkTorpedoInit(c) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Errorf("error during InitInstance"),
		})
		return
	}
	if err := c.BindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tests.Inst().AppList = requestBody.AppList
	options := tests.CreateScheduleOptions(requestBody.NamespaceSuffix)
	context, err := tests.Inst().S.Schedule(requestBody.NamespaceSuffix, options)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	for _, ctx := range context {
		tests.ValidateContext(ctx, &errChan)
	}
	for err = range errChan {
		errors = append(errors, err)
	}
	for _, err = range errors {
		if err != nil {
			errStrings = append(errStrings, err.Error())
		}
	}
	if len(errStrings) > 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errStrings})
		return
	}
	namespacesList := make([]string, 0)
	for _, ctx := range context {
		namespace := tests.GetAppNamespace(ctx, requestBody.NamespaceSuffix)
		namespacesList = append(namespacesList, namespace)
	}
	c.JSON(http.StatusOK, gin.H{
		"message":   "App is created and validated successfully",
		"namespace": namespacesList,
	})
}

// GetPxVersion This function returns the current Px Version in the Target Cluster
func GetPxVersion(c *gin.Context) {
	if !checkTorpedoInit(c) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Errorf("Error happened while doing InitInstance()"),
		})
		return
	}
	version, err := tests.Inst().V.GetDriverVersion()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"version": version,
		})
	}
}

// IsPxInstalled This funtion returns true if Px is Installed
func IsPxInstalled(c *gin.Context) {
	if !checkTorpedoInit(c) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Errorf("Error happened while doing InitInstance()"),
		})
		return
	}
	nodes := node.GetWorkerNodes()
	var pxinstalled bool
	var err error
	for _, n := range nodes {
		pxinstalled, err = tests.Inst().V.IsDriverInstalled(n)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err,
			})
			return
		}
		if !pxinstalled {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Px is not Installed",
				"node":  n.Name,
			})
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "Px installed on all nodes successfully",
	})
}

// GetPxctlStatusOutput This function is used to return all elements in pxctl status output
func GetPxctlStatusOutput(c *gin.Context) {
	if !checkTorpedoInit(c) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Errorf("Error happened while doing InitInstance()"),
		})
		return
	}
	cmd := "status"
	nodes := node.GetWorkerNodes()
	out, err := tests.Inst().V.GetPxctlCmdOutput(nodes[0], cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err,
		})
	} else {
		var status PxctlStatus
		// Extract Status
		re := regexp.MustCompile(`Status: (.+)`)
		match := re.FindStringSubmatch(out)
		if len(match) > 0 {
			status.Status = match[1]
		}
		// Extract Node ID
		re = regexp.MustCompile(`Node ID: (.+)`)
		match = re.FindStringSubmatch(out)
		if len(match) > 0 {
			status.NodeID = match[1]
		}
		// Extract IP
		re = regexp.MustCompile(`IP: (.+)`)
		match = re.FindStringSubmatch(out)
		if len(match) > 0 {
			status.IP = match[1]
		}
		// Extract Cluster ID
		re = regexp.MustCompile(`Cluster ID: (.+)`)
		match = re.FindStringSubmatch(out)
		if len(match) > 0 {
			status.ClusterID = match[1]
		}
		// Extract Cluster UUID
		re = regexp.MustCompile(`Cluster UUID: (.+)`)
		match = re.FindStringSubmatch(out)
		if len(match) > 0 {
			status.ClusterUUID = match[1]
		}
		// Extract Total Used
		re = regexp.MustCompile(`Total Used\s+: (.+)`)
		match = re.FindStringSubmatch(out)
		if len(match) > 0 {
			status.TotalUsed = match[1]
		}
		// Extract Total Capacity
		re = regexp.MustCompile(`Total Capacity\s+: (.+)`)
		match = re.FindStringSubmatch(out)
		if len(match) > 0 {
			status.TotalCapacity = match[1]
		}
		c.JSON(http.StatusOK, gin.H{
			"output": status,
		})
	}
}

// GetVMsInNamespaces gets the list of Virtual Machines in the given namespaces
func GetVMsInNamespaces(c *gin.Context) {
	var requestBody struct {
		Namespaces []string `json:"namespaces"`
	}
	var vms []kubevirtv1.VirtualMachine
	type VM struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		Status    string `json:"status"`
	}
	var vmResponse []VM

	if !checkTorpedoInit(c) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Errorf("error in InitInstance()"),
		})
		return
	}

	if err := c.BindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(requestBody.Namespaces) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "namespaces cannot be empty"})
		return
	}

	for _, ns := range requestBody.Namespaces {
		vmList, err := tests.GetAllVMsInNamespace(ns)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		vms = append(vms, vmList...)
	}

	for _, v := range vms {
		vmResponse = append(vmResponse, VM{
			Name:      v.Name,
			Namespace: v.Namespace,
			Status:    string(v.Status.PrintableStatus),
		})
	}

	// Return the list of VMs
	c.JSON(http.StatusOK, vmResponse)
}

// GetVMsWithNamespaceLabels gets the list of Virtual Machines in the namespaces with the given labels
func GetVMsWithNamespaceLabels(c *gin.Context) {
	var requestBody struct {
		NamespaceLabels map[string]string `json:"namespaceLabels"`
	}
	var vms []kubevirtv1.VirtualMachine
	type VM struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		Status    string `json:"status"`
	}
	var vmResponse []VM
	if !checkTorpedoInit(c) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Errorf("error in InitInstance()"),
		})
		return
	}
	if err := c.BindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(requestBody.NamespaceLabels) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "namespace labels cannot be empty"})
		return
	}
	vms, err := tests.GetAllVMsInNamespacesWithLabel(requestBody.NamespaceLabels)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	for _, v := range vms {
		vmResponse = append(vmResponse, VM{
			Name:      v.Name,
			Namespace: v.Namespace,
			Status:    string(v.Status.PrintableStatus),
		})
	}

	// Return the list of VMs
	c.JSON(http.StatusOK, vmResponse)

}

// AddNSLabel adds the label to the namespaces with the given label
func AddNSLabel(c *gin.Context) {
	log.Infof("Adding label to NS ")
	type LabelUpdateResponse struct {
		Success map[string]string `json:"success"`
		Failed  map[string]string `json:"failed"`
	}

	var NamespaceLabelRequest struct {
		Namespaces []string          `json:"namespaces" binding:"required"`
		Label      map[string]string `json:"ns_label" binding:"required"`
	}
	success := make(map[string]string)
	failed := make(map[string]string)
	if !checkTorpedoInit(c) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Errorf("error in InitInstance()"),
		})
		return
	}
	if err := c.BindJSON(&NamespaceLabelRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(NamespaceLabelRequest.Label) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "namespace labels cannot be empty"})
		return
	}
	log.Infof("NamespaceLabelRequest", NamespaceLabelRequest)
	for _, namespace := range NamespaceLabelRequest.Namespaces {
		err := tests.Inst().S.AddNamespaceLabel(namespace, NamespaceLabelRequest.Label)
		if err != nil {
			failed[namespace] = err.Error()
		} else {
			success[namespace] = "Label added successfully"
		}
	}

	response := LabelUpdateResponse{
		Success: success,
		Failed:  failed,
	}
	c.JSON(http.StatusOK, gin.H{
		"message": response,
	})

}

// UpgradeStork upgrades the stork to given version
func UpgradeStork(c *gin.Context) {
	var requestBody struct {
		Version string `json:"version"`
	}
	if err := c.BindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Infof("%s", requestBody.Version)
	if !checkTorpedoInit(c) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Errorf("error in InitInstance()"),
		})
		return
	}
	err := tests.UpgradeStorkVersion(requestBody.Version)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "Stork upgraded successfully",
	})
}

// DeletePod deletes the pods with given label
func DeletePod(c *gin.Context) {
	log.Infof("Deleting pods with given label")
	var deletePodRequest struct {
		Namespace   string            `json:"namespace"`
		Label       map[string]string `json:"label"`
		PodList     []string          `json:"podList"`
		IgnoreLabel bool              `json:"ignoreLabel"`
	}
	if !checkTorpedoInit(c) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error in init": fmt.Errorf("error in InitInstance()"),
		})
		return
	}

	if err := c.BindJSON(&deletePodRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error for request": err.Error()})
		return
	}

	if len(deletePodRequest.Namespace) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "namespace cannot be empty", "ns": deletePodRequest.Namespace})
		return
	}
	if len(deletePodRequest.Label) == 0 && len(deletePodRequest.PodList) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "both pod label and pod list cannot be empty", "label": deletePodRequest.Label})
		return
	}
	if len(deletePodRequest.Label) != 0 {
		err := tests.DeletePodWithWithoutLabelInNamespace(deletePodRequest.Namespace, deletePodRequest.Label, deletePodRequest.IgnoreLabel)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else {
		for _, pod := range deletePodRequest.PodList {
			err := core.Instance().DeletePod(pod, deletePodRequest.Namespace, false)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error in deleting pod": err.Error(), "pod name": pod})
				return
			}
		}
	}
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Pods list %v with label %s in namespace %s deleted successfully", deletePodRequest.PodList, deletePodRequest.Label, deletePodRequest.Namespace)})
}

// GetPxBackupNamespace returns the namespace in which px-backup is deployed
func GetPxBackupNamespace(c *gin.Context) {
	log.Infof("Getting px-backup namespace")
	var namespace string
	type PxBackupNamespacePodResponse struct {
		Namespace string `json:"namespace"`
	}
	if !checkTorpedoInit(c) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Errorf("error in InitInstance()"),
		})
		return
	}
	ns, err := backup.GetPxBackupNamespace()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	namespace = ns
	response := PxBackupNamespacePodResponse{
		Namespace: namespace,
	}
	// Return the namespace in which px-backup is deployed
	c.JSON(http.StatusOK, response)
}

func RunSetupTeardownTestGin(c *gin.Context) {
	gomega.RegisterFailHandler(func(m string, _ ...int) { c.String(500, m) })
	if !checkTorpedoInit(c) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Errorf("Error happened while doing InitInstance()"),
		})
		return
	}
	err, res := tests.RunSetupTeardownTest()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if res == "Test executed successfully" {
		c.JSON(http.StatusOK, res)
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failure"})
		return
	}
}

func CreateLargeNumberOfVolumesTestGin(c *gin.Context) {
	if !checkTorpedoInit(c) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Errorf("Error happened while doing InitInstance()"),
		})
		return
	}
	err, res := tests.CreateLargeNumberOfVolumesTest()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if res == "Test executed successfully" {
		c.JSON(http.StatusOK, res)
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failure"})
		return
	}
}

func ChainedLocalSnapAndValidateRestoreTestGin(c *gin.Context) {
	if !checkTorpedoInit(c) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Errorf("Error happened while doing InitInstance()"),
		})
		return
	}
	err, res := tests.ChainedLocalSnapAndValidateRestoreTest()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if res == "Test executed successfully" {
		c.JSON(http.StatusOK, res)
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failure"})
		return
	}
}

func AppScaleUpAndDownTestGin(c *gin.Context) {
	gomega.RegisterFailHandler(func(m string, _ ...int) { c.String(500, m) })
	if !checkTorpedoInit(c) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Errorf("Error happened while doing InitInstance()"),
		})
		return
	}
	err, res := tests.AppScaleUpAndDownTest()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if res == "Test executed successfully" {
		c.JSON(http.StatusOK, res)
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failure"})
		return
	}
}

func AutoPilotPvcPoolExpandGin(c *gin.Context) {
	gomega.RegisterFailHandler(func(m string, _ ...int) { c.String(500, m) })
	if !checkTorpedoInit(c) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Errorf("Error happened while doing InitInstance()"),
		})
		return
	}
	err, res := tests.AutoPilotPvcPoolExpand()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if res == "Test executed successfully" {
		c.JSON(http.StatusOK, res)
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failure"})
		return
	}
}

func LocalSkinnySnapGin(c *gin.Context) {
	if !checkTorpedoInit(c) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Errorf("Error happened while doing InitInstance()"),
		})
		return
	}
	err, res := tests.LocalSkinnySnap()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if res == "Test executed successfully" {
		c.JSON(http.StatusOK, res)
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failure"})
		return
	}
}

func MultiVolumeMountsForSharedV4Gin(c *gin.Context) {
	if !checkTorpedoInit(c) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Errorf("Error happened while doing InitInstance()"),
		})
		return
	}
	err, res := tests.MultiVolumeMountsForSharedV4()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if res == "Test executed successfully" {
		c.JSON(http.StatusOK, res)
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failure"})
		return
	}
}

func StoragePoolExpandDiskAutoGin(c *gin.Context) {
	if !checkTorpedoInit(c) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Errorf("Error happened while doing InitInstance()"),
		})
		return
	}
	err, res := tests.StoragePoolExpandDiskAuto()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if res == "Test executed successfully" {
		c.JSON(http.StatusOK, res)
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failure"})
		return
	}
}

func EnableTrashCanDeleteVolGin(c *gin.Context) {
	if !checkTorpedoInit(c) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Errorf("Error happened while doing InitInstance()"),
		})
		return
	}
	err, res := tests.EnableTrashCanDeleteVol()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if res == "Test executed successfully" {
		c.JSON(http.StatusOK, res)
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failure"})
		return
	}
}

func AutopilotPvcResizeGin(c *gin.Context) {
	if !checkTorpedoInit(c) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Errorf("Error happened while doing InitInstance()"),
		})
		return
	}
	err, res := tests.AutopilotPvcResize()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if res == "Test executed successfully" {
		c.JSON(http.StatusOK, res)
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failure"})
		return
	}
}

func AutopilotPvcVolDetachedGin(c *gin.Context) {
	if !checkTorpedoInit(c) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Errorf("Error happened while doing InitInstance()"),
		})
		return
	}
	err, res := tests.AutopilotPvcVolDetached()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if res == "Test executed successfully" {
		c.JSON(http.StatusOK, res)
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failure"})
		return
	}
}

func AutToggleAutopilotGin(c *gin.Context) {
	if !checkTorpedoInit(c) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Errorf("Error happened while doing InitInstance()"),
		})
		return
	}
	err, res := tests.AutToggleAutopilot()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if res == "Test executed successfully" {
		c.JSON(http.StatusOK, res)
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failure"})
		return
	}
}

func VolHAIncreaseAllVolumesGin(c *gin.Context) {
	if !checkTorpedoInit(c) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Errorf("Error happened while doing InitInstance()"),
		})
		return
	}
	err, res := tests.VolHAIncreaseAllVolumes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if res == "Test executed successfully" {
		c.JSON(http.StatusOK, res)
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failure"})
		return
	}
}

func ValidateSvMotionGin(c *gin.Context) {
	if !checkTorpedoInit(c) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Errorf("Error happened while doing InitInstance()"),
		})
		return
	}
	err, res := tests.ValidateSvMotion()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if res == "Test executed successfully" {
		c.JSON(http.StatusOK, res)
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failure"})
		return
	}
}

func LoggingTestGin(c *gin.Context) {
	gomega.RegisterFailHandler(func(m string, _ ...int) { c.String(500, m) })
	if !checkTorpedoInit(c) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Errorf("Error happened while doing InitInstance()"),
		})
		return
	}
	err, res := tests.LoggingTest()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if res == "Test executed successfully" {
		c.JSON(http.StatusOK, res)
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failure"})
		return
	}
}

func VerifyNoPxRestartDueToPxPodRestartGin(c *gin.Context) {
	gomega.RegisterFailHandler(func(m string, _ ...int) { c.String(500, m) })
	if !checkTorpedoInit(c) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Errorf("Error happened while doing InitInstance()"),
		})
		return
	}
	err, res := tests.VerifyNoPxRestartDueToPxPodRestart()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if res == "Test executed successfully" {
		c.JSON(http.StatusOK, res)
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failure"})
		return
	}
}

func VolumeDriverDownGin(c *gin.Context) {
	if !checkTorpedoInit(c) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Errorf("Error happened while doing InitInstance()"),
		})
		return
	}
	err, res := tests.VolumeDriverDown()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if res == "Test executed successfully" {
		c.JSON(http.StatusOK, res)
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failure"})
		return
	}
}

func VolumeIOThrottleGin(c *gin.Context) {
	if !checkTorpedoInit(c) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Errorf("Error happened while doing InitInstance()"),
		})
		return
	}
	err, res := tests.VolumeIOThrottle()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if res == "Test executed successfully" {
		c.JSON(http.StatusOK, res)
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failure"})
		return
	}
}

func StickyVolumeTestGin(c *gin.Context) {
	if !checkTorpedoInit(c) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Errorf("Error happened while doing InitInstance()"),
		})
		return
	}
	err, res := tests.StickyVolumeTest()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if res == "Test executed successfully" {
		c.JSON(http.StatusOK, res)
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failure"})
		return
	}
}

func ResizeVolumeAfterFullGin(c *gin.Context) {
	if !checkTorpedoInit(c) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Errorf("Error happened while doing InitInstance()"),
		})
		return
	}
	err, res := tests.ResizeVolumeAfterFull()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if res == "Test executed successfully" {
		c.JSON(http.StatusOK, res)
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failure"})
		return
	}
}

func ResizeDiskVolUpdateGin(c *gin.Context) {
	if !checkTorpedoInit(c) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Errorf("Error happened while doing InitInstance()"),
		})
		return
	}
	err, res := tests.ResizeDiskVolUpdate()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if res == "Test executed successfully" {
		c.JSON(http.StatusOK, res)
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failure"})
		return
	}
}
