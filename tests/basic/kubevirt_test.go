package tests

import (
	context1 "context"
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	"github.com/portworx/sched-ops/k8s/kubevirt"
	"github.com/portworx/sched-ops/task"
	"github.com/portworx/torpedo/drivers/node"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/pkg/log"
	v1 "kubevirt.io/api/core/v1"
	"strconv"
	"time"

	. "github.com/portworx/torpedo/tests"
)

var _ = Describe("{AddNewDiskToKubevirtVM}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("AddNewDiskToKubevirtVM", "Add a new disk to a kubevirtVM", nil, 0)
	})
	var contexts []*scheduler.Context
	var bootTime = 3 * time.Minute
	var diskCount int

	itLog := "Add a new disk to a kubevirtVM"
	It(itLog, func() {
		appList := Inst().AppList
		defer func() {
			Inst().AppList = appList
		}()
		numberOfVolumes := 1
		Inst().AppList = []string{"kubevirt-cirros-cd-with-pvc"}
		stepLog := "schedule a kubevirtVM"
		Step(stepLog, func() {
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				contexts = append(contexts, ScheduleApplications("test")...)
			}
		})
		ValidateApplications(contexts)

		Step("Before adding disk count number of disks in the vms", func() {
			log.InfoD("Collecting the number of disks in the VMs")

			vms, err := GetAllVMsFromScheduledContexts(contexts)
			log.FailOnError(err, "Failed to get VMs from scheduled contexts")
			for _, v := range vms {
				t := func() (interface{}, bool, error) {
					diskCountOutput, err := GetNumberOfDisksInVM(v)
					if err != nil {
						return nil, false, fmt.Errorf("failed to get number of disks in VM [%s] in namespace [%s]", v.Name, v.Namespace)
					}
					// Total disks will be numberOfVolumes plus the container disk
					if diskCountOutput != numberOfVolumes+1 {
						return nil, true, fmt.Errorf("expected number of disks in VM [%s] in namespace [%s] is [%d] but got [%d]", v.Name, v.Namespace, numberOfVolumes+1, diskCountOutput)
					}
					return diskCountOutput, false, nil
				}
				d, err := task.DoRetryWithTimeout(t, 10*time.Minute, 30*time.Second)
				log.FailOnError(err, "Failed to get number of disks in VM [%s] in namespace [%s] after retry", v.Name, v.Namespace)
				diskCount = d.(int)
				log.InfoD("Number of disks in VM [%s] in namespace [%s] is [%d]", v.Name, v.Namespace, diskCount)
			}
		})

		stepLog = "Create a PVC to be added to the kube-virtVM"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, appCtx := range contexts {
				vms, err := GetAllVMsFromScheduledContexts([]*scheduler.Context{appCtx})
				log.FailOnError(err, "Failed to get VMs from scheduled contexts")
				for _, v := range vms {
					pvcs, err := CreatePVCsForVM(v, 1, "kubevirt-sc-for-cirros-cd", "5Gi")
					log.FailOnError(err, "Failed to create PVCs for VM [%s] in namespace [%s]", v.Name, v.Namespace)
					specListInterfaces := make([]interface{}, len(pvcs))
					for i, pvc := range pvcs {
						// Converting each PVC to interface for appending to SpecList
						specListInterfaces[i] = pvc
					}
					appCtx.App.SpecList = append(appCtx.App.SpecList, specListInterfaces...)

					err = AddPVCsToVirtualMachine(v, pvcs)
					log.FailOnError(err, "Failed to add PVCs to VM [%s] in namespace [%s]", v.Name, v.Namespace)

					err = RestartKubevirtVM(v.Name, v.Namespace, true)
					log.FailOnError(err, "Failed to restart VM [%s] in namespace [%s]", v.Name, v.Namespace)

					// Perhaps moving it outside all loops is more efficient.
					log.Infof("Waiting for VM [%s] in namespace [%s] to boot. Sleeping for %v minutes...", v.Name, v.Namespace, bootTime)
					time.Sleep(bootTime)
				}
			}
		})

		stepLog = "Validate the PVCs are added to the kube-virtVM"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, appCtx := range contexts {
				vms, err := GetAllVMsFromScheduledContexts([]*scheduler.Context{appCtx})
				log.FailOnError(err, "Failed to get VMs from scheduled contexts")
				for _, v := range vms {
					t := func() (interface{}, bool, error) {
						diskCountOutput, err := GetNumberOfDisksInVM(v)
						if err != nil {
							return nil, false, fmt.Errorf("failed to get number of disks in VM [%s] in namespace [%s]", v.Name, v.Namespace)
						}
						// Total disks will be numberOfVolumes plus the container disk
						if diskCountOutput != diskCount+1 {
							return nil, true, fmt.Errorf("expected number of disks in VM [%s] in namespace [%s] is [%d] but got [%d]", v.Name, v.Namespace, diskCount+1, diskCountOutput)
						}
						return diskCountOutput, false, nil
					}
					d, err := task.DoRetryWithTimeout(t, bootTime, 30*time.Second)
					log.FailOnError(err, "Failed to get number of disks in VM [%s] in namespace [%s] after retry", v.Name, v.Namespace)
					diskCountAfterAddingDisks := d.(int)
					log.InfoD("Number of disks in VM [%s] in namespace [%s] is [%d] after adding [%d] disks", v.Name, v.Namespace, diskCountAfterAddingDisks, 1)
				}
			}
		})

	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts)
	})
})

// GetNumberOfDisksInVM returns the number of disks in the VM
func getNumberOfDisksInVM(vm v1.VirtualMachine, port int) (int, error) {
	log.InfoD("Running the command on the worker node to get the number of disks in the VM")
	workerNode := node.GetWorkerNodes()[0]
	password := "gocubsgo"
	username := "cirros"
	k8sKubevirt := kubevirt.Instance()

	vmInstance, err := k8sKubevirt.GetVirtualMachineInstance(context1.TODO(), vm.Name, vm.Namespace)
	ipAddress := vmInstance.Status.Interfaces[0].IP
	cmd := "lsblk -d | grep disk | wc -l"
	sshCmd := fmt.Sprintf("sshpass -p '%s' ssh -o StrictHostKeyChecking=no %s@%s -p %v %s", password, username, ipAddress, port, cmd)
	log.InfoD("Running the command [%s] on the worker node [%s]", sshCmd, workerNode)
	output, err := runCmd(sshCmd, workerNode)
	log.FailOnError(err, "Failed to get number of disks in VM")
	log.InfoD("Number of disks in the VM is %s", output)
	numberOfDisks, err := strconv.Atoi(output)
	if err != nil {
		log.InfoD("Failed to convert the output to integer")
		return 0, err
	}
	return numberOfDisks, nil

}
