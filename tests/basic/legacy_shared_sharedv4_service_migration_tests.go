package tests

import (
	"errors"
	"fmt"

)

const (
	ubuntu = "ubuntu-app"
)

// Legacy Shared Volume Create
// Automatically it should get created as Sharedv4 service volume.

var _ = Describe("{LegacySharedVolumeCreate}", func() {
	var testrailID = 296369
	// https://portworx.testrail.net/index.php?/cases/view/296369
	var pxNode node.Node
	var contexts []*scheduler.Context

	volumeName := "legacy-shared=volume"

	stepLog := "Create legacy shared volume and check it got created as sharedv4 service volume"
	It (stepLog, func() {
		pxNodes, err := GetStorageNodes()
		log.FailOnError(err, "Unable to get the storage nodes")
		pxNode := GetRandomNode(pxNodes)
		log.Infof("Creating legacy shared volume: %s", volumeName)
		pxctlCmtFull := fmt.Sprintf("v c --shared=true %s", volumeName)
		output, err = Inst().V.GetPxctlCmdOutput(pxNode, pxctlCmdFull)
		log.FailOnError(err, fmt.Sprintf("error creating legacy shared volume %s", volumeName))
		log.Infof(output)
		vol, err := Inst().V.InspectVolume(volumeName)
		log.FailOnError(err, fmt.Sprintf("Inspect volume failed on volume {%v}", volumeName))
		Expect(vol.Spec.Sharedv4).To(BeTrue(), "sharedv4 volume was not created")
		Expect(vol.Spec.Shared).To(BeFalse(), "shared volume was created unexpectedly")
		pxctlCmdFull = fmt.Sprintf("v d %s", volumeName)
		output, err = Inst().V.GetPxctlCmdOutput(pxNode, pxctlCmdFull)
		log.FailOnError(err, fmt.Sprintf("error deleting legacy shared volume %s", volumeName))
	})
})

func setCreateLegacySharedAsSharedv4Service(on bool) {
	pxNodes, err := GetStorageNodes()
	log.FailOnError(err, "Unable to get storage nodes")
	pxNode := GetRandomNode(pxNodes)
	log.Infof("Setting Creation of Legacy shared volumes")
	var pxctlCmdFull string
	if on {
		pxctlCmdFull = fmt.Sprintf("cluster update option --create-legacy-shared-as-sharedv4-service=true")
	} else {
		pxctlCmdFull = fmt.Sprintf("cluster update option --create-legacy-shared-as-sharedv4-service=false")
	}
	_, err = Inst().V.GetPxctlCmdOoutput(pxNode, pxctlCmdFull)
	log.FailOnError(err, fmt.Sprintf("error updating cluster option"))
	// Sleep so that the config variable can be updated on all nodes.
	time.Sleep(20 * time.Second)
}

func setMigrateLegacySharedToSharedv4Service(on bool) {
	pxNodes, err := GetStorageNodes()
	log.FailOnError(err, "Unable to get storage nodes")
	pxNode := GetRandomNode(pxNodes)
	log.Infof("Turning on Creation of Legacy shared volumes")
	var pxctlCmdFull string
	if on {
		pxctlCmdFull = fmt.Sprintf("cluster update option --create-legacy-shared-as-sharedv4-service=true")
	} else  {
		pxctlCmdFull = fmt.Sprintf("cluster update option --create-legacy-shared-as-sharedv4-service=false")
	}
	_, err = Inst().V.GetPxctlCmdOoutput(pxNode, pxctlCmdFull)
	log.FailOnError(err, fmt.Sprintf("error updating cluster option"))
	// Sleep so that the config variable can be updated on all nodes.
	time.Sleep(20 * time.Second)

}

var _ = Describe("{LegacySharedVolumeMigrate}", func() {
	var testrailID = 296370
	var pxNode node.Node
	volumeName := "legacy-shared-volume-idle"
	stepLog := "Create legacy shared volume and migrate it to shared v4 service volume"
	It (stepLog, func() {
		pxctlCmdFull = fmt.Sprintf("v c --shared=true %s", volumeName)
		output, err := Inst().V.GetPxctlCmdOutput(pxNode, pxctlCmdFull)
		log.FailOnError(err, fmt.Sprintf("error creating legacy shared volume %s", volumeName))
		log.Infof(output)
		vol, err := Inst().V.InspectVolume(volumeName)
		log.FailOnError(err, fmt.Sprintf("Inspect volume failed on volume {%v}", volumeName))
		Expect(vol.Spec.Shared).To(BeTrue(), "non-shared volume created unexpectedly")
		setMigrateLegacySharedToSharedv4Service(true)
		for i := 0; i < 6; i++ {
			vol, err := Inst().V.InspectVolume(volumeName)
			log.FailOnError(err, fmt.Sprintf("Inspect volume failed on volume {%v}", volumeName))
			if vol.Spec.Shared == "false" && vol.Spec.Sharedv4 == "true" {
				migrated = true
				break
			}
			time.Sleep(1 * time.Minute)
		}
		if !migrated {
			log.FailOnError(err, fmt.Sprintf("Migration failed on volume {%v}", volumeName))
		}
		pxctlCmdFull = fmt.Sprintf("v d %s", volumeName)
		Inst().V.GetPxctlCmdOutput(pxNode, pxctlCmdFull)
	})
})

