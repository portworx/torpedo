/*
public/portworx/platform/targetcluster/application/apiv1/application.proto

No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)

API version: version not set
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package targetcluster

import (
	"encoding/json"
	"time"
)

// checks if the PlatformTargetClusterv1Status type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &PlatformTargetClusterv1Status{}

// PlatformTargetClusterv1Status Status of the Target Cluster.
type PlatformTargetClusterv1Status struct {
	Metadata *V1Metadata `json:"metadata,omitempty"`
	Phase *V1TargetClusterPhasePhase `json:"phase,omitempty"`
	// Timestamp of cluster's last status update to control plane.
	LastStatusUpdateTime *time.Time `json:"lastStatusUpdateTime,omitempty"`
	PlatformAgent *V1ApplicationPhasePhase `json:"platformAgent,omitempty"`
	// Status of applications running in the target cluster eg: BAAS, PDS, MPXE.
	Applications *map[string]V1ApplicationPhasePhase `json:"applications,omitempty"`
}

// NewPlatformTargetClusterv1Status instantiates a new PlatformTargetClusterv1Status object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewPlatformTargetClusterv1Status() *PlatformTargetClusterv1Status {
	this := PlatformTargetClusterv1Status{}
	var phase V1TargetClusterPhasePhase = V1TARGETCLUSTERPHASEPHASE_PHASE_UNSPECIFIED
	this.Phase = &phase
	var platformAgent V1ApplicationPhasePhase = V1APPLICATIONPHASEPHASE_PHASE_UNSPECIFIED
	this.PlatformAgent = &platformAgent
	return &this
}

// NewPlatformTargetClusterv1StatusWithDefaults instantiates a new PlatformTargetClusterv1Status object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewPlatformTargetClusterv1StatusWithDefaults() *PlatformTargetClusterv1Status {
	this := PlatformTargetClusterv1Status{}
	var phase V1TargetClusterPhasePhase = V1TARGETCLUSTERPHASEPHASE_PHASE_UNSPECIFIED
	this.Phase = &phase
	var platformAgent V1ApplicationPhasePhase = V1APPLICATIONPHASEPHASE_PHASE_UNSPECIFIED
	this.PlatformAgent = &platformAgent
	return &this
}

// GetMetadata returns the Metadata field value if set, zero value otherwise.
func (o *PlatformTargetClusterv1Status) GetMetadata() V1Metadata {
	if o == nil || IsNil(o.Metadata) {
		var ret V1Metadata
		return ret
	}
	return *o.Metadata
}

// GetMetadataOk returns a tuple with the Metadata field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *PlatformTargetClusterv1Status) GetMetadataOk() (*V1Metadata, bool) {
	if o == nil || IsNil(o.Metadata) {
		return nil, false
	}
	return o.Metadata, true
}

// HasMetadata returns a boolean if a field has been set.
func (o *PlatformTargetClusterv1Status) HasMetadata() bool {
	if o != nil && !IsNil(o.Metadata) {
		return true
	}

	return false
}

// SetMetadata gets a reference to the given V1Metadata and assigns it to the Metadata field.
func (o *PlatformTargetClusterv1Status) SetMetadata(v V1Metadata) {
	o.Metadata = &v
}

// GetPhase returns the Phase field value if set, zero value otherwise.
func (o *PlatformTargetClusterv1Status) GetPhase() V1TargetClusterPhasePhase {
	if o == nil || IsNil(o.Phase) {
		var ret V1TargetClusterPhasePhase
		return ret
	}
	return *o.Phase
}

// GetPhaseOk returns a tuple with the Phase field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *PlatformTargetClusterv1Status) GetPhaseOk() (*V1TargetClusterPhasePhase, bool) {
	if o == nil || IsNil(o.Phase) {
		return nil, false
	}
	return o.Phase, true
}

// HasPhase returns a boolean if a field has been set.
func (o *PlatformTargetClusterv1Status) HasPhase() bool {
	if o != nil && !IsNil(o.Phase) {
		return true
	}

	return false
}

// SetPhase gets a reference to the given V1TargetClusterPhasePhase and assigns it to the Phase field.
func (o *PlatformTargetClusterv1Status) SetPhase(v V1TargetClusterPhasePhase) {
	o.Phase = &v
}

// GetLastStatusUpdateTime returns the LastStatusUpdateTime field value if set, zero value otherwise.
func (o *PlatformTargetClusterv1Status) GetLastStatusUpdateTime() time.Time {
	if o == nil || IsNil(o.LastStatusUpdateTime) {
		var ret time.Time
		return ret
	}
	return *o.LastStatusUpdateTime
}

// GetLastStatusUpdateTimeOk returns a tuple with the LastStatusUpdateTime field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *PlatformTargetClusterv1Status) GetLastStatusUpdateTimeOk() (*time.Time, bool) {
	if o == nil || IsNil(o.LastStatusUpdateTime) {
		return nil, false
	}
	return o.LastStatusUpdateTime, true
}

// HasLastStatusUpdateTime returns a boolean if a field has been set.
func (o *PlatformTargetClusterv1Status) HasLastStatusUpdateTime() bool {
	if o != nil && !IsNil(o.LastStatusUpdateTime) {
		return true
	}

	return false
}

// SetLastStatusUpdateTime gets a reference to the given time.Time and assigns it to the LastStatusUpdateTime field.
func (o *PlatformTargetClusterv1Status) SetLastStatusUpdateTime(v time.Time) {
	o.LastStatusUpdateTime = &v
}

// GetPlatformAgent returns the PlatformAgent field value if set, zero value otherwise.
func (o *PlatformTargetClusterv1Status) GetPlatformAgent() V1ApplicationPhasePhase {
	if o == nil || IsNil(o.PlatformAgent) {
		var ret V1ApplicationPhasePhase
		return ret
	}
	return *o.PlatformAgent
}

// GetPlatformAgentOk returns a tuple with the PlatformAgent field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *PlatformTargetClusterv1Status) GetPlatformAgentOk() (*V1ApplicationPhasePhase, bool) {
	if o == nil || IsNil(o.PlatformAgent) {
		return nil, false
	}
	return o.PlatformAgent, true
}

// HasPlatformAgent returns a boolean if a field has been set.
func (o *PlatformTargetClusterv1Status) HasPlatformAgent() bool {
	if o != nil && !IsNil(o.PlatformAgent) {
		return true
	}

	return false
}

// SetPlatformAgent gets a reference to the given V1ApplicationPhasePhase and assigns it to the PlatformAgent field.
func (o *PlatformTargetClusterv1Status) SetPlatformAgent(v V1ApplicationPhasePhase) {
	o.PlatformAgent = &v
}

// GetApplications returns the Applications field value if set, zero value otherwise.
func (o *PlatformTargetClusterv1Status) GetApplications() map[string]V1ApplicationPhasePhase {
	if o == nil || IsNil(o.Applications) {
		var ret map[string]V1ApplicationPhasePhase
		return ret
	}
	return *o.Applications
}

// GetApplicationsOk returns a tuple with the Applications field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *PlatformTargetClusterv1Status) GetApplicationsOk() (*map[string]V1ApplicationPhasePhase, bool) {
	if o == nil || IsNil(o.Applications) {
		return nil, false
	}
	return o.Applications, true
}

// HasApplications returns a boolean if a field has been set.
func (o *PlatformTargetClusterv1Status) HasApplications() bool {
	if o != nil && !IsNil(o.Applications) {
		return true
	}

	return false
}

// SetApplications gets a reference to the given map[string]V1ApplicationPhasePhase and assigns it to the Applications field.
func (o *PlatformTargetClusterv1Status) SetApplications(v map[string]V1ApplicationPhasePhase) {
	o.Applications = &v
}

func (o PlatformTargetClusterv1Status) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o PlatformTargetClusterv1Status) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.Metadata) {
		toSerialize["metadata"] = o.Metadata
	}
	if !IsNil(o.Phase) {
		toSerialize["phase"] = o.Phase
	}
	if !IsNil(o.LastStatusUpdateTime) {
		toSerialize["lastStatusUpdateTime"] = o.LastStatusUpdateTime
	}
	if !IsNil(o.PlatformAgent) {
		toSerialize["platformAgent"] = o.PlatformAgent
	}
	if !IsNil(o.Applications) {
		toSerialize["applications"] = o.Applications
	}
	return toSerialize, nil
}

type NullablePlatformTargetClusterv1Status struct {
	value *PlatformTargetClusterv1Status
	isSet bool
}

func (v NullablePlatformTargetClusterv1Status) Get() *PlatformTargetClusterv1Status {
	return v.value
}

func (v *NullablePlatformTargetClusterv1Status) Set(val *PlatformTargetClusterv1Status) {
	v.value = val
	v.isSet = true
}

func (v NullablePlatformTargetClusterv1Status) IsSet() bool {
	return v.isSet
}

func (v *NullablePlatformTargetClusterv1Status) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullablePlatformTargetClusterv1Status(val *PlatformTargetClusterv1Status) *NullablePlatformTargetClusterv1Status {
	return &NullablePlatformTargetClusterv1Status{value: val, isSet: true}
}

func (v NullablePlatformTargetClusterv1Status) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullablePlatformTargetClusterv1Status) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

