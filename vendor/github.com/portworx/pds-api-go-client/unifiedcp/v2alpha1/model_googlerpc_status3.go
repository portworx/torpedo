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

// GooglerpcStatus3 struct for GooglerpcStatus3
type GooglerpcStatus3 struct {
	Code *int32 `json:"code,omitempty"`
	Message *string `json:"message,omitempty"`
	Details []ProtobufAny3 `json:"details,omitempty"`
}

// NewGooglerpcStatus3 instantiates a new GooglerpcStatus3 object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewGooglerpcStatus3() *GooglerpcStatus3 {
	this := GooglerpcStatus3{}
	return &this
}

// NewGooglerpcStatus3WithDefaults instantiates a new GooglerpcStatus3 object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewGooglerpcStatus3WithDefaults() *GooglerpcStatus3 {
	this := GooglerpcStatus3{}
	return &this
}

// GetCode returns the Code field value if set, zero value otherwise.
func (o *GooglerpcStatus3) GetCode() int32 {
	if o == nil || o.Code == nil {
		var ret int32
		return ret
	}
	return *o.Code
}

// GetCodeOk returns a tuple with the Code field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *GooglerpcStatus3) GetCodeOk() (*int32, bool) {
	if o == nil || o.Code == nil {
		return nil, false
	}
	return o.Code, true
}

// HasCode returns a boolean if a field has been set.
func (o *GooglerpcStatus3) HasCode() bool {
	if o != nil && o.Code != nil {
		return true
	}

	return false
}

// SetCode gets a reference to the given int32 and assigns it to the Code field.
func (o *GooglerpcStatus3) SetCode(v int32) {
	o.Code = &v
}

// GetMessage returns the Message field value if set, zero value otherwise.
func (o *GooglerpcStatus3) GetMessage() string {
	if o == nil || o.Message == nil {
		var ret string
		return ret
	}
	return *o.Message
}

// GetMessageOk returns a tuple with the Message field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *GooglerpcStatus3) GetMessageOk() (*string, bool) {
	if o == nil || o.Message == nil {
		return nil, false
	}
	return o.Message, true
}

// HasMessage returns a boolean if a field has been set.
func (o *GooglerpcStatus3) HasMessage() bool {
	if o != nil && o.Message != nil {
		return true
	}

	return false
}

// SetMessage gets a reference to the given string and assigns it to the Message field.
func (o *GooglerpcStatus3) SetMessage(v string) {
	o.Message = &v
}

// GetDetails returns the Details field value if set, zero value otherwise.
func (o *GooglerpcStatus3) GetDetails() []ProtobufAny3 {
	if o == nil || o.Details == nil {
		var ret []ProtobufAny3
		return ret
	}
	return o.Details
}

// GetDetailsOk returns a tuple with the Details field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *GooglerpcStatus3) GetDetailsOk() ([]ProtobufAny3, bool) {
	if o == nil || o.Details == nil {
		return nil, false
	}
	return o.Details, true
}

// HasDetails returns a boolean if a field has been set.
func (o *GooglerpcStatus3) HasDetails() bool {
	if o != nil && o.Details != nil {
		return true
	}

	return false
}

// SetDetails gets a reference to the given []ProtobufAny3 and assigns it to the Details field.
func (o *GooglerpcStatus3) SetDetails(v []ProtobufAny3) {
	o.Details = v
}

func (o GooglerpcStatus3) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if o.Code != nil {
		toSerialize["code"] = o.Code
	}
	if o.Message != nil {
		toSerialize["message"] = o.Message
	}
	if o.Details != nil {
		toSerialize["details"] = o.Details
	}
	return json.Marshal(toSerialize)
}

type NullableGooglerpcStatus3 struct {
	value *GooglerpcStatus3
	isSet bool
}

func (v NullableGooglerpcStatus3) Get() *GooglerpcStatus3 {
	return v.value
}

func (v *NullableGooglerpcStatus3) Set(val *GooglerpcStatus3) {
	v.value = val
	v.isSet = true
}

func (v NullableGooglerpcStatus3) IsSet() bool {
	return v.isSet
}

func (v *NullableGooglerpcStatus3) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableGooglerpcStatus3(val *GooglerpcStatus3) *NullableGooglerpcStatus3 {
	return &NullableGooglerpcStatus3{value: val, isSet: true}
}

func (v NullableGooglerpcStatus3) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableGooglerpcStatus3) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

