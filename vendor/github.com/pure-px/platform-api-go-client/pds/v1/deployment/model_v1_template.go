/*
public/portworx/pds/deployment/apiv1/deployment.proto

No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)

API version: version not set
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package deployment

import (
	"encoding/json"
)

// checks if the V1Template type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &V1Template{}

// V1Template Template.
type V1Template struct {
	// UID of the Template.
	Id *string `json:"id,omitempty"`
	// Resource version of the template.
	ResourceVersion *string `json:"resourceVersion,omitempty"`
	// Values required for template.
	Values *map[string]string `json:"values,omitempty"`
}

// NewV1Template instantiates a new V1Template object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewV1Template() *V1Template {
	this := V1Template{}
	return &this
}

// NewV1TemplateWithDefaults instantiates a new V1Template object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewV1TemplateWithDefaults() *V1Template {
	this := V1Template{}
	return &this
}

// GetId returns the Id field value if set, zero value otherwise.
func (o *V1Template) GetId() string {
	if o == nil || IsNil(o.Id) {
		var ret string
		return ret
	}
	return *o.Id
}

// GetIdOk returns a tuple with the Id field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1Template) GetIdOk() (*string, bool) {
	if o == nil || IsNil(o.Id) {
		return nil, false
	}
	return o.Id, true
}

// HasId returns a boolean if a field has been set.
func (o *V1Template) HasId() bool {
	if o != nil && !IsNil(o.Id) {
		return true
	}

	return false
}

// SetId gets a reference to the given string and assigns it to the Id field.
func (o *V1Template) SetId(v string) {
	o.Id = &v
}

// GetResourceVersion returns the ResourceVersion field value if set, zero value otherwise.
func (o *V1Template) GetResourceVersion() string {
	if o == nil || IsNil(o.ResourceVersion) {
		var ret string
		return ret
	}
	return *o.ResourceVersion
}

// GetResourceVersionOk returns a tuple with the ResourceVersion field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1Template) GetResourceVersionOk() (*string, bool) {
	if o == nil || IsNil(o.ResourceVersion) {
		return nil, false
	}
	return o.ResourceVersion, true
}

// HasResourceVersion returns a boolean if a field has been set.
func (o *V1Template) HasResourceVersion() bool {
	if o != nil && !IsNil(o.ResourceVersion) {
		return true
	}

	return false
}

// SetResourceVersion gets a reference to the given string and assigns it to the ResourceVersion field.
func (o *V1Template) SetResourceVersion(v string) {
	o.ResourceVersion = &v
}

// GetValues returns the Values field value if set, zero value otherwise.
func (o *V1Template) GetValues() map[string]string {
	if o == nil || IsNil(o.Values) {
		var ret map[string]string
		return ret
	}
	return *o.Values
}

// GetValuesOk returns a tuple with the Values field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1Template) GetValuesOk() (*map[string]string, bool) {
	if o == nil || IsNil(o.Values) {
		return nil, false
	}
	return o.Values, true
}

// HasValues returns a boolean if a field has been set.
func (o *V1Template) HasValues() bool {
	if o != nil && !IsNil(o.Values) {
		return true
	}

	return false
}

// SetValues gets a reference to the given map[string]string and assigns it to the Values field.
func (o *V1Template) SetValues(v map[string]string) {
	o.Values = &v
}

func (o V1Template) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o V1Template) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.Id) {
		toSerialize["id"] = o.Id
	}
	if !IsNil(o.ResourceVersion) {
		toSerialize["resourceVersion"] = o.ResourceVersion
	}
	if !IsNil(o.Values) {
		toSerialize["values"] = o.Values
	}
	return toSerialize, nil
}

type NullableV1Template struct {
	value *V1Template
	isSet bool
}

func (v NullableV1Template) Get() *V1Template {
	return v.value
}

func (v *NullableV1Template) Set(val *V1Template) {
	v.value = val
	v.isSet = true
}

func (v NullableV1Template) IsSet() bool {
	return v.isSet
}

func (v *NullableV1Template) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableV1Template(val *V1Template) *NullableV1Template {
	return &NullableV1Template{value: val, isSet: true}
}

func (v NullableV1Template) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableV1Template) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

