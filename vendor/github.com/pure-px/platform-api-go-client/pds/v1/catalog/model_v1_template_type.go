/*
public/portworx/pds/catalog/dataservices/apiv1/dataservices.proto

No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)

API version: version not set
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package catalog

import (
	"encoding/json"
)

// checks if the V1TemplateType type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &V1TemplateType{}

// V1TemplateType Template type containing id, mame and description.
type V1TemplateType struct {
	// UID of the template type.
	Uid *string `json:"uid,omitempty"`
	// Name of the template type.
	Name *string `json:"name,omitempty"`
	// Description of the template type.
	Description *string `json:"description,omitempty"`
}

// NewV1TemplateType instantiates a new V1TemplateType object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewV1TemplateType() *V1TemplateType {
	this := V1TemplateType{}
	return &this
}

// NewV1TemplateTypeWithDefaults instantiates a new V1TemplateType object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewV1TemplateTypeWithDefaults() *V1TemplateType {
	this := V1TemplateType{}
	return &this
}

// GetUid returns the Uid field value if set, zero value otherwise.
func (o *V1TemplateType) GetUid() string {
	if o == nil || IsNil(o.Uid) {
		var ret string
		return ret
	}
	return *o.Uid
}

// GetUidOk returns a tuple with the Uid field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1TemplateType) GetUidOk() (*string, bool) {
	if o == nil || IsNil(o.Uid) {
		return nil, false
	}
	return o.Uid, true
}

// HasUid returns a boolean if a field has been set.
func (o *V1TemplateType) HasUid() bool {
	if o != nil && !IsNil(o.Uid) {
		return true
	}

	return false
}

// SetUid gets a reference to the given string and assigns it to the Uid field.
func (o *V1TemplateType) SetUid(v string) {
	o.Uid = &v
}

// GetName returns the Name field value if set, zero value otherwise.
func (o *V1TemplateType) GetName() string {
	if o == nil || IsNil(o.Name) {
		var ret string
		return ret
	}
	return *o.Name
}

// GetNameOk returns a tuple with the Name field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1TemplateType) GetNameOk() (*string, bool) {
	if o == nil || IsNil(o.Name) {
		return nil, false
	}
	return o.Name, true
}

// HasName returns a boolean if a field has been set.
func (o *V1TemplateType) HasName() bool {
	if o != nil && !IsNil(o.Name) {
		return true
	}

	return false
}

// SetName gets a reference to the given string and assigns it to the Name field.
func (o *V1TemplateType) SetName(v string) {
	o.Name = &v
}

// GetDescription returns the Description field value if set, zero value otherwise.
func (o *V1TemplateType) GetDescription() string {
	if o == nil || IsNil(o.Description) {
		var ret string
		return ret
	}
	return *o.Description
}

// GetDescriptionOk returns a tuple with the Description field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1TemplateType) GetDescriptionOk() (*string, bool) {
	if o == nil || IsNil(o.Description) {
		return nil, false
	}
	return o.Description, true
}

// HasDescription returns a boolean if a field has been set.
func (o *V1TemplateType) HasDescription() bool {
	if o != nil && !IsNil(o.Description) {
		return true
	}

	return false
}

// SetDescription gets a reference to the given string and assigns it to the Description field.
func (o *V1TemplateType) SetDescription(v string) {
	o.Description = &v
}

func (o V1TemplateType) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o V1TemplateType) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.Uid) {
		toSerialize["uid"] = o.Uid
	}
	if !IsNil(o.Name) {
		toSerialize["name"] = o.Name
	}
	if !IsNil(o.Description) {
		toSerialize["description"] = o.Description
	}
	return toSerialize, nil
}

type NullableV1TemplateType struct {
	value *V1TemplateType
	isSet bool
}

func (v NullableV1TemplateType) Get() *V1TemplateType {
	return v.value
}

func (v *NullableV1TemplateType) Set(val *V1TemplateType) {
	v.value = val
	v.isSet = true
}

func (v NullableV1TemplateType) IsSet() bool {
	return v.isSet
}

func (v *NullableV1TemplateType) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableV1TemplateType(val *V1TemplateType) *NullableV1TemplateType {
	return &NullableV1TemplateType{value: val, isSet: true}
}

func (v NullableV1TemplateType) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableV1TemplateType) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

