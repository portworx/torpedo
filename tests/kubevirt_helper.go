package tests

import (
	context1 "context"
	"encoding/hex"
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/api/resource"
	cdiv1 "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"

	"github.com/libopenstorage/openstorage/api"
	. "github.com/onsi/ginkgo/v2"
	"github.com/portworx/sched-ops/k8s/core"
	"github.com/portworx/sched-ops/k8s/kubevirt"
	kubevirtdy "github.com/portworx/sched-ops/k8s/kubevirt-dynamic"
	"github.com/portworx/sched-ops/task"
	"github.com/pure-px/torpedo/drivers/node"
	"github.com/pure-px/torpedo/drivers/scheduler"
	"github.com/pure-px/torpedo/drivers/volume"
	"github.com/pure-px/torpedo/pkg/log"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubevirtv1 "kubevirt.io/api/core/v1"
	v1alpha1 "kubevirt.io/api/migrations/v1alpha1"
)

const (
	mountTypeBind = "bind"
	mountTypeNFS  = "nfs"

	kubevirtTemplates                     = "kubevirt-templates"
	kubevirtTemplateNamespace             = "openshift-virtualization-os-images"
	kubevirtCDIStorageConditionAnnotation = "cdi.kubevirt.io/storage.condition.running.reason"
	kubevirtCDIStoragePodPhaseAnnotation  = "cdi.kubevirt.io/storage.pod.phase"
	sshUserName                           = "root"
	postCopyMigrationPolicy               = "allow-post-copy-migration"
)

var (
	defaultVmMountCheckTimeout       = 15 * time.Minute
	defaultVmMountCheckRetryInterval = 30 * time.Second
	k8sKubevirt                      = kubevirt.Instance()
	importerPodCompletionTimeout     = 30 * time.Minute
	importerPodRetryInterval         = 20 * time.Second
	defaultMigrationTimeout          = 30 * time.Minute
	defaultMigrationRetryInterval    = 30 * time.Second
	// RebootFlag sends the signal to start the reboot operation of node or px during hot-plug disk
	RebootFlag chan struct{}
	/*
		SignalSent flag is set to true once the signal for reboot is sent during hot-plug disk so that
		the signal is sent only once even if there are multiple DVs to be added (for loop).
		Make sure to set this to false in JustAfterEach block of test once test completes
	*/
	SignalSent bool
)

// AddDisksToKubevirtVM is a function which takes number of disks to add and adds them to the kubevirt VMs passed (Please provide size in Gi)
func AddDisksToKubevirtVM(virtualMachines []*scheduler.Context, numberOfDisks int, size string) (bool, error) {
	var (
		newDiskCount     int
		initialDiskCount int
	)
	log.InfoD("create config map")
	CreateConfigMap()

	for _, appCtx := range virtualMachines {
		vms, err := GetAllVMsFromScheduledContexts([]*scheduler.Context{appCtx})
		if err != nil {
			return false, err
		}
		for _, v := range vms {
			initialDiskCount, err = GetNumberOfDrivesInVM(v)
			if err != nil {
				return false, fmt.Errorf("failed to get initial number of disks in VM [%s]: %v", v.Name, err)
			}
			log.InfoD("Initial Number of disks in VM [%s] in namespace [%s] is [%d]", v.Name, v.Namespace, initialDiskCount)

			// Before we add the pvc we need to get storage class of the pvc
			storageClass, err := GetStorageClassOfVmPVC(appCtx)
			if err != nil {
				return false, err
			}
			log.InfoD("Storage class of PVC attached to VM [%s] in namespace [%s] is [%s]", v.Name, v.Namespace, storageClass)

			// Add the disks to the VM
			pvcs, err := CreatePVCsForVM(v, 1, storageClass, size)

			if err != nil {
				return false, err
			}

			specListInterfaces := make([]interface{}, len(pvcs))
			for i, pvc := range pvcs {
				// Converting each PVC to interface for appending to SpecList
				specListInterfaces[i] = pvc
			}
			appCtx.App.SpecList = append(appCtx.App.SpecList, specListInterfaces...)

			err = AddPVCsToVirtualMachine(v, pvcs)
			if err != nil {
				return false, err
			}

			err = RestartKubevirtVM(v.Name, v.Namespace, true)
			if err != nil {
				return false, err
			}
			log.InfoD("Sleep for 30 seconds for vm to come up")
			time.Sleep(30 * time.Second)

			//After adding the pvcs check the number of disks in the VM
			vms, err := GetAllVMsFromScheduledContexts([]*scheduler.Context{appCtx})
			if err != nil {
				return false, err
			}
			for _, v := range vms {
				t := func() (interface{}, bool, error) {
					newDiskCount, err = GetNumberOfDrivesInVM(v)
					if err != nil {
						return nil, true, err
					}
					if newDiskCount < initialDiskCount+numberOfDisks {
						return nil, true, fmt.Errorf(
							"Expected at least [%d] disks, found only [%d]",
							initialDiskCount+numberOfDisks, newDiskCount)
					}
					return newDiskCount, false, nil
				}
				d, err := task.DoRetryWithTimeout(t, 10*time.Minute, 30*time.Second)
				if err != nil {
					return false, err
				}
				if newDiskCount == initialDiskCount {
					return false, fmt.Errorf("number of disks in VM [%s] in namespace [%s] is same as before adding disks", v.Name, v.Namespace)
				}
				newDiskCount = d.(int)
				log.InfoD("Final Number of disks in VM [%s] in namespace [%s] is [%d]", v.Name, v.Namespace, newDiskCount)
			}
		}
	}
	log.InfoD("Number of disks after cold add disk is [%d] and total number of disks before cold add disk [%v]", newDiskCount, initialDiskCount)
	return true, nil
}

// RunCmdInVirtLauncherPod runs a command in the virt-launcher pod of the VM
func RunCmdInVirtLauncherPod(virtualMachineCtx *scheduler.Context, cmd []string) (string, error) {
	vols, err := Inst().S.GetVolumes(virtualMachineCtx)
	if err != nil {
		return "", err
	}

	// Check if the pod is in the 'Running' state before executing the command
	_, err = task.DoRetryWithTimeout(func() (interface{}, bool, error) {
		vmPod, err := GetVirtLauncherPodForVM(virtualMachineCtx, vols[0])
		if err != nil {
			return nil, true, fmt.Errorf("failed to get pod: %s", err)
		}
		if vmPod.Status.Phase != corev1.PodRunning {
			return nil, true, fmt.Errorf("pod %s is not in running state, current state: %s", vmPod.Name, vmPod.Status.Phase)
		}
		return nil, false, nil
	}, 10*time.Minute, 10*time.Second)

	vmPod, err := GetVirtLauncherPodForVM(virtualMachineCtx, vols[0])
	if err != nil {
		return "", err
	}

	output, err := core.Instance().RunCommandInPod(cmd, vmPod.Name, "compute", vmPod.Namespace)
	if err != nil {
		return "", err
	}
	log.InfoD("Output of command: %s", output)
	return core.Instance().RunCommandInPod(cmd, vmPod.Name, "compute", vmPod.Namespace)

}

// GetNumberOfDisksInVMViaVirtLauncherPod gets the number of disks in the VM via the virt-launcher pod
func GetNumberOfDisksInVMViaVirtLauncherPod(virtualMachineCtx *scheduler.Context) (int, error) {
	cmd := []string{"lsblk"}

	t := func() (interface{}, bool, error) {
		output, err := RunCmdInVirtLauncherPod(virtualMachineCtx, cmd)
		log.InfoD("Output of command: %s", output)
		if err != nil {
			return 0, false, err
		}
		// Splitting the output into lines
		lines := strings.Split(output, "\n")

		// Count of lines containing "/run/kubevirt-private/vmi-disks/" and "pxd"
		numberOfDisks := 0

		// Loop through each line
		for _, line := range lines {
			// Check if the line contains "/run/kubevirt-private/vmi-disks/" and "pxd"
			fmt.Println(line)
			if strings.Contains(line, "/run/kubevirt-private/vmi-disks/") && strings.Contains(line, "pxd") {
				// Increment count if both conditions are met
				numberOfDisks++
			}
		}

		if err != nil {
			return 0, false, err
		}
		if numberOfDisks == 0 {
			return 0, true, nil
		}
		return numberOfDisks, numberOfDisks == 0, nil
	}
	d, err := task.DoRetryWithTimeout(t, 10*time.Minute, 30*time.Second)
	if err != nil {
		return 0, err
	}
	return d.(int), nil
}

// GetStorageClassOfVmPVC returns the storage class of pvc attached to the VM
func GetStorageClassOfVmPVC(vm *scheduler.Context) (string, error) {
	// Get the PVC object from the VM
	nameSpace := vm.App.NameSpace
	pvcs, err := core.Instance().GetPersistentVolumeClaims(nameSpace, nil)
	if err != nil {
		return "", err
	}

	// Get the PVCs attached to the VM
	for _, pvc := range pvcs.Items {
		ScName, err := core.Instance().GetStorageClassForPVC(&pvc)
		if err != nil {
			return "", err
		}
		return ScName.Name, nil
	}
	return "", fmt.Errorf("failed to get storage class of PVC attached to VM [%s] in namespace [%s]", vm.App.Key, vm.App.NameSpace)
}

