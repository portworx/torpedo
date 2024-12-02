package tests

import (
	ctxt "context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ghodss/yaml"
	snapv1 "github.com/kubernetes-incubator/external-storage/snapshot/pkg/apis/crd/v1"
	apapi "github.com/libopenstorage/autopilot-api/pkg/apis/autopilot/v1alpha1"
	"github.com/libopenstorage/openstorage/api"
	storkv1 "github.com/pure-px/stork/pkg/apis/stork/v1alpha1"
	"github.com/pborman/uuid"
	"github.com/portworx/sched-ops/k8s/batch"
	"github.com/portworx/sched-ops/k8s/core"
	storkops "github.com/pure-px/stork/pkg/crud/stork"
	"github.com/portworx/sched-ops/task"
	"github.com/pure-px/torpedo/drivers/node"
	"github.com/pure-px/torpedo/drivers/scheduler"
	"github.com/pure-px/torpedo/drivers/volume"
	"github.com/pure-px/torpedo/drivers/volume/portworx"
	"github.com/pure-px/torpedo/pkg/aututils"
	"github.com/pure-px/torpedo/pkg/log"
	"github.com/pure-px/torpedo/pkg/osutils"
	"github.com/pure-px/torpedo/pkg/pureutils"
	"github.com/pure-px/torpedo/pkg/units"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func RunSetupTeardownTest() (error, string) {
	log.SetTestName("SetupTearDown")
	var contexts []*scheduler.Context
	contexts = make([]*scheduler.Context, 0)
	Inst().AppList = []string{"fio"}
	appNamespace := fmt.Sprintf("%v-%v", "setupteardown", time.Now().Unix())
	for i := 0; i < Inst().GlobalScaleFactor; i++ {
		contexts = append(contexts, ScheduleApplicationsOnNamespace(appNamespace, fmt.Sprintf("setupteardown-%d", i))...)
	}
	ValidateApplications(contexts)
	opts := make(map[string]bool)
	opts[scheduler.OptionsWaitForResourceLeakCleanup] = true
	opts["SkipClusterScopedObjects"] = true
	for _, ctx := range contexts {
		TearDownContext(ctx, opts)
	}
	return nil, "Test executed successfully"
}

func CreateLargeNumberOfVolumesTest() (error, string) {
	log.SetTestName("CreateLargeNumberOfVolumesTest")
	var contexts []*scheduler.Context
	var totalVolumesToCreate = 25
	var maxVolumesToAttach = 15
	var volumesCurrentlyAttached = 0
	var newVolumeIDs []string
	var attachedVolumes []string
	terminate := false
	Inst().AppList = []string{"nginx"}
	for i := 0; i < Inst().GlobalScaleFactor; i++ {
		contexts = append(contexts, ScheduleApplications(fmt.Sprintf("createmaxvolume-%d", i))...)
	}
	deleteVolumes := func() {
		terminate = true
		for _, each := range newVolumeIDs {
			if IsVolumeExits(each) {
				log.InfoD(fmt.Sprintf("delete volume [%v]", each))
				err := Inst().V.DetachVolume(each)
				if err != nil {
					log.Errorf("Failed to detach volume [%v]", each)
					return
				}
				time.Sleep(500 * time.Millisecond)
				err = Inst().V.DeleteVolume(each)
				if err != nil {
					log.Errorf("Delete volume with ID [%v] failed", each)
					return
				}
			}
		}
	}

	appsValidateAndDestroy := func(contexts []*scheduler.Context) {
		opts := make(map[string]bool)
		opts[scheduler.OptionsWaitForResourceLeakCleanup] = true

		Step("validate apps", func() {
			log.InfoD("Validating apps")
			for _, ctx := range contexts {
				ctx.ReadinessTimeout = 15 * time.Minute
				ValidateContext(ctx)
			}
		})

		Step("destroy apps", func() {
			log.InfoD("Destroying apps")
			for _, ctx := range contexts {
				TearDownContext(ctx, opts)
			}
		})
	}
	ValidateApplications(contexts)
	defer appsValidateAndDestroy(contexts)
	defer deleteVolumes()

	// Get list of all volumes present in the cluster
	log.InfoD("Listing all the volumes present in the cluster")
	allVolumeIds, err := Inst().V.ListAllVolumes()
	if err != nil {
		return err, "failed to list all the volume"
	}
	log.Info(fmt.Sprintf("total number of volumes present in the cluster [%v]", len(allVolumeIds)))

	if len(allVolumeIds) >= totalVolumesToCreate {
		return fmt.Errorf("exceeded total volume count limit.. exiting [%d]", len(allVolumeIds)),
			"Total volume count exceeded "
	}

	// Get Total number of already attached volumes
	for _, each := range allVolumeIds {
		vol, err := Inst().V.InspectVolume(each)
		if err != nil {
			return err, "inspect returned error ?"
		}
		if vol.State.String() == "VOLUME_STATE_ATTACHED" {
			volumesCurrentlyAttached = volumesCurrentlyAttached + 1
		}
	}

	// Run inspect continuously in the background
	log.InfoD("start attach volume in the backend while more than 100 volumes got created")
	go func(volumeIds []string) {
		//defer GinkgoRecover()
		attachedCount := 0
		for {
			if terminate == true {
				break
			}
			if len(newVolumeIDs) > 100 {
				for _, each := range newVolumeIDs {
					if attachedCount < (maxVolumesToAttach - volumesCurrentlyAttached) {
						_, err := Inst().V.AttachVolume(each)
						if err != nil {
							log.Infof("attaching volume failed")
							return
						}
						attachedCount += 1
						attachedVolumes = append(attachedVolumes, each)
						time.Sleep(2 * time.Second)
					}
				}
			}
		}
	}(newVolumeIDs)

	volumesToBeCreated := totalVolumesToCreate - len(allVolumeIds)
	log.InfoD(fmt.Sprintf("Total number of new volumes to be created in the cluster [%v]", volumesToBeCreated))

	// Create volumes in the cluster till it reaches maximum count
	for initVol := 0; initVol < volumesToBeCreated; initVol++ {
		id := uuid.New()
		volName := fmt.Sprintf("volume_%s", id[:8])
		log.InfoD(fmt.Sprintf("Volume [%v] will be created with name [%v]", initVol, volName))

		// get size of the volume from size 1GiB till 100GiB
		minSize := 1
		maxSize := 100
		randSize := uint64(rand.Intn(maxSize-minSize) + minSize)

		// Pick HA Update from 1 to 3
		haUpdate := int64(rand.Intn(3-1) + 1)

		volId, err := Inst().V.CreateVolume(volName, randSize, haUpdate)
		if err != nil {
			terminate = true
			return err, fmt.Sprintf("Failed to create volume with vol Name [%v]", volName)
		}
		log.InfoD("Volume Created with ID [%v]", volId)
		newVolumeIDs = append(newVolumeIDs, volId)
	}

	// Validate Volume Attached status
	for _, eachVol := range attachedVolumes {
		vol, err := Inst().V.InspectVolume(eachVol)
		if err != nil {
			return err, fmt.Sprintf("Inspect volume failed on volume [%v]", eachVol)
		}
		if vol.State.String() != "VOLUME_STATE_ATTACHED" {
			return fmt.Errorf(" volume [%v] state is [%v]", eachVol, vol.State.String()), "Volume not attached"
		}
	}
	return nil, "Test executed successfully"
}

func ChainedLocalSnapAndValidateRestoreTest() (error, string) {
	log.SetTestName("ChainedLocalSnapAndValidateRestore")
	var contexts []*scheduler.Context
	const defaultCommandTimeout time.Duration = 1 * time.Minute
	const defaultReadynessTimeout time.Duration = 2 * time.Minute
	contexts = make([]*scheduler.Context, 0)
	retain := 8
	interval := 3
	var writeAppNS string
	contexts = make([]*scheduler.Context, 0)
	policyName := "localintervalpolicydv"
	chainSize := 2
	latestApp := make([]*scheduler.Context, 0)

	log.Infof("create schedule policy %s for local snapshots", policyName)
	schedPolicy, err := storkops.Instance().GetSchedulePolicy(policyName)
	if err != nil {
		log.InfoD("Creating a interval schedule policy %v with interval %v minutes", policyName, interval)
		schedPolicy = &storkv1.SchedulePolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name: policyName,
			},
			Policy: storkv1.SchedulePolicyItem{
				Interval: &storkv1.IntervalPolicy{
					Retain:          storkv1.Retain(retain),
					IntervalMinutes: interval,
				},
			}}
		_, err = storkops.Instance().CreateSchedulePolicy(schedPolicy)
		if err != nil {
			return err, fmt.Sprintf("error creating a SchedulePolicy [%s]", policyName)
		}
	}

	log.Infof("Create an Application that does writes for snapshots to happen")
	appList := Inst().AppList
	defer func() {
		Inst().AppList = appList
	}()
	Inst().AppList = []string{"data-validation-write-job"}
	for i := 0; i < Inst().GlobalScaleFactor; i++ {
		contexts = append(contexts, ScheduleApplications(fmt.Sprintf("localsnaprestore-%d", i))...)
	}

	for _, ctx := range contexts {
		writeAppNS = ctx.App.Key + "-" + ctx.UID
		err := batch.Instance().ValidateJob("write-and-checksum-job", writeAppNS, 20*time.Minute)
		if err != nil {
			log.FailOnError(err, "Write job has failed")
		}
		log.Infof("Write job is completed....")
		latestApp = append(latestApp, ctx)
	}
	log.Infof("Waiting for 5 minutes to let one more snapshot happen")
	time.Sleep(5 * time.Minute)

	for i := 0; i < chainSize; i++ {
		volSnapMap := make(map[string]map[*volume.Volume]*storkv1.ScheduledVolumeSnapshotStatus)
		for _, ctx := range latestApp {
			var appVolumes []*volume.Volume
			var err error
			appNamespace := writeAppNS
			log.Infof("Namespace: %v", appNamespace)
			log.Infof("Getting app volumes for volume %s", ctx.App.Key)
			appVolumes, err = Inst().S.GetVolumes(ctx)
			if err != nil {
				return err, fmt.Sprintf("error getting volumes for [%s]", ctx.App.Key)
			}

			if len(appVolumes) == 0 {
				return fmt.Errorf("no volumes found for [%s]", ctx.App.Key), fmt.Sprintf("no volumes found for [%s]", ctx.App.Key)
			}
			log.Infof("Got volume count : %v", len(appVolumes))

			snapMap := make(map[*volume.Volume]*storkv1.ScheduledVolumeSnapshotStatus)
			for _, v := range appVolumes {
				snapshotScheduleName := v.Name + "-interval-schedule"
				log.InfoD("snapshotScheduleName : %v for volume: %s", snapshotScheduleName, v.Name)

				var latestSnapshot *storkv1.ScheduledVolumeSnapshotStatus
				_, err = task.DoRetryWithTimeout(func() (interface{}, bool, error) {
					resp, err := storkops.Instance().GetSnapshotSchedule(snapshotScheduleName, appNamespace)
					if err != nil {
						return nil, false, fmt.Errorf("error getting snapshot schedule for %s, volume:%s in namespace %s", snapshotScheduleName, v.Name, v.Namespace)
					}
					if len(resp.Status.Items) == 0 {
						return nil, true, fmt.Errorf("waiting for new snapshot schedules for %s, volume:%s in namespace %s", snapshotScheduleName, v.Name, v.Namespace)
					}

					for _, item := range resp.Status.Items {
						for _, status := range item {
							if latestSnapshot == nil || status.CreationTimestamp.After(latestSnapshot.CreationTimestamp.Time) {
								latestSnapshot = status
							}
						}
					}

					if latestSnapshot != nil && latestSnapshot.Status == snapv1.VolumeSnapshotConditionReady {
						return latestSnapshot, false, nil
					}
					return nil, true, fmt.Errorf("latest snapshot %s is not ready yet", latestSnapshot.Name)
				}, time.Duration(5*15)*defaultCommandTimeout, defaultReadynessTimeout)
				if err != nil {
					return err, fmt.Sprintf("Failed to get a ready snapshot")
				}

				log.Infof("Latest snapshot %s in ready state for volume: %s", latestSnapshot.Name, v.Name)
				snapMap[v] = latestSnapshot
			}
			volSnapMap[appNamespace] = snapMap
		}

		log.Infof("Verify local snap restore to new PVC")
		storageClassName := "data-validation-sc-ss"
		randomSuffix := rand.Intn(10000)
		pvcName := fmt.Sprintf("data-validation-verify-pvc-%d", randomSuffix)
		for _, volSnap := range volSnapMap {
			for vol, snap := range volSnap {
				pvc := corev1.PersistentVolumeClaim{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
						Kind:       "PersistentVolumeClaim",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      pvcName,
						Namespace: vol.Namespace,
						Annotations: map[string]string{
							"snapshot.alpha.kubernetes.io/snapshot": snap.Name,
						},
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						AccessModes:      []corev1.PersistentVolumeAccessMode{"ReadWriteMany"},
						StorageClassName: &storageClassName,
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceStorage: resource.MustParse("120Gi"),
							},
						},
					},
				}
				currentDir, err := os.Getwd()
				appFilePath := filepath.Join(currentDir, "..", "drivers", "scheduler", "k8s", "specs", "data-validation-verify-job", "dd-app.yaml")
				pvcFilePath := filepath.Join(currentDir, "..", "drivers", "scheduler", "k8s", "specs", "data-validation-verify-job", "pvc.yaml")
				log.Infof("Current path is: %v", currentDir)
				log.Infof("Abs App Path is: %v", appFilePath)
				err = dumpPVCToYAML(&pvc, pvcFilePath)
				if err != nil {
					return err, fmt.Sprintf("Failed to dump PVC to yaml")
				}

				err = modifyPVCName(appFilePath, pvcName)
				if err != nil {
					return err, fmt.Sprintf("Failed to modify PVC name in the app yaml")
				}

				err = Inst().S.RescanSpecs(Inst().SpecDir, Inst().V.String())
				if err != nil {
					return err, fmt.Sprintf("Failed to rescan specs")
				}

				Inst().AppList = []string{"data-validation-verify-job"}
				newApp := ScheduleApplicationsOnNamespace(writeAppNS, "restorelocalsnap")
				contexts = append(contexts, newApp...)
				latestApp = newApp
				err = batch.Instance().ValidateJob(pvcName, writeAppNS, 20*time.Minute)
				if err != nil {
					return err, fmt.Sprintf("Verify job has failed")
				}
				log.Infof("Verify Job is successfully completed....")
				log.Infof("Waiting for 5 minutes to let one more snapshot happen")
				time.Sleep(5 * time.Minute)
			}
		}
	}

	log.Infof("Destroying apps")
	DestroyApps(contexts, make(map[string]bool))
	return nil, "Test executed successfully"

}

