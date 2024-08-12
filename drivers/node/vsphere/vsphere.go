package vsphere

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	pxutil "github.com/libopenstorage/operator/drivers/storage/portworx/util"
	corev1 "github.com/libopenstorage/operator/pkg/apis/core/v1"
	operatorcorev1 "github.com/libopenstorage/operator/pkg/apis/core/v1"
	coreops "github.com/portworx/sched-ops/k8s/core"
	"github.com/portworx/sched-ops/k8s/operator"
	"github.com/portworx/sched-ops/task"
	"github.com/portworx/torpedo/drivers/node"
	"github.com/portworx/torpedo/drivers/node/ssh"
	"github.com/portworx/torpedo/pkg/log"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
	v1 "k8s.io/api/core/v1"
)

const (
	// DriverName is the name of the vsphere driver
	DriverName = "vsphere"
	// Protocol is the protocol used
	Protocol = "https://"
)

const (
	vsphereUname      = "VSPHERE_USER"
	vspherePwd        = "VSPHERE_PWD"
	vsphereIP         = "VSPHERE_HOST_IP"
	vsphereDatacenter = "VSPHERE_DATACENTER"
)

const (
	// DefaultUsername is the default username used for ssh operations
	DefaultUsername = "root"
	// VMReadyTimeout Timeout for checking VM power state
	VMReadyTimeout = 3 * time.Minute
	// VMReadyRetryInterval interval for retry when checking VM power state
	VMReadyRetryInterval = 5 * time.Second
)

type DriveSet struct {
	// Configs describes the configuration of the drives present in this set
	// The key is the volumeID
	Configs map[string]DriveConfig
	// NodeID is the id of the node where the drive set is being used/last
	// used
	NodeID string
	// ReservedInstanceID if set is the instance ID of the node that's attempting to transfer the driveset to itself
	ReservedInstanceID string
	// SchedulerNodeName is the name of the node in scheduler context
	SchedulerNodeName string
	// NodeIndex is the index of the node where the drive set is being
	// used/last used
	NodeIndex int
	// CreateTimestamp is the timestamp when the drive set was created
	CreateTimestamp time.Time
	// InstanceID is the cloud provider id of the instance using this drive set
	InstanceID string
	// Zone defines the zone in which the node exists
	Zone string
	// State state of the drive set from the well defined states
	State string
	// Labels associated with this drive set
	Labels *map[string]string `json:"labels"`
}

// DriveConfig defines the configuration for a cloud drive
type DriveConfig struct {
	// Type defines the type of cloud drive
	Type string
	// Size defines the size of the cloud drive in Gi
	Size int64
	// ID is the cloud drive id
	ID string
	// Path is the path where the drive is attached
	Path string
	// Iops is the iops that the drive supports
	Iops int64
	// Vpus provide a measure of disk resources available for
	// performance (IOPS/GBs) of Oracle drives.
	// Oracle uses VPU in lieu of disk types.
	Vpus int64
	// PXType indicates how this drive is being used by PX
	PXType string
	// State state of the drive config from the well defined states
	State string
	// Labels associated with this drive config
	Labels map[string]string `json:"labels"`
	// AttachOptions for cloud drives to be attached
	AttachOptions map[string]string
	// Provisioner is a name of provisioner which was used to create a drive
	Provisioner string
	// Encryption Key string to be passed in device specs
	EncryptionKeyInfo string
	// UUID of VMDK
	DiskUUID string
}

// DrivePaths stores the device paths of the disks which will be used by PX.
type DrivePaths struct {
	// Storage drives
	Storage []string
	// Journal drive
	Journal string
	// Metadata drive
	Metadata string
	// Kvdb drive
	Kvdb string
}

// Vsphere ssh driver
type vsphere struct {
	ssh.SSH
	vsphereUsername string
	vspherePassword string
	vsphereHostIP   string
	ctx             context.Context
	cancel          context.CancelFunc
}

var (
	vmMap = make(map[string]*object.VirtualMachine)
)

func (v *vsphere) String() string {
	return DriverName
}

// InitVsphere initializes the vsphere driver for ssh
func (v *vsphere) Init(nodeOpts node.InitOptions) error {
	log.Infof("Using the vsphere node driver")

	v.vsphereUsername = DefaultUsername
	username := os.Getenv(vsphereUname)
	if len(username) != 0 {
		v.vsphereUsername = username
	}

	v.vspherePassword = os.Getenv(vspherePwd)
	if len(v.vspherePassword) == 0 {
		return fmt.Errorf("Vsphere password not provided as env var: %s", vspherePwd)
	}

	v.vsphereHostIP = os.Getenv(vsphereIP)
	if len(v.vsphereHostIP) == 0 {
		return fmt.Errorf("Vsphere host IP not provided as env var: %s", vsphereIP)
	}
	err := v.connect()
	if err != nil {
		return err
	}
	err = v.SSH.Init(nodeOpts)
	if err != nil {
		return err
	}
	return nil
}

