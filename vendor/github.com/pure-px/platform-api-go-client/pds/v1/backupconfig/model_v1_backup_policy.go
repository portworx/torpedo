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

// checks if the V1BackupPolicy type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &V1BackupPolicy{}

// V1BackupPolicy BackupPolicy associated with the backup config.
type V1BackupPolicy struct {
	// UID of the backup policy associated with the backup configuration.
	Id *string `json:"id,omitempty"`
	// Resource version of the backup policy.
	ResourceVersion *string `json:"resourceVersion,omitempty"`
}

// NewV1BackupPolicy instantiates a new V1BackupPolicy object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewV1BackupPolicy() *V1BackupPolicy {
	this := V1BackupPolicy{}
	return &this
}

// NewV1BackupPolicyWithDefaults instantiates a new V1BackupPolicy object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewV1BackupPolicyWithDefaults() *V1BackupPolicy {
	this := V1BackupPolicy{}
	return &this
}

// GetId returns the Id field value if set, zero value otherwise.
func (o *V1BackupPolicy) GetId() string {
	if o == nil || IsNil(o.Id) {
		var ret string
		return ret
	}
	return *o.Id
}

// GetIdOk returns a tuple with the Id field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1BackupPolicy) GetIdOk() (*string, bool) {
	if o == nil || IsNil(o.Id) {
		return nil, false
	}
	return o.Id, true
}

// HasId returns a boolean if a field has been set.
func (o *V1BackupPolicy) HasId() bool {
	if o != nil && !IsNil(o.Id) {
		return true
	}

	return false
}

// SetId gets a reference to the given string and assigns it to the Id field.
func (o *V1BackupPolicy) SetId(v string) {
	o.Id = &v
}

// GetResourceVersion returns the ResourceVersion field value if set, zero value otherwise.
func (o *V1BackupPolicy) GetResourceVersion() string {
	if o == nil || IsNil(o.ResourceVersion) {
		var ret string
		return ret
	}
	return *o.ResourceVersion
}

// GetResourceVersionOk returns a tuple with the ResourceVersion field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1BackupPolicy) GetResourceVersionOk() (*string, bool) {
	if o == nil || IsNil(o.ResourceVersion) {
		return nil, false
	}
	return o.ResourceVersion, true
}

// HasResourceVersion returns a boolean if a field has been set.
func (o *V1BackupPolicy) HasResourceVersion() bool {
	if o != nil && !IsNil(o.ResourceVersion) {
		return true
	}

	return false
}

// SetResourceVersion gets a reference to the given string and assigns it to the ResourceVersion field.
func (o *V1BackupPolicy) SetResourceVersion(v string) {
	o.ResourceVersion = &v
}

func (o V1BackupPolicy) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o V1BackupPolicy) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.Id) {
		toSerialize["id"] = o.Id
	}
	if !IsNil(o.ResourceVersion) {
		toSerialize["resourceVersion"] = o.ResourceVersion
	}
	return toSerialize, nil
}

type NullableV1BackupPolicy struct {
	value *V1BackupPolicy
	isSet bool
}

func (v NullableV1BackupPolicy) Get() *V1BackupPolicy {
	return v.value
}

func (v *NullableV1BackupPolicy) Set(val *V1BackupPolicy) {
	v.value = val
	v.isSet = true
}

func (v NullableV1BackupPolicy) IsSet() bool {
	return v.isSet
}

func (v *NullableV1BackupPolicy) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableV1BackupPolicy(val *V1BackupPolicy) *NullableV1BackupPolicy {
	return &NullableV1BackupPolicy{value: val, isSet: true}
}

func (v NullableV1BackupPolicy) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableV1BackupPolicy) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

