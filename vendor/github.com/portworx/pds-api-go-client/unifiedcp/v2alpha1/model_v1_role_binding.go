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

// V1RoleBinding RoleBinding represents the tenant/project/namespace level role bindings and resource IDS.
type V1RoleBinding struct {
	// Role name represents the role for a tenant/project/namespace.
	RoleName *string `json:"roleName,omitempty"`
	// Resource IDs represent the IDs bounded for the given role.
	ResourceIds []string `json:"resourceIds,omitempty"`
}

// NewV1RoleBinding instantiates a new V1RoleBinding object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewV1RoleBinding() *V1RoleBinding {
	this := V1RoleBinding{}
	return &this
}

// NewV1RoleBindingWithDefaults instantiates a new V1RoleBinding object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewV1RoleBindingWithDefaults() *V1RoleBinding {
	this := V1RoleBinding{}
	return &this
}

// GetRoleName returns the RoleName field value if set, zero value otherwise.
func (o *V1RoleBinding) GetRoleName() string {
	if o == nil || o.RoleName == nil {
		var ret string
		return ret
	}
	return *o.RoleName
}

// GetRoleNameOk returns a tuple with the RoleName field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1RoleBinding) GetRoleNameOk() (*string, bool) {
	if o == nil || o.RoleName == nil {
		return nil, false
	}
	return o.RoleName, true
}

// HasRoleName returns a boolean if a field has been set.
func (o *V1RoleBinding) HasRoleName() bool {
	if o != nil && o.RoleName != nil {
		return true
	}

	return false
}

// SetRoleName gets a reference to the given string and assigns it to the RoleName field.
func (o *V1RoleBinding) SetRoleName(v string) {
	o.RoleName = &v
}

// GetResourceIds returns the ResourceIds field value if set, zero value otherwise.
func (o *V1RoleBinding) GetResourceIds() []string {
	if o == nil || o.ResourceIds == nil {
		var ret []string
		return ret
	}
	return o.ResourceIds
}

// GetResourceIdsOk returns a tuple with the ResourceIds field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1RoleBinding) GetResourceIdsOk() ([]string, bool) {
	if o == nil || o.ResourceIds == nil {
		return nil, false
	}
	return o.ResourceIds, true
}

// HasResourceIds returns a boolean if a field has been set.
func (o *V1RoleBinding) HasResourceIds() bool {
	if o != nil && o.ResourceIds != nil {
		return true
	}

	return false
}

// SetResourceIds gets a reference to the given []string and assigns it to the ResourceIds field.
func (o *V1RoleBinding) SetResourceIds(v []string) {
	o.ResourceIds = v
}

func (o V1RoleBinding) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if o.RoleName != nil {
		toSerialize["roleName"] = o.RoleName
	}
	if o.ResourceIds != nil {
		toSerialize["resourceIds"] = o.ResourceIds
	}
	return json.Marshal(toSerialize)
}

type NullableV1RoleBinding struct {
	value *V1RoleBinding
	isSet bool
}

func (v NullableV1RoleBinding) Get() *V1RoleBinding {
	return v.value
}

func (v *NullableV1RoleBinding) Set(val *V1RoleBinding) {
	v.value = val
	v.isSet = true
}

func (v NullableV1RoleBinding) IsSet() bool {
	return v.isSet
}

func (v *NullableV1RoleBinding) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableV1RoleBinding(val *V1RoleBinding) *NullableV1RoleBinding {
	return &NullableV1RoleBinding{value: val, isSet: true}
}

func (v NullableV1RoleBinding) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableV1RoleBinding) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