// TestConnection tests the connection to the given node
func (v *vsphere) TestConnection(n node.Node, options node.ConnectionOpts) error {
	var err error
	log.Infof("Testing vsphere driver connection by checking state of the VMs in the vsphere")
	//Reestablishing the connection where we saw session getting NotAuthenticated issue in Longevity
	if err = v.connect(); err != nil {
		return err
	}
	// If n.Name is not in vmMap after the first attempt, wait and try to connect again.
	if _, ok := vmMap[n.Name]; !ok {
		time.Sleep(2 * time.Minute) // Wait for 2 minutes before retrying.

		// Attempt to reconnect.
		if err = v.connect(); err != nil {
			return err
		}

		// Check if n.Name is in vmMap after reconnecting.
		if _, ok = vmMap[n.Name]; !ok {
			return fmt.Errorf("Failed to get VM: %s", n.Name)
		}
	}
	vm := vmMap[n.Name]
	cmd := "hostname"
	t := func() (interface{}, bool, error) {
		powerState, err := vm.PowerState(v.ctx)
		log.Infof("Power state of VM : %s state %v ", vm.Name(), powerState)
		if err != nil || powerState != types.VirtualMachinePowerStatePoweredOn {
			return nil, true, &node.ErrFailedToTestConnection{
				Node:  n,
				Cause: fmt.Sprintf("Failed to test connection to VM: %s Current Status: %v, error: %v", vm.Name(), powerState, err),
			}
		}

		return nil, false, nil
	}
	if _, err := task.DoRetryWithTimeout(t, VMReadyTimeout, VMReadyRetryInterval); err != nil {
		return err
	}
	// Check if VM is not just powered on but also usable
	_, err = v.RunCommand(n, cmd, node.ConnectionOpts{
		Timeout:         VMReadyTimeout,
		TimeBeforeRetry: VMReadyRetryInterval,
	})
	return err
}

// getVMFinder return find.Finder instance
func (v *vsphere) getVMFinder() (*find.Finder, error) {
	login := fmt.Sprintf("%s%s:%s@%s/sdk", Protocol, v.vsphereUsername, v.vspherePassword, v.vsphereHostIP)
	log.Infof("Logging in to Virtual Center using: %s", login)
	u, err := url.Parse(login)
	if err != nil {
		return nil, fmt.Errorf("error parsing url %s", login)
	}

	v.ctx, v.cancel = context.WithCancel(context.Background())
	//defer cancel()

	c, err := govmomi.NewClient(v.ctx, u, true)
	if err != nil {
		return nil, fmt.Errorf("logging in error: %s", err.Error())
	}
	log.Infof("Log in successful to vsphere:  %s:\n", v.vsphereHostIP)

	f := find.NewFinder(c.Client, true)

	// Find one and only datacenter
	dc, err := f.DefaultDatacenter(v.ctx)
	if err != nil && !strings.Contains(err.Error(), "default datacenter resolves to multiple instances") {
		return nil, fmt.Errorf("Failed to find data center: %v", err)
	}
	if dc == nil {
		vcDatacenter := os.Getenv(vsphereDatacenter)
		if len(vcDatacenter) == 0 {
			return nil, fmt.Errorf("default datacenter resolves to multiple instances. please pass env variable VSPHERE_DATACENTER")
		}

		dcList, err := f.DatacenterList(v.ctx, "*")
		if err != nil {
			return nil, err
		}

		for _, dcObj := range dcList {
			log.Debugf("checking dc name: %s", dcObj.Name())
			if dcObj.Name() == vcDatacenter {
				dc = dcObj
				break
			}
		}

	}
	// Make future calls local to this datacenter
	f.SetDatacenter(dc)
	return f, nil

}

// GetCompatibleDatastores get matching prefix datastores
func (v *vsphere) GetCompatibleDatastores(portworxNamespace string, datastoreNames []string) ([]*object.Datastore, error) {
	var err error
	datastores, err := v.GetDatastoresFromDatacenter()
	if err != nil {
		return nil, err
	}
	var stc *operatorcorev1.StorageCluster
	pxOperator := operator.Instance()
	stcList, err := pxOperator.ListStorageClusters(portworxNamespace)
	if err != nil {
		return nil, fmt.Errorf("Failed to find storage clusters %v ", err)
	}
	var selectedDatastore []*object.Datastore
	stc, err = pxOperator.GetStorageCluster(stcList.Items[0].Name, stcList.Items[0].Namespace)
	if err != nil {
		return nil, fmt.Errorf("Failed to find storage cluster %v  in namespace  %s ", err, portworxNamespace)
	}
	var envVariables []v1.EnvVar
	envVariables = stc.Spec.CommonConfig.Env
	var prefixName string
	for _, envVar := range envVariables {
		if envVar.Name == "VSPHERE_DATASTORE_PREFIX" {
			prefixName = envVar.Value
			log.Infof("prefixName   %s ", prefixName)
		}
	}
	if prefixName == "" {
		return nil, fmt.Errorf("Failed to find VSPHERE_DATASTORE_PREFIX  prefix ")
	}
	for _, ds := range datastores {
		if strings.HasPrefix(ds.Name(), prefixName) {
			log.Infof("Prefix match found for datastore Name %v ", ds.Name())
			selectedDatastore = append(selectedDatastore, ds)
		}
	}
	if len(selectedDatastore) == 0 {
		return nil, fmt.Errorf("All datastores are not available, available are  %v , but expected are : %v", selectedDatastore, datastoreNames)
	}
	return selectedDatastore, nil
}

