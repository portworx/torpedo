package vsphere

const (
	// DiskAttachMode for attaching vmdk to vms
	// persistent, independent-persistent, independent-persistent
	DiskAttachMode              = "DiskAttachMode"
	diskDirectory               = "osd-provisioned-disks"
	dummyDiskName               = "kube-dummyDisk.vmdk"
	diskByIDPath                = "/dev/disk/by-id/"
	diskSCSIPrefix              = "wwn-0x"
	keepAfterDeleteVMApiVersion = "6.7.3"

	VCenterEnvKey     = "VSPHERE_VCENTER"
	VCenterPortEnvKey = "VSPHERE_VCENTER_PORT"
	UserEnvKey        = "VSPHERE_USER"
	PasswordEnvKey    = "VSPHERE_PASSWORD"
	InsecureEnvKey    = "VSPHERE_INSECURE"

	// for tests
	VMUUIDEnvKey        = "VSPHERE_VM_UUID"
	TestDatastoreEnvKey = "VSPHERE_TEST_DATASTORE"
)