// StartAndWaitForVMIMigration starts the VM migration and waits for the VM to be in running state in the new node
func StartAndWaitForVMIMigration(virtualMachineCtx *scheduler.Context, ctx context1.Context) error {
	log.InfoD("Initiating VM migration for VM [%s] in namespace [%s]", virtualMachineCtx.App.Key, virtualMachineCtx.App.NameSpace)
	vms, err := GetAllVMsFromScheduledContexts([]*scheduler.Context{virtualMachineCtx})
	if err != nil {
		return err
	}
	if len(vms) == 0 {
		return fmt.Errorf("No VMs found for VM [%s] in namespace [%s]", virtualMachineCtx.App.Key, virtualMachineCtx.App.NameSpace)
	}
	log.Infof("Total number of VMs [%v] in namespace [%s]", len(vms), virtualMachineCtx.App.NameSpace)

	for _, vm := range vms {
		vmiNamespace := vm.Namespace
		vmiName := vm.Name

		//Get the node where the vm is scheduled before the migration
		nodeName, err := GetNodeOfVM(vm)
		if err != nil {
			return err
		}
		log.Infof("VM [%s] in namespace [%s] is scheduled on node [%s]", vmiName, vmiNamespace, nodeName)

		// Start the VM migration
		migration, err := kubevirtdy.Instance().CreateVirtualMachineInstanceMigration(ctx, vmiNamespace, vmiName)
		if err != nil {
			return err
		}
		log.Infof("VM migration created for VM [%s] in namespace [%s]", vmiName, vmiNamespace)

		// get volumes from app context
		vols, err := Inst().S.GetVolumes(virtualMachineCtx)
		if err != nil {
			return err
		}

		t := func() (interface{}, bool, error) {
			var migr *kubevirtdy.VirtualMachineInstanceMigration
			migr, err = kubevirtdy.Instance().GetVirtualMachineInstanceMigration(ctx, vmiNamespace, migration.Name)
			if err != nil {
				log.InfoD("Error: %v", err)
				return "", false, fmt.Errorf("failed to get migration for VM [%s] in namespace [%s]", vmiName, vmiNamespace)
			}
			if !(migr.Phase == "Succeeded") {
				return "", true, fmt.Errorf("waiting for migration to complete for VM [%s] in namespace [%s]", vmiName, vmiNamespace)
			}

			// wait until there is only one pod in the running state
			//TODO https://purestorage.atlassian.net/browse/PTX-23166 - This is a temporary fix to get the pod of the VM
			testPod, err := GetVirtLauncherPodForVM(virtualMachineCtx, vols[0])
			if err != nil {
				return "", true, err
			}

			//Get the node where the vm is scheduled after the migration
			nodeNameAfterMigration := testPod.Spec.NodeName

			if nodeName == nodeNameAfterMigration {
				return "", false, fmt.Errorf("VM pod live migrated [%s] in namespace [%s] but is still on the same node [%s]", testPod.Name, testPod.Namespace, nodeName)
			}
			log.InfoD("VM pod live migrated to node: [%s]", nodeNameAfterMigration)
			return "", false, nil
		}
		_, err = task.DoRetryWithTimeout(t, defaultMigrationTimeout, defaultMigrationRetryInterval)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetVirtLauncherPodForVM returns the virt-launcher pod for the VM
func GetVirtLauncherPodForVM(virtualMachineCtx *scheduler.Context, vol *volume.Volume) (*corev1.Pod, error) {
	pods, err := core.Instance().GetPodsUsingPV(vol.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get pods for volume %s of context %s: %w", vol.ID, virtualMachineCtx.App.Key, err)
	}

	var found corev1.Pod
	for _, pod := range pods {
		if pod.Labels["kubevirt.io"] == "virt-launcher" && pod.Status.Phase == corev1.PodRunning {
			if found.Name != "" {
				// there should be only one VM pod in the running state (otherwise live migration is in progress)
				return nil, fmt.Errorf("more than 1 KubeVirt pods (%s, %s) are in running state for volume %s",
					found.Name, pod.Name, vol.ID)
			}
			found = pod
		}
	}
	if found.Name == "" {
		return nil, fmt.Errorf("failed to find a running pod for volume %s", vol.ID)
	}
	return &found, nil
}

// IsVMBindMounted checks if the volumes are bind mounted to the VM
func IsVMBindMounted(virtualMachineCtx *scheduler.Context, wait bool) (bool, error) {
	vols, err := Inst().S.GetVolumes(virtualMachineCtx)
	if err != nil {
		return false, err
	}
	log.InfoD("Length of volumes: %d", len(vols))

	// Get the node where pod is scheduled
	vms, err := GetAllVMsFromScheduledContexts([]*scheduler.Context{virtualMachineCtx})
	if err != nil {
		return false, err
	}
	// TODO Add support for multiple vm per spec/Context : https://purestorage.atlassian.net/browse/PTX-23167
	if len(vms) != 1 {
		return false, fmt.Errorf("expected 1 VM for VM [%s] in namespace [%s] but got [%d]", virtualMachineCtx.App.Key, virtualMachineCtx.App.NameSpace, len(vms))
	}

	vm := vms[0]
	vmNodeName, err := GetNodeOfVM(vm)
	if err != nil {
		return false, err
	}
	log.Infof("VM [%s] is deployed on node [%s]", virtualMachineCtx.App.Key, vmNodeName)

	// Keep a track of replicaset and consider this as the source of truth
	volInspect, err := Inst().V.InspectVolume(vols[0].ID)
	if err != nil {
		return false, fmt.Errorf("failed to inspect volume [%s]: %w", vols[0].ID, err)
	}
	globalReplicSet := volInspect.ReplicaSets
	log.InfoD("Length of replicaset: %d", len(globalReplicSet))

	// The criteria to call the bind mount successful is to check if the replicaset of all volumes should be same and should be locally attached to the node
	// check if the replicaset values is same as the global replicaset
	// Here we are considering globalreplicaset to be source of truth for comparison
	var vmPod *corev1.Pod
	for _, vol := range vols {
		if vmPod == nil {
			vmPod, _ = GetVirtLauncherPodForVM(virtualMachineCtx, vol)
		}
		if val, exists := vol.Labels["pure_direct_access"]; exists && val == "fada" {
			log.Infof("Skipping Co-location check as it's a direct attached volume")
			continue
		} else {
			err := IsVolumeBindMounted(virtualMachineCtx, vmNodeName, vol, wait, vmPod)
			if err != nil {
				return false, err
			}
			err = AreVolumeReplicasCollocated(vol, globalReplicSet)
			if err != nil {
				return false, err
			}
		}
	}
	log.Infof("Successfully verified bind mount for VM [%s] in namespace [%s]", virtualMachineCtx.App.Key, virtualMachineCtx.App.NameSpace)
	return true, nil
}

// GetNodeOfVM returns nodename on which VM is running
func GetNodeOfVM(virtualMachineCtx kubevirtv1.VirtualMachine) (string, error) {
	vmi, err := kubevirt.Instance().GetVirtualMachineInstance(context1.TODO(), virtualMachineCtx.Name, virtualMachineCtx.Namespace)
	if err != nil {
		return "", err
	}
	log.InfoD("NodeName: %s", vmi.Status.NodeName)
	return vmi.Status.NodeName, nil
}

// ReplicaSetsMatch verifies if the replicaset nodes are present in global replicaset
func ReplicaSetsMatch(replicaset []*api.ReplicaSet, globalReplicSet []*api.ReplicaSet) error {
	// Considering aggregation is 1
	if len(replicaset[0].Nodes) != len(globalReplicSet[0].Nodes) {
		return fmt.Errorf("the number of nodes in the replicaset is not same as the global replicaset")
	}

	replicasetNodes := make(map[string]bool)

	for _, rs := range replicaset {
		for _, rsNode := range rs.Nodes {
			replicasetNodes[rsNode] = true
		}
	}

	for _, grs := range globalReplicSet {
		for _, grsNode := range grs.Nodes {
			if _, ok := replicasetNodes[grsNode]; !ok {
				return fmt.Errorf("replicaset mismatch node not found in global replicaset")
			}
		}
	}
	log.Infof("Replicaset matches with global replicaset")
	return nil
}

// getVMDiskMountType Gets mount type (nfs or bind) of the VM disk
func getVMDiskMountType(pod *corev1.Pod, vmDisk *volume.Volume, diskName string) (string, error) {
	podNamespacedName := pod.Namespace + "/" + pod.Name
	log.Infof("Checking the mount type of %s in pod %s", vmDisk, podNamespacedName)

	// Sample output if the volume is bind-mounted: (vmDisk.diskName is "rootdisk" in this example)
	// $ kubectl exec -it virt-launcher-fedora-communist-toucan-jfw7n -- mount
	// ...
	// /dev/pxd/pxd365793461222635857 on /run/kubevirt-private/vmi-disks/rootdisk type ext4 (rw,relatime,seclabel,discard)
	// ...
	volInspect, err := Inst().V.InspectVolume(vmDisk.ID)
	if err != nil {
		return "", fmt.Errorf("failed to inspect volume %s: %v", vmDisk.ID, err)
	}

	bindMountRE := regexp.MustCompile(fmt.Sprintf("/dev/pxd/pxd%s on .*%s type (ext4|xfs)",
		volInspect.Id, diskName))

	// Sample output if the volume is nfs-mounted: (vmDisk.diskName is "rootdisk" in this example)
	// $ kubectl exec -it virt-launcher-fedora-communist-toucan-bqcrp -- mount
	// ...
	// 172.30.194.11:/var/lib/osd/pxns/365793461222635857 on /run/kubevirt-private/vmi-disks/rootdisk type nfs (...)
	// ...
	nfsMountRE := regexp.MustCompile(fmt.Sprintf(":/var/lib/osd/pxns/%s on .*%s type nfs",
		volInspect.Id, diskName))

	cmd := []string{"mount"}
	output, err := core.Instance().RunCommandInPod(cmd, pod.Name, "compute", pod.Namespace)
	if err != nil {
		return "", fmt.Errorf("failed to run command %v inside the pod %s, error: %v", cmd, podNamespacedName, err)
	}
	var foundBindMount, foundNFSMount bool
	for _, line := range strings.Split(output, "\n") {
		if bindMountRE.MatchString(line) {
			if foundBindMount || foundNFSMount {
				return "", fmt.Errorf("multiple mounts found for %s: %s", vmDisk, output)
			}
			foundBindMount = true
			log.Infof("Found %s bind mounted for VM pod %s: %s", vmDisk, podNamespacedName, line)
		}

		if nfsMountRE.MatchString(line) {
			if foundBindMount || foundNFSMount {
				return "", fmt.Errorf("multiple mounts found for %s: %s", vmDisk, output)
			}
			foundNFSMount = true
			log.Infof("Found %s nfs mounted for VM pod %s: %s", vmDisk, podNamespacedName, line)
		}
	}
	if !foundBindMount && !foundNFSMount {
		return "", fmt.Errorf("no mount for %s in pod %s: %s", vmDisk, podNamespacedName, output)
	}
	if foundBindMount {
		return mountTypeBind, nil
	}
	return mountTypeNFS, nil
}

// IsVolumeBindMounted verifies if the volume is bind mounted on the VM pod or not
func IsVolumeBindMounted(virtualMachineCtx *scheduler.Context, vmNodeName string, vol *volume.Volume, wait bool, vmPod *corev1.Pod) error {
	volInspect, err := Inst().V.InspectVolume(vol.ID)
	if err != nil {
		return fmt.Errorf("failed to inspect volume [%s]: %w", vol.ID, err)
	}
	nodeIpAttachedOn := volInspect.AttachedOn
	nodeNameAttachedOn, err := node.GetNodeByIP(nodeIpAttachedOn)
	if err != nil {
		return fmt.Errorf("failed to get node name by IP [%s]: %w", nodeIpAttachedOn, err)
	}
	log.Infof("Volume [%s] is attached on node [%s]", vol.ID, nodeNameAttachedOn.Name)
	if nodeNameAttachedOn.Name != vmNodeName {
		return fmt.Errorf("volume [%s] is attached on node [%s] instead of node [%s]", vol.ID, nodeNameAttachedOn.Name, vmNodeName)
	}

	isBindMounted := false
	t := func() (interface{}, bool, error) {
		if vmPod == nil {
			vmPod, err = GetVirtLauncherPodForVM(virtualMachineCtx, vol)
			if err != nil {
				// this is expected while the live migration is running since there will be 2 VM pods
				log.Infof("Could not get VM pod for %s for context %s: %v", vol.Name, virtualMachineCtx.App.Key, err)
				return false, false, nil
			}
		}
		log.Infof("Verifying bind mount for %s", vol)
		diskName := ""
		for _, vmVol := range vmPod.Spec.Volumes {
			pvcName := volInspect.Locator.VolumeLabels["pvc"]
			if vmVol.PersistentVolumeClaim != nil && vmVol.PersistentVolumeClaim.ClaimName == pvcName {
				diskName = vmVol.Name
				break
			}
		}
		mountType, err := getVMDiskMountType(vmPod, vol, diskName)
		if err != nil {
			log.Warnf("Failed to get mount type of %s for context %s: %v", vol, virtualMachineCtx.App.Key, err)
			return false, false, nil
		}
		log.Infof("Mount type of %s for context %s: %s", vol.Name, virtualMachineCtx.App.Key, mountType)
		if mountType != mountTypeBind {
			if wait {
				log.Warnf("Waiting for %s for context %s to switch to bind-mount from %q",
					vol.Name, virtualMachineCtx.App.Key, mountType)
			}
			return false, false, nil
		}
		isBindMounted = true
		return true, false, nil
	}
	if !wait {
		// initial check is done only once
		_, _, err := t()
		if err != nil {
			return err
		}
		if !isBindMounted {
			return fmt.Errorf("volume [%s] is not bind mounted", vol.ID)
		}
	} else {
		_, err = task.DoRetryWithTimeout(t, defaultVmMountCheckTimeout, defaultVmMountCheckRetryInterval)
		if err != nil {
			return err
		}
	}
	return nil
}

// AreVolumeReplicasCollocated verifies if the volume replicas are collocated on the same set of nodes
func AreVolumeReplicasCollocated(vol *volume.Volume, globalReplicSet []*api.ReplicaSet) error {
	// Check if volumes have replicas on the same set of nodes
	volInspect, err := Inst().V.InspectVolume(vol.ID)
	if err != nil {
		return fmt.Errorf("failed to inspect volume [%s]: %w", vol.ID, err)
	}

	replicaset := volInspect.ReplicaSets

	// check if the replicaset size is same as the global replicaset size
	if len(replicaset) != len(globalReplicSet) {
		return fmt.Errorf("replicaset count mismatch for volume [%s] and for volume [%s]", vol.ID, volInspect.Id)
	}
	err = ReplicaSetsMatch(replicaset, globalReplicSet)
	if err != nil {
		return fmt.Errorf("replicaset mismatch for volume [%s] and volume [%s]", vol.ID, volInspect.Id)
	}
	return nil
}

// CreateConfigMap creates configmap for vm creds
func CreateConfigMap() error {
	// Check if a config map named kubevirt-creds exist
	configMap, err := k8sCore.GetConfigMap("kubevirt-creds", "default")
	log.Infof("configMap: %v", configMap)
	log.Infof("err: %v", err)
	if err != nil {
		if errors.IsNotFound(err) {
			// ConfigMap does not exist, so create it
			configMap = &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "kubevirt-creds",
					Namespace: "default",
				},
				Data: map[string]string{
					"fio-vm-multi-disk": "ubuntu",
				},
			}
			_, err = k8sCore.CreateConfigMap(configMap)
			if err != nil {
				log.Infof("Failed to create config map kubevirt-creds: %v", err)
				return err
			}
			log.Infof("Created config map kubevirt-creds")
		} else {
			// An unexpected error occurred when retrieving the ConfigMap
			log.Infof("Failed to retrieve config map kubevirt-creds: %v", err)
			return err
		}
	} else {
		log.Infof("Config map kubevirt-creds already exists")
	}
	return nil
}