func (v *vsphere) GetDatastoresFromDatacenter() ([]*object.Datastore, error) {
	var finder *find.Finder
	finder, err := v.getVMFinder()
	if err != nil {
		return nil, fmt.Errorf("Failed to find  getVMFinder err:: %+v", err)
	}
	datastores, err := finder.DatastoreList(v.ctx, "*")
	if err != nil {
		return nil, fmt.Errorf("Failed to get all the datastores. err: %+v", err)
	}
	return datastores, nil
}

func (v *vsphere) connect() error {
	var f *find.Finder

	// Getting finder instance
	f, err := v.getVMFinder()
	if err != nil {
		return err
	}
	// vmMap Reset to get the new valid VMs info.
	vmMap = make(map[string]*object.VirtualMachine)
	// Find virtual machines in datacenter
	vms, err := f.VirtualMachineList(v.ctx, "*")

	if err != nil {
		return fmt.Errorf("failed to find any virtual machines on %s: %v", v.vsphereHostIP, err)
	}

	nodes := node.GetNodes()
	if nodes == nil {
		return fmt.Errorf("nodes not found")
	}

	for _, vm := range vms {
		var vmMo mo.VirtualMachine
		err = vm.Properties(v.ctx, vm.Reference(), []string{"guest"}, &vmMo)
		if err != nil {
			re, regErr := regexp.Compile(".*has already been deleted or has not been completely created.*")
			if regErr != nil {
				return regErr
			}
			if re.MatchString(fmt.Sprintf("%v", err)) {
				log.Errorf("%v", err)
				continue
			} else {
				log.Errorf("failed to get properties: %v", err)
				return err
			}
		}

		// Get the hostname
		hostname := vmMo.Guest.HostName
		if hostname == "" {
			continue
		}
		log.Debugf("hostname for vm %v: %v", vm.Name(), hostname)

		for _, n := range nodes {
			if hostname == n.Name {
				if _, ok := vmMap[hostname]; !ok {
					vmMap[hostname] = vm
				}
			}
		}
	}
	return nil
}

// DetachDisk vdisks from node.
func (v *vsphere) DetachDrivesFromVM(stc *corev1.StorageCluster, nodeName string) error {
	configData, err := GetCloudDriveConfigmapData(stc)
	if err != nil {
		err = fmt.Errorf("Failed to find configData: err %w", err)
		return err
	}
	//Find out the instance VMUUID and then dettach.
	for _, nodeConfigData := range configData {
		if nodeName == nodeConfigData.SchedulerNodeName {
			allDiskPaths := GetDiskPaths(nodeConfigData)
			instanceId := nodeConfigData.InstanceID
			for i := 0; i < len(allDiskPaths); i++ {
				log.Infof("Diskpath for %v is %v and instance id is %v", nodeConfigData.NodeID, allDiskPaths[i], instanceId)
				err = v.DetachDisk(instanceId, allDiskPaths[i])
				if err != nil {
					//log.InfoD("Detach drives from the node failed %v", err)
					err = fmt.Errorf("Detaching disk: %s on node %s failed: %w", allDiskPaths[i], nodeName, err)
					return err
				}
			}
		} else {
			log.Infof(" Node Name from config %s, expected %s ", nodeConfigData.SchedulerNodeName, nodeName)
		}
	}
	return nil
}

func (v *vsphere) DetachDisk(vmUuid string, path string) error {
	// Getting finder instance
	f, err := v.getVMFinder()
	if err != nil {
		return err
	}
	// vmMap Reset to get the new valid VMs info.
	vmMap = make(map[string]*object.VirtualMachine)
	// Find virtual machines in datacenter
	vms, err := f.VirtualMachineList(v.ctx, "*")
	var vmMo *object.VirtualMachine
	if err != nil {
		return fmt.Errorf("failed to find any virtual machines on %s: %v", v.vsphereHostIP, err)
	}
	for _, vm := range vms {
		if vm.UUID(v.ctx) == vmUuid {
			//Found
			vmMo = vm
			log.Infof("VM found %v", vm)
			break
		}
	}
	//Error if not found
	if vmMo == nil {
		return fmt.Errorf("Virtual machine not found")
	}
	//Remove device and detach VM
	var selectedDevice types.BaseVirtualDevice
	deviceList, err := vmMo.Device(v.ctx)
	if err != nil {
		return fmt.Errorf("Failed to get the devices for VM: %q. err: %+v", vmMo, err)
	}
	log.Infof("All devices %v", deviceList)
	// filter vm devices to retrieve device for the given vmdk file identified by disk path
	for _, device := range deviceList {
		if deviceList.TypeName(device) == "VirtualDisk" {
			virtualDevice := device.GetVirtualDevice()
			if backing, ok := virtualDevice.Backing.(*types.VirtualDiskFlatVer2BackingInfo); ok {
				if matchVirtualDiskAndVolPath(backing.FileName, path) {
					log.Infof("Found VirtualDisk backing with filename %q for diskPath %q", backing.FileName, path)
					selectedDevice = device
				}
			}
		}
	}
	if selectedDevice != nil {
		log.Infof("Selected device %v", selectedDevice)
		return vmMo.RemoveDevice(v.ctx, true, selectedDevice)
	}
	return fmt.Errorf("No device selected for VM: %q", vmMo)
}

// Match the paths between fileNamePath and absolute vmdk path
func matchVirtualDiskAndVolPath(diskPath, volPath string) bool {
	diskPath = strings.TrimSuffix(diskPath, filepath.Ext(diskPath))
	volPath = strings.TrimSuffix(volPath, filepath.Ext(volPath))
	return diskPath == volPath
}

