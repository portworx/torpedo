/*
public/portworx/pds/backupconfig/apiv1/backupconfig.proto

No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)

API version: version not set
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package openapi

import (
	"encoding/json"
)

// V1TargetCluster TargetCluster is a high level entity that represents one large company(e.g. a Pure).
type V1TargetCluster struct {
	Meta *V1Meta `json:"meta,omitempty"`
	Config map[string]interface{} `json:"config,omitempty"`
	Status *PlatformTargetClusterv1Status `json:"status,omitempty"`
}

// NewV1TargetCluster instantiates a new V1TargetCluster object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewV1TargetCluster() *V1TargetCluster {
	this := V1TargetCluster{}
	return &this
}

// NewV1TargetClusterWithDefaults instantiates a new V1TargetCluster object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewV1TargetClusterWithDefaults() *V1TargetCluster {
	this := V1TargetCluster{}
	return &this
}

// GetMeta returns the Meta field value if set, zero value otherwise.
func (o *V1TargetCluster) GetMeta() V1Meta {
	if o == nil || o.Meta == nil {
		var ret V1Meta
		return ret
	}
	return *o.Meta
}

// GetMetaOk returns a tuple with the Meta field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1TargetCluster) GetMetaOk() (*V1Meta, bool) {
	if o == nil || o.Meta == nil {
		return nil, false
	}
	return o.Meta, true
}

// HasMeta returns a boolean if a field has been set.
func (o *V1TargetCluster) HasMeta() bool {
	if o != nil && o.Meta != nil {
		return true
	}

	return false
}

// SetMeta gets a reference to the given V1Meta and assigns it to the Meta field.
func (o *V1TargetCluster) SetMeta(v V1Meta) {
	o.Meta = &v
}

// GetConfig returns the Config field value if set, zero value otherwise.
func (o *V1TargetCluster) GetConfig() map[string]interface{} {
	if o == nil || o.Config == nil {
		var ret map[string]interface{}
		return ret
	}
	return o.Config
}

// GetConfigOk returns a tuple with the Config field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1TargetCluster) GetConfigOk() (map[string]interface{}, bool) {
	if o == nil || o.Config == nil {
		return nil, false
	}
	return o.Config, true
}

// HasConfig returns a boolean if a field has been set.
func (o *V1TargetCluster) HasConfig() bool {
	if o != nil && o.Config != nil {
		return true
	}

	return false
}

// SetConfig gets a reference to the given map[string]interface{} and assigns it to the Config field.
func (o *V1TargetCluster) SetConfig(v map[string]interface{}) {
	o.Config = v
}

// GetStatus returns the Status field value if set, zero value otherwise.
func (o *V1TargetCluster) GetStatus() PlatformTargetClusterv1Status {
	if o == nil || o.Status == nil {
		var ret PlatformTargetClusterv1Status
		return ret
	}
	return *o.Status
}

// GetStatusOk returns a tuple with the Status field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1TargetCluster) GetStatusOk() (*PlatformTargetClusterv1Status, bool) {
	if o == nil || o.Status == nil {
		return nil, false
	}
	return o.Status, true
}

// HasStatus returns a boolean if a field has been set.
func (o *V1TargetCluster) HasStatus() bool {
	if o != nil && o.Status != nil {
		return true
	}

	return false
}

// SetStatus gets a reference to the given PlatformTargetClusterv1Status and assigns it to the Status field.
func (o *V1TargetCluster) SetStatus(v PlatformTargetClusterv1Status) {
	o.Status = &v
}

func (o V1TargetCluster) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if o.Meta != nil {
		toSerialize["meta"] = o.Meta
	}
	if o.Config != nil {
		toSerialize["config"] = o.Config
	}
	if o.Status != nil {
		toSerialize["status"] = o.Status
	}
	return json.Marshal(toSerialize)
}

type NullableV1TargetCluster struct {
	value *V1TargetCluster
	isSet bool
}

func (v NullableV1TargetCluster) Get() *V1TargetCluster {
	return v.value
}

func (v *NullableV1TargetCluster) Set(val *V1TargetCluster) {
	v.value = val
	v.isSet = true
}

func (v NullableV1TargetCluster) IsSet() bool {
	return v.isSet
}

func (v *NullableV1TargetCluster) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableV1TargetCluster(val *V1TargetCluster) *NullableV1TargetCluster {
	return &NullableV1TargetCluster{value: val, isSet: true}
}

func (v NullableV1TargetCluster) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableV1TargetCluster) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