func dumpPVCToYAML(pvc *corev1.PersistentVolumeClaim, filePath string) error {
	yamlData, err := yaml.Marshal(pvc)
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, yamlData, 0644)
}

func modifyPVCName(filePath, newPVCName string) error {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	var config map[string]interface{}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return err
	}

	if metadata, ok := config["metadata"].(map[string]interface{}); ok {
		metadata["name"] = newPVCName
	}

	spec := config["spec"].(map[string]interface{})
	template := spec["template"].(map[string]interface{})
	templateSpec := template["spec"].(map[string]interface{})
	volumes := templateSpec["volumes"].([]interface{})
	for _, v := range volumes {
		volume := v.(map[string]interface{})
		if volume["name"] == "data-volume" {
			pvc := volume["persistentVolumeClaim"].(map[string]interface{})
			pvc["claimName"] = newPVCName
		}
	}

	updatedData, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	err = os.WriteFile(filePath, updatedData, 0644)
	if err != nil {
		return err
	}

	return nil
}

// To be run on Pure FA Backend only
func EnableTrashCanDeleteVol() (error, string) {
	var contexts []*scheduler.Context
	log.Infof("Enable TrashCan and Delete Volume")
	cluster, err := Inst().V.InspectCurrentCluster()
	if err != nil {
		return err, fmt.Sprintf("failed to inspect current cluster")
	}
	log.Infof("Current cluster [%s] UID: [%s]", cluster.Cluster.Name, cluster.Cluster.Id)

	clusterUIDPrefix := strings.Split(cluster.Cluster.Id, "-")[0]

	getPureVolName := func(volName string) string {
		return "px_" + clusterUIDPrefix + "-" + volName
	}

	GetVolumeNameFromPvc := func(namespace string) []string {
		pvclist := make([]string, 0)
		allPvcList, err := core.Instance().GetPersistentVolumeClaims(namespace, nil)
		if err != nil {
			log.Infof("Error is: %v", err)
			return nil
		}

		for _, p := range allPvcList.Items {
			pvclist = append(pvclist, p.Spec.VolumeName)
		}
		return pvclist
	}
	defer func() {
		log.Infof("Disable trashcan")
		//Disable trashcan in the node
		err := Inst().V.SetClusterOptsWithConfirmation(node.GetStorageNodes()[0], map[string]string{
			"--volume-expiration-minutes": "0",
		})
		if err != nil {
			log.Infof("error while disabling trashcan")
		} else {
			log.InfoD("Trashcan is successfully disabled")
		}
	}()

	log.Infof("Enabling Trashcan")
	err = Inst().V.SetClusterOptsWithConfirmation(node.GetStorageNodes()[0], map[string]string{
		"--volume-expiration-minutes": "600",
	})
	if err != nil {
		return err, fmt.Sprintf("error while enabling trashcan")
	}
	log.InfoD("Trashcan is successfully enabled")

	log.Infof("Scheduling Apps")
	Inst().AppList = []string{"sysbench"}
	for i := 0; i < Inst().GlobalScaleFactor; i++ {
		ns := fmt.Sprintf("enabletrashcan-%d", i)
		contexts = append(contexts, ScheduleApplications(ns)...)
	}
	ValidateApplications(contexts)
	// before destroying the apps, get the volume list
	var volumeList []string
	for _, ctx := range contexts {
		vol := GetVolumeNameFromPvc(ctx.App.NameSpace)
		volumeList = append(volumeList, vol...)
	}
	log.Infof("Destroy apps and let it's volumes should not be placed in trashcan")
	DestroyApps(contexts, nil)

	log.Infof("Check if the volumes are placed in trashcan and not deleted from the FA")
	//Get pure secrets
	volDriverNamespace, err := Inst().V.GetVolumeDriverNamespace()
	if err != nil {
		return err, fmt.Sprintf("failed to get volume driver [%s] namespace", Inst().V.String())
	}

	pxPureSecret, err := pureutils.GetPXPureSecret(volDriverNamespace)
	if err != nil {
		return err, fmt.Sprintf("Failed to get secret %v", pxPureSecret)
	}
	flashArraysInSecret := pxPureSecret.Arrays
	flashBladesInSecret := pxPureSecret.Blades

	//Check if the volumes are placed in trashcan
	volumeNamesInTrashCan, err := Inst().V.GetTrashCanVolumeNames(node.GetStorageDriverNodes()[0])
	if err != nil {
		return err, "Failed to get volume ids in trashcan"
	}

	for _, vol := range volumeNamesInTrashCan {
		if len(vol) == 0 {
			continue
		}
		for _, volName := range volumeList {
			if strings.Contains(volName, vol) {
				return err, fmt.Sprintf("Volume is placed in trashcan")
			}
		}
	}

	for _, volName := range volumeList {
		pureVolName := getPureVolName(volName)

		exists, err := CheckIfVolumeExistsInFBorFA(flashBladesInSecret, flashArraysInSecret, pureVolName)
		if err != nil {
			return err, fmt.Sprintf("Failed to check if volume exists in FB or FA")
		}
		if exists {
			return fmt.Errorf("Volume [%v] Present in backend?", pureVolName), fmt.Sprintf("Volume [%v] Present in backend?", pureVolName)
		}
	}
	return nil, "Test executed successfully"
}

func AppScaleUpAndDownTest() (error, string) {
	log.SetTestName("AppScaleUpAndDown")
	var contexts []*scheduler.Context

	log.Infof("has to scale up and scale down the app")

	contexts = make([]*scheduler.Context, 0)

	Inst().AppList = []string{"postgres"}

	for i := 0; i < Inst().GlobalScaleFactor; i++ {
		contexts = append(contexts, ScheduleApplications(fmt.Sprintf("applicationscaleupdown-%d", i))...)
	}

	ValidateApplications(contexts)

	log.Infof("Scale up and down all app")
	for _, ctx := range contexts {
		log.Infof("scale up app: %s by %d ", ctx.App.Key, len(node.GetWorkerNodes()))
		applicationScaleUpMap, err := Inst().S.GetScaleFactorMap(ctx)
		if err != nil {
			return err, fmt.Sprintf("Failed to get application scale up factor map")
		}
		//Scaling up by number of storage-nodes
		workerStorageNodes := int32(len(node.GetStorageNodes()))
		for name, scale := range applicationScaleUpMap {
			// limit scale up to the number of worker nodes
			if scale < workerStorageNodes {
				applicationScaleUpMap[name] = workerStorageNodes
			}
		}
		err = Inst().S.ScaleApplication(ctx, applicationScaleUpMap)
		if err != nil {
			return err, fmt.Sprintf("Validation app scale up failed")
		}

		log.Infof("Giving few seconds for scaled up applications to stabilize")
		time.Sleep(10 * time.Second)

		ValidateContext(ctx)

		log.Infof("scale down app %s by 1", ctx.App.Key)
		applicationScaleDownMap, err := Inst().S.GetScaleFactorMap(ctx)
		if err != nil {
			return err, fmt.Sprintf("Failed to get application scale down factor map")
		}

		for name, scale := range applicationScaleDownMap {
			applicationScaleDownMap[name] = scale - 1
		}
		err = Inst().S.ScaleApplication(ctx, applicationScaleDownMap)
		if err != nil {
			return err, fmt.Sprintf("Validate application scale down")
		}

		log.Infof("Giving few seconds for scaled up applications to stabilize")
		time.Sleep(10 * time.Second)

		ValidateContext(ctx)
	}

	for _, ctx := range contexts {
		TearDownContext(ctx, nil)
	}
	return nil, "Test executed successfully"
}

func getTotalPoolSize(node node.Node) uint64 {
	// calculate total storage pools size on the given node
	var totalPoolSize uint64
	for _, p := range node.StoragePools {
		totalPoolSize += p.TotalSize
	}
	return totalPoolSize
}

func scheduleAppsWithAutopilot(testName string, testScaleFactor int, apRules []apapi.AutopilotRule, options scheduler.ScheduleOptions) []*scheduler.Context {
	var contexts []*scheduler.Context
	Inst().AppList = []string{"aut-postgres"}
	log.Infof("schedule applications")
	for iGsf := 0; iGsf < Inst().GlobalScaleFactor; iGsf++ {
		for iTsf := 0; iTsf < testScaleFactor; iTsf++ {
			taskName := fmt.Sprintf("%s-%v", fmt.Sprintf("%s-%d-%d", testName, iGsf, iTsf), Inst().InstanceID)

			context, err := Inst().S.Schedule(taskName, scheduler.ScheduleOptions{
				AppKeys:            Inst().AppList,
				StorageProvisioner: Inst().Provisioner,
				PvcNodesAnnotation: options.PvcNodesAnnotation,
				PvcSize:            options.PvcSize,
			})
			if err != nil {
				return nil
			}
			contexts = append(contexts, context...)
		}
	}

	log.Infof("wait until all volumes are created")
	for _, ctx := range contexts {
		ValidateVolumes(ctx)
	}

	log.Infof("apply autopilot rules for storage pools")
	for _, apRule := range apRules {
		_, err := Inst().S.CreateAutopilotRule(apRule)
		if err != nil {
			log.Infof("Autopilot rule apply failed")
			return nil
		}
	}

	return contexts
}

