package tests

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	"github.com/portworx/sched-ops/k8s/storage"
	"github.com/portworx/torpedo/drivers/scheduler/k8s"
	"github.com/portworx/torpedo/drivers/vcluster"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
	v1 "k8s.io/api/core/v1"
	storageApi "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"path/filepath"
	"time"
)

var _ = Describe("NewCreatevCluster", func() {
	vc := &vcluster.VCluster{}
	var scName string
	var pvcName string
	var appNS string
	fioOptions := vcluster.FIOOptions{
		Name:      "mytest",
		IOEngine:  "libaio",
		RW:        "randwrite",
		BS:        "4k",
		NumJobs:   1,
		Size:      "500m",
		TimeBased: true,
		Runtime:   "600s",
		Filename:  "/data/fiotest",
		EndFsync:  1,
	}
	JustBeforeEach(func() {
		StartTorpedoTest("CreateAndRunFioOnVcluster", "Create, Connect and run FIO Application on Vcluster", nil, 0)
		vc = vcluster.NewVCluster("my-vcluster1")
		err := vc.CreateAndWaitVCluster()
		log.FailOnError(err, "Failed to create VCluster")
	})
	It("Create FIO app on VCluster and run it for 10 minutes", func() {
		// Create Storage Class on Host Cluster
		scName = fmt.Sprintf("fio-app-sc-%v", time.Now().Unix())
		err = CreateStorageClass(scName)
		log.FailOnError(err, "Error creating Storageclass")
		log.Infof("Successfully created StorageClass with name: %v", scName)
		// Create PVC on VCluster
		appNS = scName + "-ns"
		pvcName, err = vc.CreatePVC(scName, appNS)
		log.FailOnError(err, fmt.Sprintf("Error creating PVC with Storageclass name %v", scName))
		log.Infof("Successfully created PVC with name: %v", pvcName)
		// Create FIO Deployment on VCluster using the above PVC
		err = vc.CreateFIODeployment(pvcName, appNS, fioOptions)
		log.FailOnError(err, "Error in creating FIO Application")
		log.Infof("Successfully ran FIO on Vcluster")
	})
	JustAfterEach(func() {
		EndTorpedoTest()
		// VCluster, StorageClass and Namespace cleanup
		//err := vc.VClusterCleanup(scName)
		//if err != nil {
		//	log.Errorf("Problem in Cleanup: %v", err)
		//} else {
		//	log.Infof("Cleanup successfully done.")
		//}
	})
})

// CreateStorageClass method creates a storageclass using host's k8s clientset on host cluster
func CreateStorageClass(scName string) error {
	params := make(map[string]string)
	params["repl"] = "2"
	params["priority_io"] = "high"
	params["io_profile"] = "auto"
	v1obj := metav1.ObjectMeta{
		Name: scName,
	}
	reclaimPolicyDelete := v1.PersistentVolumeReclaimDelete
	bindMode := storageApi.VolumeBindingImmediate
	scObj := storageApi.StorageClass{
		ObjectMeta:        v1obj,
		Provisioner:       k8s.CsiProvisioner,
		Parameters:        params,
		ReclaimPolicy:     &reclaimPolicyDelete,
		VolumeBindingMode: &bindMode,
	}
	k8sStorage := storage.Instance()
	if _, err := k8sStorage.CreateStorageClass(&scObj); err != nil {
		return err
	}
	return nil
}

var _ = Describe("CreatevCluster", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("VclusterOperations", "Create  Vcluster", nil, 0)
	})
	var vclusterNames = []string{"my-vcluster2"}
	It("Create and connect to vclusters and run a sample method", func() {
		steplog := "Create vClusters"
		log.InfoD(steplog)
		Step(steplog, func() {
			for _, name := range vclusterNames {
				currentDir, err := os.Getwd()
				log.FailOnError(err, "Could not get absolute path to current Dir")
				vClusterPath := filepath.Join(currentDir, "..", "deployments", "customconfigs", "vcluster.yaml")
				absPath, err := filepath.Abs(vClusterPath)
				log.FailOnError(err, "Could not get absolute path to vcluster.yaml")
				err = vcluster.CreateVCluster(name, absPath)
				log.FailOnError(err, "Failed to create vCluster")
			}
		})
		steplog = "Wait for all vClusters to come up in Running State"
		log.InfoD(steplog)
		Step(steplog, func() {
			for _, name := range vclusterNames {
				err := vcluster.WaitForVClusterRunning(name, 10*time.Minute)
				log.FailOnError(err, "Vcluster did not come up in time")
				err = vcluster.GetVClusterSecret("vc-my-vcluster2", "vcluster-my-vcluster2")
				log.FailOnError(err, "error while getting the vcluster secret")
			}
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		//for _, name := range vclusterNames {
		//	vcluster.DeleteVCluster(name)
		//}
	})
})
