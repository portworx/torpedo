/*
public/portworx/pds/catalog/dataservices/apiv1/dataservices.proto

No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)

API version: version not set
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package catalog

import (
	"encoding/json"
)

// checks if the V1Info2 type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &V1Info2{}

// V1Info2 Information related to the data service image.
type V1Info2 struct {
	References *V1References `json:"references,omitempty"`
	// Image registry where the image is stored.
	Registry *string `json:"registry,omitempty"`
	// Image registry namespace where the image is stored.
	Namespace *string `json:"namespace,omitempty"`
	// Tag associated with the image.
	Tag *string `json:"tag,omitempty"`
	// Build version of the image.
	Build *string `json:"build,omitempty"`
	// Flag indicating if TLS is supported for a data service using this image.
	TlsSupport *bool `json:"tlsSupport,omitempty"`
	// Capabilities associated with this image.
	Capabilities *map[string]string `json:"capabilities,omitempty"`
	// Additional images associated with this data service image.
	AdditionalImages *map[string]string `json:"additionalImages,omitempty"`
}

// NewV1Info2 instantiates a new V1Info2 object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewV1Info2() *V1Info2 {
	this := V1Info2{}
	return &this
}

// NewV1Info2WithDefaults instantiates a new V1Info2 object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewV1Info2WithDefaults() *V1Info2 {
	this := V1Info2{}
	return &this
}

// GetReferences returns the References field value if set, zero value otherwise.
func (o *V1Info2) GetReferences() V1References {
	if o == nil || IsNil(o.References) {
		var ret V1References
		return ret
	}
	return *o.References
}

// GetReferencesOk returns a tuple with the References field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1Info2) GetReferencesOk() (*V1References, bool) {
	if o == nil || IsNil(o.References) {
		return nil, false
	}
	return o.References, true
}

// HasReferences returns a boolean if a field has been set.
func (o *V1Info2) HasReferences() bool {
	if o != nil && !IsNil(o.References) {
		return true
	}

	return false
}

// SetReferences gets a reference to the given V1References and assigns it to the References field.
func (o *V1Info2) SetReferences(v V1References) {
	o.References = &v
}

// GetRegistry returns the Registry field value if set, zero value otherwise.
func (o *V1Info2) GetRegistry() string {
	if o == nil || IsNil(o.Registry) {
		var ret string
		return ret
	}
	return *o.Registry
}

// GetRegistryOk returns a tuple with the Registry field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1Info2) GetRegistryOk() (*string, bool) {
	if o == nil || IsNil(o.Registry) {
		return nil, false
	}
	return o.Registry, true
}

// HasRegistry returns a boolean if a field has been set.
func (o *V1Info2) HasRegistry() bool {
	if o != nil && !IsNil(o.Registry) {
		return true
	}

	return false
}

// SetRegistry gets a reference to the given string and assigns it to the Registry field.
func (o *V1Info2) SetRegistry(v string) {
	o.Registry = &v
}

// GetNamespace returns the Namespace field value if set, zero value otherwise.
func (o *V1Info2) GetNamespace() string {
	if o == nil || IsNil(o.Namespace) {
		var ret string
		return ret
	}
	return *o.Namespace
}

// GetNamespaceOk returns a tuple with the Namespace field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1Info2) GetNamespaceOk() (*string, bool) {
	if o == nil || IsNil(o.Namespace) {
		return nil, false
	}
	return o.Namespace, true
}

// HasNamespace returns a boolean if a field has been set.
func (o *V1Info2) HasNamespace() bool {
	if o != nil && !IsNil(o.Namespace) {
		return true
	}

	return false
}

// SetNamespace gets a reference to the given string and assigns it to the Namespace field.
func (o *V1Info2) SetNamespace(v string) {
	o.Namespace = &v
}

// GetTag returns the Tag field value if set, zero value otherwise.
func (o *V1Info2) GetTag() string {
	if o == nil || IsNil(o.Tag) {
		var ret string
		return ret
	}
	return *o.Tag
}

// GetTagOk returns a tuple with the Tag field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1Info2) GetTagOk() (*string, bool) {
	if o == nil || IsNil(o.Tag) {
		return nil, false
	}
	return o.Tag, true
}

// HasTag returns a boolean if a field has been set.
func (o *V1Info2) HasTag() bool {
	if o != nil && !IsNil(o.Tag) {
		return true
	}

	return false
}

// SetTag gets a reference to the given string and assigns it to the Tag field.
func (o *V1Info2) SetTag(v string) {
	o.Tag = &v
}

// GetBuild returns the Build field value if set, zero value otherwise.
func (o *V1Info2) GetBuild() string {
	if o == nil || IsNil(o.Build) {
		var ret string
		return ret
	}
	return *o.Build
}

// GetBuildOk returns a tuple with the Build field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1Info2) GetBuildOk() (*string, bool) {
	if o == nil || IsNil(o.Build) {
		return nil, false
	}
	return o.Build, true
}

// HasBuild returns a boolean if a field has been set.
func (o *V1Info2) HasBuild() bool {
	if o != nil && !IsNil(o.Build) {
		return true
	}

	return false
}

// SetBuild gets a reference to the given string and assigns it to the Build field.
func (o *V1Info2) SetBuild(v string) {
	o.Build = &v
}

// GetTlsSupport returns the TlsSupport field value if set, zero value otherwise.
func (o *V1Info2) GetTlsSupport() bool {
	if o == nil || IsNil(o.TlsSupport) {
		var ret bool
		return ret
	}
	return *o.TlsSupport
}

// GetTlsSupportOk returns a tuple with the TlsSupport field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1Info2) GetTlsSupportOk() (*bool, bool) {
	if o == nil || IsNil(o.TlsSupport) {
		return nil, false
	}
	return o.TlsSupport, true
}

// HasTlsSupport returns a boolean if a field has been set.
func (o *V1Info2) HasTlsSupport() bool {
	if o != nil && !IsNil(o.TlsSupport) {
		return true
	}

	return false
}

// SetTlsSupport gets a reference to the given bool and assigns it to the TlsSupport field.
func (o *V1Info2) SetTlsSupport(v bool) {
	o.TlsSupport = &v
}

// GetCapabilities returns the Capabilities field value if set, zero value otherwise.
func (o *V1Info2) GetCapabilities() map[string]string {
	if o == nil || IsNil(o.Capabilities) {
		var ret map[string]string
		return ret
	}
	return *o.Capabilities
}

// GetCapabilitiesOk returns a tuple with the Capabilities field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1Info2) GetCapabilitiesOk() (*map[string]string, bool) {
	if o == nil || IsNil(o.Capabilities) {
		return nil, false
	}
	return o.Capabilities, true
}

// HasCapabilities returns a boolean if a field has been set.
func (o *V1Info2) HasCapabilities() bool {
	if o != nil && !IsNil(o.Capabilities) {
		return true
	}

	return false
}

// SetCapabilities gets a reference to the given map[string]string and assigns it to the Capabilities field.
func (o *V1Info2) SetCapabilities(v map[string]string) {
	o.Capabilities = &v
}

// GetAdditionalImages returns the AdditionalImages field value if set, zero value otherwise.
func (o *V1Info2) GetAdditionalImages() map[string]string {
	if o == nil || IsNil(o.AdditionalImages) {
		var ret map[string]string
		return ret
	}
	return *o.AdditionalImages
}

// GetAdditionalImagesOk returns a tuple with the AdditionalImages field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1Info2) GetAdditionalImagesOk() (*map[string]string, bool) {
	if o == nil || IsNil(o.AdditionalImages) {
		return nil, false
	}
	return o.AdditionalImages, true
}

// HasAdditionalImages returns a boolean if a field has been set.
func (o *V1Info2) HasAdditionalImages() bool {
	if o != nil && !IsNil(o.AdditionalImages) {
		return true
	}

	return false
}

// SetAdditionalImages gets a reference to the given map[string]string and assigns it to the AdditionalImages field.
func (o *V1Info2) SetAdditionalImages(v map[string]string) {
	o.AdditionalImages = &v
}

func (o V1Info2) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o V1Info2) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.References) {
		toSerialize["references"] = o.References
	}
	if !IsNil(o.Registry) {
		toSerialize["registry"] = o.Registry
	}
	if !IsNil(o.Namespace) {
		toSerialize["namespace"] = o.Namespace
	}
	if !IsNil(o.Tag) {
		toSerialize["tag"] = o.Tag
	}
	if !IsNil(o.Build) {
		toSerialize["build"] = o.Build
	}
	if !IsNil(o.TlsSupport) {
		toSerialize["tlsSupport"] = o.TlsSupport
	}
	if !IsNil(o.Capabilities) {
		toSerialize["capabilities"] = o.Capabilities
	}
	if !IsNil(o.AdditionalImages) {
		toSerialize["additionalImages"] = o.AdditionalImages
	}
	return toSerialize, nil
}

type NullableV1Info2 struct {
	value *V1Info2
	isSet bool
}

func (v NullableV1Info2) Get() *V1Info2 {
	return v.value
}

func (v *NullableV1Info2) Set(val *V1Info2) {
	v.value = val
	v.isSet = true
}

func (v NullableV1Info2) IsSet() bool {
	return v.isSet
}

func (v *NullableV1Info2) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableV1Info2(val *V1Info2) *NullableV1Info2 {
	return &NullableV1Info2{value: val, isSet: true}
}

func (v NullableV1Info2) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableV1Info2) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

