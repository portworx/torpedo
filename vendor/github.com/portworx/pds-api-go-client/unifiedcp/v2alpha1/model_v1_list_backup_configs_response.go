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

// V1ListBackupConfigsResponse Response of list of backup configurations.
type V1ListBackupConfigsResponse struct {
	// The list of backup configurations.
	BackupConfigs []V1BackupConfig `json:"backupConfigs,omitempty"`
	Pagination *V1PageBasedPaginationResponse `json:"pagination,omitempty"`
}

// NewV1ListBackupConfigsResponse instantiates a new V1ListBackupConfigsResponse object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewV1ListBackupConfigsResponse() *V1ListBackupConfigsResponse {
	this := V1ListBackupConfigsResponse{}
	return &this
}

// NewV1ListBackupConfigsResponseWithDefaults instantiates a new V1ListBackupConfigsResponse object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewV1ListBackupConfigsResponseWithDefaults() *V1ListBackupConfigsResponse {
	this := V1ListBackupConfigsResponse{}
	return &this
}

// GetBackupConfigs returns the BackupConfigs field value if set, zero value otherwise.
func (o *V1ListBackupConfigsResponse) GetBackupConfigs() []V1BackupConfig {
	if o == nil || o.BackupConfigs == nil {
		var ret []V1BackupConfig
		return ret
	}
	return o.BackupConfigs
}

// GetBackupConfigsOk returns a tuple with the BackupConfigs field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1ListBackupConfigsResponse) GetBackupConfigsOk() ([]V1BackupConfig, bool) {
	if o == nil || o.BackupConfigs == nil {
		return nil, false
	}
	return o.BackupConfigs, true
}

// HasBackupConfigs returns a boolean if a field has been set.
func (o *V1ListBackupConfigsResponse) HasBackupConfigs() bool {
	if o != nil && o.BackupConfigs != nil {
		return true
	}

	return false
}

// SetBackupConfigs gets a reference to the given []V1BackupConfig and assigns it to the BackupConfigs field.
func (o *V1ListBackupConfigsResponse) SetBackupConfigs(v []V1BackupConfig) {
	o.BackupConfigs = v
}

// GetPagination returns the Pagination field value if set, zero value otherwise.
func (o *V1ListBackupConfigsResponse) GetPagination() V1PageBasedPaginationResponse {
	if o == nil || o.Pagination == nil {
		var ret V1PageBasedPaginationResponse
		return ret
	}
	return *o.Pagination
}

// GetPaginationOk returns a tuple with the Pagination field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1ListBackupConfigsResponse) GetPaginationOk() (*V1PageBasedPaginationResponse, bool) {
	if o == nil || o.Pagination == nil {
		return nil, false
	}
	return o.Pagination, true
}

// HasPagination returns a boolean if a field has been set.
func (o *V1ListBackupConfigsResponse) HasPagination() bool {
	if o != nil && o.Pagination != nil {
		return true
	}

	return false
}

// SetPagination gets a reference to the given V1PageBasedPaginationResponse and assigns it to the Pagination field.
func (o *V1ListBackupConfigsResponse) SetPagination(v V1PageBasedPaginationResponse) {
	o.Pagination = &v
}

func (o V1ListBackupConfigsResponse) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if o.BackupConfigs != nil {
		toSerialize["backupConfigs"] = o.BackupConfigs
	}
	if o.Pagination != nil {
		toSerialize["pagination"] = o.Pagination
	}
	return json.Marshal(toSerialize)
}

type NullableV1ListBackupConfigsResponse struct {
	value *V1ListBackupConfigsResponse
	isSet bool
}

func (v NullableV1ListBackupConfigsResponse) Get() *V1ListBackupConfigsResponse {
	return v.value
}

func (v *NullableV1ListBackupConfigsResponse) Set(val *V1ListBackupConfigsResponse) {
	v.value = val
	v.isSet = true
}

func (v NullableV1ListBackupConfigsResponse) IsSet() bool {
	return v.isSet
}

func (v *NullableV1ListBackupConfigsResponse) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableV1ListBackupConfigsResponse(val *V1ListBackupConfigsResponse) *NullableV1ListBackupConfigsResponse {
	return &NullableV1ListBackupConfigsResponse{value: val, isSet: true}
}

func (v NullableV1ListBackupConfigsResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableV1ListBackupConfigsResponse) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