// WriteFilesAndStoreMD5InVM write few files in the VM and calculates the md5sum
func WriteFilesAndStoreMD5InVM(virtualMachines []*scheduler.Context, namespace string, fileCount int, maxFileSize int) error {
	log.Infof("Creating %d files and storing their MD5 checksums in namespace %s", fileCount, namespace)
	createFilesCmd := fmt.Sprintf("mkdir -p ~/testfiles && cd ~/testfiles && "+
		"rm -f ~/file_checksums.md5 && for i in $(seq 1 %d); do "+
		"head -c $(($RANDOM %% %d)) </dev/urandom >file$i; md5sum file$i >> ~/file_checksums.md5; done", fileCount, maxFileSize)
	for _, appCtx := range virtualMachines {
		vms, err := GetAllVMsFromScheduledContexts([]*scheduler.Context{appCtx})
		if err != nil {
			return err
		}
		for _, v := range vms {
			_, err := RunCmdInVM(v, createFilesCmd, context1.TODO())
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// ValidateFileIntegrityInVM validates the md5sum of files that we wrote in the VM
func ValidateFileIntegrityInVM(virtualMachines []*scheduler.Context, namespace string) error {
	log.Infof("Validating file integrity in namespace %s", namespace)
	validateFilesCmd := "cd ~/testfiles && md5sum -c ~/file_checksums.md5"
	for _, appCtx := range virtualMachines {
		vms, err := GetAllVMsFromScheduledContexts([]*scheduler.Context{appCtx})
		if err != nil {
			return err
		}
		for _, v := range vms {
			_, err := RunCmdInVM(v, validateFilesCmd, context1.TODO())
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func CreateSSHPod() error {
	_, err := k8sCore.GetPodByName(sshPodName, "default")
	if err == nil {
		log.Infof("Ssh pod already running")
		return nil
	}
	err = initSSHPod("default")
	if err != nil {
		return err
	}
	return nil
}

// ListEvents lists all events in a namespace in logs.
func ListEvents(namespace string) error {
	eventList, err := k8sCore.ListEvents(namespace, metav1.ListOptions{})
	if err != nil {
		log.Infof("Failed to list events in namespace %s: %v", namespace, err)
		return err
	}
	log.Infof("Events in namespace %s:", namespace)
	for _, event := range eventList.Items {
		log.Infof("Time: %v, Event: %s, Type: %s, Reason: %s, Object: %s/%s, Message: %s",
			event.FirstTimestamp, event.Name, event.Type, event.Reason, event.InvolvedObject.Kind, event.InvolvedObject.Name, event.Message)
	}
	return nil
}

// HotAddPVCsToKubevirtVM hot adds disk to a running VM
func HotAddPVCsToKubevirtVM(virtualMachines []*scheduler.Context, numberOfDisks int, size string) error {
	for _, appCtx := range virtualMachines {
		vms, err := GetAllVMsFromScheduledContexts([]*scheduler.Context{appCtx})
		if err != nil {
			return fmt.Errorf("failed to get VMs from context: %w", err)
		}
		for _, v := range vms {
			diskCountOutput, err := GetNumberOfDisksInVMViaVirtLauncherPod(appCtx)
			if err != nil {
				return fmt.Errorf("failed to get number of disks in VM [%s] in namespace [%s]: %w", v.Name, v.Namespace, err)
			}
			log.Infof("Currently having %v number of disks in the VM", diskCountOutput)
			storageClass, err := GetStorageClassOfVmPVC(appCtx)
			if err != nil {
				return fmt.Errorf("failed to get storage class for VM [%s]: %w", v.Name, err)
			}
			pvcs, err := CreatePVCsForVM(v, numberOfDisks, storageClass, size)
			if err != nil {
				return fmt.Errorf("failed to create PVCs for VM [%s]: %w", v.Name, err)
			}

			vm, err := k8sKubevirt.GetVirtualMachine(v.Name, v.Namespace)
			if err != nil {
				return fmt.Errorf("failed to get VM [%s]: %w", v.Name, err)
			}

			for _, pvc := range pvcs {
				diskName := fmt.Sprintf("disk-%s", pvc.Name)
				// Appending the disk and volume definitions to the VM spec
				vm.Spec.Template.Spec.Domain.Devices.Disks = append(vm.Spec.Template.Spec.Domain.Devices.Disks, kubevirtv1.Disk{
					Name: diskName,
					DiskDevice: kubevirtv1.DiskDevice{
						Disk: &kubevirtv1.DiskTarget{
							Bus: "virtio",
						},
					},
				})

				vm.Spec.Template.Spec.Volumes = append(vm.Spec.Template.Spec.Volumes, kubevirtv1.Volume{
					Name: diskName,
					VolumeSource: kubevirtv1.VolumeSource{
						PersistentVolumeClaim: &kubevirtv1.PersistentVolumeClaimVolumeSource{
							PersistentVolumeClaimVolumeSource: corev1.PersistentVolumeClaimVolumeSource{
								ClaimName: pvc.Name,
							},
							Hotpluggable: true,
						},
					},
				})
			}

			_, err = k8sKubevirt.UpdateVirtualMachine(vm)
			if err != nil {
				return fmt.Errorf("failed to update VM [%s]: %w", vm.Name, err)
			}

			err = RestartKubevirtVM(v.Name, v.Namespace, true)
			if err != nil {
				return err
			}
			log.InfoD("Sleep for 5mins for vm to come up")
			time.Sleep(5 * time.Minute)
			NewDiskCountOutput, err := GetNumberOfDisksInVMViaVirtLauncherPod(appCtx)
			if err != nil {
				return fmt.Errorf("failed to get number of disks in VM [%s] in namespace [%s]", v.Name, v.Namespace)
			}
			if NewDiskCountOutput == diskCountOutput {
				log.Infof("Disk successfully added")
			} else {
				return fmt.Errorf("Disk cannot be added.")
			}
		}
	}
	return nil
}

func DeployVMTemplatesAndValidate() error {
	_, err := Inst().S.Schedule("",
		scheduler.ScheduleOptions{
			AppKeys:   []string{kubevirtTemplates},
			Namespace: kubevirtTemplateNamespace,
		})

	// if new templates are deployed, this function will wait for them to get imported else it will exit
	waitForCompletedAnnotations := func() (interface{}, bool, error) {
		// Loop through all PVCs and check for annotations that signify existing downloaded templates
		pvcTemplates, err := core.Instance().GetPersistentVolumeClaims(kubevirtTemplateNamespace, nil)
		if err != nil {
			return nil, true, fmt.Errorf("failed to get any PVCs in namespace: %s. Retrying.", kubevirtTemplateNamespace)
		}
		for _, pvc := range pvcTemplates.Items {
			if pvc.ObjectMeta.Annotations[kubevirtCDIStorageConditionAnnotation] != "Completed" {
				return nil, true, fmt.Errorf("storage condition is not completed on pvc %s. Status: %s. Retrying.",
					pvc.Name, pvc.ObjectMeta.Annotations[kubevirtCDIStorageConditionAnnotation])
			}
			if pvc.ObjectMeta.Annotations[kubevirtCDIStoragePodPhaseAnnotation] != "Succeeded" {
				return nil, true, fmt.Errorf("pod phase has not succeeded on pvc %s. Phase: %s. Retrying.",
					pvc.Name, pvc.ObjectMeta.Annotations[kubevirtCDIStoragePodPhaseAnnotation])
			}
		}
		log.Infof("All templates are downloaded.")
		return "", false, nil
	}
	_, err = task.DoRetryWithTimeout(waitForCompletedAnnotations, importerPodCompletionTimeout, importerPodRetryInterval)
	return err
}

// GetReplicaNodesOfVM returns a list of nodes where the replica of the volumes are present
func GetReplicaNodesOfVM(virtualMachineCtx *scheduler.Context) ([]string, error) {
	vols, err := Inst().S.GetVolumes(virtualMachineCtx)
	if err != nil {
		return nil, err
	}
	// Get replica nodes where the volumes are present
	replicaNodes, err := getReplicaNodes(vols[0])
	if err != nil {
		return nil, err
	}
	replicaNodesName := []string{}
	for _, replicaNode := range replicaNodes {
		id, err := node.GetNodeDetailsByNodeID(replicaNode)
		log.FailOnError(err, "Failed to get node details by node id")
		replicaNodesName = append(replicaNodesName, id.Name)
	}

	return replicaNodesName, nil
}

func GetNonReplicaNodesOfVM(virtualMachineCtx *scheduler.Context) ([]string, error) {
	vols, err := Inst().S.GetVolumes(virtualMachineCtx)
	if err != nil {
		return nil, err
	}
	// Get replica nodes where the volumes are present
	replicaNodes, err := getReplicaNodes(vols[0])
	if err != nil {
		return nil, err
	}
	for _, replicaNode := range replicaNodes {
		log.Infof("Replica node: %s", replicaNode)
	}

	replicaNodesName := []string{}
	for _, replicaNode := range replicaNodes {
		id, err := node.GetNodeDetailsByNodeID(replicaNode)
		log.FailOnError(err, "Failed to get node details by node id")
		replicaNodesName = append(replicaNodesName, id.Name)
	}

	// Get all nodes in the cluster
	allNodes, err := GetStorageNodes()
	if err != nil {
		return nil, err
	}
	for _, node := range allNodes {
		log.Infof("All nodes: %s", node.Name)
	}

	// Get non-replica nodes
	nonReplicaNodes := []string{}

	for _, node := range allNodes {
		flag := false
		for _, replicaNode := range replicaNodesName {
			if node.Name == replicaNode {
				flag = true
			}
		}
		if !flag {
			nonReplicaNodes = append(nonReplicaNodes, node.Name)
		}
	}
	for _, nonReplicaNode := range nonReplicaNodes {
		log.Infof("Non-replica node: %s", nonReplicaNode)
	}

	return nonReplicaNodes, nil
}

func FillRootDiskOfVM(virtualMachineCtx *scheduler.Context) error {
	vols, err := Inst().S.GetVolumes(virtualMachineCtx)
	if err != nil {
		return err
	}
	vmPod, err := GetVirtLauncherPodForVM(virtualMachineCtx, vols[0])
	if err != nil {
		return err
	}
	rootDiskPath, err := GetVMRootDiskPath(virtualMachineCtx)
	if err != nil {
		return err
	}

	output, err := core.Instance().RunCommandInPod([]string{"dd", "if=/dev/zero", "of=" + rootDiskPath + "/fill", "bs=100M", "count=100"}, vmPod.Name, "compute", vmPod.Namespace)
	if err != nil && strings.Contains(output, "No space left on device") {
		return nil
	}
	if err != nil {
		return err
	}
	log.Infof("Output of dd command: %s", output)
	return nil
}

func RestartVM(virtualMachineCtx *scheduler.Context) error {
	vms, err := GetAllVMsFromScheduledContexts([]*scheduler.Context{virtualMachineCtx})
	if err != nil {
		return err
	}
	for _, v := range vms {
		err = RestartKubevirtVM(v.Name, v.Namespace, true)
		if err != nil {
			return err
		}
	}
	return nil
}

func GetVMRootDiskPath(virtualMachineCtx *scheduler.Context) (string, error) {
	vols, err := Inst().S.GetVolumes(virtualMachineCtx)
	if err != nil {
		return "", err
	}
	vmPod, err := GetVirtLauncherPodForVM(virtualMachineCtx, vols[0])
	if err != nil {
		return "", err
	}
	output, err := core.Instance().RunCommandInPod([]string{"lsblk"}, vmPod.Name, "compute", vmPod.Namespace)
	if err != nil {
		return "", err
	}
	log.Infof("Output of lsblk command: %s", output)
	// Get the line which has rootdisk as the disk name
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "rootdisk") {
			// Split the line to get the disk path
			fields := strings.Fields(line)
			return fields[len(fields)-1], nil
		}
	}

	return "", fmt.Errorf("rootdisk not found in lsblk output")
}

func GetVMIPAddress(vm kubevirtv1.VirtualMachine) (string, error) {
	t := func() (interface{}, bool, error) {
		vmInstance, err := k8sKubevirt.GetVirtualMachineInstance(context1.TODO(), vm.Name, vm.Namespace)
		if err != nil {
			return "", false, err
		}
		if len(vmInstance.Status.Interfaces) == 0 {
			return "", true, fmt.Errorf("no interfaces found in the VM [%s] in namespace [%s]", vm.Name, vm.Namespace)
		}
		return vmInstance.Status.Interfaces[0].IP, false, nil
	}
	result, err := task.DoRetryWithTimeout(t, 5*time.Minute, 30*time.Second)
	if err != nil {
		return "", err
	}
	ipAddress, ok := result.(string)
	if !ok {
		return "", fmt.Errorf("failed to get IP address of VM [%s] in namespace [%s]", vm.Name, vm.Namespace)
	}
	return ipAddress, nil
}

func TestSSHConnectivity(ipAddress string) error {
	sshPwd, present := os.LookupEnv("KUBEVIRT_VM_PWD")
	if !present {
		return fmt.Errorf("Please set KUBEVIRT_VM_PWD to login inside the Kubevirt VM")
	}
	testCmdArgs := getSSHCommandArgs(sshUserName, sshPwd, ipAddress, "hostname")
	t := func() (interface{}, bool, error) {
		output, err := k8sCore.RunCommandInPod(testCmdArgs, sshPodName, "ssh-container", "default")
		if err != nil {
			log.Infof("Error encountered during SSH connection test")
			if isConnectionError(err.Error()) {
				log.Infof("Test connection output - \n%s", output)
				return "", true, err
			} else {
				return "", false, err
			}
		}
		log.Infof("SSH connection successful. Output - \n%s", output)
		return "", false, nil
	}
	_, err := task.DoRetryWithTimeout(t, 10*time.Minute, 30*time.Second)
	return err
}

func RunCommandInVM(ipAddress, command string) (string, error) {
	sshPwd, present := os.LookupEnv("KUBEVIRT_VM_PWD")
	if !present {
		return "", fmt.Errorf("Please set KUBEVIRT_VM_PWD to login inside the Kubevirt VM")
	}
	cmdArgs := getSSHCommandArgs(sshUserName, sshPwd, ipAddress, command)
	output, err := k8sCore.RunCommandInPod(cmdArgs, sshPodName, "ssh-container", "default")
	if err != nil {
		log.Errorf("Error executing command %s - \n%s", command, output)
		return "", err
	}
	log.Infof("Output of command %s - \n%s", command, output)
	return output, nil
}

func CheckFioIsRunningInVM(vm kubevirtv1.VirtualMachine) error {
	ipAddress, err := GetVMIPAddress(vm)
	if err != nil {
		return err
	}
	log.Infof("VM Name - %s", vm.Name)
	log.Infof("IP Address - %s", ipAddress)

	err = TestSSHConnectivity(ipAddress)
	if err != nil {
		return err
	}

	cmd := "ps -ef | grep fio | grep -v grep"
	output, err := RunCommandInVM(ipAddress, cmd)
	if err != nil {
		return err
	}
	output = strings.TrimSpace(output)
	if output == "" {
		return fmt.Errorf("fio is not running in VM [%s] in namespace [%s]", vm.Name, vm.Namespace)
	} else {
		log.Infof("fio is running in VM [%s]", vm.Name)
	}
	return nil
}

func GetNumberOfDrivesInVM(vm kubevirtv1.VirtualMachine) (int, error) {
	var numDrives int
	t := func() (interface{}, bool, error) {
		ipAddress, err := GetVMIPAddress(vm)
		if err != nil {
			return nil, true, err
		}
		log.Infof("VM Name - %s", vm.Name)
		log.Infof("IP Address - %s", ipAddress)

		err = TestSSHConnectivity(ipAddress)
		if err != nil {
			log.Warnf("SSH connectivity failed to VM [%s]: %v", vm.Name, err)
			return nil, true, err
		}

		cmd := "lsblk | grep -E '^(vd|sd)' | grep disk | wc -l"
		output, err := RunCommandInVM(ipAddress, cmd)
		if err != nil {
			log.Warnf("Failed to execute command in VM [%s]: %v", vm.Name, err)
			return nil, true, err
		}

		output = strings.TrimSpace(output)
		numDrives, err = strconv.Atoi(output)
		if err != nil {
			log.Warnf("Failed to parse number of drives from output [%s]: %v", output, err)
			return nil, true, err
		}
		return numDrives, false, nil
	}
	_, err := task.DoRetryWithTimeout(t, 5*time.Minute, 30*time.Second)
	if err != nil {
		return 0, fmt.Errorf("failed to get number of drives in VM [%s]: %v", vm.Name, err)
	}
	log.Infof("Number of drives in VM [%s]: %d", vm.Name, numDrives)
	return numDrives, nil
}

// AddFadaDriveToKubevirtVM adds additional drives to KubeVirt VMs.
func AddRawBlockDriveToKubevirtVM(virtualMachines []*scheduler.Context, numberOfDisks int, size string) (bool, error) {

	for _, appCtx := range virtualMachines {
		vms, err := GetAllVMsFromScheduledContexts([]*scheduler.Context{appCtx})
		if err != nil {
			return false, fmt.Errorf("failed to get VMs from scheduled contexts: %v", err)
		}

		for _, v := range vms {
			// Get the initial number of disks
			initialDiskCount, err := GetNumberOfDrivesInVM(v)
			if err != nil {
				return false, fmt.Errorf("failed to get initial number of disks in VM [%s]: %v", v.Name, err)
			}
			log.Infof("Initial number of disks in VM [%s]: %d", v.Name, initialDiskCount)

			// Get the storage class of the existing VM PVC
			storageClass, err := GetStorageClassOfVmPVC(appCtx)
			if err != nil {
				return false, fmt.Errorf("failed to get storage class of VM PVC: %v", err)
			}
			log.Infof("Storage class of PVC attached to VM [%s]: %s", v.Name, storageClass)

			// Create the new PVCs
			pvcs, err := CreateBlockModePVCsForVM(v, numberOfDisks, storageClass, size, corev1.PersistentVolumeBlock)
			if err != nil {
				return false, fmt.Errorf("failed to create PVCs for VM [%s]: %v", v.Name, err)
			}

			// Add the new PVCs to the app context's spec list
			for _, pvc := range pvcs {
				appCtx.App.SpecList = append(appCtx.App.SpecList, pvc)
			}

			// Add the PVCs to the VM
			err = AddPVCsToVirtualMachine(v, pvcs)
			if err != nil {
				return false, fmt.Errorf("failed to add PVCs to VM [%s]: %v", v.Name, err)
			}

			// Restart the VM
			err = RestartKubevirtVM(v.Name, v.Namespace, true)
			if err != nil {
				return false, fmt.Errorf("failed to restart VM [%s]: %v", v.Name, err)
			}

			// Wait for VM to be ready
			err = WaitForVMToBeReady(v.Name, v.Namespace)
			if err != nil {
				return false, fmt.Errorf("VM [%s] did not become ready: %v", v.Name, err)
			}

			// Verify the new number of disks
			expectedDiskCount := initialDiskCount + numberOfDisks
			t := func() (interface{}, bool, error) {
				newDiskCount, err := GetNumberOfDrivesInVM(v)
				if err != nil {
					log.Warnf("Failed to get number of disks in VM [%s]: %v", v.Name, err)
					return nil, true, err
				}
				if newDiskCount != expectedDiskCount {
					err := fmt.Errorf("number of disks in VM [%s] is %d; expected %d", v.Name, newDiskCount, expectedDiskCount)
					log.Warnf(err.Error())
					return nil, true, err
				}
				return newDiskCount, false, nil
			}
			_, err = task.DoRetryWithTimeout(t, 5*time.Minute, 30*time.Second)
			if err != nil {
				return false, fmt.Errorf("failed to verify number of disks in VM [%s]: %v", v.Name, err)
			}
			log.Infof("Successfully verified number of disks in VM [%s]", v.Name)
		}
	}
	return true, nil
}

func CreateBlockModePVCsForVM(vm kubevirtv1.VirtualMachine, numberOfPVCs int, storageClassName, resourceStorage string, volumeMode corev1.PersistentVolumeMode) ([]*corev1.PersistentVolumeClaim, error) {
	pvcs := make([]*corev1.PersistentVolumeClaim, 0)
	for i := 0; i < numberOfPVCs; i++ {
		pvcName := fmt.Sprintf("%s-%s-%d", "pvc-new", vm.Name, rand.Intn(10000))
		pvcSpec := &corev1.PersistentVolumeClaim{
			TypeMeta: metav1.TypeMeta{
				Kind: "PersistentVolumeClaim",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      pvcName,
				Namespace: vm.Namespace,
			},
			Spec: corev1.PersistentVolumeClaimSpec{
				AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany},
				StorageClassName: &storageClassName,
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: resource.MustParse(resourceStorage),
					},
				},
				VolumeMode: &volumeMode,
			},
		}
		pvc, err := core.Instance().CreatePersistentVolumeClaim(pvcSpec)
		if err != nil {
			return nil, err
		}
		pvc.Kind = "PersistentVolumeClaim"
		pvcs = append(pvcs, pvc)
	}
	return pvcs, nil
}

func WaitForVMToBeReady(vmName string, namespace string) error {
	k8sKubevirt := kubevirt.Instance()
	var ipAddress string

	t := func() (interface{}, bool, error) {
		vmInstance, err := k8sKubevirt.GetVirtualMachineInstance(context1.TODO(), vmName, namespace)
		if err != nil {
			log.Warnf("Failed to get VM instance [%s]: %v", vmName, err)
			return nil, true, err // Retry
		}

		if vmInstance.Status.Phase != kubevirtv1.Running {
			log.Warnf("VM [%s] is not running yet. Current phase: %s", vmName, vmInstance.Status.Phase)
			return nil, true, fmt.Errorf("VM not running")
		}

		if len(vmInstance.Status.Interfaces) == 0 || vmInstance.Status.Interfaces[0].IP == "" {
			log.Warnf("VM [%s] does not have an IP address yet", vmName)
			return nil, true, fmt.Errorf("VM does not have an IP address yet")
		}
		ipAddress = vmInstance.Status.Interfaces[0].IP
		log.Infof("VM [%s] has IP address: %s", vmName, ipAddress)

		err = TestSSHConnectivity(ipAddress)
		if err != nil {
			log.Warnf("SSH connectivity test failed for VM [%s]: %v", vmName, err)
			return nil, true, err
		}

		cmd := "hostname"
		output, err := RunCommandInVM(ipAddress, cmd)
		if err != nil {
			log.Warnf("Failed to run command in VM [%s]: %v", vmName, err)
			return nil, true, err
		}
		log.Infof("VM [%s] is ready. Hostname: %s", vmName, strings.TrimSpace(output))
		return nil, false, nil
	}

	_, err := task.DoRetryWithTimeout(t, 20*time.Minute, 30*time.Second)
	if err != nil {
		return fmt.Errorf("VM [%s] did not become ready within timeout: %v", vmName, err)
	}
	return nil
}

// CheckVMState checks the state of the VM after migration
func CheckVMState(vm kubevirtv1.VirtualMachine) error {
	vmi, err := kubevirt.Instance().GetVirtualMachineInstance(context1.TODO(), vm.Name, vm.Namespace)
	if err != nil {
		return fmt.Errorf("failed to get VM instance: %v", err)
	}
	if vmi.Status.Phase != kubevirtv1.VirtualMachineInstancePhase(kubevirtv1.Running) {
		return fmt.Errorf("VM %s is not running, current phase: %s", vm.Name, vmi.Status.Phase)
	}
	log.Infof("VM %s is in phase %s", vm.Name, vmi.Status.Phase)
	return nil
}

func CheckVMUptime(vm kubevirtv1.VirtualMachine, initialUptime map[string]time.Duration) error {
	currentUptime, err := GetVMUptime(vm)
	if err != nil {
		return err
	}
	vmKey := fmt.Sprintf("%s/%s", vm.Namespace, vm.Name)
	initialUptimeVM, ok := initialUptime[vmKey]
	if !ok {
		return fmt.Errorf("Initial uptime not found for VM %s", vmKey)
	}
	acceptableDelta := 1 * time.Second
	if currentUptime >= initialUptimeVM-acceptableDelta {
		log.Infof("VM %s has not restarted. Initial uptime: %v, Current uptime: %v", vmKey, initialUptimeVM, currentUptime)
	} else {
		return fmt.Errorf("VM %s has restarted. Initial uptime: %v, Current uptime: %v", vmKey, initialUptimeVM, currentUptime)
	}
	return nil
}

func GetVMUptime(vm kubevirtv1.VirtualMachine) (time.Duration, error) {
	ipAddress, err := GetVMIPAddress(vm)
	if err != nil {
		return 0, err
	}
	log.Infof("VM Name - %s", vm.Name)
	log.Infof("IP Address - %s", ipAddress)

	err = TestSSHConnectivity(ipAddress)
	if err != nil {
		return 0, err
	}

	cmd := "cat /proc/uptime"
	output, err := RunCommandInVM(ipAddress, cmd)
	if err != nil {
		return 0, err
	}
	output = strings.TrimSpace(output)
	parts := strings.Fields(output)
	if len(parts) < 1 {
		return 0, fmt.Errorf("Unexpected output from uptime command: %s", output)
	}
	uptimeSecondsStr := parts[0]
	uptimeSeconds, err := strconv.ParseFloat(uptimeSecondsStr, 64)
	if err != nil {
		return 0, fmt.Errorf("Failed to parse uptime seconds: %v", err)
	}
	uptimeDuration := time.Duration(uptimeSeconds * float64(time.Second))
	return uptimeDuration, nil
}

func GetPVCsAttachedToVM(vm kubevirtv1.VirtualMachine) []string {
	pvcNames := k8sKubevirt.GetVMPersistentVolumeClaims(&vm)
	return pvcNames
}

// CreateBlankDataVolume creates a blank data volume with a given storageclass
func CreateBlankDataVolume(namespace string, dvName string, storageClassName string, size string, volumeMode string) (*cdiv1.DataVolume, error) {
	kvClient := k8sKubevirt.GetKubevirtClient()
	var VolumeMode *corev1.PersistentVolumeMode
	if volumeMode == "Block" {
		VolumeMode = &[]corev1.PersistentVolumeMode{corev1.PersistentVolumeBlock}[0]
	} else {
		VolumeMode = &[]corev1.PersistentVolumeMode{corev1.PersistentVolumeFilesystem}[0]
	}
	dv := &cdiv1.DataVolume{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DataVolume",
			APIVersion: "cdi.kubevirt.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      dvName,
			Namespace: namespace,
		},
		Spec: cdiv1.DataVolumeSpec{
			Source: &cdiv1.DataVolumeSource{
				Blank: &cdiv1.DataVolumeBlankImage{},
			},
			Storage: &cdiv1.StorageSpec{
				StorageClassName: &storageClassName,
				AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany},
				VolumeMode:       VolumeMode,
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: resource.MustParse(size),
					},
				},
			},
		},
	}

	createdDV, err := kvClient.
		CdiClient().
		CdiV1beta1().
		DataVolumes(namespace).
		Create(context1.TODO(), dv, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create DataVolume [%s/%s]: %w", namespace, dvName, err)
	}

	log.Infof("Created DataVolume [%s/%s], waiting for it to become Ready", namespace, dvName)

	return createdDV, nil
}