// Get virtual disk path.
// TODO need to filter only of type: DrivePaths
func GetDiskPaths(driveset DriveSet) []string {
	diskPaths := []string{}
	for vmdkPath, configs := range driveset.Configs {
		//TODO need to change later
		log.InfoD("PX type %s ", configs.PXType)
		if configs.PXType == "data" {
			diskPath := vmdkPath
			datastore := GetDatastore(configs)
			openBracketIndex := strings.Index(diskPath, "[")
			closeBracketIndex := strings.Index(diskPath, "]")
			// Extract the substring inside the square brackets
			substring := diskPath[openBracketIndex+1 : closeBracketIndex]
			// Replace the substring inside the square brackets with datastore
			diskPath = strings.Replace(diskPath, substring, datastore, 1)
			diskPaths = append(diskPaths, diskPath)
			log.Infof("diskPath %s is of type data ", diskPath)
		}
	}
	return diskPaths
}

// GetDatastore
func GetDatastore(configs DriveConfig) string {
	for key, val := range configs.Labels {
		if key == "datastore" {
			return val
		}
	}
	return ""
}

// GetCloudDriveConfigmapData Get clouddrive configMap data.
func GetCloudDriveConfigmapData(cluster *corev1.StorageCluster) (map[string]DriveSet, error) {
	cloudDriveConfigmapName := pxutil.GetCloudDriveConfigMapName(cluster)
	var PortworxNamespace = "kube-system"
	cloudDriveConfifmap, _ := coreops.Instance().GetConfigMap(cloudDriveConfigmapName, PortworxNamespace)
	var configData map[string]DriveSet
	err := json.Unmarshal([]byte(cloudDriveConfifmap.Data["cloud-drive"]), &configData)
	if err != nil {
		return nil, err
	}
	return configData, nil
}

// AddVM adds a new VM object to vmMap
func (v *vsphere) AddMachine(vmName string) error {
	var f *find.Finder

	log.Infof("Adding VM: %s into vmMap  ", vmName)

	f, err := v.getVMFinder()
	if err != nil {
		return err
	}

	vm, err := f.VirtualMachine(v.ctx, vmName)
	if err != nil {
		return err
	}

	var vmMo mo.VirtualMachine
	err = vm.Properties(v.ctx, vm.Reference(), []string{"guest.hostName"}, &vmMo)
	if err != nil {
		return err
	}

	if vmMo.Guest == nil {
		return fmt.Errorf("failed to find guest info for virtual machine %s", vmName)
	}

	// Get the hostname
	hostname := vmMo.Guest.HostName
	log.Debugf("hostname: %v", hostname)
	if hostname == "" {
		return fmt.Errorf("Failed to find hostname for  virtual machine on %s: %v", vm.Name(), err)
	}

	vmMap[hostname] = vm
	return nil
}

// RebootVM reboots vsphere VM
func (v *vsphere) RebootNode(n node.Node, options node.RebootNodeOpts) error {
	//Reestblish connection to avoid session timeout.
	err := v.connect()
	if err != nil {
		return err
	}
	if _, ok := vmMap[n.Name]; !ok {
		return fmt.Errorf("could not fetch VM for node: %s", n.Name)
	}

	vm := vmMap[n.Name]
	log.Infof("Rebooting VM: %s  ", vm.Name())
	err = vm.RebootGuest(v.ctx)
	if err != nil {
		return &node.ErrFailedToRebootNode{
			Node:  n,
			Cause: fmt.Sprintf("failed to reboot VM %s. cause %v", vm.Name(), err),
		}
	}
	return nil
}

// powerOnVM powers on VM by providing VM object
func (v *vsphere) powerOnVM(vm *object.VirtualMachine) error {

	// Checking the VM state before powering it On
	powerState, err := vm.PowerState(v.ctx)
	if err != nil {
		return err
	}

	if powerState == types.VirtualMachinePowerStatePoweredOn {
		log.Warn("VM is already in powered-on state: ", vm.Name())
		return nil
	}

	tsk, err := vm.PowerOn(v.ctx)
	if err != nil {
		return fmt.Errorf("failed to power on %s: %v", vm.Name(), err)
	}
	if _, err := tsk.WaitForResult(v.ctx); err != nil {
		return fmt.Errorf("failed to power on VM %s. cause %v", vm.Name(), err)
	}
	return nil
}

// PowerOnVM powers on the VM if not already on
func (v *vsphere) PowerOnVM(n node.Node) error {
	var err error
	//Reestblish connection to avoid session timeout.
	err = v.connect()
	if err != nil {
		return err
	}

	vm, ok := vmMap[n.Name]

	if !ok {
		return fmt.Errorf("could not fetch VM for node: %s to power on", n.Name)
	}
	log.Infof("Powering on VM: %s  ", vm.Name())
	if err = v.powerOnVM(vm); err != nil {
		return &node.ErrFailedToRebootNode{
			Node:  n,
			Cause: fmt.Sprintf("failed to power on VM %s. cause %v", vm.Name(), err),
		}
	}

	return nil
}

