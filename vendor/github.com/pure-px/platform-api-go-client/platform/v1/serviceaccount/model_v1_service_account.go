/*
public/portworx/platform/serviceaccount/apiv1/serviceaccount.proto

No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)

API version: version not set
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package serviceaccount

import (
	"encoding/json"
)

// checks if the V1ServiceAccount type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &V1ServiceAccount{}

// V1ServiceAccount Service account represents a group of attributes using which a service can consume platform apis.
type V1ServiceAccount struct {
	Meta *V1Meta `json:"meta,omitempty"`
	Config *V1Config `json:"config,omitempty"`
	Status *Serviceaccountv1Status `json:"status,omitempty"`
}

// NewV1ServiceAccount instantiates a new V1ServiceAccount object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewV1ServiceAccount() *V1ServiceAccount {
	this := V1ServiceAccount{}
	return &this
}

// NewV1ServiceAccountWithDefaults instantiates a new V1ServiceAccount object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewV1ServiceAccountWithDefaults() *V1ServiceAccount {
	this := V1ServiceAccount{}
	return &this
}

// GetMeta returns the Meta field value if set, zero value otherwise.
func (o *V1ServiceAccount) GetMeta() V1Meta {
	if o == nil || IsNil(o.Meta) {
		var ret V1Meta
		return ret
	}
	return *o.Meta
}

// GetMetaOk returns a tuple with the Meta field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1ServiceAccount) GetMetaOk() (*V1Meta, bool) {
	if o == nil || IsNil(o.Meta) {
		return nil, false
	}
	return o.Meta, true
}

// HasMeta returns a boolean if a field has been set.
func (o *V1ServiceAccount) HasMeta() bool {
	if o != nil && !IsNil(o.Meta) {
		return true
	}

	return false
}

// SetMeta gets a reference to the given V1Meta and assigns it to the Meta field.
func (o *V1ServiceAccount) SetMeta(v V1Meta) {
	o.Meta = &v
}

// GetConfig returns the Config field value if set, zero value otherwise.
func (o *V1ServiceAccount) GetConfig() V1Config {
	if o == nil || IsNil(o.Config) {
		var ret V1Config
		return ret
	}
	return *o.Config
}

// GetConfigOk returns a tuple with the Config field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1ServiceAccount) GetConfigOk() (*V1Config, bool) {
	if o == nil || IsNil(o.Config) {
		return nil, false
	}
	return o.Config, true
}

// HasConfig returns a boolean if a field has been set.
func (o *V1ServiceAccount) HasConfig() bool {
	if o != nil && !IsNil(o.Config) {
		return true
	}

	return false
}

// SetConfig gets a reference to the given V1Config and assigns it to the Config field.
func (o *V1ServiceAccount) SetConfig(v V1Config) {
	o.Config = &v
}

// GetStatus returns the Status field value if set, zero value otherwise.
func (o *V1ServiceAccount) GetStatus() Serviceaccountv1Status {
	if o == nil || IsNil(o.Status) {
		var ret Serviceaccountv1Status
		return ret
	}
	return *o.Status
}

// GetStatusOk returns a tuple with the Status field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1ServiceAccount) GetStatusOk() (*Serviceaccountv1Status, bool) {
	if o == nil || IsNil(o.Status) {
		return nil, false
	}
	return o.Status, true
}

// HasStatus returns a boolean if a field has been set.
func (o *V1ServiceAccount) HasStatus() bool {
	if o != nil && !IsNil(o.Status) {
		return true
	}

	return false
}

// SetStatus gets a reference to the given Serviceaccountv1Status and assigns it to the Status field.
func (o *V1ServiceAccount) SetStatus(v Serviceaccountv1Status) {
	o.Status = &v
}

func (o V1ServiceAccount) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o V1ServiceAccount) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.Meta) {
		toSerialize["meta"] = o.Meta
	}
	if !IsNil(o.Config) {
		toSerialize["config"] = o.Config
	}
	if !IsNil(o.Status) {
		toSerialize["status"] = o.Status
	}
	return toSerialize, nil
}

type NullableV1ServiceAccount struct {
	value *V1ServiceAccount
	isSet bool
}

func (v NullableV1ServiceAccount) Get() *V1ServiceAccount {
	return v.value
}

func (v *NullableV1ServiceAccount) Set(val *V1ServiceAccount) {
	v.value = val
	v.isSet = true
}

func (v NullableV1ServiceAccount) IsSet() bool {
	return v.isSet
}

func (v *NullableV1ServiceAccount) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableV1ServiceAccount(val *V1ServiceAccount) *NullableV1ServiceAccount {
	return &NullableV1ServiceAccount{value: val, isSet: true}
}

func (v NullableV1ServiceAccount) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableV1ServiceAccount) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

