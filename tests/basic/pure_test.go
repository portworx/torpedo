package tests

import (
	"fmt"
	"github.com/libopenstorage/openstorage/api"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/portworx/sched-ops/k8s/core"
	"github.com/portworx/sched-ops/task"
	"github.com/portworx/torpedo/drivers/volume"
	"github.com/portworx/torpedo/drivers/volume/portworx"
	"github.com/portworx/torpedo/pkg/log"
	"github.com/portworx/torpedo/pkg/testrailuttils"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/portworx/torpedo/drivers/node"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/drivers/scheduler/k8s"
	"github.com/portworx/torpedo/pkg/pureutils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/portworx/torpedo/tests"
)

const (
	secretNamespace = "kube-system"

	// fbS3CredentialName is the name of the credential object created in pxctl
	// see also formattingPxctlEstablishBackupCredential
	fbS3CredentialName = "fbS3bucket"

	// formattingPxctlEstablishBackupCredential is the command template used to
	// create the S3 credentials object in Portworx
	formattingPxctlEstablishBackupCredential = "pxctl credentials create --provider s3 --s3-access-key %s --s3-secret-key %s --s3-region us-east-1 --s3-endpoint %s --s3-storage-class STANDARD %s"

	// formattingPxctlDeleteFBBackupCredential is the command template used to
	// delete the S3 credentials object in Portworx
	formattingPxctlDeleteFBBackupCredential = "pxctl credentials delete %s"
)

func createCloudsnapCredential() {
	fbConfigs, err := pureutils.GetS3Secret(secretNamespace)
	Expect(err).NotTo(HaveOccurred())
	nodes := node.GetStorageDriverNodes()
	_, err = Inst().N.RunCommand(nodes[0], fmt.Sprintf(formattingPxctlEstablishBackupCredential, fbConfigs.Blades[0].S3AccessKey, fbConfigs.Blades[0].S3SecretKey, fbConfigs.Blades[0].ObjectStoreEndpoint, fbS3CredentialName), node.ConnectionOpts{
		Timeout:         k8s.DefaultTimeout,
		TimeBeforeRetry: k8s.DefaultRetryInterval,
		Sudo:            true,
	})
	// if the cloudsnap credentials already exist, just leave them there
	if err != nil && strings.Contains(err.Error(), "already exist") {
		err = nil
	}
	Expect(err).NotTo(HaveOccurred(), "unexpected error creating cloudsnap credential")
}

func deleteCloudsnapCredential() {
	nodes := node.GetStorageDriverNodes()
	_, err := Inst().N.RunCommand(nodes[0], fmt.Sprintf(formattingPxctlDeleteFBBackupCredential, fbS3CredentialName), node.ConnectionOpts{
		Timeout:         k8s.DefaultTimeout,
		TimeBeforeRetry: k8s.DefaultRetryInterval,
		Sudo:            true,
	})
	Expect(err).NotTo(HaveOccurred(), "unexpected error deleting cloudsnap credential")
}

// This test performs basic tests making sure Pure direct access are running as expected
var _ = Describe("{PureVolumeCRUDWithSDK}", func() {
	var contexts []*scheduler.Context
	JustBeforeEach(func() {
		StartTorpedoTest("PureVolumeCRUDWithSDK", "Test pure volumes on applications, run CRUD", nil, 0)
	})

	It("schedule pure volumes on applications, run CRUD, tear down", func() {
		Step("setup credential necessary for cloudsnap", createCloudsnapCredential)
		contexts = make([]*scheduler.Context, 0)

		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("purevolumestest-%d", i))...)
		}
		ValidateApplicationsPureSDK(contexts)
		opts := make(map[string]bool)
		opts[scheduler.OptionsWaitForResourceLeakCleanup] = true

		for _, ctx := range contexts {
			TearDownContext(ctx, opts)
		}
		Step("delete credential used for cloudsnap", deleteCloudsnapCredential)
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts)
	})
})

// This test performs basic tests making sure Pure direct access volumes are running as expected
var _ = Describe("{PureVolumeCRUDWithPXCTL}", func() {
	var contexts []*scheduler.Context
	JustBeforeEach(func() {
		StartTorpedoTest("PureVolumeCRUDWithPXCTL", "Test pure volumes on applications, run CRUD using pxctl", nil, 0)
	})
	It("schedule pure volumes on applications, run CRUD, tear down", func() {
		Step("setup credential necessary for cloudsnap", createCloudsnapCredential)
		contexts = make([]*scheduler.Context, 0)

		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("purevolumestest-%d", i))...)
		}
		ValidateApplicationsPurePxctl(contexts)
		opts := make(map[string]bool)
		opts[scheduler.OptionsWaitForResourceLeakCleanup] = true

		for _, ctx := range contexts {
			TearDownContext(ctx, opts)
		}
		Step("delete credential used for cloudsnap", deleteCloudsnapCredential)
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts)
	})
})

