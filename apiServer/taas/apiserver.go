package main

import (
	"github.com/gin-gonic/gin"
	"github.com/pure-px/torpedo/apiServer/taas/utils"
	"log"
)

// We will define all API calls here.
// Once Gin Server starts, it will initialise all APIs it contains.
// Future work : To have segregated APIs based on need -> We will have to create multiple main calls for initialising.
func main() {
	router := gin.Default()
	router.DELETE("taas/deletens/:namespace", utils.DeleteNS)
	router.POST("taas/createns", utils.CreateNS)
	router.POST("taas/inittorpedo", utils.InitializeDrivers)
	router.GET("taas/getnodes", utils.GetNodes)
	router.POST("taas/rebootnode/:nodename", utils.RebootNode)
	router.GET("taas/storagenodes", utils.GetStorageNodes)
	router.GET("taas/storagelessnodes", utils.GetStorageLessNodes)
	router.POST("taas/collectsupport", utils.CollectSupport)
	router.POST("taas/scheduleapps", utils.ScheduleAppsAndValidate)
	router.POST("taas/deploypxagent", utils.ExecuteHelmCmd)
	router.GET("taas/getclusterid/:namespace", utils.GetNamespaceID)
	router.GET("taas/getclusternodestatus", utils.GetNodeStatus)
	router.POST("taas/runhelmcmd", utils.ExecuteHelmCmd)
	router.GET("taas/pxversion", utils.GetPxVersion)
	router.GET("taas/ispxinstalled", utils.IsPxInstalled)
	router.GET("taas/getpxctloutput", utils.GetPxctlStatusOutput)
	router.GET("taas/getkubevirtvmsbyns", utils.GetVMsInNamespaces)
	router.GET("taas/getkubevirtvmsbynslabels", utils.GetVMsWithNamespaceLabels)
	router.POST("taas/namespaces/addLabel", utils.AddNSLabel)
	router.POST("taas/stork/upgrade", utils.UpgradeStork)
	router.DELETE("taas/deletepod", utils.DeletePod)
	router.GET("taas/getpxbackupnamespace", utils.GetPxBackupNamespace)
	router.POST("taas/createvolumesnapshotclass", utils.CreateVolumeSnapshotClass)
	router.POST("taas/RunSetupTeardownTest", utils.RunSetupTeardownTestGin)
	router.POST("taas/CreateLargeNumberOfVolumesTest", utils.CreateLargeNumberOfVolumesTestGin)
	router.POST("taas/EnableTrashCanDeleteVol", utils.EnableTrashCanDeleteVolGin)
	router.POST("taas/ChainedLocalSnapAndValidateRestoreTest", utils.ChainedLocalSnapAndValidateRestoreTestGin)
	router.POST("taas/AppScaleUpAndDownTest", utils.AppScaleUpAndDownTestGin)
	router.POST("taas/AutoPilotPvcPoolExpand", utils.AutoPilotPvcPoolExpandGin)
	router.POST("taas/LocalSkinnySnap", utils.LocalSkinnySnapGin)
	router.POST("taas/MultiVolumeMountsForSharedV4", utils.MultiVolumeMountsForSharedV4Gin)
	router.POST("taas/StoragePoolExpandDiskAuto", utils.StoragePoolExpandDiskAutoGin)
	router.POST("taas/AutopilotPvcResize", utils.AutopilotPvcResizeGin)
	router.POST("taas/AutopilotPvcVolDetached", utils.AutopilotPvcVolDetachedGin)
	router.POST("taas/AutToggleAutopilot", utils.AutToggleAutopilotGin)
	router.POST("taas/VolHAIncreaseAllVolumes", utils.VolHAIncreaseAllVolumesGin)
	router.POST("taas/ValidateSvMotion", utils.ValidateSvMotionGin)
	router.POST("taas/VolumeIOThrottle", utils.VolumeIOThrottleGin)
	router.POST("taas/StickyVolumeTest", utils.StickyVolumeTestGin)
	router.POST("taas/ResizeVolumeAfterFull", utils.ResizeVolumeAfterFullGin)
	router.POST("taas/ResizeDiskVolUpdate", utils.ResizeDiskVolUpdateGin)
	router.POST("taas/LoggingTest", utils.LoggingTestGin)
	router.POST("taas/VerifyNoPxRestartDueToPxPodRestart", utils.VerifyNoPxRestartDueToPxPodRestartGin)
	router.POST("taas/VolumeDriverDown", utils.VolumeDriverDownGin)
	router.POST("taas/runCompositeJob", utils.RunCompositeJob)
	router.GET("taas/compositeJobStatus/:job_id", utils.GetCompositeJobStatus)
	log.Fatal(router.Run(":8080"))
}
