/*
public/portworx/pds/backup/apiv1/backup.proto

No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)

API version: version not set
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package backup

import (
	"encoding/json"
)

// checks if the V1DeploymentMetaData type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &V1DeploymentMetaData{}

// V1DeploymentMetaData Deployment Meta Data contains the details of the deployment associated with the backup.
type V1DeploymentMetaData struct {
	// Name of the deployment.
	Name *string `json:"name,omitempty"`
	// Custom Resource Name is the kubernetes resource name for the deployment meta data.
	CustomResourceName *string `json:"customResourceName,omitempty"`
	// Deployment Target Name associated with the backup.
	DeploymentTargetName *string `json:"deploymentTargetName,omitempty"`
	// Namespace name to which the backup is associated.
	NamespaceName *string `json:"namespaceName,omitempty"`
	// Flag to check whether Transport Layer Security support is enabled or not.
	TlsEnabled *bool `json:"tlsEnabled,omitempty"`
	// Name of the data service of deployment.
	DataServiceName *string `json:"dataServiceName,omitempty"`
	// Name of the version of data service.
	DataServiceVersion *string `json:"dataServiceVersion,omitempty"`
	// build tag of the image for the data service version.
	ImageBuild *string `json:"imageBuild,omitempty"`
}

// NewV1DeploymentMetaData instantiates a new V1DeploymentMetaData object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewV1DeploymentMetaData() *V1DeploymentMetaData {
	this := V1DeploymentMetaData{}
	return &this
}

// NewV1DeploymentMetaDataWithDefaults instantiates a new V1DeploymentMetaData object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewV1DeploymentMetaDataWithDefaults() *V1DeploymentMetaData {
	this := V1DeploymentMetaData{}
	return &this
}

// GetName returns the Name field value if set, zero value otherwise.
func (o *V1DeploymentMetaData) GetName() string {
	if o == nil || IsNil(o.Name) {
		var ret string
		return ret
	}
	return *o.Name
}

// GetNameOk returns a tuple with the Name field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1DeploymentMetaData) GetNameOk() (*string, bool) {
	if o == nil || IsNil(o.Name) {
		return nil, false
	}
	return o.Name, true
}

// HasName returns a boolean if a field has been set.
func (o *V1DeploymentMetaData) HasName() bool {
	if o != nil && !IsNil(o.Name) {
		return true
	}

	return false
}

// SetName gets a reference to the given string and assigns it to the Name field.
func (o *V1DeploymentMetaData) SetName(v string) {
	o.Name = &v
}

// GetCustomResourceName returns the CustomResourceName field value if set, zero value otherwise.
func (o *V1DeploymentMetaData) GetCustomResourceName() string {
	if o == nil || IsNil(o.CustomResourceName) {
		var ret string
		return ret
	}
	return *o.CustomResourceName
}

// GetCustomResourceNameOk returns a tuple with the CustomResourceName field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1DeploymentMetaData) GetCustomResourceNameOk() (*string, bool) {
	if o == nil || IsNil(o.CustomResourceName) {
		return nil, false
	}
	return o.CustomResourceName, true
}

// HasCustomResourceName returns a boolean if a field has been set.
func (o *V1DeploymentMetaData) HasCustomResourceName() bool {
	if o != nil && !IsNil(o.CustomResourceName) {
		return true
	}

	return false
}

// SetCustomResourceName gets a reference to the given string and assigns it to the CustomResourceName field.
func (o *V1DeploymentMetaData) SetCustomResourceName(v string) {
	o.CustomResourceName = &v
}

// GetDeploymentTargetName returns the DeploymentTargetName field value if set, zero value otherwise.
func (o *V1DeploymentMetaData) GetDeploymentTargetName() string {
	if o == nil || IsNil(o.DeploymentTargetName) {
		var ret string
		return ret
	}
	return *o.DeploymentTargetName
}

// GetDeploymentTargetNameOk returns a tuple with the DeploymentTargetName field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1DeploymentMetaData) GetDeploymentTargetNameOk() (*string, bool) {
	if o == nil || IsNil(o.DeploymentTargetName) {
		return nil, false
	}
	return o.DeploymentTargetName, true
}

// HasDeploymentTargetName returns a boolean if a field has been set.
func (o *V1DeploymentMetaData) HasDeploymentTargetName() bool {
	if o != nil && !IsNil(o.DeploymentTargetName) {
		return true
	}

	return false
}

// SetDeploymentTargetName gets a reference to the given string and assigns it to the DeploymentTargetName field.
func (o *V1DeploymentMetaData) SetDeploymentTargetName(v string) {
	o.DeploymentTargetName = &v
}

// GetNamespaceName returns the NamespaceName field value if set, zero value otherwise.
func (o *V1DeploymentMetaData) GetNamespaceName() string {
	if o == nil || IsNil(o.NamespaceName) {
		var ret string
		return ret
	}
	return *o.NamespaceName
}

// GetNamespaceNameOk returns a tuple with the NamespaceName field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1DeploymentMetaData) GetNamespaceNameOk() (*string, bool) {
	if o == nil || IsNil(o.NamespaceName) {
		return nil, false
	}
	return o.NamespaceName, true
}

// HasNamespaceName returns a boolean if a field has been set.
func (o *V1DeploymentMetaData) HasNamespaceName() bool {
	if o != nil && !IsNil(o.NamespaceName) {
		return true
	}

	return false
}

// SetNamespaceName gets a reference to the given string and assigns it to the NamespaceName field.
func (o *V1DeploymentMetaData) SetNamespaceName(v string) {
	o.NamespaceName = &v
}

// GetTlsEnabled returns the TlsEnabled field value if set, zero value otherwise.
func (o *V1DeploymentMetaData) GetTlsEnabled() bool {
	if o == nil || IsNil(o.TlsEnabled) {
		var ret bool
		return ret
	}
	return *o.TlsEnabled
}

// GetTlsEnabledOk returns a tuple with the TlsEnabled field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1DeploymentMetaData) GetTlsEnabledOk() (*bool, bool) {
	if o == nil || IsNil(o.TlsEnabled) {
		return nil, false
	}
	return o.TlsEnabled, true
}

// HasTlsEnabled returns a boolean if a field has been set.
func (o *V1DeploymentMetaData) HasTlsEnabled() bool {
	if o != nil && !IsNil(o.TlsEnabled) {
		return true
	}

	return false
}

// SetTlsEnabled gets a reference to the given bool and assigns it to the TlsEnabled field.
func (o *V1DeploymentMetaData) SetTlsEnabled(v bool) {
	o.TlsEnabled = &v
}

// GetDataServiceName returns the DataServiceName field value if set, zero value otherwise.
func (o *V1DeploymentMetaData) GetDataServiceName() string {
	if o == nil || IsNil(o.DataServiceName) {
		var ret string
		return ret
	}
	return *o.DataServiceName
}

// GetDataServiceNameOk returns a tuple with the DataServiceName field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1DeploymentMetaData) GetDataServiceNameOk() (*string, bool) {
	if o == nil || IsNil(o.DataServiceName) {
		return nil, false
	}
	return o.DataServiceName, true
}

// HasDataServiceName returns a boolean if a field has been set.
func (o *V1DeploymentMetaData) HasDataServiceName() bool {
	if o != nil && !IsNil(o.DataServiceName) {
		return true
	}

	return false
}

// SetDataServiceName gets a reference to the given string and assigns it to the DataServiceName field.
func (o *V1DeploymentMetaData) SetDataServiceName(v string) {
	o.DataServiceName = &v
}

// GetDataServiceVersion returns the DataServiceVersion field value if set, zero value otherwise.
func (o *V1DeploymentMetaData) GetDataServiceVersion() string {
	if o == nil || IsNil(o.DataServiceVersion) {
		var ret string
		return ret
	}
	return *o.DataServiceVersion
}

// GetDataServiceVersionOk returns a tuple with the DataServiceVersion field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1DeploymentMetaData) GetDataServiceVersionOk() (*string, bool) {
	if o == nil || IsNil(o.DataServiceVersion) {
		return nil, false
	}
	return o.DataServiceVersion, true
}

// HasDataServiceVersion returns a boolean if a field has been set.
func (o *V1DeploymentMetaData) HasDataServiceVersion() bool {
	if o != nil && !IsNil(o.DataServiceVersion) {
		return true
	}

	return false
}

// SetDataServiceVersion gets a reference to the given string and assigns it to the DataServiceVersion field.
func (o *V1DeploymentMetaData) SetDataServiceVersion(v string) {
	o.DataServiceVersion = &v
}

// GetImageBuild returns the ImageBuild field value if set, zero value otherwise.
func (o *V1DeploymentMetaData) GetImageBuild() string {
	if o == nil || IsNil(o.ImageBuild) {
		var ret string
		return ret
	}
	return *o.ImageBuild
}

// GetImageBuildOk returns a tuple with the ImageBuild field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1DeploymentMetaData) GetImageBuildOk() (*string, bool) {
	if o == nil || IsNil(o.ImageBuild) {
		return nil, false
	}
	return o.ImageBuild, true
}

// HasImageBuild returns a boolean if a field has been set.
func (o *V1DeploymentMetaData) HasImageBuild() bool {
	if o != nil && !IsNil(o.ImageBuild) {
		return true
	}

	return false
}

// SetImageBuild gets a reference to the given string and assigns it to the ImageBuild field.
func (o *V1DeploymentMetaData) SetImageBuild(v string) {
	o.ImageBuild = &v
}

func (o V1DeploymentMetaData) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o V1DeploymentMetaData) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.Name) {
		toSerialize["name"] = o.Name
	}
	if !IsNil(o.CustomResourceName) {
		toSerialize["customResourceName"] = o.CustomResourceName
	}
	if !IsNil(o.DeploymentTargetName) {
		toSerialize["deploymentTargetName"] = o.DeploymentTargetName
	}
	if !IsNil(o.NamespaceName) {
		toSerialize["namespaceName"] = o.NamespaceName
	}
	if !IsNil(o.TlsEnabled) {
		toSerialize["tlsEnabled"] = o.TlsEnabled
	}
	if !IsNil(o.DataServiceName) {
		toSerialize["dataServiceName"] = o.DataServiceName
	}
	if !IsNil(o.DataServiceVersion) {
		toSerialize["dataServiceVersion"] = o.DataServiceVersion
	}
	if !IsNil(o.ImageBuild) {
		toSerialize["imageBuild"] = o.ImageBuild
	}
	return toSerialize, nil
}

type NullableV1DeploymentMetaData struct {
	value *V1DeploymentMetaData
	isSet bool
}

func (v NullableV1DeploymentMetaData) Get() *V1DeploymentMetaData {
	return v.value
}

func (v *NullableV1DeploymentMetaData) Set(val *V1DeploymentMetaData) {
	v.value = val
	v.isSet = true
}

func (v NullableV1DeploymentMetaData) IsSet() bool {
	return v.isSet
}

func (v *NullableV1DeploymentMetaData) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableV1DeploymentMetaData(val *V1DeploymentMetaData) *NullableV1DeploymentMetaData {
	return &NullableV1DeploymentMetaData{value: val, isSet: true}
}

func (v NullableV1DeploymentMetaData) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableV1DeploymentMetaData) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

