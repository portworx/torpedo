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

// V1PageBasedPaginationRequest struct for V1PageBasedPaginationRequest
type V1PageBasedPaginationRequest struct {
	PageNumber *string `json:"pageNumber,omitempty"`
	PageSize *string `json:"pageSize,omitempty"`
}

// NewV1PageBasedPaginationRequest instantiates a new V1PageBasedPaginationRequest object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewV1PageBasedPaginationRequest() *V1PageBasedPaginationRequest {
	this := V1PageBasedPaginationRequest{}
	return &this
}

// NewV1PageBasedPaginationRequestWithDefaults instantiates a new V1PageBasedPaginationRequest object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewV1PageBasedPaginationRequestWithDefaults() *V1PageBasedPaginationRequest {
	this := V1PageBasedPaginationRequest{}
	return &this
}

// GetPageNumber returns the PageNumber field value if set, zero value otherwise.
func (o *V1PageBasedPaginationRequest) GetPageNumber() string {
	if o == nil || o.PageNumber == nil {
		var ret string
		return ret
	}
	return *o.PageNumber
}

// GetPageNumberOk returns a tuple with the PageNumber field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1PageBasedPaginationRequest) GetPageNumberOk() (*string, bool) {
	if o == nil || o.PageNumber == nil {
		return nil, false
	}
	return o.PageNumber, true
}

// HasPageNumber returns a boolean if a field has been set.
func (o *V1PageBasedPaginationRequest) HasPageNumber() bool {
	if o != nil && o.PageNumber != nil {
		return true
	}

	return false
}

// SetPageNumber gets a reference to the given string and assigns it to the PageNumber field.
func (o *V1PageBasedPaginationRequest) SetPageNumber(v string) {
	o.PageNumber = &v
}

// GetPageSize returns the PageSize field value if set, zero value otherwise.
func (o *V1PageBasedPaginationRequest) GetPageSize() string {
	if o == nil || o.PageSize == nil {
		var ret string
		return ret
	}
	return *o.PageSize
}

// GetPageSizeOk returns a tuple with the PageSize field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1PageBasedPaginationRequest) GetPageSizeOk() (*string, bool) {
	if o == nil || o.PageSize == nil {
		return nil, false
	}
	return o.PageSize, true
}

// HasPageSize returns a boolean if a field has been set.
func (o *V1PageBasedPaginationRequest) HasPageSize() bool {
	if o != nil && o.PageSize != nil {
		return true
	}

	return false
}

// SetPageSize gets a reference to the given string and assigns it to the PageSize field.
func (o *V1PageBasedPaginationRequest) SetPageSize(v string) {
	o.PageSize = &v
}

func (o V1PageBasedPaginationRequest) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if o.PageNumber != nil {
		toSerialize["pageNumber"] = o.PageNumber
	}
	if o.PageSize != nil {
		toSerialize["pageSize"] = o.PageSize
	}
	return json.Marshal(toSerialize)
}

type NullableV1PageBasedPaginationRequest struct {
	value *V1PageBasedPaginationRequest
	isSet bool
}

func (v NullableV1PageBasedPaginationRequest) Get() *V1PageBasedPaginationRequest {
	return v.value
}

func (v *NullableV1PageBasedPaginationRequest) Set(val *V1PageBasedPaginationRequest) {
	v.value = val
	v.isSet = true
}

func (v NullableV1PageBasedPaginationRequest) IsSet() bool {
	return v.isSet
}

func (v *NullableV1PageBasedPaginationRequest) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableV1PageBasedPaginationRequest(val *V1PageBasedPaginationRequest) *NullableV1PageBasedPaginationRequest {
	return &NullableV1PageBasedPaginationRequest{value: val, isSet: true}
}

func (v NullableV1PageBasedPaginationRequest) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableV1PageBasedPaginationRequest) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