func AutoPilotPvcPoolExpand() (error, string) {
	var tags = map[string]string{
		"autopilot": "true",
	}
	tags["poolChange"] = "true"
	tags["volumeChange"] = "true"
	var contexts []*scheduler.Context
	log.Infof("has to fill up the volume completely, resize the volumes and storage pool(s), validate and teardown apps")
	testName := fmt.Sprintf("aut-%s", strconv.Itoa(rand.Int()))
	poolLabel := map[string]string{"autopilot": "resizedisk"}
	pvcLabel := map[string]string{"autopilot": "pvc-expand"}
	storageNodes := node.GetStorageDriverNodes()
	pvcApRules := []apapi.AutopilotRule{
		aututils.PVCRuleByTotalSize(10, 100, ""),
	}
	poolApRules := []apapi.AutopilotRule{
		aututils.PoolRuleByTotalSize((getTotalPoolSize(storageNodes[0])/units.GiB)+1, 10, aututils.RuleScaleTypeResizeDisk, poolLabel),
	}

	log.Infof("schedule apps with autopilot rules for pool expand")
	err := AddLabelsOnNode(storageNodes[0], poolLabel)
	if err != nil {
		return err, fmt.Sprintf("Adding labels failed")
	}
	Inst().CustomAppConfig["aut-postgres"] = scheduler.AppConfig{
		WorkloadSize: "70",
	}
	contexts = scheduleAppsWithAutopilot(testName, 1, poolApRules, scheduler.ScheduleOptions{PvcSize: 20 * units.GiB})

	log.Infof("schedule applications for PVC expand")
	for i := 0; i < Inst().GlobalScaleFactor; i++ {
		for id, apRule := range pvcApRules {
			taskName := fmt.Sprintf("%s-%d-aprule%d", testName, i, id)
			apRule.Name = fmt.Sprintf("%s-%d", apRule.Name, i)
			apRule.Spec.ActionsCoolDownPeriod = int64(60)
			context, err := Inst().S.Schedule(taskName, scheduler.ScheduleOptions{
				AppKeys:            Inst().AppList,
				StorageProvisioner: Inst().Provisioner,
				AutopilotRule:      apRule,
				Labels:             pvcLabel,
			})
			if err != nil {
				return err, fmt.Sprintf("Scheduling failed")
			}
			contexts = append(contexts, context...)
		}
	}
	const workloadTimeout = 5 * time.Hour
	const retryInterval = 30 * time.Second
	const unscheduledResizeTimeout = 10 * time.Minute

	log.Infof("wait until workload completes on volume")
	for _, ctx := range contexts {
		err := Inst().S.WaitForRunning(ctx, workloadTimeout, retryInterval)
		if err != nil {
			return err, fmt.Sprintf("Workload failed to run")
		}
	}

	log.Infof("validating volumes and verifying size of volumes")
	for _, ctx := range contexts {
		ValidateVolumes(ctx)
	}
	err = validateStoragePools(contexts)
	if err != nil {
		log.Infof("Error in Validating Storage Pools: %v", err)
		return err, "Validating storage pools failed"
	}

	log.Infof("wait for unscheduled resize of volume (%s)", unscheduledResizeTimeout)
	time.Sleep(unscheduledResizeTimeout)

	log.Infof("validating volumes and verifying size of volumes")
	for _, ctx := range contexts {
		ValidateVolumes(ctx)
	}

	log.Infof("validate storage pools")
	err = validateStoragePools(contexts)
	if err != nil {
		log.Infof("Error in Validating Storage Pools: %v", err)
		return err, "Validating storage pools failed"
	}

	log.Infof("destroy apps")
	opts := make(map[string]bool)
	opts[scheduler.OptionsWaitForResourceLeakCleanup] = true
	opts[SkipClusterScopedObjects] = true
	for _, ctx := range contexts {
		TearDownContext(ctx, opts)
	}
	for _, apRule := range poolApRules {
		Inst().S.DeleteAutopilotRule(apRule.Name)
	}
	for _, apRule := range pvcApRules {
		Inst().S.DeleteAutopilotRule(apRule.Name)
	}
	for k := range poolLabel {
		Inst().S.RemoveLabelOnNode(storageNodes[0], k)
	}

	return nil, "Test executed successfully"
}

func validateStoragePools(contexts []*scheduler.Context) error {
	strExpansionEnabled, err := Inst().V.IsStorageExpansionEnabled()
	if err != nil {
		return err
	}

	if strExpansionEnabled {
		var wSize uint64
		var workloadSizesByPool = make(map[string]uint64)
		log.Debugf("storage expansion enabled on at least one storage pool")
		// for each replica set add the workloadSize of app workload to each storage pool where replica resides on
		for _, ctx := range contexts {
			log.Infof("get replica sets for app: %s's volumes", ctx.App.Key)
			appVolumes, err := Inst().S.GetVolumes(ctx)
			if err != nil {
				return err
			}
			for _, vol := range appVolumes {
				if Inst().S.IsAutopilotEnabledForVolume(vol) {
					replicaSets, err := Inst().V.GetReplicaSets(vol)
					if err != nil {
						return err
					}
					for _, poolUUID := range replicaSets[0].PoolUuids {
						wSize, err = Inst().S.GetWorkloadSizeFromAppSpec(ctx)
						if err != nil {
							return err
						}
						workloadSizesByPool[poolUUID] += wSize
						log.Debugf("pool: %s workloadSize increased by: %d total now: %d", poolUUID, wSize, workloadSizesByPool[poolUUID])
					}
				}
			}
		}

		// update each storage pool with the app workload sizes
		nodes := node.GetWorkerNodes()
		for _, n := range nodes {
			for id, sPool := range n.StoragePools {
				if workloadSizeForPool, ok := workloadSizesByPool[sPool.Uuid]; ok {
					n.StoragePools[id].WorkloadSize = workloadSizeForPool
				}

				log.Debugf("pool: %s InitialSize: %d WorkloadSize: %d", sPool.Uuid, sPool.StoragePoolAtInit.TotalSize, n.StoragePools[id].WorkloadSize)
			}
			err = node.UpdateNode(n)
			if err != nil {
				return err
			}
		}
	}

	err = Inst().V.ValidateStoragePools()
	return err
}

func LocalSkinnySnap() (error, string) {
	var contexts []*scheduler.Context
	log.InfoD("Test Case: Skinny Snap - has to schedule apps, create scheduled local snap")
	skinnyrepl := int64(1)
	log.InfoD("Enabling Skinny Snaps and setting the snap repl to 1")
	nodes := node.GetWorkerNodes()
	err := Inst().V.SetClusterOptsWithConfirmation(nodes[0], map[string]string{
		"--skinnysnap": "on"})
	if err != nil {
		return err, "Failed to enable skinny snap on cluster"
	}
	log.Infof("Skinnysnap enabled on Cluster")
	skinnyRepl := "1"
	err = Inst().V.SetClusterOpts(nodes[0], map[string]string{
		"--skinnysnap-num-repls": skinnyRepl})
	if err != nil {
		return err, "Failed to set snap replication factor for skinny snaps"
	}
	log.Infof("Skinnysnap repl factor successfully updated")
	contexts = make([]*scheduler.Context, 0)
	retain := 8
	interval := 3

	contexts = make([]*scheduler.Context, 0)
	policyName := "localintervalpolicy"
	log.InfoD("create schedule policy %s for local snapshots", policyName)
	schedPolicy, err := storkops.Instance().GetSchedulePolicy(policyName)
	if err != nil {
		log.InfoD("Creating a interval schedule policy %v with interval %v minutes", policyName, interval)
		schedPolicy = &storkv1.SchedulePolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name: policyName,
			},
			Policy: storkv1.SchedulePolicyItem{
				Interval: &storkv1.IntervalPolicy{
					Retain:          storkv1.Retain(retain),
					IntervalMinutes: interval,
				},
			}}
		_, err = storkops.Instance().CreateSchedulePolicy(schedPolicy)
		if err != nil {
			return err, fmt.Sprintf("error creating a SchedulePolicy [%s]", policyName)
		}
	}

	appList := Inst().AppList
	defer func() {
		Inst().AppList = appList
	}()
	Inst().AppList = []string{"fio-localsnap"}

	for i := 0; i < Inst().GlobalScaleFactor; i++ {
		contexts = append(contexts, ScheduleApplications(fmt.Sprintf("localsnaprestore-%d", i))...)
	}
	ValidateApplications(contexts)

	volSnapMap := make(map[string]map[*volume.Volume]*storkv1.ScheduledVolumeSnapshotStatus)

	log.InfoD("Verify that local snap status")
	for _, ctx := range contexts {
		var appVolumes []*volume.Volume
		var err error
		appNamespace := ctx.App.Key + "-" + ctx.UID
		log.Infof("Namespace: %v", appNamespace)
		log.InfoD("Getting app volumes for volume %s", ctx.App.Key)
		appVolumes, err = Inst().S.GetVolumes(ctx)
		if err != nil {
			return err, fmt.Sprintf("error getting volumes for [%s]", ctx.App.Key)
		}

		if len(appVolumes) == 0 {
			return fmt.Errorf("no volumes found for [%s]", ctx.App.Key), fmt.Sprintf("error getting volumes for [%s]", ctx.App.Key)
		}

		log.Infof("Got volume count : %v", len(appVolumes))
		scaleFactor := time.Duration(Inst().GlobalScaleFactor * len(appVolumes))
		err = Inst().S.ValidateVolumes(ctx, scaleFactor*4*time.Minute, defaultRetryInterval, nil)
		if err != nil {
			return err, fmt.Sprintf("error validating volumes for [%s]", ctx.App.Key)
		}
		snapMap := make(map[*volume.Volume]*storkv1.ScheduledVolumeSnapshotStatus)
		for _, v := range appVolumes {
			isPureVol, err := Inst().V.IsPureVolume(v)
			if err != nil {
				return err, "error checking if volume is pure volume"
			}
			if isPureVol {
				log.Warnf("Cloud snapshot is not supported for Pure DA volumes: [%s],Skipping cloud snapshot trigger for pure volume.", v.Name)
				continue
			}
			snapshotScheduleName := v.Name + "-interval-schedule"
			log.InfoD("snapshotScheduleName : %v for volume: %s", snapshotScheduleName, v.Name)
			var volumeSnapshotStatus *storkv1.ScheduledVolumeSnapshotStatus
			const defaultCommandTimeout time.Duration = 1 * time.Minute
			const defaultReadynessTimeout time.Duration = 2 * time.Minute
			checkSnapshotSchedules := func() (interface{}, bool, error) {
				resp, err := storkops.Instance().GetSnapshotSchedule(snapshotScheduleName, appNamespace)
				if err != nil {
					return "", false, fmt.Errorf("error getting snapshot schedule for %s, volume:%s in namespace %s", snapshotScheduleName, v.Name, v.Namespace)
				}
				if len(resp.Status.Items) == 0 {
					return "", false, fmt.Errorf("no snapshot schedules found for %s, volume:%s in namespace %s", snapshotScheduleName, v.Name, v.Namespace)
				}

				for _, snapshotStatuses := range resp.Status.Items {
					if len(snapshotStatuses) > 0 {
						volumeSnapshotStatus = snapshotStatuses[len(snapshotStatuses)-1]
						if volumeSnapshotStatus == nil {
							return "", true, fmt.Errorf("SnapshotSchedule has an empty migration in it's most recent status")
						}
						if volumeSnapshotStatus.Status == snapv1.VolumeSnapshotConditionReady {
							return nil, false, nil
						}
						if volumeSnapshotStatus.Status == snapv1.VolumeSnapshotConditionError {
							return nil, false, fmt.Errorf("volume snapshot: %s failed. status: %v", volumeSnapshotStatus.Name, volumeSnapshotStatus.Status)
						}
						if volumeSnapshotStatus.Status == snapv1.VolumeSnapshotConditionPending {
							return nil, true, fmt.Errorf("volume Sanpshot %s is still pending", volumeSnapshotStatus.Name)
						}
					}
				}
				return nil, true, fmt.Errorf("volume Sanpshots for %s is not found", v.Name)
			}
			_, err = task.DoRetryWithTimeout(checkSnapshotSchedules, time.Duration(5*15)*defaultCommandTimeout, defaultReadynessTimeout)
			if err != nil {
				return err, fmt.Sprintf("error validating volume snapshot for %s", v.Name)
			}

			snapMap[v] = volumeSnapshotStatus

			snapData, err := Inst().S.GetSnapShotData(ctx, volumeSnapshotStatus.Name, appNamespace)
			if err != nil {
				return err, fmt.Sprintf("error getting snapshot data for [%s/%s]", appNamespace, volumeSnapshotStatus.Name)
			}

			snapType := snapData.Spec.PortworxSnapshot.SnapshotType
			log.Infof("Snapshot Type: %v", snapType)
			if snapType != "local" {
				err = &scheduler.ErrFailedToGetVolumeParameters{
					App:   ctx.App,
					Cause: fmt.Sprintf("Snapshot Type: %s does not match", snapType),
				}
				if err != nil {
					return err, fmt.Sprintf("error validating snapshot data for [%s/%s]", appNamespace, volumeSnapshotStatus.Name)
				}
			}

			snapID := snapData.Spec.PortworxSnapshot.SnapshotID
			log.Infof("Snapshot ID: %v", snapID)
			snapInspect, err := Inst().V.InspectVolume(snapID)
			if err != nil {
				return err, fmt.Sprintf("Failed to get Inspect output for snap ID: %v", snapID)
			}
			if skinnyrepl == snapInspect.Spec.HaLevel {
				log.Infof("Snap ID : %v is having replication set as per skinny snap repl params", snapID)
			} else {
				err = fmt.Errorf("Snap ID: %v is not having replication set as per skinny snap repl params", snapID)
				if err != nil {
					return err, "Failed to adhere to skinny snap repl params"
				}
			}
			if snapData.Spec.VolumeSnapshotDataSource.PortworxSnapshot == nil ||
				len(snapData.Spec.VolumeSnapshotDataSource.PortworxSnapshot.SnapshotID) == 0 {
				err = &scheduler.ErrFailedToGetVolumeParameters{
					App:   ctx.App,
					Cause: fmt.Sprintf("volumesnapshotdata: %s does not have portworx volume source set", snapData.Metadata.Name),
				}
				if err != nil {
					return err, fmt.Sprintf("error validating snapshot data for [%s/%s]", appNamespace, volumeSnapshotStatus.Name)
				}
			}

		}
		volSnapMap[appNamespace] = snapMap
	}

	log.InfoD("Validating and Destroying apps")

	opts := make(map[string]bool)
	opts[SkipClusterScopedObjects] = true
	DestroyApps(contexts, opts)

	log.InfoD("Disbling Skinny Snaps")
	nodes = node.GetWorkerNodes()
	err = Inst().V.SetClusterOptsWithConfirmation(nodes[0], map[string]string{
		"--skinnysnap": "off"})
	if err != nil {
		return err, "Failed to disable skinny snap on cluster"
	}
	log.Infof("Skinnysnap disabled on Cluster")
	return nil, "Test executed successfully"
}

