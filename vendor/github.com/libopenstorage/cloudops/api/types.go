package api

import (
	"time"
)

// CloudVolumeSpec defines volume spec.
type CloudVolumeSpec struct {
	// Name volume name
	Name *string `locationName:"name" type:"string" required:"true"`
	// Type the volume type (e.g. GP2 on AWS)
	Type *string `locationName:"type" type:"string" required:"true"`
	// Size in GiBs.
	Size *int64 `locationName:"size" type:"string" required:"true"`
	// Region in which to create the volume.
	Region *string `locationName:"region" type:"string" required:"false"`
	// Zone in which to create the volume.
	Zone *string `locationName:"zone" type:"string" required:"true"`
	// Iops desired IOPS.
	Iops *int64 `type:"integer" required:"false"`
	// Encrypted specifies whether the volume should be encrypted.
	Encrypted *bool `locationName:"encrypted" type:"boolean" required:"false"`
	// EncryptionKeyID identifier for the Key Management Service
	EncryptionKeyID *string `locationName:"encryption_key_id" type:"string" required:"false"`
	// SnapshotID the snapshot from which to create the volume. Optional
	SnapshotID *string `locationName:"snapshot_id" type:"string" required:"false"`
	// The tags to apply to the volume during creation. Optional
	Labels map[string]string `locationName:"Labels" required:"false"`
}

// VolumeAttachmentState enum for current volume attachment state.
type VolumeAttachmentState string

const (
	// VolumeAttachmentStateAttaching volume is attaching.
	VolumeAttachmentStateAttaching VolumeAttachmentState = "attaching"
	// VolumeAttachmentStateAttached volume is attached.
	VolumeAttachmentStateAttached VolumeAttachmentState = "attached"
	// VolumeAttachmentStateDetaching volume is detaching.
	VolumeAttachmentStateDetaching VolumeAttachmentState = "detaching"
	// VolumeAttachmentStateDetached volume is detached.
	VolumeAttachmentStateDetached VolumeAttachmentState = "detached"
)

// CloudVolumeAttachment runtime volume attachment status.
type CloudVolumeAttachment struct {
	// AttachTime the time stamp when the attachment initiated.
	AttachTime *time.Time `locationName:"attachTime" type:"timestamp"`
	// DeviceName the device name.
	DeviceName *string `locationName:"device" type:"string"`
	// InstanceID unique instance identifier.
	InstanceID *string `locationName:"instanceID" type:"string"`
	// State the attachment state of the volume.
	State *string `locationName:"status" type:"string" enum:"VolumeAttachmentState"`
	// VolumeID unique identifier for volume.
	VolumeID *string `locationName:"volumeID" type:"string"`
}

// CloudVolume runtime status of CloudVolume.
type CloudVolume struct {
	// VolumeID unique identifier for the volume.
	VolumeID *string `locationName:"volumeID" type:"string"`
	// Attachement information
	Attachment *CloudVolumeAttachment `locationName:"attachment"`
	// Zone for the volume.
	Zone *string `locationName:"zone" type:"string"`
	// Region for the volume.
	Region *string `locationName:"region" type:"string"`
	// CreateTime the time stamp when volume creation was initiated.
	CreateTime *time.Time `locationName:"createTime" type:"timestamp"`
	// Encrypted indicates whether the volume will be encrypted.
	Encrypted *bool `locationName:"encrypted" type:"boolean"`
	// Iops provisioned for this volume.
	Iops *int64 `locationName:"iops" type:"integer"`
	// EncryptionKeyID the full ARN of the Key Management Service.
	EncryptionKeyID *string `locationName:"kmsKeyID" type:"string"`
	// Size in GiBs.
	Size *int64 `locationName:"size" type:"integer"`
	// SnapshotID from which the volume was created, if applicable.
	SnapshotID *string `locationName:"snapshotID" type:"string"`
	// State The volume state.
	State *string `locationName:"status" type:"string" enum:"VolumeState"`
	// Labels assigned to the volume.
	Labels map[string]string `locationName:"labels" locationNameList:"item" type:"list"`
	// VolumeType the type of the volume e.g. GP2j
	VolumeType *string `locationName:"volumeType" type:"string" enum:"VolumeType"`
}