// PowerOnVMByName powers on VM by using name
func (v *vsphere) PowerOnVMByName(vmName string) error {
	// Make sure vmName is part of vmMap before using this method

	var err error
	//Reestablish connection to avoid session timeout.
	err = v.connect()
	if err != nil {
		return err
	}
	vm, ok := vmMap[vmName]

	if !ok {
		//this is to handle the case for OCP set up where we add nodes to vmMap before adding to storage nodes list
		err = v.AddMachine(vmName)
		if err != nil {
			return err
		}
	}
	vm, ok = vmMap[vmName]
	if !ok {
		return fmt.Errorf("could not fetch VM for node: %s to power on", vmName)
	}

	log.Infof("Powering on VM: %s  ", vm.Name())
	if err = v.powerOnVM(vm); err != nil {
		return err
	}
	return nil
}

// PowerOffVM powers off the VM if not already off
func (v *vsphere) PowerOffVM(n node.Node) error {
	var err error
	//Reestblish connection to avoid session timeout.
	err = v.connect()
	if err != nil {
		return err
	}
	vm, ok := vmMap[n.Name]
	if !ok {
		return fmt.Errorf("could not fetch VM for node: %s to power off", n.Name)
	}

	log.Infof("\nPowering off VM: %s  ", vm.Name())
	tsk, err := vm.PowerOff(v.ctx)
	if err != nil {
		return fmt.Errorf("Failed to power off %s: %v", vm.Name(), err)
	}
	if _, err := tsk.WaitForResult(v.ctx); err != nil {
		return &node.ErrFailedToShutdownNode{
			Node:  n,
			Cause: fmt.Sprintf("failed to power off  VM %s. cause %v", vm.Name(), err),
		}
	}

	return nil
}

// DestroyVM powers off the VM if not already off
func (v *vsphere) DestroyVM(n node.Node) error {
	var err error
	//Reestblish connection to avoid session timeout.
	err = v.connect()
	if err != nil {
		return err
	}
	vm, ok := vmMap[n.Name]
	if !ok {
		return fmt.Errorf("could not fetch VM for node: %s to destroy", n.Name)
	}

	log.Infof("\nDestroying VM: %s  ", vm.Name())
	tsk, err := vm.Destroy(v.ctx)

	if err != nil {
		return fmt.Errorf("Failed to destroy %s: %v", vm.Name(), err)
	}
	if _, err := tsk.WaitForResult(v.ctx); err != nil {

		return &node.ErrFailedToDeleteNode{
			Node:  n,
			Cause: fmt.Sprintf("failed to destroy VM %s. cause %v", vm.Name(), err),
		}
	}

	return nil
}

// ShutdownNode shutsdown the vsphere VM
func (v *vsphere) ShutdownNode(n node.Node, options node.ShutdownNodeOpts) error {
	//Reestblish connection to avoid session timeout.
	err := v.connect()
	if err != nil {
		return err
	}
	if _, ok := vmMap[n.Name]; !ok {
		return fmt.Errorf("Could not fetch VM for node: %s", n.Name)
	}

	vm, ok := vmMap[n.Name]
	if !ok {
		return fmt.Errorf("could not fetch VM for node: %s to shutdown", n.Name)
	}

	log.Infof("Shutting down VM: %s  ", vm.Name())
	err = vm.ShutdownGuest(v.ctx)
	if err != nil {
		return &node.ErrFailedToShutdownNode{
			Node:  n,
			Cause: fmt.Sprintf("failed to shutdown VM %s. cause %v", vm.Name(), err),
		}
	}
	return nil
}

func init() {
	v := &vsphere{
		SSH: *ssh.New(),
	}

	node.Register(DriverName, v)
}

func (v *vsphere) GetSupportedDriveTypes() ([]string, error) {
	return []string{"thin", "zeroedthick", "eagerzeroedthick", "lazyzeroedthick"}, nil
}