// HotPlugDataVolumesToKubevirtVM main trigger to hot plug volumes to a running VM
func HotPlugDataVolumesToKubevirtVM(virtualMachines []*scheduler.Context, numberOfDVs int, size string, volumeMode string, persist bool, opts ...int) (bool, error) {
	var (
		newDiskCount     int
		initialDiskCount int
	)
	log.InfoD("Beginning hot-plug of [%d] DataVolume(s) to each VM (size=%s, volumeMode=%s)",
		numberOfDVs, size, volumeMode)

	numberOfVMs := -1 // Default: no limit
	if len(opts) > 0 {
		numberOfVMs = opts[0] // Use the first optional parameter as the number of VMs
	}
	vmCount := 0

	for _, appCtx := range virtualMachines {
		vms, err := GetAllVMsFromScheduledContexts([]*scheduler.Context{appCtx})
		if err != nil {
			return false, fmt.Errorf("failed to get VMs from scheduled contexts: %v", err)
		}

		for _, vm := range vms {
			// Check if we've reached the limit of VMs to process
			if numberOfVMs != -1 && vmCount >= numberOfVMs {
				log.Infof("Processed the specified number of VMs [%d]. Stopping further processing.", numberOfVMs)
				return true, nil
			}
			storageClass, err := GetStorageClassOfVmPVC(appCtx)
			if err != nil {
				return false, fmt.Errorf("failed to get storage class for VM [%s/%s]: %v", vm.Namespace, vm.Name, err)
			}
			log.Infof("Using storageClass=[%s] for new DataVolumes for VM [%s/%s]", storageClass, vm.Namespace, vm.Name)

			err = WaitForVMToBeReady(vm.Name, vm.Namespace)
			if err != nil {
				return false, fmt.Errorf("VM [%s/%s] not ready: %v", vm.Namespace, vm.Name, err)
			}

			initialDiskCount, err = GetNumberOfDrivesInVM(vm)
			if err != nil {
				return false, fmt.Errorf("failed to get initial number of disks in VM [%s]: %v", vm.Name, err)
			}
			log.Infof("Initial number of disks in VM [%s]: %d", vm.Name, initialDiskCount)

			for i := 0; i < numberOfDVs; i++ {
				dvName := fmt.Sprintf("hotplug-dv-%s-%v-%d", vm.Name, time.Now().Unix(), i)
				log.Infof("Creating blank DataVolume [%s/%s] with size=[%s]", vm.Namespace, dvName, size)

				dv, err := CreateBlankDataVolume(vm.Namespace, dvName, storageClass, size, volumeMode)
				if err != nil {
					return false, fmt.Errorf("failed to create DV [%s/%s]: %v", vm.Namespace, dvName, err)
				}
				log.Infof("Data Volume Created. Hard Sleep for 30 seconds for DV to settle down")
				time.Sleep(30 * time.Second)
				// send signal to start the reboot operation of PX or Node reboot during hot-plug disk
				if RebootFlag != nil && !SignalSent {
					log.Infof("Reboot flag initiated during hot-plug")
					RebootFlag <- struct{}{}
					SignalSent = true
				}
				err = HotPlugDVToVM(vm.Name, vm.Namespace, dv.Name, persist)
				if err != nil {
					return false, fmt.Errorf("failed to hotplug DV [%s/%s] into VM [%s/%s]: %v",
						dv.Namespace, dv.Name, vm.Namespace, vm.Name, err)
				}
				err = WaitForHotplugVolumeReady(vm.Namespace, vm.Name, dv.Name, 5*time.Minute, 10*time.Second)
				if err != nil {
					return false, fmt.Errorf(
						"failed waiting for DV [%s/%s] to become Ready in VM [%s/%s]: %v",
						vm.Namespace, dv.Name, vm.Namespace, vm.Name, err,
					)
				}
				log.Infof("Successfully hot-plugged DV [%s/%s] into VM [%s/%s]", dv.Namespace, dv.Name, vm.Namespace, vm.Name)
			}

			t := func() (interface{}, bool, error) {
				newDiskCount, err = GetNumberOfDrivesInVM(vm)
				if err != nil {
					return nil, true, err
				}
				if newDiskCount < initialDiskCount+numberOfDVs {
					return nil, true, fmt.Errorf(
						"Expected at least [%d] disks, found only [%d]",
						initialDiskCount+numberOfDVs, newDiskCount)
				}
				return newDiskCount, false, nil
			}
			log.Infof("Number of disks after adding Hot Pluggable disk [%v] vs initial disks before adding disk [%v]", newDiskCount, initialDiskCount)
			_, err = task.DoRetryWithTimeout(t, 5*time.Minute, 20*time.Second)
			if err != nil {
				return false, fmt.Errorf("failed to confirm new disks in VM [%s/%s]: %v", vm.Namespace, vm.Name, err)
			}
		}
		vmCount++
	}
	log.Infof("Total number of disks after adding Hot Pluggable disk [%v] and total number of disks before adding Hot pluggable disk [%v] ", newDiskCount, initialDiskCount)
	return true, nil
}