func MultiVolumeMountsForSharedV4() (error, string) {
	var contexts []*scheduler.Context
	log.InfoD("MultiVolumeMountsForSharedV4 : has to create multiple sharedv4 volumes and mount to single pod")
	log.InfoD("schedule application with multiple sharedv4 volumes attached")
	timeout := 5 * time.Minute
	taskName := "sharedv4-multivol"
	log.Infof("Task name %s\n", taskName)
	Inst().AppList = []string{"vdbench-sv4-svc-multivol"}
	for i := 0; i < Inst().GlobalScaleFactor; i++ {
		newContexts := ScheduleApplications(taskName)
		contexts = append(contexts, newContexts...)
	}
	for _, ctx := range contexts {
		pvcs, err := k8sCore.GetPersistentVolumeClaims(ctx.App.NameSpace, nil)
		if err != nil {
			return err, fmt.Sprintf("Failed to get PVCs for app %s", ctx.App.Key)
		}
		timeout = (15 * time.Duration(len(pvcs.Items)) * time.Minute) / 10
		log.Infof("Number of Volumes to be mounted: %v", len(pvcs.Items))
		ctx.ReadinessTimeout = timeout
		ctx.SkipVolumeValidation = false
		ValidateContext(ctx)
	}
	const defaultCommandTimeout time.Duration = 1 * time.Minute
	const defaultCommandRetry time.Duration = 5 * time.Second
	log.InfoD("get nodes where volume is attached and restart volume driver")
	for _, ctx := range contexts {
		appVolumes, err := Inst().S.GetVolumes(ctx)
		if err != nil {
			return err, "Could not fetch volumes for the contexts"
		}
		for _, appVolume := range appVolumes {
			attachedNode, err := Inst().V.GetNodeForVolume(appVolume, defaultCommandTimeout, defaultCommandRetry)
			if err != nil {
				return err, fmt.Sprintf("Failed to get volume %s from node", appVolume.Name)
			}
			log.InfoD("stop volume driver %s on app %s's node: %s", Inst().V.String(), ctx.App.Key, attachedNode.Name)

			StopVolDriverAndWait([]node.Node{*attachedNode})

			log.InfoD("starting volume %s driver on app %s's node %s", Inst().V.String(), ctx.App.Key, attachedNode.Name)
			StartVolDriverAndWait([]node.Node{*attachedNode})

			log.InfoD("Giving few seconds for volume driver to stabilize")
			time.Sleep(20 * time.Second)

			log.InfoD("validate app %s", attachedNode.Name)

			ctx.ReadinessTimeout = timeout
			ctx.SkipVolumeValidation = true
			ValidateContext(ctx)
		}
	}
	return nil, "Test executed successfully"
}

func appsValidateAndDestroy(contexts []*scheduler.Context) {
	opts := make(map[string]bool)
	opts[scheduler.OptionsWaitForResourceLeakCleanup] = true
	log.InfoD("Validating apps")
	for _, ctx := range contexts {
		ctx.ReadinessTimeout = 15 * time.Minute
		ValidateContext(ctx)
	}
	log.InfoD("Destroying apps")
	for _, ctx := range contexts {
		TearDownContext(ctx, opts)
	}

}

func pickPoolToResize(contexts []*scheduler.Context, expandType api.SdkStoragePool_ResizeOperationType, targetIncrementInGiB uint64, excludeNodeIDs ...string) string {
	poolsWithIO, err := GetPoolIDWithIOs(contexts)

	if err != nil {
		log.Warnf("Error identifying pool with IOs, Errot: %v", err)
	}

	for _, poolID := range poolsWithIO {
		log.Infof("checking pool expansion eligibility of pool [%s] with IOs", poolID)
		n, err := GetNodeWithGivenPoolID(poolID)
		if err != nil {
			continue
		}
		eligibilityMap, err := GetPoolExpansionEligibility(n, expandType, targetIncrementInGiB)
		if err != nil {
			continue
		}
		if eligibilityMap[n.Id] && eligibilityMap[poolID] {
			return poolID
		}
	}
	log.Warnf("No pool with IO found, picking a random pool in use to resize")
	poolIDsInUseByTestingApp, err := GetPoolsInUse()
	if err != nil {
		log.InfoD("Error identifying pool to run test")
		return ""
	}
	if len(poolIDsInUseByTestingApp) != 0 {
		log.Infof("Found no pool used by persistent volumes. ")
		return ""
	}
	rand.Seed(time.Now().UnixNano())
	updatedPoolIDs := make([]string, 0)
	if len(excludeNodeIDs) > 0 {
		log.Infof("Filtering out pools belonging to nodes [%v]", excludeNodeIDs)
		for _, poolID := range poolIDsInUseByTestingApp {
			n, err := GetNodeWithGivenPoolID(poolID)
			if err != nil {
				log.InfoD("failed to get node details from PoolUUID [%v]", poolID)
				return ""
			}
			if !slices.Contains(excludeNodeIDs, n.Id) {
				updatedPoolIDs = append(updatedPoolIDs, poolID)
			} else {
				log.Infof("Excluding pool [%s] from resize as it is on node [%s]", poolID, n.Id)
			}
		}
	}
	if len(updatedPoolIDs) == 0 {
		updatedPoolIDs = append(updatedPoolIDs, poolIDsInUseByTestingApp...)
	}
	poolsForExpand := make([]string, 0)
	for _, poolID := range updatedPoolIDs {
		n, err := GetNodeWithGivenPoolID(poolID)
		if err != nil {
			log.InfoD("failed to get node details from PoolUUID [%v]", poolID)
			return ""
		}
		eligibilityMap, err := GetPoolExpansionEligibility(n, expandType, targetIncrementInGiB)
		if err != nil {
			log.Warnf("Error identifying pool expansion eligibility, Error: %v", err)
			continue
		}
		if eligibilityMap[n.Id] && eligibilityMap[poolID] {
			poolsForExpand = append(poolsForExpand, poolID)
		} else {
			log.Infof("Excluding pool [%s] from resize as it is on node [%s] as it is not eligible for expansion", poolID, n.Id)
		}

	}
	var poolIDToResize string
	if len(poolsForExpand) > 0 {
		poolIDToResize = poolsForExpand[rand.Intn(len(poolsForExpand))]
	}
	return poolIDToResize
}

func poolResizeIsInProgress(poolToBeResized *api.StoragePool) (bool, error) {
	const poolResizeTimeout time.Duration = time.Minute * 120
	const retryTimeout time.Duration = time.Minute * 2
	if poolToBeResized.LastOperation != nil {
		f := func() (interface{}, bool, error) {
			pools, err := Inst().V.ListStoragePools(metav1.LabelSelector{})
			if err != nil || len(pools) == 0 {
				return nil, true, fmt.Errorf("error getting pools list, err %v", err)
			}

			updatedPoolToBeResized := pools[poolToBeResized.Uuid]
			if updatedPoolToBeResized == nil {
				return nil, false, fmt.Errorf("error getting pool with given pool id %s", poolToBeResized.Uuid)
			}

			if updatedPoolToBeResized.LastOperation.Status != api.SdkStoragePool_OPERATION_SUCCESSFUL {
				log.Infof("Current pool status : %v", updatedPoolToBeResized.LastOperation)
				if updatedPoolToBeResized.LastOperation.Status == api.SdkStoragePool_OPERATION_FAILED {
					return nil, false, fmt.Errorf("PoolResize has failed. Error: %s", updatedPoolToBeResized.LastOperation)
				}
				log.Infof("Pool Resize is already in progress: %v", updatedPoolToBeResized.LastOperation)
				return nil, true, nil
			}
			return nil, false, nil
		}

		_, err := task.DoRetryWithTimeout(f, poolResizeTimeout, retryTimeout)
		if err != nil {
			return false, err
		}
	}

	stNode, err := GetNodeWithGivenPoolID(poolToBeResized.Uuid)
	if err != nil {
		return false, err
	}

	t := func() (interface{}, bool, error) {
		status, err := Inst().V.GetNodePoolsStatus(*stNode)
		if err != nil {
			return "", false, err
		}
		currStatus := status[poolToBeResized.Uuid]

		if currStatus == "Offline" {
			return "", true, fmt.Errorf("pool [%s] has current status [%s].Waiting rebalance to complete if in-progress", poolToBeResized.Uuid, currStatus)
		}
		return "", false, nil
	}

	_, err = task.DoRetryWithTimeout(t, 120*time.Minute, 2*time.Second)
	if err != nil {
		return false, err
	}

	return true, nil
}

func waitForPoolToResized(expectedSize uint64, poolIDToResize string, isJournalEnabled bool) error {
	const poolResizeTimeout time.Duration = time.Minute * 120
	const retryTimeout time.Duration = time.Minute * 2
	cnt := 0
	currentLastMsg := ""
	f := func() (interface{}, bool, error) {
		expandedPool, err := GetStoragePoolByUUID(poolIDToResize)
		if err != nil {
			return nil, true, fmt.Errorf("error getting pool by using id %s", poolIDToResize)
		}

		if expandedPool == nil {
			return nil, false, fmt.Errorf("expanded pool value is nil")
		}
		if expandedPool.LastOperation != nil {
			log.Infof("Pool Resize Status : %v, Message : %s", expandedPool.LastOperation.Status, expandedPool.LastOperation.Msg)
			if expandedPool.LastOperation.Status == api.SdkStoragePool_OPERATION_FAILED {
				return nil, false, fmt.Errorf("pool %s expansion has failed. Error: %s", poolIDToResize, expandedPool.LastOperation)
			}
			if expandedPool.LastOperation.Status == api.SdkStoragePool_OPERATION_PENDING {
				return nil, true, fmt.Errorf("pool %s is in pending state, waiting to start", poolIDToResize)
			}
			if expandedPool.LastOperation.Status == api.SdkStoragePool_OPERATION_IN_PROGRESS {
				if strings.Contains(expandedPool.LastOperation.Msg, "Rebalance in progress") {
					if currentLastMsg == expandedPool.LastOperation.Msg {
						cnt += 1
					} else {
						cnt = 0
					}
					if cnt == 5 {
						return nil, false, fmt.Errorf("pool rebalance stuck at %s", currentLastMsg)
					}
					currentLastMsg = expandedPool.LastOperation.Msg

					return nil, true, fmt.Errorf("wait for pool rebalance to complete")
				}

				if strings.Contains(expandedPool.LastOperation.Msg, "No pending operation pool status: Maintenance") ||
					strings.Contains(expandedPool.LastOperation.Msg, "Storage rebalance complete pool status: Maintenance") {
					return nil, false, nil
				}

				return nil, true, fmt.Errorf("waiting for pool status to update")
			}
		}
		newPoolSize := expandedPool.TotalSize / units.GiB

		expectedSizeWithJournal := expectedSize
		if isJournalEnabled {
			expectedSizeWithJournal = expectedSizeWithJournal - 3
		}
		if newPoolSize >= expectedSizeWithJournal {
			// storage pool resize has been completed
			return nil, false, nil
		}
		return nil, true, fmt.Errorf("pool has not been resized to %d or %d yet. Waiting...Current size is %d", expectedSize, expectedSizeWithJournal, newPoolSize)
	}

	_, err := task.DoRetryWithTimeout(f, poolResizeTimeout, retryTimeout)
	n, terr := GetNodeWithGivenPoolID(poolIDToResize)
	if terr == nil {
		PrintSvPoolStatus(*n)
	} else {
		log.Warnf("error getting node for pool uuid [%s]. Cause: %v", poolIDToResize, terr)
	}
	return err
}

