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

// V1ListDeploymentConfigUpdatesResponse struct for V1ListDeploymentConfigUpdatesResponse
type V1ListDeploymentConfigUpdatesResponse struct {
	DeploymentConfigUpdates []V1DeploymentConfigUpdate `json:"deploymentConfigUpdates,omitempty"`
	Pagination *V1PageBasedPaginationResponse `json:"pagination,omitempty"`
}

// NewV1ListDeploymentConfigUpdatesResponse instantiates a new V1ListDeploymentConfigUpdatesResponse object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewV1ListDeploymentConfigUpdatesResponse() *V1ListDeploymentConfigUpdatesResponse {
	this := V1ListDeploymentConfigUpdatesResponse{}
	return &this
}

// NewV1ListDeploymentConfigUpdatesResponseWithDefaults instantiates a new V1ListDeploymentConfigUpdatesResponse object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewV1ListDeploymentConfigUpdatesResponseWithDefaults() *V1ListDeploymentConfigUpdatesResponse {
	this := V1ListDeploymentConfigUpdatesResponse{}
	return &this
}

// GetDeploymentConfigUpdates returns the DeploymentConfigUpdates field value if set, zero value otherwise.
func (o *V1ListDeploymentConfigUpdatesResponse) GetDeploymentConfigUpdates() []V1DeploymentConfigUpdate {
	if o == nil || o.DeploymentConfigUpdates == nil {
		var ret []V1DeploymentConfigUpdate
		return ret
	}
	return o.DeploymentConfigUpdates
}

// GetDeploymentConfigUpdatesOk returns a tuple with the DeploymentConfigUpdates field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1ListDeploymentConfigUpdatesResponse) GetDeploymentConfigUpdatesOk() ([]V1DeploymentConfigUpdate, bool) {
	if o == nil || o.DeploymentConfigUpdates == nil {
		return nil, false
	}
	return o.DeploymentConfigUpdates, true
}

// HasDeploymentConfigUpdates returns a boolean if a field has been set.
func (o *V1ListDeploymentConfigUpdatesResponse) HasDeploymentConfigUpdates() bool {
	if o != nil && o.DeploymentConfigUpdates != nil {
		return true
	}

	return false
}

// SetDeploymentConfigUpdates gets a reference to the given []V1DeploymentConfigUpdate and assigns it to the DeploymentConfigUpdates field.
func (o *V1ListDeploymentConfigUpdatesResponse) SetDeploymentConfigUpdates(v []V1DeploymentConfigUpdate) {
	o.DeploymentConfigUpdates = v
}

// GetPagination returns the Pagination field value if set, zero value otherwise.
func (o *V1ListDeploymentConfigUpdatesResponse) GetPagination() V1PageBasedPaginationResponse {
	if o == nil || o.Pagination == nil {
		var ret V1PageBasedPaginationResponse
		return ret
	}
	return *o.Pagination
}

// GetPaginationOk returns a tuple with the Pagination field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1ListDeploymentConfigUpdatesResponse) GetPaginationOk() (*V1PageBasedPaginationResponse, bool) {
	if o == nil || o.Pagination == nil {
		return nil, false
	}
	return o.Pagination, true
}

// HasPagination returns a boolean if a field has been set.
func (o *V1ListDeploymentConfigUpdatesResponse) HasPagination() bool {
	if o != nil && o.Pagination != nil {
		return true
	}

	return false
}

// SetPagination gets a reference to the given V1PageBasedPaginationResponse and assigns it to the Pagination field.
func (o *V1ListDeploymentConfigUpdatesResponse) SetPagination(v V1PageBasedPaginationResponse) {
	o.Pagination = &v
}

func (o V1ListDeploymentConfigUpdatesResponse) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if o.DeploymentConfigUpdates != nil {
		toSerialize["deploymentConfigUpdates"] = o.DeploymentConfigUpdates
	}
	if o.Pagination != nil {
		toSerialize["pagination"] = o.Pagination
	}
	return json.Marshal(toSerialize)
}

type NullableV1ListDeploymentConfigUpdatesResponse struct {
	value *V1ListDeploymentConfigUpdatesResponse
	isSet bool
}

func (v NullableV1ListDeploymentConfigUpdatesResponse) Get() *V1ListDeploymentConfigUpdatesResponse {
	return v.value
}

func (v *NullableV1ListDeploymentConfigUpdatesResponse) Set(val *V1ListDeploymentConfigUpdatesResponse) {
	v.value = val
	v.isSet = true
}

func (v NullableV1ListDeploymentConfigUpdatesResponse) IsSet() bool {
	return v.isSet
}

func (v *NullableV1ListDeploymentConfigUpdatesResponse) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableV1ListDeploymentConfigUpdatesResponse(val *V1ListDeploymentConfigUpdatesResponse) *NullableV1ListDeploymentConfigUpdatesResponse {
	return &NullableV1ListDeploymentConfigUpdatesResponse{value: val, isSet: true}
}

func (v NullableV1ListDeploymentConfigUpdatesResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableV1ListDeploymentConfigUpdatesResponse) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