// WaitForHotplugVolumeReady waits for hp-volume pod to become ready
func WaitForHotplugVolumeReady(namespace, vmName, dvName string, timeout, retryInterval time.Duration) error {
	f := func() (interface{}, bool, error) {
		kvClient := k8sKubevirt.GetKubevirtClient()
		vmi, err := kvClient.VirtualMachineInstance(namespace).Get(context1.TODO(), vmName, &metav1.GetOptions{})
		if err != nil {
			return nil, true, fmt.Errorf("failed to get VMI [%s/%s]: %w", namespace, vmName, err)
		}

		var (
			attachPodName string
			found         bool
		)

		for _, vs := range vmi.Status.VolumeStatus {
			if vs.Name == dvName {
				found = true
				if vs.HotplugVolume != nil {
					attachPodName = vs.HotplugVolume.AttachPodName
				}
				if vs.Phase == kubevirtv1.VolumeReady {
					log.Infof("Volume [%s] is VolumeReady in VMI [%s/%s]", dvName, namespace, vmName)
					if attachPodName != "" {
						pod, err := k8sCore.GetPodByName(attachPodName, namespace)
						if err != nil {
							return nil, true, fmt.Errorf("failed to get hotplug pod [%s]: %w", attachPodName, err)
						}
						if k8sCore.IsPodRunning(*pod) {
							return nil, false, nil
						}
						return nil, true, fmt.Errorf("hotplug pod [%s] is not running yet", attachPodName)
					}
					return nil, true, fmt.Errorf("volume is Ready but attachPodName is empty")
				}
			}
		}

		if !found {
			log.Infof("Volume [%s] not yet in VMI status for VMI [%s/%s]. Retrying...", dvName, namespace, vmName)
			return nil, true, fmt.Errorf("volume [%s] not found in volumeStatus", dvName)
		}

		return nil, true, fmt.Errorf("volume [%s] found but not Ready yet in VMI [%s/%s]", dvName, namespace, vmName)
	}

	_, err := task.DoRetryWithTimeout(f, timeout, retryInterval)
	if err != nil {
		return fmt.Errorf("volume [%s] didn't become Ready in VMI [%s/%s] within %v: %v",
			dvName, namespace, vmName, timeout, err)
	}
	return nil
}