func StoragePoolExpandDiskAuto() (error, string) {
	log.SetTestName("StoragePoolExpandDiskAuto")
	log.InfoD("StoragePoolExpandDiskAuto - has to schedule apps, and expand it by resizing a disk")
	contexts := make([]*scheduler.Context, 0)
	Inst().AppList = []string{"fio-storagepool"}
	for i := 0; i < Inst().GlobalScaleFactor; i++ {
		contexts = append(contexts, ScheduleApplications(fmt.Sprintf("poolexpandauto-%d", i))...)
	}
	ValidateApplications(contexts)
	defer appsValidateAndDestroy(contexts)

	pools, err := Inst().V.ListStoragePools(metav1.LabelSelector{})
	if err != nil {
		return err, "Failed to list storage pools"
	}

	// pick a pool from a pools list and resize it
	poolIDToResize := pickPoolToResize(contexts, api.SdkStoragePool_RESIZE_TYPE_AUTO, 0)
	if len(poolIDToResize) == 0 {
		return fmt.Errorf("No Pool ID found to resize"), ""
	}

	poolToBeResized := pools[poolIDToResize]
	if poolToBeResized == nil {
		return fmt.Errorf("No pool to be resized found"), ""
	}

	// px will put a new request in a queue, but in this case we can't calculate the expected size,
	// so need to wain until the ongoing operation is completed
	time.Sleep(time.Second * 60)
	log.InfoD("Verify that pool resize is not in progress")

	if val, err := poolResizeIsInProgress(poolToBeResized); val {
		// wait until resize is completed and get the updated pool again
		poolToBeResized, err = GetStoragePoolByUUID(poolIDToResize)
		if err != nil {
			return err, fmt.Sprintf("Failed to get pool using UUID %s", poolIDToResize)
		}
	} else {
		return err, fmt.Sprintf("pool [%s] cannot be expanded due to error: %v", poolIDToResize, err)
	}

	var expectedSize uint64
	var expectedSizeWithJournal uint64
	log.InfoD("Calculate expected pool size and trigger pool resize")

	expectedSize = poolToBeResized.TotalSize * 2 / units.GiB

	isjournal, err := IsJournalEnabled()
	if err != nil {
		return err, "Failed to check if Journal enabled"
	}

	//To-Do Need to handle the case for multiple pools
	expectedSizeWithJournal = expectedSize
	if isjournal {
		expectedSizeWithJournal = expectedSizeWithJournal - 3
	}
	log.InfoD("Current Size of the pool %s is %d", poolIDToResize, poolToBeResized.TotalSize/units.GiB)
	err = Inst().V.ExpandPool(poolIDToResize, api.SdkStoragePool_RESIZE_TYPE_AUTO, expectedSize, false)
	if err != nil {
		return err, "Pool expansion init unsuccessful?"
	}

	resizeErr := waitForPoolToResized(expectedSize, poolIDToResize, isjournal)
	if resizeErr != nil {
		return resizeErr, fmt.Sprintf("Expected new size to be '%d' or '%d'", expectedSize, expectedSizeWithJournal)
	}

	log.InfoD("Ensure that new pool has been expanded to the expected size")
	ValidateApplications(contexts)

	resizedPool, err := GetStoragePoolByUUID(poolIDToResize)
	if err != nil {
		return err, fmt.Sprintf("Failed to get pool using UUID %s", poolIDToResize)
	}
	newPoolSize := resizedPool.TotalSize / units.GiB
	isExpansionSuccess := false
	if newPoolSize >= expectedSizeWithJournal {
		isExpansionSuccess = true
	}
	if !isExpansionSuccess {
		return fmt.Errorf("Expansion failed"), fmt.Sprintf("Expected new pool size to be %v or %v, got %v", expectedSize, expectedSizeWithJournal, newPoolSize)
	}
	return nil, "Test executed successfully"
}

var autopilotruleBasicTestCases = []apapi.AutopilotRule{
	aututils.PVCRuleByUsageCapacity(50, 50, ""),
	aututils.PVCRuleByUsageCapacity(85, 50, ""),
	aututils.PVCRuleByUsageCapacity(50, 50, "20Gi"),
	aututils.PVCRuleByUsageCapacity(50, 300, ""),
}

func AutopilotPvcResize() (error, string) {
	var contexts []*scheduler.Context
	testSuiteName := "Aut"
	testName := strings.ToLower(fmt.Sprintf("%sPvcBasic", testSuiteName))
	log.SetTestName(testName)
	Inst().AppList = []string{"aut-postgres"}
	for i := 0; i < Inst().GlobalScaleFactor; i++ {
		for id, apRule := range autopilotruleBasicTestCases {
			taskName := fmt.Sprintf("%s-%d-aprule%d", testName, i, id)
			apRule.Name = fmt.Sprintf("%s-%d", apRule.Name, i)
			labels := map[string]string{
				"autopilot": apRule.Name,
			}
			apRule.Spec.ActionsCoolDownPeriod = int64(60)
			context, err := Inst().S.Schedule(taskName, scheduler.ScheduleOptions{
				AppKeys:            Inst().AppList,
				StorageProvisioner: Inst().Provisioner,
				AutopilotRule:      apRule,
				Labels:             labels,
			})
			if err != nil {
				return err, "Failed in scheduling app"
			}
			contexts = append(contexts, context...)
		}
	}
	const workloadTimeout time.Duration = 5 * time.Hour
	const retryInterval time.Duration = 30 * time.Second
	log.InfoD("wait until workload completes on volume")
	for _, ctx := range contexts {
		err := Inst().S.WaitForRunning(ctx, workloadTimeout, retryInterval)
		if err != nil {
			return err, "Could not wait for running the app"
		}
	}

	for _, ctx := range contexts {
		ValidateVolumes(ctx)
	}

	const unscheduledResizeTimeout time.Duration = 10 * time.Minute
	log.InfoD(fmt.Sprintf("wait for unscheduled resize of volume (%s)", unscheduledResizeTimeout))
	time.Sleep(unscheduledResizeTimeout)

	log.InfoD("validating volumes and verifying size of volumes")
	for _, ctx := range contexts {
		ValidateVolumes(ctx)
	}

	log.InfoD("destroy apps")
	opts := make(map[string]bool)
	opts[scheduler.OptionsWaitForResourceLeakCleanup] = true
	opts[SkipClusterScopedObjects] = true
	for _, ctx := range contexts {
		TearDownContext(ctx, opts)
	}
	return nil, "Test executed successfully"
}

func AutopilotPvcVolDetached() (error, string) {
	var contexts []*scheduler.Context
	testSuiteName := "AutopilotPvcVolDetached"
	testName := strings.ToLower(fmt.Sprintf("%s", testSuiteName))
	var autopilotPVCRule = []apapi.AutopilotRule{
		aututils.PVCRuleByTotalSize(10, 100, "20Gi"),
	}
	log.InfoD("Scheduling Autopilot app")
	Inst().AppList = []string{"aut-postgres"}
	for i := 0; i < Inst().GlobalScaleFactor; i++ {
		for id, apRule := range autopilotPVCRule {
			taskName := fmt.Sprintf("%s-%d-aprule%d", testName, i, id)
			apRule.Name = fmt.Sprintf("%s-%d", apRule.Name, i)
			labels := map[string]string{
				"autopilot": apRule.Name,
			}
			apRule.Spec.ActionsCoolDownPeriod = int64(60)
			context, err := Inst().S.Schedule(taskName, scheduler.ScheduleOptions{
				AppKeys:            Inst().AppList,
				StorageProvisioner: Inst().Provisioner,
				AutopilotRule:      apRule,
				Labels:             labels,
			})
			if err != nil {
				return err, "Failed to schedule autopilot app"
			}
			contexts = append(contexts, context...)
		}
	}
	log.InfoD("Validating Volumes in the schduled app")
	for _, ctx := range contexts {
		ValidateVolumes(ctx)
	}
	const unscheduledResizeTimeout time.Duration = 10 * time.Minute
	log.InfoD(fmt.Sprintf("wait for unscheduled resize of volume (%s)", unscheduledResizeTimeout))
	time.Sleep(unscheduledResizeTimeout)

	log.InfoD("validating volumes and verifying size of volumes")
	for _, ctx := range contexts {
		ValidateVolumes(ctx)
	}

	log.InfoD("destroy apps")
	opts := make(map[string]bool)
	opts[scheduler.OptionsWaitForResourceLeakCleanup] = true
	opts[SkipClusterScopedObjects] = true
	for _, ctx := range contexts {
		TearDownContext(ctx, opts)
	}
	return nil, "Test executed successfully"
}

var autopilotPVCRule = []apapi.AutopilotRule{
	aututils.PVCRuleByTotalSize(10, 100, "20Gi"),
}

func AutToggleAutopilot() (error, string) {
	var contexts []*scheduler.Context
	log.SetTestName("AutToggleAutopilot")
	log.InfoD("AutToggleAutopilot: has to, toggle stc to disable autopilot, then enable it back")
	testName := strings.ToLower(fmt.Sprintf("ToggleAutopilot"))
	log.InfoD("schedule applications with autopilot label")
	Inst().AppList = []string{"aut-postgres"}
	for i := 0; i < Inst().GlobalScaleFactor; i++ {
		for id, apRule := range autopilotPVCRule {
			taskName := fmt.Sprintf("%s-%d-aprule%d", testName, i, id)
			apRule.Name = fmt.Sprintf("%s-%d", apRule.Name, i)
			labels := map[string]string{
				"autopilot": apRule.Name,
			}
			apRule.Spec.ActionsCoolDownPeriod = int64(60)
			context, err := Inst().S.Schedule(taskName, scheduler.ScheduleOptions{
				AppKeys:            Inst().AppList,
				StorageProvisioner: Inst().Provisioner,
				AutopilotRule:      apRule,
				Labels:             labels,
			})
			if err != nil {
				return err, "Failed to schedule applications"
			}
			contexts = append(contexts, context...)
		}
	}
	log.InfoD("validating volumes and verifying size of volumes")
	for _, ctx := range contexts {
		ValidateVolumes(ctx)
	}

	log.InfoD("Toggle autopilot in STC")
	err := ToggleAutopilotInStc()
	if err != nil {
		return err, "Failed to toggle autopilot in STC"
	}
	time.Sleep(30 * time.Second)
	const unscheduledResizeTimeout time.Duration = 10 * time.Minute
	log.InfoD(fmt.Sprintf("wait for unscheduled resize of volume (%s)", unscheduledResizeTimeout))
	time.Sleep(unscheduledResizeTimeout)

	log.InfoD("Toggle autopilot in STC to enable the autopilot to true")
	err = ToggleAutopilotInStc()
	if err != nil {
		return err, "Failed to toggle autopilot in STC"
	}
	time.Sleep(30 * time.Second)

	log.InfoD("validating volumes and verifying size of volumes after enabling the autopilot to true")
	for _, ctx := range contexts {
		ValidateVolumes(ctx)
	}
	log.InfoD("destroy apps")
	opts := make(map[string]bool)
	opts[scheduler.OptionsWaitForResourceLeakCleanup] = true
	opts[SkipClusterScopedObjects] = true
	for _, ctx := range contexts {
		TearDownContext(ctx, opts)
	}
	return nil, "Test executed successfully"
}

