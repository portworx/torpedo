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

// Platforminvitationv1Config struct for Platforminvitationv1Config
type Platforminvitationv1Config struct {
	UserEmail string `json:"userEmail"`
	AccessPolicy V1AccessPolicy `json:"accessPolicy"`
	AccountId *string `json:"accountId,omitempty"`
	TenantId *string `json:"tenantId,omitempty"`
	ProjectId *string `json:"projectId,omitempty"`
}

// NewPlatforminvitationv1Config instantiates a new Platforminvitationv1Config object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewPlatforminvitationv1Config(userEmail string, accessPolicy V1AccessPolicy) *Platforminvitationv1Config {
	this := Platforminvitationv1Config{}
	this.UserEmail = userEmail
	this.AccessPolicy = accessPolicy
	return &this
}

// NewPlatforminvitationv1ConfigWithDefaults instantiates a new Platforminvitationv1Config object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewPlatforminvitationv1ConfigWithDefaults() *Platforminvitationv1Config {
	this := Platforminvitationv1Config{}
	return &this
}

// GetUserEmail returns the UserEmail field value
func (o *Platforminvitationv1Config) GetUserEmail() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.UserEmail
}

// GetUserEmailOk returns a tuple with the UserEmail field value
// and a boolean to check if the value has been set.
func (o *Platforminvitationv1Config) GetUserEmailOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.UserEmail, true
}

// SetUserEmail sets field value
func (o *Platforminvitationv1Config) SetUserEmail(v string) {
	o.UserEmail = v
}

// GetAccessPolicy returns the AccessPolicy field value
func (o *Platforminvitationv1Config) GetAccessPolicy() V1AccessPolicy {
	if o == nil {
		var ret V1AccessPolicy
		return ret
	}

	return o.AccessPolicy
}

// GetAccessPolicyOk returns a tuple with the AccessPolicy field value
// and a boolean to check if the value has been set.
func (o *Platforminvitationv1Config) GetAccessPolicyOk() (*V1AccessPolicy, bool) {
	if o == nil {
		return nil, false
	}
	return &o.AccessPolicy, true
}

// SetAccessPolicy sets field value
func (o *Platforminvitationv1Config) SetAccessPolicy(v V1AccessPolicy) {
	o.AccessPolicy = v
}

// GetAccountId returns the AccountId field value if set, zero value otherwise.
func (o *Platforminvitationv1Config) GetAccountId() string {
	if o == nil || o.AccountId == nil {
		var ret string
		return ret
	}
	return *o.AccountId
}

// GetAccountIdOk returns a tuple with the AccountId field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Platforminvitationv1Config) GetAccountIdOk() (*string, bool) {
	if o == nil || o.AccountId == nil {
		return nil, false
	}
	return o.AccountId, true
}

// HasAccountId returns a boolean if a field has been set.
func (o *Platforminvitationv1Config) HasAccountId() bool {
	if o != nil && o.AccountId != nil {
		return true
	}

	return false
}

// SetAccountId gets a reference to the given string and assigns it to the AccountId field.
func (o *Platforminvitationv1Config) SetAccountId(v string) {
	o.AccountId = &v
}

// GetTenantId returns the TenantId field value if set, zero value otherwise.
func (o *Platforminvitationv1Config) GetTenantId() string {
	if o == nil || o.TenantId == nil {
		var ret string
		return ret
	}
	return *o.TenantId
}

// GetTenantIdOk returns a tuple with the TenantId field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Platforminvitationv1Config) GetTenantIdOk() (*string, bool) {
	if o == nil || o.TenantId == nil {
		return nil, false
	}
	return o.TenantId, true
}

// HasTenantId returns a boolean if a field has been set.
func (o *Platforminvitationv1Config) HasTenantId() bool {
	if o != nil && o.TenantId != nil {
		return true
	}

	return false
}

// SetTenantId gets a reference to the given string and assigns it to the TenantId field.
func (o *Platforminvitationv1Config) SetTenantId(v string) {
	o.TenantId = &v
}

// GetProjectId returns the ProjectId field value if set, zero value otherwise.
func (o *Platforminvitationv1Config) GetProjectId() string {
	if o == nil || o.ProjectId == nil {
		var ret string
		return ret
	}
	return *o.ProjectId
}

// GetProjectIdOk returns a tuple with the ProjectId field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Platforminvitationv1Config) GetProjectIdOk() (*string, bool) {
	if o == nil || o.ProjectId == nil {
		return nil, false
	}
	return o.ProjectId, true
}

// HasProjectId returns a boolean if a field has been set.
func (o *Platforminvitationv1Config) HasProjectId() bool {
	if o != nil && o.ProjectId != nil {
		return true
	}

	return false
}

// SetProjectId gets a reference to the given string and assigns it to the ProjectId field.
func (o *Platforminvitationv1Config) SetProjectId(v string) {
	o.ProjectId = &v
}

func (o Platforminvitationv1Config) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if true {
		toSerialize["userEmail"] = o.UserEmail
	}
	if true {
		toSerialize["accessPolicy"] = o.AccessPolicy
	}
	if o.AccountId != nil {
		toSerialize["accountId"] = o.AccountId
	}
	if o.TenantId != nil {
		toSerialize["tenantId"] = o.TenantId
	}
	if o.ProjectId != nil {
		toSerialize["projectId"] = o.ProjectId
	}
	return json.Marshal(toSerialize)
}

type NullablePlatforminvitationv1Config struct {
	value *Platforminvitationv1Config
	isSet bool
}

func (v NullablePlatforminvitationv1Config) Get() *Platforminvitationv1Config {
	return v.value
}

func (v *NullablePlatforminvitationv1Config) Set(val *Platforminvitationv1Config) {
	v.value = val
	v.isSet = true
}

func (v NullablePlatforminvitationv1Config) IsSet() bool {
	return v.isSet
}

func (v *NullablePlatforminvitationv1Config) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullablePlatforminvitationv1Config(val *Platforminvitationv1Config) *NullablePlatforminvitationv1Config {
	return &NullablePlatforminvitationv1Config{value: val, isSet: true}
}

func (v NullablePlatforminvitationv1Config) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullablePlatforminvitationv1Config) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