// HotPlugDVToVM method triggers hot pluging of given datavolume to given VM
func HotPlugDVToVM(vmName, namespace, dvName string, persist bool) error {
	kvClient := k8sKubevirt.GetKubevirtClient()
	bytes := make([]byte, 10)
	serial := hex.EncodeToString(bytes)
	addVolumeOptions := &kubevirtv1.AddVolumeOptions{
		Name: dvName,
		Disk: &kubevirtv1.Disk{
			DiskDevice: kubevirtv1.DiskDevice{
				Disk: &kubevirtv1.DiskTarget{Bus: kubevirtv1.DiskBusSCSI},
			},
			Serial: serial,
		},
		VolumeSource: &kubevirtv1.HotplugVolumeSource{
			DataVolume: &kubevirtv1.DataVolumeSource{Name: dvName},
		},
	}

	err := kvClient.VirtualMachineInstance(namespace).AddVolume(context1.TODO(), vmName, addVolumeOptions)
	if err != nil {
		return err
	}
	if persist {
		vm, err := kvClient.VirtualMachine(namespace).Get(vmName, &metav1.GetOptions{})
		if err != nil {
			return err
		}
		vm.Spec.Template.Spec.Domain.Devices.Disks = append(
			vm.Spec.Template.Spec.Domain.Devices.Disks,
			kubevirtv1.Disk{
				Name: dvName,
				DiskDevice: kubevirtv1.DiskDevice{
					Disk: &kubevirtv1.DiskTarget{Bus: kubevirtv1.DiskBusSCSI},
				},
			},
		)
		vm.Spec.Template.Spec.Volumes = append(
			vm.Spec.Template.Spec.Volumes,
			kubevirtv1.Volume{
				Name: dvName,
				VolumeSource: kubevirtv1.VolumeSource{
					DataVolume: &kubevirtv1.DataVolumeSource{Name: dvName},
				},
			},
		)
		_, err = kvClient.VirtualMachine(namespace).Update(vm)
		if err != nil {
			return err
		}
	}
	return nil
}