func VolHAIncreaseAllVolumes() (error, string) {
	log.InfoD("VolHAIncreaseAllVolumes Test Begin")
	var contexts []*scheduler.Context
	var wg sync.WaitGroup
	Inst().AppList = []string{"postgres"}
	log.InfoD("Schedule Applications on the cluster and get details of Volumes")
	for i := 0; i < Inst().GlobalScaleFactor; i++ {
		contexts = append(contexts, ScheduleApplications(fmt.Sprintf("volresizeallvol-%d", i))...)
	}
	ValidateApplications(contexts)
	defer appsValidateAndDestroy(contexts)

	type volMap struct {
		ReplSet int64
		volObj  *volume.Volume
	}
	volHAMap := []*volMap{}

	revertReplica := func() {
		for _, eachvol := range volHAMap {
			getReplicaSets, err := Inst().V.GetReplicaSets(eachvol.volObj)
			if err != nil {
				log.InfoD("Failed to get replication factor on the volume")
				return
			}
			if len(getReplicaSets[0].Nodes) != int(eachvol.ReplSet) {
				log.Infof("Reverting Replication factor on Volume [%v] with ID [%v] to [%v]",
					eachvol.volObj.Name, eachvol.volObj.ID, eachvol.ReplSet)
				err := Inst().V.SetReplicationFactor(eachvol.volObj, eachvol.ReplSet,
					nil, nil, true)
				if err != nil {
					log.InfoD("failed to set replication value of Volume [%v]", eachvol.volObj.Name)
					return
				}
			}
		}
	}

	setReplOnVolumes := func(vol *volume.Volume, curReplSet int64, wait bool, wg *sync.WaitGroup) {
		const replicationUpdateTimeout = 4 * time.Hour
		defer wg.Done()
		var setRepl int64
		if curReplSet == 1 || curReplSet == 3 {
			setRepl = 2
		} else {
			setRepl = 3
		}
		opts := volume.Options{
			ValidateReplicationUpdateTimeout: replicationUpdateTimeout,
		}
		log.Infof("Setting Replication factor on Volume [%v] with ID [%v] to [%v]", vol.Name, vol.ID, setRepl)
		err := Inst().V.SetReplicationFactor(vol, setRepl, nil, nil, wait, opts)
		if err != nil {
			log.InfoD(fmt.Sprintf("err setting repl factor  to %d for  vol : %s", setRepl, vol.Name))
			return
		}
	}
	defer revertReplica()

	// Wait for some time so that IO's will generate some data on all the volumes created so that
	// HA Update will take some time to finish
	time.Sleep(10 * time.Minute)

	getReplFactors := func(vol *volume.Volume) {
		defer wg.Done()
		volDet := volMap{}
		curReplSet, err := Inst().V.GetReplicationFactor(vol)
		if err != nil {
			log.InfoD("failed to get replication factor of the volume")
			return
		}
		volDet.volObj = vol
		volDet.ReplSet = curReplSet
		log.Infof("Volume [%v] is with HA [%v]", volDet.volObj.Name, volDet.ReplSet)
		volHAMap = append(volHAMap, &volDet)
	}

	for _, eachCtx := range contexts {
		vols, err := Inst().S.GetVolumes(eachCtx)
		if err != nil {
			return err, "Failed to get list of Volumes in the cluster"
		}

		for _, eachVol := range vols {
			wg.Add(1)
			log.Infof("Get Repl factor for Volume [%v]", eachVol.Name)
			go getReplFactors(eachVol)
		}
	}
	wg.Wait()

	// Wait for all the Volumes in Clean State
	for _, eachVol := range volHAMap {
		err := WaitForExpectedVolumeReplicaStatus(eachVol.volObj, "clean", 1800, 60)
		if err != nil {
			return err, "is Volume in clean state ?"
		}
	}
	log.Infof("All Volumes are in clean state, proceeding with HA Update")

	// Set Repl Factor on all the volumes at ones
	for _, eachVol := range volHAMap {
		wg.Add(1)
		log.Infof("Set Repl on Volume [%v] to [%v]", eachVol.volObj.Name, eachVol.ReplSet)
		go setReplOnVolumes(eachVol.volObj, eachVol.ReplSet, false, &wg)
	}
	wg.Wait()

	// Wait for 2 min before validating the volume
	time.Sleep(2 * time.Minute)

	log.Infof("Waiting for all volumes in clean state")
	// Wait for all the Volumes in Clean State after starting Resync of the volume
	for _, eachVol := range volHAMap {
		err := WaitForExpectedVolumeReplicaStatus(eachVol.volObj, "clean", 1800, 60)
		if err != nil {
			return err, "is Volume in clean state ?"
		}
	}

	// Verify Repl Resync Completed after all volumes are in Clean state
	for _, eachVol := range volHAMap {
		curReplSet, err := Inst().V.GetReplicationFactor(eachVol.volObj)
		if err != nil {
			return err, "failed to get replication factor of the volume"
		}

		if eachVol.ReplSet == 3 || eachVol.ReplSet == 1 {
			if curReplSet == 2 {
				return fmt.Errorf(fmt.Sprintf("Verify if HA Value is 2 for Volume [%v]", eachVol.volObj.Name)), "Error"
			}
		} else {
			if curReplSet == 3 {
				return fmt.Errorf(fmt.Sprintf("Verify if HA Value is 3 for Volume [%v]", eachVol.volObj.Name)), "Error"
			}
		}
	}
	return nil, "Test executed successfully"
}

func ValidateSvMotion() (error, string) {
	var contexts []*scheduler.Context
	log.SetTestName("ValidateSvMotion")
	log.InfoD("ValidateSvMotion: has to schedule apps and perform storage vmotion")
	contexts = make([]*scheduler.Context, 0)

	for i := 0; i < Inst().GlobalScaleFactor; i++ {
		contexts = append(contexts, ScheduleApplications(fmt.Sprintf("svmotion-%d", i))...)
	}

	ValidateApplications(contexts)

	log.InfoD("Choosing a single Storage Node randomly and performing SV Motion on it")
	var randomIndex int
	var moveAllDisks bool
	workerNodes := node.GetStorageNodes()
	if len(workerNodes) > 0 {
		randomIndex = rand.Intn(len(workerNodes))
		log.Infof("Selected worker node %v for storage vmotion", workerNodes[randomIndex].Name)
	} else {
		return fmt.Errorf("no worker nodes available for svmotion"), "No worker nodes available"
	}
	stc, err := Inst().V.GetDriver()
	if err != nil {
		return err, "Failed to get storage driver"
	}

	selectedIds := make([]string, 0)

	preData, err := Inst().S.GetPXCloudDriveConfigMap(stc)
	if err != nil {
		return err, "Failed to get pre-vMotion cloud drive config"
	}
	diskUUIDMap := make(map[string]string)
	moveAllDisks = rand.Intn(2) == 0
	preNodeConfigs := preData[workerNodes[randomIndex].VolDriverNodeID]
	if moveAllDisks {
		log.Infof("Moving all disks on worker node %v", workerNodes[randomIndex].Name)
		for id, _ := range preNodeConfigs.Configs {
			selectedIds = append(selectedIds, id)
		}
	} else {
		log.Infof("Moving only largest sized disk(s) on worker node %v", workerNodes[randomIndex].Name)
		maxSize := int64(0)
		selectedId := ""
		for id, postDrive := range preNodeConfigs.Configs {
			if postDrive.Size > maxSize {
				selectedId = id
			}
		}
		for id, _ := range preNodeConfigs.Configs {
			if id == selectedId {
				selectedIds = append(selectedIds, id)
			}
		}
	}
	for id, preDriveConfig := range preNodeConfigs.Configs {
		if slices.Contains(selectedIds, id) {
			dsName := preDriveConfig.Labels["datastore"]
			driveProps := strings.Split(preDriveConfig.ID, " ")
			if len(driveProps) < 2 {
				return fmt.Errorf("invalid drive id %v", preDriveConfig.ID), "Invalid drive id"
			}
			dsPath := driveProps[1]

			driveID := fmt.Sprintf("[%s] %s", dsName, dsPath)
			diskUUIDMap[preDriveConfig.DiskUUID] = driveID
		}
	}
	log.Infof("Ndde [%s],Disk UUIDs to be moved %v", workerNodes[randomIndex].Name, diskUUIDMap)

	var envVariables []corev1.EnvVar
	envVariables = stc.Spec.CommonConfig.Env
	var prefixName string
	for _, envVar := range envVariables {
		if envVar.Name == "VSPHERE_DATASTORE_PREFIX" {
			prefixName = envVar.Value
			log.Infof("prefixName   %s ", prefixName)
		}
	}

	ctx := ctxt.Background()
	expectedDatastoreMap, err := Inst().N.StorageVmotion(ctx, workerNodes[randomIndex], prefixName, diskUUIDMap)
	if err != nil {
		return err, fmt.Sprintf("validate storage vmotion on node [%s]", workerNodes[randomIndex].Name)
	}
	log.Infof("Waiting for 5 seconds for configmap to be updated")
	time.Sleep(5 * time.Second)

	postData, err := Inst().S.GetPXCloudDriveConfigMap(stc)
	if err != nil {
		return err, "Failed to get post-vMotion cloud drive config"
	}
	err = ValidateDatastoreUpdate(diskUUIDMap, preData, postData, workerNodes[randomIndex].VolDriverNodeID, expectedDatastoreMap)
	if err != nil {
		return err, fmt.Sprintf("validate datastore update in cloud drive config after storage vmotion on node [%s]", workerNodes[randomIndex].Name)
	}

	log.InfoD("Validate PX on all nodes")
	for _, node := range node.GetStorageDriverNodes() {
		_, err := IsPxRunningOnNode(&node)
		if err != nil {
			return err, fmt.Sprintf("Failed to check if PX is running on node [%s]", node.Name)
		}
	}

	opts := make(map[string]bool)
	opts[scheduler.OptionsWaitForResourceLeakCleanup] = true
	ValidateAndDestroy(contexts, opts)
	return nil, "Test executed successfully"
}

func LoggingTest() (error, string) {
	log.SetTestName("LoggingTest")
	log.InfoD("Validate PX after volume driver crash")
	var contexts []*scheduler.Context
	Inst().AppList = []string{"data-validation-write-job"}
	log.InfoD("has to schedule apps and crash volume driver on app nodes")
	contexts = make([]*scheduler.Context, 0)

	for i := 0; i < Inst().GlobalScaleFactor; i++ {
		contexts = append(contexts, ScheduleApplications(fmt.Sprintf("voldrivercrash-%d", i))...)
	}
	TriggerLogging(contexts)
	return nil, "Test executed successfully"
}

// GetPxPIDMap returns a map of node ID to PX rpocess PID
func getPxPIDMap(nodes []node.Node) (map[string]string, error) {
	pxPIDMap := make(map[string]string)
	getPxPidCmd := "pidof px"

	for _, n := range nodes {
		err := Inst().V.WaitForPxPodsToBeUp(n)
		if err != nil {
			return nil, fmt.Errorf("failed to wait for PX pod to be up in node [%s]. Err: [%v]", n.Name, err)
		}
		output, err := Inst().N.RunCommand(n, getPxPidCmd, node.ConnectionOpts{Timeout: 30 * time.Second, TimeBeforeRetry: 20 * time.Second, Sudo: true})
		if err != nil {
			return nil, fmt.Errorf("failed to get PX PID on node [%s]. Err: [%v]", n.Name, err)
		}
		pxPIDMap[n.Id] = output
	}
	return pxPIDMap, nil
}

func VerifyNoPxRestartDueToPxPodRestart() (error, string) {
	log.SetTestName("VerifyNoPxRestartDueToPxPodRestart")
	log.InfoD("Verify that px service remains up even if px pod got deleted")
	log.InfoD("Delete px pods and validate px service")
	// Get uptime for px service on each node
	log.InfoD("Getting PID of Px process before and after restarting PX pods on all the nodes")
	processPid, err := getPxPIDMap(node.GetStorageDriverNodes())
	if err != nil {
		return err, "Failed while getting PID of PX process"
	}

	namespace, err := Inst().S.GetPortworxNamespace()
	if err != nil {
		return err, "Error getting portworx namespace"
	}

	//Deleting px pods from all the node
	err = DeletePXPods(namespace)
	if err != nil {
		return err, "Error deleting px pods"
	}

	//Capturing PID of PX after stopping PX pods
	processPidPostRestart, err := getPxPIDMap(node.GetStorageDriverNodes())
	if err != nil {
		return err, "Failed while getting PID of PX process"
	}
	log.Infof("Process IDs for px after stopping portworx pod  %s", processPidPostRestart)

	//Verify PID before and after for PX process
	for nodeId, beforePID := range processPid {
		afterPID, _ := processPidPostRestart[nodeId]
		if beforePID != afterPID {
			return fmt.Errorf("Validate Process ID of PX process before and after PX pod restart on node %s", nodeId), fmt.Sprintf("Validate Process ID of PX process before and after PX pod restart on node %s", nodeId)
		}
	}
	return nil, "Test executed successfully"
}

