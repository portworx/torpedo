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

// V1ConnectionInfo Connection Information for the Deployment Topology.
type V1ConnectionInfo struct {
	// Ready pods.
	Pods []V1PodInfo `json:"pods,omitempty"`
	// Pods that are not ready.
	NotReadyPods []V1PodInfo `json:"notReadyPods,omitempty"`
	ConnectionDetails *V1ConnectionDetails `json:"connectionDetails,omitempty"`
	// Stores details about the cluster.
	ClusterDetails *map[string]ProtobufAny2 `json:"clusterDetails,omitempty"`
}

// NewV1ConnectionInfo instantiates a new V1ConnectionInfo object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewV1ConnectionInfo() *V1ConnectionInfo {
	this := V1ConnectionInfo{}
	return &this
}

// NewV1ConnectionInfoWithDefaults instantiates a new V1ConnectionInfo object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewV1ConnectionInfoWithDefaults() *V1ConnectionInfo {
	this := V1ConnectionInfo{}
	return &this
}

// GetPods returns the Pods field value if set, zero value otherwise.
func (o *V1ConnectionInfo) GetPods() []V1PodInfo {
	if o == nil || o.Pods == nil {
		var ret []V1PodInfo
		return ret
	}
	return o.Pods
}

// GetPodsOk returns a tuple with the Pods field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1ConnectionInfo) GetPodsOk() ([]V1PodInfo, bool) {
	if o == nil || o.Pods == nil {
		return nil, false
	}
	return o.Pods, true
}

// HasPods returns a boolean if a field has been set.
func (o *V1ConnectionInfo) HasPods() bool {
	if o != nil && o.Pods != nil {
		return true
	}

	return false
}

// SetPods gets a reference to the given []V1PodInfo and assigns it to the Pods field.
func (o *V1ConnectionInfo) SetPods(v []V1PodInfo) {
	o.Pods = v
}

// GetNotReadyPods returns the NotReadyPods field value if set, zero value otherwise.
func (o *V1ConnectionInfo) GetNotReadyPods() []V1PodInfo {
	if o == nil || o.NotReadyPods == nil {
		var ret []V1PodInfo
		return ret
	}
	return o.NotReadyPods
}

// GetNotReadyPodsOk returns a tuple with the NotReadyPods field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1ConnectionInfo) GetNotReadyPodsOk() ([]V1PodInfo, bool) {
	if o == nil || o.NotReadyPods == nil {
		return nil, false
	}
	return o.NotReadyPods, true
}

// HasNotReadyPods returns a boolean if a field has been set.
func (o *V1ConnectionInfo) HasNotReadyPods() bool {
	if o != nil && o.NotReadyPods != nil {
		return true
	}

	return false
}

// SetNotReadyPods gets a reference to the given []V1PodInfo and assigns it to the NotReadyPods field.
func (o *V1ConnectionInfo) SetNotReadyPods(v []V1PodInfo) {
	o.NotReadyPods = v
}

// GetConnectionDetails returns the ConnectionDetails field value if set, zero value otherwise.
func (o *V1ConnectionInfo) GetConnectionDetails() V1ConnectionDetails {
	if o == nil || o.ConnectionDetails == nil {
		var ret V1ConnectionDetails
		return ret
	}
	return *o.ConnectionDetails
}

// GetConnectionDetailsOk returns a tuple with the ConnectionDetails field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1ConnectionInfo) GetConnectionDetailsOk() (*V1ConnectionDetails, bool) {
	if o == nil || o.ConnectionDetails == nil {
		return nil, false
	}
	return o.ConnectionDetails, true
}

// HasConnectionDetails returns a boolean if a field has been set.
func (o *V1ConnectionInfo) HasConnectionDetails() bool {
	if o != nil && o.ConnectionDetails != nil {
		return true
	}

	return false
}

// SetConnectionDetails gets a reference to the given V1ConnectionDetails and assigns it to the ConnectionDetails field.
func (o *V1ConnectionInfo) SetConnectionDetails(v V1ConnectionDetails) {
	o.ConnectionDetails = &v
}

// GetClusterDetails returns the ClusterDetails field value if set, zero value otherwise.
func (o *V1ConnectionInfo) GetClusterDetails() map[string]ProtobufAny2 {
	if o == nil || o.ClusterDetails == nil {
		var ret map[string]ProtobufAny2
		return ret
	}
	return *o.ClusterDetails
}

// GetClusterDetailsOk returns a tuple with the ClusterDetails field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1ConnectionInfo) GetClusterDetailsOk() (*map[string]ProtobufAny2, bool) {
	if o == nil || o.ClusterDetails == nil {
		return nil, false
	}
	return o.ClusterDetails, true
}

// HasClusterDetails returns a boolean if a field has been set.
func (o *V1ConnectionInfo) HasClusterDetails() bool {
	if o != nil && o.ClusterDetails != nil {
		return true
	}

	return false
}

// SetClusterDetails gets a reference to the given map[string]ProtobufAny2 and assigns it to the ClusterDetails field.
func (o *V1ConnectionInfo) SetClusterDetails(v map[string]ProtobufAny2) {
	o.ClusterDetails = &v
}

func (o V1ConnectionInfo) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if o.Pods != nil {
		toSerialize["pods"] = o.Pods
	}
	if o.NotReadyPods != nil {
		toSerialize["notReadyPods"] = o.NotReadyPods
	}
	if o.ConnectionDetails != nil {
		toSerialize["connectionDetails"] = o.ConnectionDetails
	}
	if o.ClusterDetails != nil {
		toSerialize["clusterDetails"] = o.ClusterDetails
	}
	return json.Marshal(toSerialize)
}

type NullableV1ConnectionInfo struct {
	value *V1ConnectionInfo
	isSet bool
}

func (v NullableV1ConnectionInfo) Get() *V1ConnectionInfo {
	return v.value
}

func (v *NullableV1ConnectionInfo) Set(val *V1ConnectionInfo) {
	v.value = val
	v.isSet = true
}

func (v NullableV1ConnectionInfo) IsSet() bool {
	return v.isSet
}

func (v *NullableV1ConnectionInfo) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableV1ConnectionInfo(val *V1ConnectionInfo) *NullableV1ConnectionInfo {
	return &NullableV1ConnectionInfo{value: val, isSet: true}
}

func (v NullableV1ConnectionInfo) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableV1ConnectionInfo) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

