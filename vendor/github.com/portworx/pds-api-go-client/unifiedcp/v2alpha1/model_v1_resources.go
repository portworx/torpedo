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

// V1Resources Infra resource are platform managed resources, used by associated applications.
type V1Resources struct {
	// Clusters represents the target k8s clusters.
	Clusters []string `json:"clusters,omitempty"`
	// Namespaces created in k8s cluster to provide the logical isolation.
	Namespaces []string `json:"namespaces,omitempty"`
	// Credentials required to connect to a backup target.
	Credentials []string `json:"credentials,omitempty"`
	// Backup locations where backups can be placed.
	BackupLocations []string `json:"backupLocations,omitempty"`
	// Templates can be used by applications to manage its resources.
	Templates []string `json:"templates,omitempty"`
}

// NewV1Resources instantiates a new V1Resources object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewV1Resources() *V1Resources {
	this := V1Resources{}
	return &this
}

// NewV1ResourcesWithDefaults instantiates a new V1Resources object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewV1ResourcesWithDefaults() *V1Resources {
	this := V1Resources{}
	return &this
}

// GetClusters returns the Clusters field value if set, zero value otherwise.
func (o *V1Resources) GetClusters() []string {
	if o == nil || o.Clusters == nil {
		var ret []string
		return ret
	}
	return o.Clusters
}

// GetClustersOk returns a tuple with the Clusters field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1Resources) GetClustersOk() ([]string, bool) {
	if o == nil || o.Clusters == nil {
		return nil, false
	}
	return o.Clusters, true
}

// HasClusters returns a boolean if a field has been set.
func (o *V1Resources) HasClusters() bool {
	if o != nil && o.Clusters != nil {
		return true
	}

	return false
}

// SetClusters gets a reference to the given []string and assigns it to the Clusters field.
func (o *V1Resources) SetClusters(v []string) {
	o.Clusters = v
}

// GetNamespaces returns the Namespaces field value if set, zero value otherwise.
func (o *V1Resources) GetNamespaces() []string {
	if o == nil || o.Namespaces == nil {
		var ret []string
		return ret
	}
	return o.Namespaces
}

// GetNamespacesOk returns a tuple with the Namespaces field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1Resources) GetNamespacesOk() ([]string, bool) {
	if o == nil || o.Namespaces == nil {
		return nil, false
	}
	return o.Namespaces, true
}

// HasNamespaces returns a boolean if a field has been set.
func (o *V1Resources) HasNamespaces() bool {
	if o != nil && o.Namespaces != nil {
		return true
	}

	return false
}

// SetNamespaces gets a reference to the given []string and assigns it to the Namespaces field.
func (o *V1Resources) SetNamespaces(v []string) {
	o.Namespaces = v
}

// GetCredentials returns the Credentials field value if set, zero value otherwise.
func (o *V1Resources) GetCredentials() []string {
	if o == nil || o.Credentials == nil {
		var ret []string
		return ret
	}
	return o.Credentials
}

// GetCredentialsOk returns a tuple with the Credentials field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1Resources) GetCredentialsOk() ([]string, bool) {
	if o == nil || o.Credentials == nil {
		return nil, false
	}
	return o.Credentials, true
}

// HasCredentials returns a boolean if a field has been set.
func (o *V1Resources) HasCredentials() bool {
	if o != nil && o.Credentials != nil {
		return true
	}

	return false
}

// SetCredentials gets a reference to the given []string and assigns it to the Credentials field.
func (o *V1Resources) SetCredentials(v []string) {
	o.Credentials = v
}

// GetBackupLocations returns the BackupLocations field value if set, zero value otherwise.
func (o *V1Resources) GetBackupLocations() []string {
	if o == nil || o.BackupLocations == nil {
		var ret []string
		return ret
	}
	return o.BackupLocations
}

// GetBackupLocationsOk returns a tuple with the BackupLocations field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1Resources) GetBackupLocationsOk() ([]string, bool) {
	if o == nil || o.BackupLocations == nil {
		return nil, false
	}
	return o.BackupLocations, true
}

// HasBackupLocations returns a boolean if a field has been set.
func (o *V1Resources) HasBackupLocations() bool {
	if o != nil && o.BackupLocations != nil {
		return true
	}

	return false
}

// SetBackupLocations gets a reference to the given []string and assigns it to the BackupLocations field.
func (o *V1Resources) SetBackupLocations(v []string) {
	o.BackupLocations = v
}

// GetTemplates returns the Templates field value if set, zero value otherwise.
func (o *V1Resources) GetTemplates() []string {
	if o == nil || o.Templates == nil {
		var ret []string
		return ret
	}
	return o.Templates
}

// GetTemplatesOk returns a tuple with the Templates field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1Resources) GetTemplatesOk() ([]string, bool) {
	if o == nil || o.Templates == nil {
		return nil, false
	}
	return o.Templates, true
}

// HasTemplates returns a boolean if a field has been set.
func (o *V1Resources) HasTemplates() bool {
	if o != nil && o.Templates != nil {
		return true
	}

	return false
}

// SetTemplates gets a reference to the given []string and assigns it to the Templates field.
func (o *V1Resources) SetTemplates(v []string) {
	o.Templates = v
}

func (o V1Resources) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if o.Clusters != nil {
		toSerialize["clusters"] = o.Clusters
	}
	if o.Namespaces != nil {
		toSerialize["namespaces"] = o.Namespaces
	}
	if o.Credentials != nil {
		toSerialize["credentials"] = o.Credentials
	}
	if o.BackupLocations != nil {
		toSerialize["backupLocations"] = o.BackupLocations
	}
	if o.Templates != nil {
		toSerialize["templates"] = o.Templates
	}
	return json.Marshal(toSerialize)
}

type NullableV1Resources struct {
	value *V1Resources
	isSet bool
}

func (v NullableV1Resources) Get() *V1Resources {
	return v.value
}

func (v *NullableV1Resources) Set(val *V1Resources) {
	v.value = val
	v.isSet = true
}

func (v NullableV1Resources) IsSet() bool {
	return v.isSet
}

func (v *NullableV1Resources) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableV1Resources(val *V1Resources) *NullableV1Resources {
	return &NullableV1Resources{value: val, isSet: true}
}

func (v NullableV1Resources) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableV1Resources) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

