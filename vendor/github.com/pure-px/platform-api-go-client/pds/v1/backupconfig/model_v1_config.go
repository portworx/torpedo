/*
public/portworx/pds/backupconfig/apiv1/backupconfig.proto

No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)

API version: version not set
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package backupconfig

import (
	"encoding/json"
)

// checks if the V1Config type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &V1Config{}

// V1Config Desired config of the backup configuration.
type V1Config struct {
	References *V1References `json:"references,omitempty"`
	// Job History Limit is a number of retained backup jobs. Must be greater than or equal to 1.
	JobHistoryLimit *int32 `json:"jobHistoryLimit,omitempty"`
	BackupPolicy *V1BackupPolicy `json:"backupPolicy,omitempty"`
	// Suspend flag is used to suspend a scheduled backup from creating new backup jobs.
	Suspend *bool `json:"suspend,omitempty"`
	BackupType *ConfigBackupType `json:"backupType,omitempty"`
	BackupLevel *ConfigBackupLevel `json:"backupLevel,omitempty"`
	ReclaimPolicy *ConfigReclaimPolicyType `json:"reclaimPolicy,omitempty"`
}

// NewV1Config instantiates a new V1Config object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewV1Config() *V1Config {
	this := V1Config{}
	var backupType ConfigBackupType = CONFIGBACKUPTYPE_BACKUP_TYPE_UNSPECIFIED
	this.BackupType = &backupType
	var backupLevel ConfigBackupLevel = CONFIGBACKUPLEVEL_BACKUP_LEVEL_UNSPECIFIED
	this.BackupLevel = &backupLevel
	var reclaimPolicy ConfigReclaimPolicyType = CONFIGRECLAIMPOLICYTYPE_RECLAIM_POLICY_TYPE_UNSPECIFIED
	this.ReclaimPolicy = &reclaimPolicy
	return &this
}

// NewV1ConfigWithDefaults instantiates a new V1Config object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewV1ConfigWithDefaults() *V1Config {
	this := V1Config{}
	var backupType ConfigBackupType = CONFIGBACKUPTYPE_BACKUP_TYPE_UNSPECIFIED
	this.BackupType = &backupType
	var backupLevel ConfigBackupLevel = CONFIGBACKUPLEVEL_BACKUP_LEVEL_UNSPECIFIED
	this.BackupLevel = &backupLevel
	var reclaimPolicy ConfigReclaimPolicyType = CONFIGRECLAIMPOLICYTYPE_RECLAIM_POLICY_TYPE_UNSPECIFIED
	this.ReclaimPolicy = &reclaimPolicy
	return &this
}

// GetReferences returns the References field value if set, zero value otherwise.
func (o *V1Config) GetReferences() V1References {
	if o == nil || IsNil(o.References) {
		var ret V1References
		return ret
	}
	return *o.References
}

// GetReferencesOk returns a tuple with the References field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1Config) GetReferencesOk() (*V1References, bool) {
	if o == nil || IsNil(o.References) {
		return nil, false
	}
	return o.References, true
}

// HasReferences returns a boolean if a field has been set.
func (o *V1Config) HasReferences() bool {
	if o != nil && !IsNil(o.References) {
		return true
	}

	return false
}

// SetReferences gets a reference to the given V1References and assigns it to the References field.
func (o *V1Config) SetReferences(v V1References) {
	o.References = &v
}

// GetJobHistoryLimit returns the JobHistoryLimit field value if set, zero value otherwise.
func (o *V1Config) GetJobHistoryLimit() int32 {
	if o == nil || IsNil(o.JobHistoryLimit) {
		var ret int32
		return ret
	}
	return *o.JobHistoryLimit
}

// GetJobHistoryLimitOk returns a tuple with the JobHistoryLimit field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1Config) GetJobHistoryLimitOk() (*int32, bool) {
	if o == nil || IsNil(o.JobHistoryLimit) {
		return nil, false
	}
	return o.JobHistoryLimit, true
}

// HasJobHistoryLimit returns a boolean if a field has been set.
func (o *V1Config) HasJobHistoryLimit() bool {
	if o != nil && !IsNil(o.JobHistoryLimit) {
		return true
	}

	return false
}

// SetJobHistoryLimit gets a reference to the given int32 and assigns it to the JobHistoryLimit field.
func (o *V1Config) SetJobHistoryLimit(v int32) {
	o.JobHistoryLimit = &v
}

// GetBackupPolicy returns the BackupPolicy field value if set, zero value otherwise.
func (o *V1Config) GetBackupPolicy() V1BackupPolicy {
	if o == nil || IsNil(o.BackupPolicy) {
		var ret V1BackupPolicy
		return ret
	}
	return *o.BackupPolicy
}

// GetBackupPolicyOk returns a tuple with the BackupPolicy field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1Config) GetBackupPolicyOk() (*V1BackupPolicy, bool) {
	if o == nil || IsNil(o.BackupPolicy) {
		return nil, false
	}
	return o.BackupPolicy, true
}

// HasBackupPolicy returns a boolean if a field has been set.
func (o *V1Config) HasBackupPolicy() bool {
	if o != nil && !IsNil(o.BackupPolicy) {
		return true
	}

	return false
}

// SetBackupPolicy gets a reference to the given V1BackupPolicy and assigns it to the BackupPolicy field.
func (o *V1Config) SetBackupPolicy(v V1BackupPolicy) {
	o.BackupPolicy = &v
}

// GetSuspend returns the Suspend field value if set, zero value otherwise.
func (o *V1Config) GetSuspend() bool {
	if o == nil || IsNil(o.Suspend) {
		var ret bool
		return ret
	}
	return *o.Suspend
}

// GetSuspendOk returns a tuple with the Suspend field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1Config) GetSuspendOk() (*bool, bool) {
	if o == nil || IsNil(o.Suspend) {
		return nil, false
	}
	return o.Suspend, true
}

// HasSuspend returns a boolean if a field has been set.
func (o *V1Config) HasSuspend() bool {
	if o != nil && !IsNil(o.Suspend) {
		return true
	}

	return false
}

// SetSuspend gets a reference to the given bool and assigns it to the Suspend field.
func (o *V1Config) SetSuspend(v bool) {
	o.Suspend = &v
}

// GetBackupType returns the BackupType field value if set, zero value otherwise.
func (o *V1Config) GetBackupType() ConfigBackupType {
	if o == nil || IsNil(o.BackupType) {
		var ret ConfigBackupType
		return ret
	}
	return *o.BackupType
}

// GetBackupTypeOk returns a tuple with the BackupType field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1Config) GetBackupTypeOk() (*ConfigBackupType, bool) {
	if o == nil || IsNil(o.BackupType) {
		return nil, false
	}
	return o.BackupType, true
}

// HasBackupType returns a boolean if a field has been set.
func (o *V1Config) HasBackupType() bool {
	if o != nil && !IsNil(o.BackupType) {
		return true
	}

	return false
}

// SetBackupType gets a reference to the given ConfigBackupType and assigns it to the BackupType field.
func (o *V1Config) SetBackupType(v ConfigBackupType) {
	o.BackupType = &v
}

// GetBackupLevel returns the BackupLevel field value if set, zero value otherwise.
func (o *V1Config) GetBackupLevel() ConfigBackupLevel {
	if o == nil || IsNil(o.BackupLevel) {
		var ret ConfigBackupLevel
		return ret
	}
	return *o.BackupLevel
}

// GetBackupLevelOk returns a tuple with the BackupLevel field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1Config) GetBackupLevelOk() (*ConfigBackupLevel, bool) {
	if o == nil || IsNil(o.BackupLevel) {
		return nil, false
	}
	return o.BackupLevel, true
}

// HasBackupLevel returns a boolean if a field has been set.
func (o *V1Config) HasBackupLevel() bool {
	if o != nil && !IsNil(o.BackupLevel) {
		return true
	}

	return false
}

// SetBackupLevel gets a reference to the given ConfigBackupLevel and assigns it to the BackupLevel field.
func (o *V1Config) SetBackupLevel(v ConfigBackupLevel) {
	o.BackupLevel = &v
}

// GetReclaimPolicy returns the ReclaimPolicy field value if set, zero value otherwise.
func (o *V1Config) GetReclaimPolicy() ConfigReclaimPolicyType {
	if o == nil || IsNil(o.ReclaimPolicy) {
		var ret ConfigReclaimPolicyType
		return ret
	}
	return *o.ReclaimPolicy
}

// GetReclaimPolicyOk returns a tuple with the ReclaimPolicy field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1Config) GetReclaimPolicyOk() (*ConfigReclaimPolicyType, bool) {
	if o == nil || IsNil(o.ReclaimPolicy) {
		return nil, false
	}
	return o.ReclaimPolicy, true
}

// HasReclaimPolicy returns a boolean if a field has been set.
func (o *V1Config) HasReclaimPolicy() bool {
	if o != nil && !IsNil(o.ReclaimPolicy) {
		return true
	}

	return false
}

// SetReclaimPolicy gets a reference to the given ConfigReclaimPolicyType and assigns it to the ReclaimPolicy field.
func (o *V1Config) SetReclaimPolicy(v ConfigReclaimPolicyType) {
	o.ReclaimPolicy = &v
}

func (o V1Config) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o V1Config) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.References) {
		toSerialize["references"] = o.References
	}
	if !IsNil(o.JobHistoryLimit) {
		toSerialize["jobHistoryLimit"] = o.JobHistoryLimit
	}
	if !IsNil(o.BackupPolicy) {
		toSerialize["backupPolicy"] = o.BackupPolicy
	}
	if !IsNil(o.Suspend) {
		toSerialize["suspend"] = o.Suspend
	}
	if !IsNil(o.BackupType) {
		toSerialize["backupType"] = o.BackupType
	}
	if !IsNil(o.BackupLevel) {
		toSerialize["backupLevel"] = o.BackupLevel
	}
	if !IsNil(o.ReclaimPolicy) {
		toSerialize["reclaimPolicy"] = o.ReclaimPolicy
	}
	return toSerialize, nil
}

type NullableV1Config struct {
	value *V1Config
	isSet bool
}

func (v NullableV1Config) Get() *V1Config {
	return v.value
}

func (v *NullableV1Config) Set(val *V1Config) {
	v.value = val
	v.isSet = true
}

func (v NullableV1Config) IsSet() bool {
	return v.isSet
}

func (v *NullableV1Config) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableV1Config(val *V1Config) *NullableV1Config {
	return &NullableV1Config{value: val, isSet: true}
}

func (v NullableV1Config) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableV1Config) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