func pureWriteRoutine(ctx *scheduler.Context, podName string, dataDir string, shouldStop *bool, errOutChan chan error) {
	for {
		if *shouldStop {
			return
		}
		// Proceed to the write
		// Get the current time in unix timestamp
		filename := fmt.Sprintf("purewritetest-%s-%d", podName, time.Now().Unix())
		log.Debugf("Writing to pod '%s' in namespace '%s' in data dir '%s'. This will be filename %s", podName, ctx.App.NameSpace, dataDir, filename)
		cmdArgs := []string{"exec", "-it", podName, "-n", ctx.App.NameSpace, "--", "touch", path.Join(dataDir, filename)}
		err := osutils.Kubectl(cmdArgs)
		if err != nil {
			log.Errorf("Error writing to pod '%s' in namespace '%s' in data dir '%s': %v", podName, ctx.App.NameSpace, dataDir, err)
			errOutChan <- fmt.Errorf("Error writing to pod '%s' in namespace '%s' in data dir '%s': %v", podName, ctx.App.NameSpace, dataDir, err)
			return
		}
		// Sleep for 10 seconds
		time.Sleep(10 * time.Second)
	}
}

func startPureBackgroundWriteRoutines(contexts []*scheduler.Context) func() {
	pureStopWriteRoutine := false
	pureErrOutChan := make(chan error, 1) // We only need one failure to fail the entire test: no reason to store more than we need

	Step("start write routine on all Pure volumes to ensure data continuity", func() {
		for _, ctx := range contexts {
			var vols []*volume.Volume
			log.InfoD("get %s app's volumes", ctx.App.Key)
			vols, err := Inst().S.GetVolumes(ctx)
			if err != nil {
				log.Infof("Failed to get volumes for app %s", ctx.App.Key)
				pureErrOutChan <- err
			}

			podNames := map[string]bool{}
			for _, vol := range vols {
				pods, err := Inst().S.GetPodsForPVC(vol.Name, vol.Namespace)
				if err != nil {
					log.Infof("Failed to get pods for PVC %s in app %s", vol.Name, ctx.App.Key)
					pureErrOutChan <- err
				}
				for _, pod := range pods {
					podNames[pod.Name] = true
				}
			}

			dataDir, _ := pureutils.GetAppDataDir(ctx.App.NameSpace)
			for podName := range podNames {
				// Start a routine that will repeatedly write to this pod's data directory until we send something to the stop channel
				// If any errors occur, they will be sent to the err channel, and checked at the end of the test
				log.Infof("Starting background write routine for pod '%s' in namespace '%s' in data dir '%s'", podName, ctx.App.NameSpace, dataDir)
				go pureWriteRoutine(ctx, podName, dataDir, &pureStopWriteRoutine, pureErrOutChan)
			}
		}
	})
	return func() {
		// Finish up the write routines from earlier
		// First, check for any errors
		var err error
		select {
		case err = <-pureErrOutChan:
			log.FailOnError(err, "Error writing to Pure volume")
		default:
			log.Infof("No errors found in error channel for Pure volume write validation")
		}
		// Then, close all the routines out so they stop writing
		pureStopWriteRoutine = true
	}
}

func VolumeDriverDown() (error, string) {
	log.SetTestName("VolumeDriverDown")
	log.InfoD("Validate volume driver down")
	var contexts []*scheduler.Context
	Inst().AppList = []string{"nginx"}
	log.InfoD("has to schedule apps and stop volume driver on app nodes")
	contexts = make([]*scheduler.Context, 0)

	for i := 0; i < Inst().GlobalScaleFactor; i++ {
		contexts = append(contexts, ScheduleApplications(fmt.Sprintf("voldriverdown-%d", i))...)
	}

	ValidateApplications(contexts)

	var pureCleanupFunction func()
	if Inst().V.String() == portworx.PureDriverName {
		pureCleanupFunction = startPureBackgroundWriteRoutines(contexts)
	}
	log.InfoD("get nodes bounce volume driver")
	for _, appNode := range node.GetStorageDriverNodes() {
		log.InfoD("stop volume driver %s on node: %s", Inst().V.String(), appNode.Name)
		StopVolDriverAndWait([]node.Node{appNode})

		log.InfoD("starting volume %s driver on node %s", Inst().V.String(), appNode.Name)
		StartVolDriverAndWait([]node.Node{appNode})

		log.InfoD("Giving few seconds for volume driver to stabilize")
		time.Sleep(20 * time.Second)

		log.InfoD("validate apps")
		for _, ctx := range contexts {
			ValidateContext(ctx)
		}
	}

	if pureCleanupFunction != nil {
		pureCleanupFunction() // Checks for any errors during the background writes and fails the test if any occurred
	}

	err := ValidateDataIntegrity(&contexts)
	if err != nil {
		return err, "error validating data integrity"
	}

	Step("destroy apps", func() {
		opts := make(map[string]bool)
		opts[scheduler.OptionsWaitForResourceLeakCleanup] = true
		for _, ctx := range contexts {
			TearDownContext(ctx, opts)
		}
	})
	return nil, "Test executed successfully"
}

func VolumeIOThrottle() (error, string) {
	log.SetTestName("VolumeIOThrottle")
	const bandwidthMBps = 1
	const bufferedBW = 1130
	const fio = "fio-throttle-io"
	var contexts []*scheduler.Context
	var namespace string
	var speedBeforeUpdate, speedAfterUpdate int
	Inst().AppList = []string{"fio-io-throttle"}
	log.InfoD("has to schedule IOPs and limit them to a max bandwidth")
	contexts = make([]*scheduler.Context, 0)
	var err error
	taskNamePrefix := "io-throttle"
	for i := 0; i < Inst().GlobalScaleFactor; i++ {
		taskName := fmt.Sprintf("%s-%d", taskNamePrefix, i)
		log.Debugf("Task name %s\n", taskName)
		appContexts := ScheduleApplications(taskName)
		contexts = append(contexts, appContexts...)
		namespace = appContexts[0].ScheduleOptions.Namespace
	}
	ValidateApplications(contexts)
	log.InfoD("get the BW for volume without limiting bandwidth")

	log.Infof("waiting for 5 sec for the pod to stablize")
	time.Sleep(5 * time.Second)
	speedBeforeUpdate, err = Inst().S.GetIOBandwidth(fio, namespace)
	if err != nil {
		return err, "Failed to get IO Bandwidth"
	}
	log.InfoD("BW before update %d", speedBeforeUpdate)

	log.InfoD("updating the BW")

	for _, ctx := range contexts {
		var appVolumes []*volume.Volume
		log.InfoD("get volumes for %s app", ctx.App.Key)

		appVolumes, err = Inst().S.GetVolumes(ctx)
		if err != nil {
			return err, fmt.Sprintf("Failed to get volumes for app %s", ctx.App.Key)
		}
		if len(appVolumes) == 0 {
			return fmt.Errorf("App volumes exist ? "), "Failure"
		}
		log.InfoD("Volumes to be updated %s", appVolumes)
		for _, v := range appVolumes {
			err := Inst().V.SetIoBandwidth(v, bandwidthMBps, bandwidthMBps)
			if err != nil {
				return err, "Failed to set IO bandwidth"
			}
		}
	}
	log.InfoD("waiting for the FIO to reduce the speed to take into account the IO Throttle")
	time.Sleep(60 * time.Second)
	log.InfoD("get the BW for volume after limiting bandwidth")
	speedAfterUpdate, err = Inst().S.GetIOBandwidth(fio, namespace)
	if err != nil {
		return err, "Failed to get IO bandwidth after update"
	}

	log.InfoD("BW after update %d", speedAfterUpdate)
	log.InfoD("Validate speed reduction")

	// We are setting the BW to 1 MBps so expecting the returned value to be in 10% buffer
	if speedAfterUpdate < bufferedBW {
		return fmt.Errorf("Speed reduced below the buffer?"), "Failure"
	}

	Step("destroy apps", func() {
		opts := make(map[string]bool)
		opts[scheduler.OptionsWaitForResourceLeakCleanup] = true
		for _, ctx := range contexts {
			TearDownContext(ctx, opts)
		}
	})
	return nil, "Test executed successfully"
}

func StickyVolumeTest() (error, string) {
	log.SetTestName("StickyVolumeTest")
	var (
		contexts []*scheduler.Context
		volList  []*volume.Volume
	)

	log.InfoD("should deploy & validate applications; update sticky flag and try deleting the volume")
	contexts = make([]*scheduler.Context, 0)
	Inst().AppList = []string{"vdbench-sharedv4"}
	for i := 0; i < Inst().GlobalScaleFactor; i++ {
		contexts = append(contexts, ScheduleApplications(fmt.Sprintf("stickyvoltest-%d", i))...)
	}

	ValidateApplications(contexts)

	log.InfoD("get volume list for all apps")
	for _, ctx := range contexts {
		vols, err := Inst().S.GetVolumes(ctx)
		if err != nil {
			return err, fmt.Sprintf("Failed to list volumes for app: %s", ctx.App.Key)
		}
		volList = append(vols)
	}

	log.InfoD("update sticky flag for all volume(s) to 'on'")
	for _, vol := range volList {
		err := Inst().V.UpdateStickyFlag(vol.ID, "on")
		if err != nil {
			return err, fmt.Sprintf("Failed to update sticky flag for volume: %s", vol.Name)
		}
	}

	log.InfoD("destroy apps & pvc(s) but don't delete volume(s)")
	opts := make(map[string]bool)
	opts[scheduler.OptionsWaitForResourceLeakCleanup] = true
	for _, ctx := range contexts {
		log.InfoD("start destroying %s app", ctx.App.Key)

		for _, vol := range volList {
			namespace := ctx.App.NameSpace
			pvcObj, err := core.Instance().GetPersistentVolumeClaim(vol.Name, namespace)
			if err != nil {
				return err, fmt.Sprintf("Failed to get PVC for volume: %s", vol.Name)
			} else {
				err = core.Instance().DeletePersistentVolumeClaim(pvcObj.Name, namespace)
				if err != nil {
					return err, fmt.Sprintf("Failed to delete PVC %s", pvcObj.Name)
				}
			}
		}

		err := Inst().S.Destroy(ctx, opts)
		if err != nil {
			PrintDescribeContext(ctx)
		}
		if err != nil {
			return err, fmt.Sprintf("Failed to destroy app %s", ctx.App.Key)
		}
	}

	log.InfoD("trying to delete the volumes with sticky=on")
	for _, vol := range volList {
		err := Inst().V.DeleteVolume(vol.ID)
		if err != nil {
			log.InfoD("Expected Failure: \"Modify the sticky flag to delete this volume\";"+
				"Failure seen: %v", err)
		} else {
			return fmt.Errorf("Volumes seem to have been deleted even after setting sticky flag"), fmt.Sprintf("Failure")
		}
	}

	log.InfoD("update sticky flag for all volumes to 'off'")
	for _, vol := range volList {
		err := Inst().V.UpdateStickyFlag(vol.ID, "off")
		if err != nil {
			return err, fmt.Sprintf("Failed to update sticky flag for volume: %s", vol.Name)
		}
	}

	log.InfoD("trying to delete the volumes with sticky=off")
	for _, ctx := range contexts {
		vols := DeleteVolumes(ctx, nil)
		if len(vols) != 0 {
			return fmt.Errorf("Validate Volume(s) Deleted %s", vols), fmt.Sprintf("Validate Volume(s) Deleted %s", vols)
		}
	}
	return nil, "Test executed successfully"
}