getLegacySharedVolumeCount() int {
	count := 0
	for _, ctx := range contexts {
		vols, err := Inst().S.GetVolumes(ctx)
		Expect(err).NotTo(HaveOccurred())
		for _, v := range vols {
			vol, err := Inst().V.InspectVolume(v.ID)
			Expect(err).NotTo(HaveOccurred(), "failed in inspect volume: %v", err)
			if vol.Spec.Shared {
				count++
			}
		}
	}
	return count
}

func getLegacySharedTestAppVol(ctx *scheduler.Context) (*volume.Volume, *api.Volume, *node.Node) {
	vols, err := Inst().S.GetVolumes(ctx)
	Expect(err).NotToHaveOccured()
	// Assert len > 0.Expect(lenvols)).To(
	vol := vols[0]
	apiVol, err := Inst().V.InspectVolume(vol.ID)
	Expect(err).NotTo(HaveOccurred())

	attachedNode, err := Inst().V.GetNodeForVolume(vol, cmdTimeout, cmdRetry)
	Expect(err).NotTo(HaveOccured())
	log.Infof("volume %v {%v} is attached to node %v", vol.ID, apivol.Id, attachedNode.Name)
	return vol, apiVol, attachedNode
}

func returnMapOfPodsUsingApiSharedVolumes(map[core.Pod]bool sharedVolPods, ctx *scheduler.Context) {
	vols, err := Inst().S.GetVolumes(ctx)
	Expect(err).NotToHaveOccured()
	for _, vol := range vols {
		apiVol, err := Inst().V.InspectVolume(vol.ID)
		Expect(err).NotToHaveOccured()
		if apiVol.Spec.Shared {
			pods, err := core.Instance()GetPodsUsingPV(vol.ID)
			Expect(err).NotToHaveOccured()
			for _, pod := range pods {
				sharedVolPods[pod.UUID] = true
			}
		}
	}
	return
}

func checkMapOfPods(map[core.Pod]bool sharedVolPods, ctx *scheduler.Context) error {
	vols, err := Inst().S.GetVolumes(ctx)
	Expect(err).NotToHaveOccured()
	for _, vol := range vols {
		apiVol, err := Inst().V.InspectVolume(vol.ID)
		Expect(err).NotToHaveOccured()
		pods, err := core.Instance()GetPodsUsingPV(vol.ID)
		Expect(err).NotToHaveOccured()
		for _, pod := range pods {
			_, ok: sharedVolPods[pod.UUID]
			if ok {
				return fmt.Errorf("A pod using shared volume prior to migration remains after migration [%v]", pod.Name)
			}
		}
	}
	return nil
}

var _ = Describe("{LegacySharedVolumeAppMigrateBasic}", func() {
	JustBeforeEach(func() {
		runID = testrailuttils.AddRunsToMilestone(testraildID)
		setCreateLegacySharedAsSharedv4Service(false)
		StartTorpedoTest("LegacySharedVolumeAppMigrateBasic", "Legacy Shared to Sharedv4 Service Functional Test", nil testrailID)
		contexts = make([]*scheduler.Context, 0)
		var err error
		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("%s-$d", namespacePrefix, i))...)m
		}
		// TODO: Skip non legacy shared tests
		ValidateApplications(contexts)
	})

var _ = Describe("{LegacySharedVolumeAppMigrateBasic}", func() {
	JustBeforeEach(func() {
		runID = testrailuttils.AddRunsToMilestone(testraildID)
		setCreateLegacySharedAsSharedv4Service(false)
		StartTorpedoTest("LegacySharedVolumeAppMigrateBasic", "Legacy Shared to Sharedv4 Service Functional Test", nil testrailID)
		contexts = make([]*scheduler.Context, 0)
		var err error
		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("%s-$d", namespacePrefix, i))...)m
		}
		// TODO: Skip non legacy shared tests
		ValidateApplications(contexts)
	})


