/*
public/portworx/pds/deploymentconfigupdate/apiv1/deploymentconfigupdate.proto

No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)

API version: version not set
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package deploymentconfigupdate

import (
	"encoding/json"
)

// checks if the Pdsdeploymentconfigupdatev1Config type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &Pdsdeploymentconfigupdatev1Config{}

// Pdsdeploymentconfigupdatev1Config Config of the desired deployment configuration.
type Pdsdeploymentconfigupdatev1Config struct {
	DeploymentMeta *V1Meta `json:"deploymentMeta,omitempty"`
	DeploymentConfig *Pdsdeploymentv1Config `json:"deploymentConfig,omitempty"`
}

// NewPdsdeploymentconfigupdatev1Config instantiates a new Pdsdeploymentconfigupdatev1Config object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewPdsdeploymentconfigupdatev1Config() *Pdsdeploymentconfigupdatev1Config {
	this := Pdsdeploymentconfigupdatev1Config{}
	return &this
}

// NewPdsdeploymentconfigupdatev1ConfigWithDefaults instantiates a new Pdsdeploymentconfigupdatev1Config object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewPdsdeploymentconfigupdatev1ConfigWithDefaults() *Pdsdeploymentconfigupdatev1Config {
	this := Pdsdeploymentconfigupdatev1Config{}
	return &this
}

// GetDeploymentMeta returns the DeploymentMeta field value if set, zero value otherwise.
func (o *Pdsdeploymentconfigupdatev1Config) GetDeploymentMeta() V1Meta {
	if o == nil || IsNil(o.DeploymentMeta) {
		var ret V1Meta
		return ret
	}
	return *o.DeploymentMeta
}

// GetDeploymentMetaOk returns a tuple with the DeploymentMeta field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Pdsdeploymentconfigupdatev1Config) GetDeploymentMetaOk() (*V1Meta, bool) {
	if o == nil || IsNil(o.DeploymentMeta) {
		return nil, false
	}
	return o.DeploymentMeta, true
}

// HasDeploymentMeta returns a boolean if a field has been set.
func (o *Pdsdeploymentconfigupdatev1Config) HasDeploymentMeta() bool {
	if o != nil && !IsNil(o.DeploymentMeta) {
		return true
	}

	return false
}

// SetDeploymentMeta gets a reference to the given V1Meta and assigns it to the DeploymentMeta field.
func (o *Pdsdeploymentconfigupdatev1Config) SetDeploymentMeta(v V1Meta) {
	o.DeploymentMeta = &v
}

// GetDeploymentConfig returns the DeploymentConfig field value if set, zero value otherwise.
func (o *Pdsdeploymentconfigupdatev1Config) GetDeploymentConfig() Pdsdeploymentv1Config {
	if o == nil || IsNil(o.DeploymentConfig) {
		var ret Pdsdeploymentv1Config
		return ret
	}
	return *o.DeploymentConfig
}

// GetDeploymentConfigOk returns a tuple with the DeploymentConfig field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Pdsdeploymentconfigupdatev1Config) GetDeploymentConfigOk() (*Pdsdeploymentv1Config, bool) {
	if o == nil || IsNil(o.DeploymentConfig) {
		return nil, false
	}
	return o.DeploymentConfig, true
}

// HasDeploymentConfig returns a boolean if a field has been set.
func (o *Pdsdeploymentconfigupdatev1Config) HasDeploymentConfig() bool {
	if o != nil && !IsNil(o.DeploymentConfig) {
		return true
	}

	return false
}

// SetDeploymentConfig gets a reference to the given Pdsdeploymentv1Config and assigns it to the DeploymentConfig field.
func (o *Pdsdeploymentconfigupdatev1Config) SetDeploymentConfig(v Pdsdeploymentv1Config) {
	o.DeploymentConfig = &v
}

func (o Pdsdeploymentconfigupdatev1Config) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o Pdsdeploymentconfigupdatev1Config) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.DeploymentMeta) {
		toSerialize["deploymentMeta"] = o.DeploymentMeta
	}
	if !IsNil(o.DeploymentConfig) {
		toSerialize["deploymentConfig"] = o.DeploymentConfig
	}
	return toSerialize, nil
}

type NullablePdsdeploymentconfigupdatev1Config struct {
	value *Pdsdeploymentconfigupdatev1Config
	isSet bool
}

func (v NullablePdsdeploymentconfigupdatev1Config) Get() *Pdsdeploymentconfigupdatev1Config {
	return v.value
}

func (v *NullablePdsdeploymentconfigupdatev1Config) Set(val *Pdsdeploymentconfigupdatev1Config) {
	v.value = val
	v.isSet = true
}

func (v NullablePdsdeploymentconfigupdatev1Config) IsSet() bool {
	return v.isSet
}

func (v *NullablePdsdeploymentconfigupdatev1Config) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullablePdsdeploymentconfigupdatev1Config(val *Pdsdeploymentconfigupdatev1Config) *NullablePdsdeploymentconfigupdatev1Config {
	return &NullablePdsdeploymentconfigupdatev1Config{value: val, isSet: true}
}

func (v NullablePdsdeploymentconfigupdatev1Config) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullablePdsdeploymentconfigupdatev1Config) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