func CreateSSHPodAndSetCanSsh() bool {
	stepLog := "Create SSH Pod"
	var canSsh bool = false
	Step(stepLog, func() {
		log.InfoD(stepLog)
		err := CreateSSHPod()
		if err == nil {
			canSsh = true
		} else {
			canSsh = false
		}
	})
	return canSsh
}

func ValidateFioInVMs(appCtxs []*scheduler.Context, canSsh bool) {
	if canSsh {
		stepLog := "Validate fio is running in the VMs"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			var wg sync.WaitGroup
			for _, appCtx := range appCtxs {
				wg.Add(1)
				go func(appCtx *scheduler.Context) {
					defer GinkgoRecover()
					defer wg.Done()
					vms, err := GetAllVMsFromScheduledContexts([]*scheduler.Context{appCtx})
					log.FailOnError(err, "Failed to get VMs from appCtx")
					for _, vm := range vms {
						err = CheckFioIsRunningInVM(vm)
						log.FailOnError(err, "Failed to validate fio in VM %s", vm.Name)
					}
				}(appCtx)
			}
			wg.Wait()
		})
	}
}

func ValidateVMUptime(appCtxs []*scheduler.Context, canSsh bool, initialUptime map[string]time.Duration) {
	if canSsh {
		stepLog := "Validate VMs have not restarted"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			var wg sync.WaitGroup
			for _, appCtx := range appCtxs {
				wg.Add(1)
				go func(appCtx *scheduler.Context) {
					defer GinkgoRecover()
					defer wg.Done()
					vms, err := GetAllVMsFromScheduledContexts([]*scheduler.Context{appCtx})
					log.FailOnError(err, "Failed to get VMs from appCtx")
					for _, vm := range vms {
						err = CheckVMUptime(vm, initialUptime)
						log.FailOnError(err, "Failed to validate uptime in VM %s", vm.Name)
					}
				}(appCtx)
			}
			wg.Wait()
		})
	}
}

func RemoveHotPluggedDiskFromVM(appCtx []*scheduler.Context) (bool, error) {
	var (
		vmName           string
		namespace        string
		dvName           string
		initialDiskCount int
		finalDiskCount   int
	)

	// Get the KubeVirt client
	kvClient := k8sKubevirt.GetKubevirtClient()

	vms, err := GetAllVMsFromScheduledContexts(appCtx)
	if err != nil {
		return false, fmt.Errorf("failed to get VMs from scheduled contexts: %v", err)
	}

	for _, vm := range vms {
		vmName = vm.Name
		namespace = vm.Namespace
		log.Infof("Checking VM [%v] for hot-plugged volumes", vmName)

		initialDiskCount, err = GetNumberOfDrivesInVM(vm)
		if err != nil {
			return false, fmt.Errorf("failed to get initial number of disks in VM [%s]: %v", vm.Name, err)
		}
		log.Infof("Initial number of disks in VM [%s]: %d", vm.Name, initialDiskCount)

		// Look for hot-plugged volumes in the VMI status
		vmi, err := kvClient.VirtualMachineInstance(namespace).Get(context1.TODO(), vmName, &metav1.GetOptions{})
		if err != nil {
			return false, fmt.Errorf("failed to get VMI for VM [%v], error : [%v]", vmName, err)
		}

		// Filter hot-plugged volumes
		for _, volumeStatus := range vmi.Status.VolumeStatus {
			if volumeStatus.HotplugVolume != nil {
				dvName = volumeStatus.Name
				log.Infof("Found hot-plugged volume [%v] for VM [%v]", dvName, vmName)

				// Prepare remove volume options
				removeVolumeOptions := &kubevirtv1.RemoveVolumeOptions{
					Name: dvName,
				}

				// Remove the volume
				err = kvClient.VirtualMachineInstance(namespace).RemoveVolume(context1.TODO(), vmName, removeVolumeOptions)
				if err != nil {
					return false, fmt.Errorf("failed to remove hot-plugged DataVolume [%v from VM [%v], error : [%v]",
						dvName, vmName, err)
				}

				log.Infof("Successfully initiated removal of hot-plugged DataVolume [%v] from VM [%v]", dvName, vmName)

				// Wait for the disk to be detached
				err = WaitForHotplugVolumeDetached(namespace, vmName, dvName, 5*time.Minute, 10*time.Second)
				if err != nil {
					return false, fmt.Errorf("failed to confirm removal of hot-plugged DataVolume [%v] from VM [%v], error : %v",
						dvName, vmName, err)
				}

				log.Infof("Successfully removed hot-plugged DataVolume [%v] from VM [%v]", dvName, vmName)
			}
		}
		finalDiskCount, err = GetNumberOfDrivesInVM(vm)
		if err != nil {
			return false, fmt.Errorf("failed to get initial number of disks in VM [%s]: %v", vm.Name, err)
		}
		log.Infof("Final number of disks [%v] to the initial number of disks [%v] in VM [%s]: %d", finalDiskCount, initialDiskCount, vmName)
	}

	return true, nil
}

func WaitForHotplugVolumeDetached(namespace, vmName, dvName string, timeout, retryInterval time.Duration) error {
	f := func() (interface{}, bool, error) {
		kvClient := k8sKubevirt.GetKubevirtClient()
		vmi, err := kvClient.VirtualMachineInstance(namespace).Get(context1.TODO(), vmName, &metav1.GetOptions{})
		if err != nil {
			return nil, true, fmt.Errorf("failed to get VMI [%v] in namespace [%v], error : [%v]", vmName, namespace, err)
		}

		if vmi == nil {
			return nil, true, fmt.Errorf("No vmi [%v] is found for namespace [%v]", vmName, namespace)
		}

		if vmi.Status.VolumeStatus == nil {
			return nil, true, fmt.Errorf("vmi.Status.VolumeStatus [%v] is nil for namespace [%v]", vmName, namespace)
		}

		for _, vs := range vmi.Status.VolumeStatus {

			if vs.Name == dvName {
				log.Infof("Volume [%s] is still attached to VM [%s/%s]. Retrying...", dvName, namespace, vmName)
				return nil, true, fmt.Errorf("volume [%s] is still attached to VM [%v] for namespace [%v]", dvName, vmName, namespace)
			}
		}

		log.Infof("Volume [%v] has been successfully detached from VM [%v] for namepsace [%v]", dvName, vmName, namespace)
		return nil, false, nil
	}

	_, err := task.DoRetryWithTimeout(f, timeout, retryInterval)
	if err != nil {
		return fmt.Errorf("volume [%s] didn't detach from VM [%v] within [%v],error : [%v]",
			dvName, vmName, timeout, err)
	}
	return nil
}

func CheckIsDiskSizeFullInVM(vm kubevirtv1.VirtualMachine) (bool, error) {
	ipAddress, err := GetVMIPAddress(vm)
	if err != nil {
		return false, fmt.Errorf("failed to get IP address: %w", err)
	}
	targetMounts := []string{"/mnt/disks/vdb", "/mnt/disks/vdc"}

	for {
		cmd := "df -kh"
		output, err := RunCommandInVM(ipAddress, cmd)
		if err != nil {
			return false, fmt.Errorf("failed to run command in VM: %w", err)
		}
		log.Infof("Disk usage output for VM [%s]:\n%s", vm.Name, output)

		for _, mount := range targetMounts {
			if isMountUsageFull(output, mount) {
				log.Infof("Disk usage for [%s] has reached 100%%", mount)
				return true, nil
			}
		}
		log.Infof("Rechecking disk usage after 10 seconds...")
		time.Sleep(10 * time.Second)
	}
}

// isMountUsageFull parses the `df -kh` output and checks if the specified mount has 100% usage.
func isMountUsageFull(output, mount string) bool {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, mount) {
			parts := strings.Fields(line)
			if len(parts) < 5 {
				continue
			}
			usageStr := parts[4] // Use% column
			if strings.HasSuffix(usageStr, "%") {
				usage, err := strconv.Atoi(strings.TrimSuffix(usageStr, "%"))
				if err == nil && usage == 100 {
					log.Infof("Mount %s is at 100%% usage.", mount)
					return true
				}
			}
		}
	}
	return false
}