// MoveDisks detaches all disks from the source VM and attaches it to the target
func (v *vsphere) MoveDisks(sourceNode node.Node, targetNode node.Node) error {
	// Reestablish connection to avoid session timeout.
	err := v.connect()
	if err != nil {
		return err
	}

	sourceVM, ok := vmMap[sourceNode.Name]
	if !ok {
		return fmt.Errorf("could not fetch VM for node: %s", sourceNode.Name)
	}

	targetVM, ok := vmMap[targetNode.Name]
	if !ok {
		return fmt.Errorf("could not fetch VM for node: %s", targetNode.Name)
	}

	devices, err := sourceVM.Device(v.ctx)
	if err != nil {
		return err
	}

	// Detach disks from source VM and attach to destination VM
	var disks []*types.VirtualDisk
	for _, device := range devices {
		if disk, ok := device.(*types.VirtualDisk); ok {
			// skip the first/root disk
			if *disk.UnitNumber == 0 {
				continue
			}
			disks = append(disks, disk)

			config := &types.VirtualMachineConfigSpec{
				DeviceChange: []types.BaseVirtualDeviceConfigSpec{
					&types.VirtualDeviceConfigSpec{
						Operation: types.VirtualDeviceConfigSpecOperationRemove,
						Device:    disk,
					},
				},
			}
			log.Debugf("Detaching disk %s from VM %s", disk.DeviceInfo.GetDescription().Label, sourceVM.Name())
			event, err := sourceVM.Reconfigure(v.ctx, *config)
			if err != nil {
				return err
			}

			err = event.Wait(v.ctx)
			if err != nil {
				return err
			}
		}
	}

	for _, disk := range disks {
		config := &types.VirtualMachineConfigSpec{
			DeviceChange: []types.BaseVirtualDeviceConfigSpec{
				&types.VirtualDeviceConfigSpec{
					Operation: types.VirtualDeviceConfigSpecOperationAdd,
					Device:    disk,
				},
			},
		}
		log.Debugf("Attaching disk %s to VM %s", disk.DeviceInfo.GetDescription().Label, targetVM.Name())
		event, err := targetVM.Reconfigure(v.ctx, *config)
		if err != nil {
			return err
		}

		err = event.Wait(v.ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

// RemoveNonRootDisks removes all disks except the root disk from the VM
func (v *vsphere) RemoveNonRootDisks(n node.Node) error {
	// Reestablish connection to avoid session timeout.
	err := v.connect()
	if err != nil {
		return err
	}

	vm, ok := vmMap[n.Name]
	if !ok {
		return fmt.Errorf("could not fetch VM for node: %s", n.Name)
	}

	devices, err := vm.Device(v.ctx)
	if err != nil {
		return err
	}

	for _, device := range devices {
		if disk, ok := device.(*types.VirtualDisk); ok {
			// skip the first/root disk
			if *disk.UnitNumber == 0 {
				continue
			}
			log.Debugf("Deleting disk %s from VM %s", disk.DeviceInfo.GetDescription().Label, vm.Name())
			err = vm.RemoveDevice(v.ctx, false, disk)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// StorageVmotion relocates the largest disks of a VM from one datastore to another within the same prefix group
// With moveAllDisks true we will be moving all disks attached to a VM onto same Datastore
// If moveAllDisks is set to False then we will choose the largest sized disk and move that only to a new Datastore
func (v *vsphere) StorageVmotion(ctx context.Context, node node.Node, portworxNamespace string, moveAllDisks bool) error {
	log.Infof("Trying to find the VM on vSphere: %v", node.Name)
	vm, err := v.FindVMByIP(node)
	if err != nil {
		return fmt.Errorf("error retrieving VM: %v", err)
	}

	var vmProps mo.VirtualMachine
	err = vm.Properties(ctx, vm.Reference(), []string{"config.hardware"}, &vmProps)
	if err != nil {
		return fmt.Errorf("error retrieving VM properties: %v", err)
	}

	log.Infof("Trying to fetch all compatible Datastores with the prefix that is set in Px Storage Class spec")
	compatibleDatastores, err := v.GetCompatibleDatastores(portworxNamespace, []string{})
	if err != nil {
		return fmt.Errorf("error retrieving compatible datastores: %v", err)
	}

	var stc *operatorcorev1.StorageCluster
	pxOperator := operator.Instance()
	stcList, err := pxOperator.ListStorageClusters(portworxNamespace)
	if err != nil {
		return fmt.Errorf("Failed to find storage clusters %v ", err)
	}
	stc, err = pxOperator.GetStorageCluster(stcList.Items[0].Name, stcList.Items[0].Namespace)
	if err != nil {
		return fmt.Errorf("Failed to find storage cluster %v  in namespace  %s ", err, portworxNamespace)
	}

	preData, err := GetCloudDriveConfigmapData(stc)
	if err != nil {
		return fmt.Errorf("error fetching pre-vMotion cloud drive config: %v", err)
	}

	originalDatastoreIDs := map[string]string{}
	for _, device := range vmProps.Config.Hardware.Device {
		if disk, ok := device.(*types.VirtualDisk); ok {
			if disk.DeviceInfo != nil {
				if backing, ok := disk.Backing.(*types.VirtualDiskFlatVer2BackingInfo); ok {
					originalDatastoreIDs[disk.DeviceInfo.GetDescription().Label] = backing.Datastore.Value
				}
			}
		}
	}

	var targetDatastores []*object.Datastore

	largestDisks := findLargestDisksOnDatastores(vmProps.Config.Hardware.Device, compatibleDatastores)
	if len(largestDisks) == 0 {
		return fmt.Errorf("no large disks found on specified prefix datastores")
	}

	sourceDatastore := object.NewDatastore(vm.Client(), largestDisks[0].Datastore)

	targetDatastores, err = filterTargetDatastores(ctx, sourceDatastore, compatibleDatastores, &vmProps)
	if err != nil {
		return fmt.Errorf("error filtering target datastores: %v", err)
	}

	if !moveAllDisks {
		log.Infof("Trying to Move largest disk on VM %v from Datastore %v to Datastore : %v", node.Name, sourceDatastore.Name(), targetDatastores[0].Name())
		err = initiateStorageVmotion(ctx, vm, largestDisks[:1], targetDatastores)
		if err != nil {
			return fmt.Errorf("error during storage vMotion: %v", err)
		}
	} else {
		log.Infof("Trying to Move all disks of %v from Datastore %v to Datastore : %v", node.Name, sourceDatastore.Name(), targetDatastores[0].Name())
		diskLocators := make([]types.VirtualMachineRelocateSpecDiskLocator, 0)
		for _, device := range vmProps.Config.Hardware.Device {
			if disk, ok := device.(*types.VirtualDisk); ok {
				diskLocators = append(diskLocators, types.VirtualMachineRelocateSpecDiskLocator{
					DiskId:    disk.Key,
					Datastore: targetDatastores[0].Reference(),
				})
			}
		}
		if len(diskLocators) == 0 {
			return fmt.Errorf("no disks found on the VM")
		}
		log.Infof("Going to trigger Storage Vmotion for %v", node.Name)
		err = initiateStorageVmotion(ctx, vm, diskLocators, targetDatastores)
		if err != nil {
			return fmt.Errorf("error during storage vMotion: %v", err)
		}
		log.Infof("Sleeping for a minute to let config map be updated with latest changes")
		time.Sleep(1 * time.Minute)
		postData, err := GetCloudDriveConfigmapData(stc)
		if err != nil {
			return fmt.Errorf("error fetching post-vMotion cloud drive config: %v", err)
		}
		if !v.ValidateDatastoreUpdate(preData, postData, node.VolDriverNodeID, targetDatastores[0].Reference().Value) {
			return fmt.Errorf("validation failed: datastore updates are not as expected")
		}
	}
	return nil
}

// Function to get the datastore's cluster
func getDatastoreCluster(ctx context.Context, ds *object.Datastore) (*object.StoragePod, error) {
	var dsProps mo.Datastore
	err := ds.Properties(ctx, ds.Reference(), []string{"parent"}, &dsProps)
	if err != nil {
		return nil, fmt.Errorf("failed to get properties for datastore: %v", err)
	}

	if dsProps.Parent != nil {
		spRef := dsProps.Parent.Reference()
		if spRef.Type == "StoragePod" {
			return object.NewStoragePod(ds.Client(), spRef), nil
		}
	}

	return nil, nil
}

// filterTargetDatastores filters from list of Datastores available with the prefix defined in Storageclass
// If source DS is in a cluster with another DS, then this method skips the other DS and chooses a DS outside this cluster
func filterTargetDatastores(ctx context.Context, sourceDatastore *object.Datastore, allDatastores []*object.Datastore, vmProps *mo.VirtualMachine) ([]*object.Datastore, error) {
	sourceCluster, err := getDatastoreCluster(ctx, sourceDatastore)
	if err != nil {
		return nil, err
	}
	var filteredDatastores []*object.Datastore

	totalDiskSizeToMove := int64(0)
	for _, device := range vmProps.Config.Hardware.Device {
		if disk, ok := device.(*types.VirtualDisk); ok {
			totalDiskSizeToMove += disk.CapacityInKB
		}
	}

	for _, ds := range allDatastores {
		if ds.Reference().Value == sourceDatastore.Reference().Value {
			continue
		}

		targetCluster, err := getDatastoreCluster(ctx, ds)
		if err != nil {
			continue
		}

		var dsProps mo.Datastore
		err = ds.Properties(ctx, ds.Reference(), []string{"summary"}, &dsProps)
		if err != nil {
			log.Errorf("Failed to get properties for datastore %v: %v", ds.Name(), err)
			continue
		}

		availableSpace := dsProps.Summary.FreeSpace / 1024
		spaceAfterMove := availableSpace - totalDiskSizeToMove
		maxAllowedUsage := (dsProps.Summary.Capacity / 1024) * 90 / 100
		log.Infof("Available Space is: %v", availableSpace)
		log.Infof("Total Disk size to move is: %v", totalDiskSizeToMove)
		log.Infof("Space after move is: %v", spaceAfterMove)
		log.Infof("Max Allowed Usage is: %v", maxAllowedUsage)

		if spaceAfterMove < 0 || ((availableSpace - spaceAfterMove) < maxAllowedUsage) {
			log.Infof("Datastore %v does not have enough space or will exceed 90 percent capacity.", ds.Name())
			continue
		}

		if sourceCluster == nil || targetCluster == nil || sourceCluster.Reference() != targetCluster.Reference() {
			filteredDatastores = append(filteredDatastores, ds)
		}
	}
	if len(filteredDatastores) == 0 {
		return nil, fmt.Errorf("no suitable datastores found that meet the space requirements")
	}
	return filteredDatastores, nil
}

// findLargestDisksOnDatastores identifies the largest disks on the specified datastores for a VM
func findLargestDisksOnDatastores(devices []types.BaseVirtualDevice, datastores []*object.Datastore) []types.VirtualMachineRelocateSpecDiskLocator {
	datastoreMap := make(map[types.ManagedObjectReference]*object.Datastore)
	for _, ds := range datastores {
		datastoreMap[ds.Reference()] = ds
	}

	var disks []struct {
		Disk     *types.VirtualDisk
		Capacity int64
	}

	for _, device := range devices {
		if disk, ok := device.(*types.VirtualDisk); ok {
			if backingInfo, ok := disk.Backing.(*types.VirtualDiskFlatVer2BackingInfo); ok {
				dsRef := backingInfo.Datastore
				if _, exists := datastoreMap[*dsRef]; exists {
					disks = append(disks, struct {
						Disk     *types.VirtualDisk
						Capacity int64
					}{
						Disk: disk, Capacity: disk.CapacityInKB,
					})
				}
			} else {
				log.Infof("Disk backing type assertion failed")
			}
		}
	}

	sort.Slice(disks, func(i, j int) bool {
		return disks[i].Capacity > disks[j].Capacity
	})

	diskLocators := []types.VirtualMachineRelocateSpecDiskLocator{}
	for _, disk := range disks {
		if backing, ok := disk.Disk.Backing.(*types.VirtualDiskFlatVer2BackingInfo); ok {
			diskLocator := types.VirtualMachineRelocateSpecDiskLocator{
				DiskId:    disk.Disk.Key,
				Datastore: *backing.Datastore,
			}
			diskLocators = append(diskLocators, diskLocator)
		} else {
			log.Infof("Disk backing type is not compatible or assertion failed\n")
		}
	}
	return diskLocators
}

// initiateStorageVmotion starts the Storage vMotion process for the disks, targeting a specific datastore.
func initiateStorageVmotion(ctx context.Context, vm *object.VirtualMachine, diskLocators []types.VirtualMachineRelocateSpecDiskLocator, datastores []*object.Datastore) error {
	var targetDatastore *object.Datastore
	if len(datastores) == 0 {
		return fmt.Errorf("no compatible datastores available for storage vMotion")
	}
	if len(datastores) > 1 {
		maxAvailableSpace := int64(-1)
		for _, ds := range datastores {
			var dsProps mo.Datastore
			if err := ds.Properties(ctx, ds.Reference(), []string{"summary"}, &dsProps); err == nil {
				if available := dsProps.Summary.FreeSpace; available > maxAvailableSpace {
					maxAvailableSpace = available
					targetDatastore = ds
				}
			}
		}
	} else {
		targetDatastore = datastores[0]
	}

	if targetDatastore == nil {
		return fmt.Errorf("failed to select a target datastore")
	}

	for i := range diskLocators {
		diskLocators[i].Datastore = targetDatastore.Reference()
	}

	relocateSpec := types.VirtualMachineRelocateSpec{
		Disk: diskLocators,
	}

	task, err := vm.Relocate(ctx, relocateSpec, types.VirtualMachineMovePriorityDefaultPriority)
	if err != nil {
		return fmt.Errorf("error initiating VM relocate: %v", err)
	}
	return task.Wait(ctx)
}

// FindVMByName finds a virtual machine by its name.
func (v *vsphere) FindVMByName(vmName string) (*object.VirtualMachine, error) {
	log.Infof("VM Name is: %v", vmName)
	finder, err := v.getVMFinder()
	if err != nil {
		return nil, err
	}
	vm, err := finder.VirtualMachine(v.ctx, vmName)
	if err != nil {
		return nil, fmt.Errorf("failed to find VM: %v", err)
	}
	return vm, nil
}

// FindDatastoreByName finds a datastore by its name.
func (v *vsphere) FindDatastoreByName(dsName string) (*object.Datastore, error) {
	finder, err := v.getVMFinder()
	if err != nil {
		return nil, err
	}
	ds, err := finder.Datastore(v.ctx, dsName)
	if err != nil {
		return nil, fmt.Errorf("failed to find datastore: %v", err)
	}
	return ds, nil
}

// FindVMByIP finds the vsphere Name of a VM through its Data IP Address
func (v *vsphere) FindVMByIP(node node.Node) (*object.VirtualMachine, error) {
	log.Infof("Searching VM by IP Addresses: %v", node.Addresses)
	finder, err := v.getVMFinder()
	if err != nil {
		return nil, fmt.Errorf("failed to get VM finder: %v", err)
	}

	vms, err := finder.VirtualMachineList(v.ctx, "*")
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve VM list: %v", err)
	}

	for _, vm := range vms {
		var vmProps mo.VirtualMachine
		err = vm.Properties(v.ctx, vm.Reference(), []string{"guest.net"}, &vmProps)
		if err != nil {
			log.Infof("failed to get properties for VM: %s, error: %v", vm.Name(), err)
			continue
		}

		for _, net := range vmProps.Guest.Net {
			for _, ip := range net.IpAddress {
				for _, nodeIP := range node.Addresses {
					if ip == nodeIP {
						log.Infof("Found VM by IP Address %s: %s", ip, vm.Name())
						return vm, nil
					}
				}
			}
		}
	}
	return nil, fmt.Errorf("no VM found with the given IP addresses: %v", node.Addresses)
}

// ValidateDatastoreUpdate validates the cloud drive configmap after Storage vmotion is done
func (v *vsphere) ValidateDatastoreUpdate(preData, postData map[string]DriveSet, nodeUUID string, targetDatastoreID string) bool {
	preNodeData, preExists := preData[nodeUUID]
	postNodeData, postExists := postData[nodeUUID]

	if !preExists || !postExists {
		return false
	}

	postDiskDatastores := make(map[string]string)
	for _, postDrive := range postNodeData.Configs {
		postDiskDatastores[postDrive.DiskUUID] = postDrive.Labels["datastore"]
	}

	allMoved := true
	for _, preDrive := range preNodeData.Configs {
		postDSName, exists := postDiskDatastores[preDrive.DiskUUID]
		if !exists {
			log.Infof("No post-migration data found for disk with UUID %v", preDrive.DiskUUID)
			return false
		}

		postDS, err := v.FindDatastoreByName(postDSName)
		if err != nil {
			log.Errorf("Failed to find datastore with name %v: %v", postDSName, err)
			return false
		}
		postDSID := postDS.Reference().Value

		if !(postDSID == targetDatastoreID || (preDrive.Labels["datastore"] == postDSName && postDSID == targetDatastoreID)) {
			log.Infof("Disk with UUID %v did not move to the target datastore %v as expected, or was not already there. This is for Node %v", preDrive.DiskUUID, targetDatastoreID, nodeUUID)
			allMoved = false
		}
	}

	if allMoved {
		log.Infof("Storage vMotion happened successfully for all disks")
	}

	return allMoved
}
