/*
public/portworx/pds/tasks/apiv1/tasks.proto

No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)

API version: version not set
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package pdsclient

import (
	"encoding/json"
)

// checks if the V1Info1 type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &V1Info1{}

// V1Info1 Desired Info of the data service.
type V1Info1 struct {
	// Short name of the data service.
	ShortName *string `json:"shortName,omitempty"`
	// Enabled flag suggests if the data service is enabled or not.
	Enabled *bool `json:"enabled,omitempty"`
	// Node limitations.
	NodesLimitations *string `json:"nodesLimitations,omitempty"`
	NodeRestrictions *V1NodeRestrictions `json:"nodeRestrictions,omitempty"`
}

// NewV1Info1 instantiates a new V1Info1 object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewV1Info1() *V1Info1 {
	this := V1Info1{}
	return &this
}

// NewV1Info1WithDefaults instantiates a new V1Info1 object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewV1Info1WithDefaults() *V1Info1 {
	this := V1Info1{}
	return &this
}

// GetShortName returns the ShortName field value if set, zero value otherwise.
func (o *V1Info1) GetShortName() string {
	if o == nil || IsNil(o.ShortName) {
		var ret string
		return ret
	}
	return *o.ShortName
}

// GetShortNameOk returns a tuple with the ShortName field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1Info1) GetShortNameOk() (*string, bool) {
	if o == nil || IsNil(o.ShortName) {
		return nil, false
	}
	return o.ShortName, true
}

// HasShortName returns a boolean if a field has been set.
func (o *V1Info1) HasShortName() bool {
	if o != nil && !IsNil(o.ShortName) {
		return true
	}

	return false
}

// SetShortName gets a reference to the given string and assigns it to the ShortName field.
func (o *V1Info1) SetShortName(v string) {
	o.ShortName = &v
}

// GetEnabled returns the Enabled field value if set, zero value otherwise.
func (o *V1Info1) GetEnabled() bool {
	if o == nil || IsNil(o.Enabled) {
		var ret bool
		return ret
	}
	return *o.Enabled
}

// GetEnabledOk returns a tuple with the Enabled field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1Info1) GetEnabledOk() (*bool, bool) {
	if o == nil || IsNil(o.Enabled) {
		return nil, false
	}
	return o.Enabled, true
}

// HasEnabled returns a boolean if a field has been set.
func (o *V1Info1) HasEnabled() bool {
	if o != nil && !IsNil(o.Enabled) {
		return true
	}

	return false
}

// SetEnabled gets a reference to the given bool and assigns it to the Enabled field.
func (o *V1Info1) SetEnabled(v bool) {
	o.Enabled = &v
}

// GetNodesLimitations returns the NodesLimitations field value if set, zero value otherwise.
func (o *V1Info1) GetNodesLimitations() string {
	if o == nil || IsNil(o.NodesLimitations) {
		var ret string
		return ret
	}
	return *o.NodesLimitations
}

// GetNodesLimitationsOk returns a tuple with the NodesLimitations field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1Info1) GetNodesLimitationsOk() (*string, bool) {
	if o == nil || IsNil(o.NodesLimitations) {
		return nil, false
	}
	return o.NodesLimitations, true
}

// HasNodesLimitations returns a boolean if a field has been set.
func (o *V1Info1) HasNodesLimitations() bool {
	if o != nil && !IsNil(o.NodesLimitations) {
		return true
	}

	return false
}

// SetNodesLimitations gets a reference to the given string and assigns it to the NodesLimitations field.
func (o *V1Info1) SetNodesLimitations(v string) {
	o.NodesLimitations = &v
}

// GetNodeRestrictions returns the NodeRestrictions field value if set, zero value otherwise.
func (o *V1Info1) GetNodeRestrictions() V1NodeRestrictions {
	if o == nil || IsNil(o.NodeRestrictions) {
		var ret V1NodeRestrictions
		return ret
	}
	return *o.NodeRestrictions
}

// GetNodeRestrictionsOk returns a tuple with the NodeRestrictions field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1Info1) GetNodeRestrictionsOk() (*V1NodeRestrictions, bool) {
	if o == nil || IsNil(o.NodeRestrictions) {
		return nil, false
	}
	return o.NodeRestrictions, true
}

// HasNodeRestrictions returns a boolean if a field has been set.
func (o *V1Info1) HasNodeRestrictions() bool {
	if o != nil && !IsNil(o.NodeRestrictions) {
		return true
	}

	return false
}

// SetNodeRestrictions gets a reference to the given V1NodeRestrictions and assigns it to the NodeRestrictions field.
func (o *V1Info1) SetNodeRestrictions(v V1NodeRestrictions) {
	o.NodeRestrictions = &v
}

func (o V1Info1) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o V1Info1) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.ShortName) {
		toSerialize["shortName"] = o.ShortName
	}
	if !IsNil(o.Enabled) {
		toSerialize["enabled"] = o.Enabled
	}
	if !IsNil(o.NodesLimitations) {
		toSerialize["nodesLimitations"] = o.NodesLimitations
	}
	if !IsNil(o.NodeRestrictions) {
		toSerialize["nodeRestrictions"] = o.NodeRestrictions
	}
	return toSerialize, nil
}

type NullableV1Info1 struct {
	value *V1Info1
	isSet bool
}

func (v NullableV1Info1) Get() *V1Info1 {
	return v.value
}

func (v *NullableV1Info1) Set(val *V1Info1) {
	v.value = val
	v.isSet = true
}

func (v NullableV1Info1) IsSet() bool {
	return v.isSet
}

func (v *NullableV1Info1) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableV1Info1(val *V1Info1) *NullableV1Info1 {
	return &NullableV1Info1{value: val, isSet: true}
}

func (v NullableV1Info1) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableV1Info1) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