func UpgradePortworxDriverForKubevirtVM(upgradeEndpoints string) error {
	// Ensure upgrade endpoints are provided
	if upgradeEndpoints == "" {
		return fmt.Errorf("no upgrade endpoints provided for PX upgrade")
	}

	storageNodes := node.GetStorageNodes()
	if len(storageNodes) == 0 {
		return fmt.Errorf("no storage nodes found in the cluster")
	}

	// Iterate over upgrade hops and perform the upgrade
	for _, upgradeHop := range strings.Split(upgradeEndpoints, ",") {
		log.Infof("Starting PX upgrade for endpoint: %s", upgradeHop)

		// Capture current PX version
		currPXVersion, err := Inst().V.GetDriverVersionOnNode(storageNodes[0])
		if err != nil {
			log.Warnf("Error getting current PX version: %v", err)
		}

		// Time tracking for the upgrade
		timeBeforeUpgrade := time.Now()

		// Perform PX upgrade
		err = Inst().V.UpgradeDriver(upgradeHop)
		if err != nil {
			return fmt.Errorf("PX upgrade failed for endpoint %s: %v", upgradeHop, err)
		}

		timeAfterUpgrade := time.Now()
		durationInMins := int(timeAfterUpgrade.Sub(timeBeforeUpgrade).Minutes())
		expectedUpgradeTime := 9 * len(node.GetStorageDriverNodes())

		log.Infof("Upgrade completed in %d minutes", durationInMins)
		if durationInMins > expectedUpgradeTime {
			log.Warnf("Upgrade took longer than expected: %d minutes (expected: %d minutes)", durationInMins, expectedUpgradeTime)
		}

		// Verify the new PX version
		updatedPXVersion, err := Inst().V.GetDriverVersionOnNode(storageNodes[0])
		if err != nil {
			log.Warnf("Error getting updated PX version: %v", err)
		}
		log.Infof("PX version upgraded from %s to %s", currPXVersion, updatedPXVersion)
	}

	log.InfoD("PX upgrade completed successfully for all endpoints")
	return nil
}

// StartAndWaitForPostCopyVMIMigration starts the VM migration in post-copy mode
func StartAndWaitForPostCopyVMIMigration(virtualMachineCtx *scheduler.Context, ctx context1.Context) error {
	log.InfoD("Initiating VM migration for VM [%s] in namespace [%s]", virtualMachineCtx.App.Key, virtualMachineCtx.App.NameSpace)
	vms, err := GetAllVMsFromScheduledContexts([]*scheduler.Context{virtualMachineCtx})
	if err != nil {
		return err
	}
	if len(vms) == 0 {
		return fmt.Errorf("No VMs found for VM [%s] in namespace [%s]", virtualMachineCtx.App.Key, virtualMachineCtx.App.NameSpace)
	}
	log.Infof("Total number of VMs [%v] in namespace [%s]", len(vms), virtualMachineCtx.App.NameSpace)

	for _, vm := range vms {
		vmiNamespace := vm.Namespace
		vmiName := vm.Name

		//Get the node where the vm is scheduled before the migration
		nodeName, err := GetNodeOfVM(vm)
		if err != nil {
			return err
		}
		log.Infof("VM [%s] in namespace [%s] is scheduled on node [%s]", vmiName, vmiNamespace, nodeName)

		// Enable post copy
		log.Infof("Enabling post copy")
		err = EnablePostCopy(ctx, vmiNamespace, vmiName)
		if err != nil {
			return err
		}

		// Start the VM migration
		migration, err := kubevirtdy.Instance().CreateVirtualMachineInstanceMigration(ctx, vmiNamespace, vmiName)
		if err != nil {
			return err
		}
		log.Infof("VM migration created for VM [%s] in namespace [%s]", vmiName, vmiNamespace)

		// get volumes from app context
		vols, err := Inst().S.GetVolumes(virtualMachineCtx)
		if err != nil {
			return err
		}

		t := func() (interface{}, bool, error) {
			var migr *kubevirtdy.VirtualMachineInstanceMigration
			migr, err = kubevirtdy.Instance().GetVirtualMachineInstanceMigration(ctx, vmiNamespace, migration.Name)
			if err != nil {
				log.InfoD("Error: %v", err)
				return "", false, fmt.Errorf("failed to get migration for VM [%s] in namespace [%s]", vmiName, vmiNamespace)
			}

			if !(migr.Phase == "Succeeded") {
				return "", true, fmt.Errorf("waiting for migration to complete for VM [%s] in namespace [%s]", vmiName, vmiNamespace)
			}

			err = validatePostCopyMigration(vmiNamespace, migration.Name)
			if err != nil {
				return "", false, fmt.Errorf("failed to validate post copy migration for VM [%s] in namespace [%s]. Error : %v", vmiName, vmiNamespace, err)
			}

			// wait until there is only one pod in the running state
			//TODO https://purestorage.atlassian.net/browse/PTX-23166 - This is a temporary fix to get the pod of the VM
			testPod, err := GetVirtLauncherPodForVM(virtualMachineCtx, vols[0])
			if err != nil {
				return "", true, err
			}

			//Get the node where the vm is scheduled after the migration
			nodeNameAfterMigration := testPod.Spec.NodeName

			if nodeName == nodeNameAfterMigration {
				return "", false, fmt.Errorf("VM pod live migrated [%s] in namespace [%s] but is still on the same node [%s]", testPod.Name, testPod.Namespace, nodeName)
			}
			log.InfoD("VM pod live migrated to node: [%s]", nodeNameAfterMigration)
			return "", false, nil
		}
		_, err = task.DoRetryWithTimeout(t, defaultMigrationTimeout, defaultMigrationRetryInterval)
		if err != nil {
			return err
		}

	}
	kvClient := k8sKubevirt.GetKubevirtClient()
	err = kvClient.MigrationPolicy().Delete(ctx, postCopyMigrationPolicy, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete migration policy: %v", err)
	}
	return nil
}

func EnablePostCopy(ctx context1.Context, vmiNamespace, vmiName string) error {
	kvClient := k8sKubevirt.GetKubevirtClient()
	kvClient.MigrationPolicy()
	// AllowPostCopy enables post-copy live migrations. If set to true, migrations will still start in pre-copy,
	// but switch to post-copy when CompletionTimeoutPerGiB triggers.
	allowPostCopy := true
	// To make likelihood of migration to become post-copy more, set CompletionTimeoutPerGiB to a low value
	var completionTimeoutPerGiB int64 = 1

	// Define the MigrationPolicy with allowPostCopy enabled
	migrationPolicy := v1alpha1.MigrationPolicy{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "migrations.kubevirt.io/v1alpha1",
			Kind:       "MigrationPolicy",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      postCopyMigrationPolicy,
			Namespace: vmiNamespace,
		},
		Spec: v1alpha1.MigrationPolicySpec{
			AllowPostCopy:           &allowPostCopy,
			CompletionTimeoutPerGiB: &completionTimeoutPerGiB,
			Selectors: &v1alpha1.Selectors{
				VirtualMachineInstanceSelector: v1alpha1.LabelSelector{
					"postCopyMigrate": "true",
				},
			},
		},
	}

	// check if migration policy is present
	migrationPolicyResult, err := kvClient.MigrationPolicy().Get(ctx, postCopyMigrationPolicy, metav1.GetOptions{})

	if err != nil {
		if errors.IsNotFound(err) {
			log.Infof("Migration policy not found, creating new...")
		} else {
			return fmt.Errorf("failed to get migration policy: %v", err)
		}
	} else {
		log.Infof("Migration policy name found: %s. Deleting the migration policy", migrationPolicyResult.Name)
		err := kvClient.MigrationPolicy().Delete(ctx, postCopyMigrationPolicy, metav1.DeleteOptions{})
		if err != nil {
			return fmt.Errorf("failed to delete migration policy: %v", err)
		}
	}

	_, err = kvClient.MigrationPolicy().Create(ctx, &migrationPolicy, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create migration policy: %v", err)
	}

	log.Infof("Successfully created MigrationPolicy to enable post-copy for VMs with label postCopyMigrate=true")

	// label the VMI to apply the migration policy
	err = labelVMI(vmiNamespace, vmiName, "postCopyMigrate", "true")
	if err != nil {
		return fmt.Errorf("failed to label VMI: %v", err)
	}

	log.Infof("Successfully labeled VMI [%s] in namespace [%s] to apply migration policy", vmiName, vmiNamespace)

	return nil
}

func labelVMI(vmiNamespace, vmiName, labelKey, labelValue string) error {
	// Prepare the label in the form of a map
	labels := map[string]string{
		labelKey: labelValue,
	}

	kvClient := k8sKubevirt.GetKubevirtClient()
	vmi, err := kvClient.VirtualMachineInstance(vmiNamespace).Get(context1.TODO(), vmiName, &metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get VMI %s in namespace %s: %v", vmiName, vmiNamespace, err)
	}

	// Set the label on the VMI
	if vmi.GetLabels() == nil {
		vmi.SetLabels(labels)
	} else {
		// If labels already exist, update the label map
		vmiLabels := vmi.GetLabels()
		vmiLabels[labelKey] = labelValue
		vmi.SetLabels(vmiLabels)
	}

	// Update the VMI with the new label
	updateResult, err := kvClient.VirtualMachineInstance(vmiNamespace).Update(context1.TODO(), vmi)
	if err != nil {
		return fmt.Errorf("failed to label VMI %s in namespace %s: %v", vmiName, vmiNamespace, err)
	}
	log.Infof("Labels are updated. Result : %v", updateResult)

	return nil
}

func validatePostCopyMigration(vmiNamespace, migrationName string) error {
	kvClient := k8sKubevirt.GetKubevirtClient()
	migrationResult, err := kvClient.VirtualMachineInstanceMigration(vmiNamespace).Get(migrationName, &metav1.GetOptions{})
	if err != nil {
		return err
	}
	log.Infof("Migration result : %+v", migrationResult)
	log.Infof("Migration policy:%v", *migrationResult.Status.MigrationState.MigrationPolicyName)
	if migrationResult.Status.MigrationState.Mode != "PostCopy" {
		return fmt.Errorf("Migration mode - Expected: PostCopy, Actual: %v", migrationResult.Status.MigrationState.Mode)
	}
	log.Infof("Migration mode - Expected: PostCopy, Actual: %v", migrationResult.Status.MigrationState.Mode)
	return nil
}
