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
	var volumrlidttr []*api.Volume

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

var _ = Describe("{LegacySharedVolumeMigrate}", func() {
	var testrailID = 296370
	var pxNode node.Node
	volumeName := "legacy-shared-volume-idle"
	stepLog := "Create legacy shared volume and migrate it to shared v4 service volume"
	It (stepLog, func() {
		pxNodes, err := GetStorageNodes()
		log.FailOnError(err, "Unable to get storange nodes")
		pxNode := GetRandomNode(pxNodes)
		log.Infof("Turning on Creation of Legacy shared volumes")
		pxctlCmdFull := fmt.Sprintf("cluster update option --create-legacy-shared-as-sharedv4-service=false")
		_, err = Inst().V.GetPxctlCmdOoutput(pxNode, pxctlCmdFull)
		log.FailOnError(err, fmt.Sprintf("error updating cluster option"))
		pxctlCmdFull = fmt.Sprintf("v c --shared=true %s", volumeName)
		output, err := Inst().V.GetPxctlCmdOoutput(pxNode, pxctlCmdFull)
		log.FailOnError(err, fmt.Sprintf("error creating legacy shared volume %s", volumeName))
		log.Infof(output)
		vol, err := Inst().V.InspectVolume(volumeName)
		log.FailOnError(err, fmt.Sprintf("Inspect volume failed on volume {%v}", volumeName))
		Expect(vol.Spec.Shared).To(BeTrue(), "non-shared volume created unexpectedly")
		pxctlCmdFull = fmt.Sprintf("cluster update option --migrate-legacy-shared-to-sharedv4-service=true")
		_, err = Inst().V.GetPxctlCmdOutput(pxNode, pxctlCmdFull)
		log.FailOnError(err, fmt.Spritnf("Error setting cluster option"))
		migrated := false
		for i := 0; i < 6; i++ {
			vol, err := Inst().V.InspectVolume(volumeName)
			log.FailOnError(err, fmt.Sprintf("Inspect volume failed on volume {%v}", volumeName))
			if vol.Spec.Shared == "false" && vol.Spec.Sharedv4 == "true" {
				migrated = true
				break
			}
			time.Slep(1 * time.Minute)
		}
		vol, err = Inst().V.InspectVolume(volumeName)
		log.FailOnError(err, fmt.Sprintf("Inspect volume failed on volume {%v}", volumeName))
		Expect(vol.Spec.Shared).To(BeFalse()", Unexpectedly volume is still shared", volumeName)
		Expect(vol.Spec.Sharedv4).To(BeTrue()", Unexpectedly volume is not sharedv4", volumeName)
		pxctlCmdFull = fmt.Sprintf("v d %s", volumeName)
		Inst().V.GetPxctlCmdOoutput(pxNode, pxctlCmdFull)
	})
})