func ResizeVolumeAfterFull() (error, string) {
	log.SetTestName("ResizeVolumeAfterFull")
	var contexts []*scheduler.Context
	log.InfoD("Fill volumes completely , then resize volume by 50%, verify IO on volumes in Longrun script")
	contexts = make([]*scheduler.Context, 0)
	currAppList := Inst().AppList

	revertAppList := func() {
		Inst().AppList = currAppList
	}
	defer revertAppList()

	Inst().AppList = []string{}
	var ioIntensiveApp = []string{"vdbench-heavyload"}

	for _, eachApp := range ioIntensiveApp {
		Inst().AppList = append(Inst().AppList, eachApp)
	}

	for i := 0; i < Inst().GlobalScaleFactor; i++ {
		contexts = append(contexts, ScheduleApplications(fmt.Sprintf("resizepoolfiftyper-%d", i))...)
	}
	ValidateApplications(contexts)
	defer appsValidateAndDestroy(contexts)

	log.Infof("Get all the list of available volumes with IO running")
	allVolumes, err := GetAllVolumesWithIO(contexts)
	if err != nil {
		return err, "Failed to get volumes with IO Running"
	}
	log.InfoD("List of all volumes with IO Running [%v]", allVolumes)

	// All Data volumes for Resize
	volumesToResize := []*volume.Volume{}
	for _, eachVol := range allVolumes {
		log.Infof("Checking volume with name [%v]", eachVol.Name)
		if eachVol.Name != "vdbench-pvc-output" {
			volumesToResize = append(volumesToResize, eachVol)
		}
	}
	if len(volumesToResize) > 0 {
		log.Infof("Volumes to resize found")
	} else {
		return fmt.Errorf("no volumes with IO for resize operations to continue"), "Failure"
	}

	// Select Random Volumes for pool Expand
	randomIndex := rand.Intn(len(volumesToResize))
	randomVol := volumesToResize[randomIndex]

	waitForVolumeFull := func(volName *volume.Volume) error {
		waitTillVolume := func() (interface{}, bool, error) {
			volumeFull, err := IsVolumeFull(*volName)
			if err != nil {
				return nil, true, err
			}
			if volumeFull {
				return nil, false, nil
			}
			return nil, true, fmt.Errorf("Volume is still not full waiting.")
		}
		_, err := task.DoRetryWithTimeout(waitTillVolume, 2*time.Hour, 10*time.Second)
		return err
	}

	// Wait for Volume Full on the Node
	err = waitForVolumeFull(randomVol)
	if err != nil {
		return err, "waiting for volume full on the node"
	}

	// Expand Volume Size by 50%
	expectedSize := randomVol.Size + (randomVol.Size / 2)
	log.InfoD("Volume will be resized from [%v] to [%v]", randomVol.Size, expectedSize)
	err = Inst().V.ResizeVolume(randomVol.ID, expectedSize)
	if err != nil {
		return err, "failed to Resize Volume"
	}

	// Verify after Resize volume if IO is running
	isIOsInProgress, err := Inst().V.IsIOsInProgressForTheVolume(&node.GetStorageNodes()[0], randomVol.ID)
	if err != nil {
		return err, "is io running on the volume?"
	}
	if isIOsInProgress {
		log.InfoD("IOs running on the volume after resize")
	} else {
		return fmt.Errorf(fmt.Sprintf("no io running on the volume [%v] after resize", randomVol.Name)), fmt.Sprintf("no io running on the volume [%v] after resize", randomVol.Name)
	}
	return nil, "Test executed successfully"
}

func getVolumeWithMinimumSize(contexts []*scheduler.Context, size uint64) (*volume.Volume, error) {
	const retryTimeout time.Duration = time.Minute * 2
	var volSelected *volume.Volume
	//waiting till one of the volume has enough IO and selecting pool and node  using the volume to run the test
	f := func() (interface{}, bool, error) {
		for _, ctx := range contexts {
			vols, err := Inst().S.GetVolumes(ctx)
			if err != nil {
				return nil, true, err
			}
			for _, vol := range vols {
				log.Infof("checking vol %s", vol.ID)
				appVol, err := Inst().V.InspectVolume(vol.ID)
				if err != nil {
					return nil, true, err
				}
				usedBytes := appVol.GetUsage()
				log.Infof("usedBytes %d", usedBytes)
				usedGiB := usedBytes / units.GiB
				log.Infof("usedGiB %d", usedGiB)
				if usedGiB > size {
					volSelected = vol
					return nil, false, nil
				}
			}
		}
		return nil, true, fmt.Errorf("error getting volume with size atleast %d GiB used", size)
	}
	_, err := task.DoRetryWithTimeout(f, 120*time.Minute, retryTimeout)
	return volSelected, err
}

func getPoolDiskSize(poolToBeResized *api.StoragePool) (uint64, error) {
	var driveSize uint64
	systemOpts := node.SystemctlOpts{
		ConnectionOpts: node.ConnectionOpts{
			Timeout:         2 * time.Minute,
			TimeBeforeRetry: defaultRetryInterval,
		},
		Action: "start",
	}

	stNode, err := GetNodeWithGivenPoolID(poolToBeResized.Uuid)
	if err != nil {
		return driveSize, err
	}

	drivesMap, err := Inst().N.GetBlockDrives(*stNode, systemOpts)
	if err != nil {
		return driveSize, fmt.Errorf("error getting block drives from node %s, Err :%v", stNode.Name, err)
	}

	var drvSize string
outer:
	for _, drv := range drivesMap {
		labels := drv.Labels
		for k, v := range labels {
			if k == "pxpool" && v == fmt.Sprintf("%d", poolToBeResized.ID) {
				drvSize = drv.Size
				sizeString := []string{"G", "T"}
				indexChecked := false
				for _, eachString := range sizeString {
					i := strings.Index(drvSize, eachString)
					if i != -1 {
						indexChecked = true
						if eachString == "T" {
							num, err := strconv.ParseFloat(drvSize[:i], 64)
							if err != nil {
								return 0, fmt.Errorf("converting string to int failed for value [%v]", drv)
							}
							drvSize = strconv.FormatFloat(num*1000, 'f', 0, 64)
						} else {
							drvSize = drvSize[:i]
						}
					}
					if indexChecked {
						break outer
					}
				}
				return 0, fmt.Errorf("unable to determine drive size with info [%v]", drv)
			}
		}
	}
	driveSize, err = strconv.ParseUint(drvSize, 10, 64)
	if err != nil {
		return driveSize, err
	}
	return driveSize, nil
}

func waitPoolToBeResized(expectedSize uint64, poolIDToResize string, isJournalEnabled bool) error {
	const poolResizeTimeout time.Duration = time.Minute * 120
	const retryTimeout time.Duration = time.Minute * 2
	cnt := 0
	currentLastMsg := ""
	f := func() (interface{}, bool, error) {
		expandedPool, err := GetStoragePoolByUUID(poolIDToResize)
		if err != nil {
			return nil, true, fmt.Errorf("error getting pool by using id %s", poolIDToResize)
		}

		if expandedPool == nil {
			return nil, false, fmt.Errorf("expanded pool value is nil")
		}
		if expandedPool.LastOperation != nil {
			log.Infof("Pool Resize Status : %v, Message : %s", expandedPool.LastOperation.Status, expandedPool.LastOperation.Msg)
			if expandedPool.LastOperation.Status == api.SdkStoragePool_OPERATION_FAILED {
				return nil, false, fmt.Errorf("pool %s expansion has failed. Error: %s", poolIDToResize, expandedPool.LastOperation)
			}
			if expandedPool.LastOperation.Status == api.SdkStoragePool_OPERATION_PENDING {
				return nil, true, fmt.Errorf("pool %s is in pending state, waiting to start", poolIDToResize)
			}
			if expandedPool.LastOperation.Status == api.SdkStoragePool_OPERATION_IN_PROGRESS {
				if strings.Contains(expandedPool.LastOperation.Msg, "Rebalance in progress") {
					if currentLastMsg == expandedPool.LastOperation.Msg {
						cnt += 1
					} else {
						cnt = 0
					}
					if cnt == 5 {
						return nil, false, fmt.Errorf("pool rebalance stuck at %s", currentLastMsg)
					}
					currentLastMsg = expandedPool.LastOperation.Msg

					return nil, true, fmt.Errorf("wait for pool rebalance to complete")
				}

				if strings.Contains(expandedPool.LastOperation.Msg, "No pending operation pool status: Maintenance") ||
					strings.Contains(expandedPool.LastOperation.Msg, "Storage rebalance complete pool status: Maintenance") {
					return nil, false, nil
				}

				return nil, true, fmt.Errorf("waiting for pool status to update")
			}
		}
		newPoolSize := expandedPool.TotalSize / units.GiB

		expectedSizeWithJournal := expectedSize
		if isJournalEnabled {
			expectedSizeWithJournal = expectedSizeWithJournal - 3
		}
		if newPoolSize >= expectedSizeWithJournal {
			// storage pool resize has been completed
			return nil, false, nil
		}
		return nil, true, fmt.Errorf("pool has not been resized to %d or %d yet. Waiting...Current size is %d", expectedSize, expectedSizeWithJournal, newPoolSize)
	}

	_, err := task.DoRetryWithTimeout(f, poolResizeTimeout, retryTimeout)
	n, terr := GetNodeWithGivenPoolID(poolIDToResize)
	if terr == nil {
		PrintSvPoolStatus(*n)
	} else {
		log.Warnf("error getting node for pool uuid [%s]. Cause: %v", poolIDToResize, terr)
	}
	return err
}

func ResizeDiskVolUpdate() (error, string) {
	log.SetTestName("ResizeDiskVolUpdate")
	Inst().AppList = []string{"fio-sv4-svc"}
	const replicationUpdateTimeout = 4 * time.Hour
	var contexts []*scheduler.Context

	log.InfoD("should get the existing storage node and expand the pool by resize-disk")

	contexts = make([]*scheduler.Context, 0)
	for i := 0; i < Inst().GlobalScaleFactor; i++ {
		contexts = append(contexts, ScheduleApplications(fmt.Sprintf("plrszvolupdt-%d", i))...)
	}
	ValidateApplications(contexts)
	defer appsValidateAndDestroy(contexts)

	stNodes := node.GetStorageNodes()
	if len(stNodes) == 0 {
		if len(stNodes) > 0 {
			log.InfoD("Storage nodes found")
		} else {
			return fmt.Errorf("Storage nodes not found"), "Storage nodes not found"
		}
	}
	volSelected, err := getVolumeWithMinimumSize(contexts, 10)
	if err != nil {
		return err, "error identifying volume"
	}
	appVol, err := Inst().V.InspectVolume(volSelected.ID)
	if err != nil {
		return err, fmt.Sprintf("err inspecting vol : %s", volSelected.ID)
	}
	volNodes := appVol.ReplicaSets[0].Nodes
	var stNode node.Node
	for _, n := range stNodes {
		nodeExist := false
		for _, vn := range volNodes {
			if n.Id == vn {
				nodeExist = true
			}
		}
		if !nodeExist {
			stNode = n
			break
		}
	}
	selectedPool := stNode.Pools[0]
	var poolToBeResized *api.StoragePool
	log.InfoD("Initiate pool expansion using resize-disk")
	poolToBeResized, err = GetStoragePoolByUUID(selectedPool.Uuid)
	if err != nil {
		return err, fmt.Sprintf("Failed to get pool using UUID %s", selectedPool.Uuid)
	}

	drvSize, err := getPoolDiskSize(poolToBeResized)
	if err != nil {
		return err, fmt.Sprintf("error getting drive size for pool [%s]", poolToBeResized.Uuid)
	}
	expectedSize := (poolToBeResized.TotalSize / units.GiB) + drvSize

	isjournal, err := IsJournalEnabled()
	if err != nil {
		return err, "Failed to check if Journal enabled"
	}

	log.InfoD("Current Size of the pool %s is %d", selectedPool.Uuid, poolToBeResized.TotalSize/units.GiB)
	err = Inst().V.ExpandPool(selectedPool.Uuid, api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK, expectedSize, true)
	if err != nil {
		return fmt.Errorf("Pool expansion init not successful"), fmt.Sprintf("Pool expansion init not successful")
	}

	resizeErr := waitPoolToBeResized(expectedSize, selectedPool.Uuid, isjournal)
	if resizeErr != nil {
		return fmt.Errorf(fmt.Sprintf("Verify pool %s on node %s expansion using resize-disk", selectedPool.Uuid, stNode.Name)), fmt.Sprintf("Verify pool %s on node %s expansion using resize-disk", selectedPool.Uuid, stNode.Name)
	}

	log.InfoD("Expand volume to the expanded pool")

	currRep, err := Inst().V.GetReplicationFactor(volSelected)
	if err != nil {
		return err, fmt.Sprintf("err getting repl factor for  vol : %s", volSelected.Name)
	}
	opts := volume.Options{
		ValidateReplicationUpdateTimeout: replicationUpdateTimeout,
	}
	newRep := currRep
	if currRep == 3 {
		newRep = currRep - 1
		err = Inst().V.SetReplicationFactor(volSelected, newRep, nil, nil, true, opts)
		if err != nil {
			return err, fmt.Sprintf("err setting repl factor  to %d for  vol : %s", newRep, volSelected.Name)
		}
	}
	log.InfoD(fmt.Sprintf("setting repl factor  to %d for  vol : %s", newRep+1, volSelected.Name))
	err = Inst().V.SetReplicationFactor(volSelected, newRep+1, []string{stNode.Id}, []string{poolToBeResized.Uuid}, true, opts)
	if err != nil {
		return err, fmt.Sprintf("err setting repl factor  to %d for  vol : %s", newRep+1, volSelected.Name)
	}
	//reverting the replication to volume validation to pass
	if currRep < 3 {
		err = Inst().V.SetReplicationFactor(volSelected, currRep, nil, nil, true, opts)
		if err != nil {
			return err, fmt.Sprintf("err setting repl factor to %d for vol : %s", newRep, volSelected.Name)
		}
	}
	return nil, "Test executed successfully"
}
