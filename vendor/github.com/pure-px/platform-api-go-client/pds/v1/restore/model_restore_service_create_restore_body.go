/*
public/portworx/pds/restore/apiv1/restore.proto

No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)

API version: version not set
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package restore

import (
	"encoding/json"
	"bytes"
	"fmt"
)

// checks if the RestoreServiceCreateRestoreBody type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &RestoreServiceCreateRestoreBody{}

// RestoreServiceCreateRestoreBody Request to create a restore.
type RestoreServiceCreateRestoreBody struct {
	// UID of the project associated with the restore.
	ProjectId string `json:"projectId"`
	Restore V1Restore `json:"restore"`
}

type _RestoreServiceCreateRestoreBody RestoreServiceCreateRestoreBody

// NewRestoreServiceCreateRestoreBody instantiates a new RestoreServiceCreateRestoreBody object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewRestoreServiceCreateRestoreBody(projectId string, restore V1Restore) *RestoreServiceCreateRestoreBody {
	this := RestoreServiceCreateRestoreBody{}
	this.ProjectId = projectId
	this.Restore = restore
	return &this
}

// NewRestoreServiceCreateRestoreBodyWithDefaults instantiates a new RestoreServiceCreateRestoreBody object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewRestoreServiceCreateRestoreBodyWithDefaults() *RestoreServiceCreateRestoreBody {
	this := RestoreServiceCreateRestoreBody{}
	return &this
}

// GetProjectId returns the ProjectId field value
func (o *RestoreServiceCreateRestoreBody) GetProjectId() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.ProjectId
}

// GetProjectIdOk returns a tuple with the ProjectId field value
// and a boolean to check if the value has been set.
func (o *RestoreServiceCreateRestoreBody) GetProjectIdOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.ProjectId, true
}

// SetProjectId sets field value
func (o *RestoreServiceCreateRestoreBody) SetProjectId(v string) {
	o.ProjectId = v
}

// GetRestore returns the Restore field value
func (o *RestoreServiceCreateRestoreBody) GetRestore() V1Restore {
	if o == nil {
		var ret V1Restore
		return ret
	}

	return o.Restore
}

// GetRestoreOk returns a tuple with the Restore field value
// and a boolean to check if the value has been set.
func (o *RestoreServiceCreateRestoreBody) GetRestoreOk() (*V1Restore, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Restore, true
}

// SetRestore sets field value
func (o *RestoreServiceCreateRestoreBody) SetRestore(v V1Restore) {
	o.Restore = v
}

func (o RestoreServiceCreateRestoreBody) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o RestoreServiceCreateRestoreBody) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["projectId"] = o.ProjectId
	toSerialize["restore"] = o.Restore
	return toSerialize, nil
}

func (o *RestoreServiceCreateRestoreBody) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"projectId",
		"restore",
	}

	allProperties := make(map[string]interface{})

	err = json.Unmarshal(data, &allProperties)

	if err != nil {
		return err;
	}

	for _, requiredProperty := range(requiredProperties) {
		if _, exists := allProperties[requiredProperty]; !exists {
			return fmt.Errorf("no value given for required property %v", requiredProperty)
		}
	}

	varRestoreServiceCreateRestoreBody := _RestoreServiceCreateRestoreBody{}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&varRestoreServiceCreateRestoreBody)

	if err != nil {
		return err
	}

	*o = RestoreServiceCreateRestoreBody(varRestoreServiceCreateRestoreBody)

	return err
}

type NullableRestoreServiceCreateRestoreBody struct {
	value *RestoreServiceCreateRestoreBody
	isSet bool
}

func (v NullableRestoreServiceCreateRestoreBody) Get() *RestoreServiceCreateRestoreBody {
	return v.value
}

func (v *NullableRestoreServiceCreateRestoreBody) Set(val *RestoreServiceCreateRestoreBody) {
	v.value = val
	v.isSet = true
}

func (v NullableRestoreServiceCreateRestoreBody) IsSet() bool {
	return v.isSet
}

func (v *NullableRestoreServiceCreateRestoreBody) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableRestoreServiceCreateRestoreBody(val *RestoreServiceCreateRestoreBody) *NullableRestoreServiceCreateRestoreBody {
	return &NullableRestoreServiceCreateRestoreBody{value: val, isSet: true}
}

func (v NullableRestoreServiceCreateRestoreBody) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableRestoreServiceCreateRestoreBody) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