// This test validates that, on an FACD installation, drives are located
// on the correct arrays that match their zone.
var _ = Describe("{PureFACDTopologyValidateDriveLocations}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("PureFACDTopologyValidateDriveLocations", "Test that FACD cloud drive volumes are located on proper FlashArrays", nil, 0)
	})
	It("installs with cloud drive volumes on the correct FlashArrays", func() {
		err := ValidatePureCloudDriveTopologies()
		Expect(err).NotTo(HaveOccurred(), "unexpected error validating Pure cloud drive topologies")
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})

// this tests brings up large number of pods on multiple namespaces and validate if there is not PANIC or nilpointer exceptions
var _ = Describe("{BringUpLargePodsVerifyNoPanic}", func() {
	/*
				https://portworx.atlassian.net/browse/PTX-18792
			    https://portworx.atlassian.net/browse/PTX-17723

				PWX :
				https://portworx.atlassian.net/browse/PWX-32190

				Bug Description :
					PX is hitting `panic: runtime error: invalid memory address or nil pointer dereference`
		when creating 250 FADA volumes

				1. Deploying nginx pods using two FADA volumes in 125 name-space simultaneously
				2. After that verify if any panic in the logs due to nil pointer deference.
	*/
	var testrailID = 0
	var runID int
	JustBeforeEach(func() {
		StartTorpedoTest("BringUpLargePodsVerifyNoPanic",
			"Validate no panics when creating more number of pods on "+
				"FADA/Generic Volumes while kvdb failover in progress", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})
	var contexts []*scheduler.Context

	stepLog := "Validate no panics when creating more number of pods on FADA/Generic " +
		"Volumes while kvdb failover in progress"
	It(stepLog, func() {
		/*
			NOTE : In order to verify https://portworx.atlassian.net/browse/PWX-32190 , please use nginx-fa-davol
				please use provisioner as portworx.PortworxCsi and storage-device to pure and application as nginx-fa-davol
			e.x : --app-list nginx-fa-davol --provisioner csi --storage-driver pure
		*/

		var wg sync.WaitGroup
		var terminate bool = false

		log.InfoD("Failover kvdb in parallel while volume creation in progress")
		go func() {
			defer GinkgoRecover()
			for {
				if terminate == true {
					break
				}
				// Wait for KVDB Members to be online
				log.FailOnError(WaitForKVDBMembers(), "failed waiting for KVDB members to be active")

				// Kill KVDB Master Node
				masterNode, err := GetKvdbMasterNode()
				log.FailOnError(err, "failed getting details of KVDB master node")

				log.InfoD("killing kvdb master node with Name [%v]", masterNode.Name)

				// Get KVDB Master PID
				pid, err := GetKvdbMasterPID(*masterNode)
				log.FailOnError(err, "failed getting PID of KVDB master node")

				log.InfoD("KVDB Master is [%v] and PID is [%v]", masterNode.Name, pid)

				// Kill kvdb master PID for regular intervals
				log.FailOnError(KillKvdbMemberUsingPid(*masterNode), "failed to kill KVDB Node")

				// Wait for some time after killing kvdb master Node
				time.Sleep(5 * time.Minute)
			}
		}()

		contexts = make([]*scheduler.Context, 0)

		// Apps list provided by user while triggering the test is considered to run the apps in parallel
		totalAppsRequested := Inst().AppList

		parallelThreads := 5
		scheduleCount := 1
		if len(totalAppsRequested) > 0 {
			for _, eachApp := range totalAppsRequested {
				if eachApp == "nginx-fa-davol" {
					if strings.ToLower(Inst().Provisioner) != fmt.Sprintf("%v", portworx.PortworxCsi) {
						log.FailOnError(fmt.Errorf("need csi provisioner to run the test , "+
							"please pass --provisioner csi "+
							"or -e provisioner=csi in the arguments"), "csi provisioner enabled?")
					}
					parallelThreads = 15
					scheduleCount = 20
				}
			}
		}

		// if app list is more than 5 we run 1 application in one point of time in parallel,
		// intention here is to run 20 applications in parallel, In any point of time max pod count doesn't exceed more than 300
		var appThreads int
		if len(totalAppsRequested) >= 5 {
			appThreads = 1
		} else {
			appThreads = parallelThreads / len(totalAppsRequested)
		}

		wg.Add(appThreads)
		scheduleAppParallel := func() {
			defer wg.Done()
			defer GinkgoRecover()
			id := uuid.New()
			nsName := fmt.Sprintf("%s", id.String()[:4])
			for i := 0; i < scheduleCount; i++ {
				contexts = append(contexts, ScheduleApplications(fmt.Sprintf(fmt.Sprintf("largenumberpods-%v-%d", nsName, i)))...)
			}
		}

		teardownContext := func() {
			opts := make(map[string]bool)
			opts[scheduler.OptionsWaitForResourceLeakCleanup] = true

			for _, ctx := range contexts {
				TearDownContext(ctx, opts)
			}
		}

		// Create apps in parallel
		for count := 0; count < appThreads; count++ {
			go scheduleAppParallel()
			time.Sleep(500 * time.Millisecond)
		}
		wg.Wait()

		allVolumes := []*volume.Volume{}
		for _, eachContext := range contexts {
			vols, err := Inst().S.GetVolumes(eachContext)
			if err != nil {
				log.Errorf("Failed to get app %s's volumes", eachContext.App.Key)
			}
			for _, eachVol := range vols {
				allVolumes = append(allVolumes, eachVol)
			}
		}

		// Funciton to validate nil pointer dereference errors
		validateNilPointerErrors := func() {
			terminate = true
			// we validate negative scenario here , function returns true if nil pointer exception is seen.
			errors := []string{}
			for _, eachNode := range node.GetStorageNodes() {
				status, output, Nodeerr := VerifyNilPointerDereferenceError(&eachNode)
				if status == true {
					log.Infof("nil pointer dereference error seen on the Node [%v]", eachNode.Name)
					log.Infof("error log [%v]", output)
					errors = append(errors, fmt.Sprintf("[%v]", eachNode.Name))
				} else if Nodeerr != nil && output == "" {
					// we just print error in case if found one
					log.InfoD(fmt.Sprintf("[%v]", Nodeerr))
				}
			}
			if len(errors) > 0 {
				log.FailOnError(fmt.Errorf("nil pointer dereference panic seen on nodes [%v]", errors),
					"nil pointer de-reference error?")
			}
		}

		// Delete all the applications
		defer teardownContext()

		// Check for nilPointer de-reference error on the nodes.
		defer validateNilPointerErrors()

		// Waiting for all pods to become ready and in running state
		waitForPodsRunning := func() (interface{}, bool, error) {
			for _, eachContext := range contexts {
				log.Infof("Verifying Context [%v]", eachContext.App.Key)
				err := Inst().S.WaitForRunning(eachContext, 5*time.Minute, 2*time.Second)
				if err != nil {
					return nil, true, err
				}
			}
			return nil, false, nil
		}
		_, err := task.DoRetryWithTimeout(waitForPodsRunning, 60*time.Minute, 10*time.Second)
		log.FailOnError(err, "Error checking pool rebalance")

		for _, eachVol := range allVolumes {
			log.InfoD("Validating Volume Status of Volume [%v]", eachVol.ID)
			status, err := IsVolumeStatusUP(eachVol)
			if err != nil {
				log.FailOnError(err, "error validating volume status")
			}
			dash.VerifyFatal(status == true, true, "is volume status up ?")
			terminate = true
		}

		terminate = true
		log.Info("all pods are up and in running state")
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

var _ = Describe("{CloneVolAndValidate}", func() {

	/*
		Testrail corresponds to:
		https://portworx.testrail.net/index.php?/tests/view/72639348
		https://portworx.testrail.net/index.php?/tests/view/72657575
	*/

	// Backend represents the cloud storage provider or platform where the volumes are provisioned
	type Backend string

	const (
		BackendPure    Backend = "PURE"
		BackendVSphere Backend = "VSPHERE"
		BackendUnknown Backend = "UNKNOWN"
	)

	// VolumeType represents the type of provisioned storage volume
	type VolumeType string

	const (
		VolumeFADA    VolumeType = "FADA"
		VolumeFBDA    VolumeType = "FBDA"
		VolumeFACD    VolumeType = "FACD"
		VolumeVsCD    VolumeType = "VsCD"
		VolumeUnknown VolumeType = "UNKNOWN"
	)

	var (
		contexts   = make([]*scheduler.Context, 0)
		namespaces = make([]string, 0)
		backend    = BackendUnknown
		volumeMap  = make(map[VolumeType][]*api.Volume)
	)

	JustBeforeEach(func() {
		StartTorpedoTest("CloneVolAndValidate", "Validate clone volumes on FADA, FBDA, and FACD", nil, 72657582)
	})
	It("Validate clone volumes on FADA, FBDA, and FACD", func() {
		// getPureVolumeType determines the type of the volume based on the proxy spec
		getPureVolumeType := func(vol *volume.Volume) (VolumeType, error) {
			proxySpec, err := Inst().V.GetProxySpecForAVolume(vol)
			if err != nil {
				return "", fmt.Errorf("failed to get proxy spec for the volume [%s/%s]. Err: [%v]", vol.Namespace, vol.Name, err)
			}
			if proxySpec != nil {
				switch proxySpec.ProxyProtocol {
				case api.ProxyProtocol_PROXY_PROTOCOL_PURE_FILE:
					return VolumeFBDA, nil
				case api.ProxyProtocol_PROXY_PROTOCOL_PURE_BLOCK:
					return VolumeFADA, nil
				default:
					return VolumeUnknown, nil
				}
			} else {
				return VolumeFACD, nil
			}
		}
		stepLog := "Schedule applications"
		Step(stepLog, func() {
			log.InfoD("Scheduling applications")
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				taskName := fmt.Sprintf("pure-test-%d", i)
				for _, ctx := range ScheduleApplications(taskName) {
					ctx.ReadinessTimeout = appReadinessTimeout
					contexts = append(contexts, ctx)
					namespaces = append(namespaces, GetAppNamespace(ctx, taskName))
				}
			}
		})
		stepLog = "Validate applications"
		Step(stepLog, func() {
			log.InfoD("Validating applications")
			ValidateApplications(contexts)
		})
		Step("Identify backend and categorize volumes", func() {
			log.InfoD("Identifying backend")
			volDriverNamespace, err := Inst().V.GetVolumeDriverNamespace()
			log.FailOnError(err, "failed to get volume driver [%s] namespace", Inst().V.String())
			secretList, err := core.Instance().ListSecret(volDriverNamespace, v1.ListOptions{})
			log.FailOnError(err, "failed to get secret list from namespace [%s]", volDriverNamespace)
			for _, secret := range secretList.Items {
				switch secret.Name {
				case PX_PURE_SECRET_NAME:
					backend = BackendPure
					break
				case PX_VSPHERE_SCERET_NAME:
					backend = BackendVSphere
					break
				}
			}
			log.InfoD("Backend: %v", backend)
			log.InfoD("Categorizing volumes")
			for _, ctx := range contexts {
				volumes, err := Inst().S.GetVolumes(ctx)
				log.FailOnError(err, "failed to get volumes for app [%s/%s]", ctx.App.NameSpace, ctx.App.Key)
				dash.VerifyFatal(len(volumes) > 0, true, "Verifying if volumes exist for resizing")
				for _, vol := range volumes {
					apiVol, err := Inst().V.InspectVolume(vol.ID)
					log.FailOnError(err, "failed to inspect volume [%s/%s]", vol.Name, vol.ID)
					switch backend {
					case BackendPure:
						volType, err := getPureVolumeType(vol)
						log.FailOnError(err, "failed to get pure volume type for volume [%+v]", vol)
						volumeMap[volType] = append(volumeMap[volType], apiVol)
					case BackendVSphere:
						volumeMap[VolumeVsCD] = append(volumeMap[VolumeVsCD], apiVol)
					default:
						volumeMap[VolumeUnknown] = append(volumeMap[VolumeUnknown], apiVol)
					}
				}
			}
		})
		stepLog = "Clone FADA,FBDA and FACD volumes and validate"
		Step(stepLog, func() {
			for key, volumes := range volumeMap {
				log.InfoD("cloning %v volumes", key)
				for i := 0; i < len(volumes); i++ {
					log.FailOnError(err, "Failed to inspect volume %v", volumes[i].Id)
					clonedVolID, err := Inst().V.CloneVolume(volumes[i].Id)
					log.FailOnError(err, "Failed to clone %v volume with volume id %v", key, volumes[i].Id)
					mountPath, err := Inst().V.AttachVolume(clonedVolID)
					log.FailOnError(err, "Failed to attach volume")
					log.InfoD("Successfully attached volume on path:%v", mountPath)
					params := make(map[string]string)
					if Inst().ConfigMap != "" {
						params["auth-token"], err = Inst().S.GetTokenFromConfigMap(Inst().ConfigMap)
					}
					err = Inst().V.ValidateCreateVolume(clonedVolID, params)
					log.FailOnError(err, "Failed to validate clone")
					log.Infof("successfully cloned and validated")
				}
			}
		})
	})

	JustAfterEach(func() {
		opts := make(map[string]bool)
		opts[SkipClusterScopedObjects] = true
		log.InfoD("Destroying applications")
		DestroyApps(contexts, opts)
		defer EndTorpedoTest()
	})
})
