/*
public/portworx/pds/backupconfig/apiv1/backupconfig.proto

No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)

API version: version not set
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package openapi

import (
	"encoding/json"
	"time"
)

// MetadataOfTheAccount Metadata of the account.
type MetadataOfTheAccount struct {
	// Name of the resource.
	Name *string `json:"name,omitempty"`
	// Description of the resource.
	Description *string `json:"description,omitempty"`
	// A string that identifies the version of this object that can be used by clients to determine when objects have changed. This value must be passed unmodified back to the server by the client.
	ResourceVersion *string `json:"resourceVersion,omitempty"`
	// Creation time of the object.
	CreateTime *time.Time `json:"createTime,omitempty"`
	// Update time of the object.
	UpdateTime *time.Time `json:"updateTime,omitempty"`
	// Labels to apply to the object.
	Labels *map[string]string `json:"labels,omitempty"`
	// Annotations for the object.
	Annotations *map[string]string `json:"annotations,omitempty"`
	ParentReference *V1Reference `json:"parentReference,omitempty"`
}

// NewMetadataOfTheAccount instantiates a new MetadataOfTheAccount object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewMetadataOfTheAccount() *MetadataOfTheAccount {
	this := MetadataOfTheAccount{}
	return &this
}

// NewMetadataOfTheAccountWithDefaults instantiates a new MetadataOfTheAccount object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewMetadataOfTheAccountWithDefaults() *MetadataOfTheAccount {
	this := MetadataOfTheAccount{}
	return &this
}

// GetName returns the Name field value if set, zero value otherwise.
func (o *MetadataOfTheAccount) GetName() string {
	if o == nil || o.Name == nil {
		var ret string
		return ret
	}
	return *o.Name
}

// GetNameOk returns a tuple with the Name field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *MetadataOfTheAccount) GetNameOk() (*string, bool) {
	if o == nil || o.Name == nil {
		return nil, false
	}
	return o.Name, true
}

// HasName returns a boolean if a field has been set.
func (o *MetadataOfTheAccount) HasName() bool {
	if o != nil && o.Name != nil {
		return true
	}

	return false
}

// SetName gets a reference to the given string and assigns it to the Name field.
func (o *MetadataOfTheAccount) SetName(v string) {
	o.Name = &v
}

// GetDescription returns the Description field value if set, zero value otherwise.
func (o *MetadataOfTheAccount) GetDescription() string {
	if o == nil || o.Description == nil {
		var ret string
		return ret
	}
	return *o.Description
}

// GetDescriptionOk returns a tuple with the Description field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *MetadataOfTheAccount) GetDescriptionOk() (*string, bool) {
	if o == nil || o.Description == nil {
		return nil, false
	}
	return o.Description, true
}

// HasDescription returns a boolean if a field has been set.
func (o *MetadataOfTheAccount) HasDescription() bool {
	if o != nil && o.Description != nil {
		return true
	}

	return false
}

// SetDescription gets a reference to the given string and assigns it to the Description field.
func (o *MetadataOfTheAccount) SetDescription(v string) {
	o.Description = &v
}

// GetResourceVersion returns the ResourceVersion field value if set, zero value otherwise.
func (o *MetadataOfTheAccount) GetResourceVersion() string {
	if o == nil || o.ResourceVersion == nil {
		var ret string
		return ret
	}
	return *o.ResourceVersion
}

// GetResourceVersionOk returns a tuple with the ResourceVersion field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *MetadataOfTheAccount) GetResourceVersionOk() (*string, bool) {
	if o == nil || o.ResourceVersion == nil {
		return nil, false
	}
	return o.ResourceVersion, true
}

// HasResourceVersion returns a boolean if a field has been set.
func (o *MetadataOfTheAccount) HasResourceVersion() bool {
	if o != nil && o.ResourceVersion != nil {
		return true
	}

	return false
}

// SetResourceVersion gets a reference to the given string and assigns it to the ResourceVersion field.
func (o *MetadataOfTheAccount) SetResourceVersion(v string) {
	o.ResourceVersion = &v
}

// GetCreateTime returns the CreateTime field value if set, zero value otherwise.
func (o *MetadataOfTheAccount) GetCreateTime() time.Time {
	if o == nil || o.CreateTime == nil {
		var ret time.Time
		return ret
	}
	return *o.CreateTime
}

// GetCreateTimeOk returns a tuple with the CreateTime field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *MetadataOfTheAccount) GetCreateTimeOk() (*time.Time, bool) {
	if o == nil || o.CreateTime == nil {
		return nil, false
	}
	return o.CreateTime, true
}

// HasCreateTime returns a boolean if a field has been set.
func (o *MetadataOfTheAccount) HasCreateTime() bool {
	if o != nil && o.CreateTime != nil {
		return true
	}

	return false
}

// SetCreateTime gets a reference to the given time.Time and assigns it to the CreateTime field.
func (o *MetadataOfTheAccount) SetCreateTime(v time.Time) {
	o.CreateTime = &v
}

// GetUpdateTime returns the UpdateTime field value if set, zero value otherwise.
func (o *MetadataOfTheAccount) GetUpdateTime() time.Time {
	if o == nil || o.UpdateTime == nil {
		var ret time.Time
		return ret
	}
	return *o.UpdateTime
}

// GetUpdateTimeOk returns a tuple with the UpdateTime field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *MetadataOfTheAccount) GetUpdateTimeOk() (*time.Time, bool) {
	if o == nil || o.UpdateTime == nil {
		return nil, false
	}
	return o.UpdateTime, true
}

// HasUpdateTime returns a boolean if a field has been set.
func (o *MetadataOfTheAccount) HasUpdateTime() bool {
	if o != nil && o.UpdateTime != nil {
		return true
	}

	return false
}

// SetUpdateTime gets a reference to the given time.Time and assigns it to the UpdateTime field.
func (o *MetadataOfTheAccount) SetUpdateTime(v time.Time) {
	o.UpdateTime = &v
}

// GetLabels returns the Labels field value if set, zero value otherwise.
func (o *MetadataOfTheAccount) GetLabels() map[string]string {
	if o == nil || o.Labels == nil {
		var ret map[string]string
		return ret
	}
	return *o.Labels
}

// GetLabelsOk returns a tuple with the Labels field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *MetadataOfTheAccount) GetLabelsOk() (*map[string]string, bool) {
	if o == nil || o.Labels == nil {
		return nil, false
	}
	return o.Labels, true
}

// HasLabels returns a boolean if a field has been set.
func (o *MetadataOfTheAccount) HasLabels() bool {
	if o != nil && o.Labels != nil {
		return true
	}

	return false
}

// SetLabels gets a reference to the given map[string]string and assigns it to the Labels field.
func (o *MetadataOfTheAccount) SetLabels(v map[string]string) {
	o.Labels = &v
}

// GetAnnotations returns the Annotations field value if set, zero value otherwise.
func (o *MetadataOfTheAccount) GetAnnotations() map[string]string {
	if o == nil || o.Annotations == nil {
		var ret map[string]string
		return ret
	}
	return *o.Annotations
}

// GetAnnotationsOk returns a tuple with the Annotations field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *MetadataOfTheAccount) GetAnnotationsOk() (*map[string]string, bool) {
	if o == nil || o.Annotations == nil {
		return nil, false
	}
	return o.Annotations, true
}

// HasAnnotations returns a boolean if a field has been set.
func (o *MetadataOfTheAccount) HasAnnotations() bool {
	if o != nil && o.Annotations != nil {
		return true
	}

	return false
}

// SetAnnotations gets a reference to the given map[string]string and assigns it to the Annotations field.
func (o *MetadataOfTheAccount) SetAnnotations(v map[string]string) {
	o.Annotations = &v
}

// GetParentReference returns the ParentReference field value if set, zero value otherwise.
func (o *MetadataOfTheAccount) GetParentReference() V1Reference {
	if o == nil || o.ParentReference == nil {
		var ret V1Reference
		return ret
	}
	return *o.ParentReference
}

// GetParentReferenceOk returns a tuple with the ParentReference field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *MetadataOfTheAccount) GetParentReferenceOk() (*V1Reference, bool) {
	if o == nil || o.ParentReference == nil {
		return nil, false
	}
	return o.ParentReference, true
}

// HasParentReference returns a boolean if a field has been set.
func (o *MetadataOfTheAccount) HasParentReference() bool {
	if o != nil && o.ParentReference != nil {
		return true
	}

	return false
}

// SetParentReference gets a reference to the given V1Reference and assigns it to the ParentReference field.
func (o *MetadataOfTheAccount) SetParentReference(v V1Reference) {
	o.ParentReference = &v
}

func (o MetadataOfTheAccount) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if o.Name != nil {
		toSerialize["name"] = o.Name
	}
	if o.Description != nil {
		toSerialize["description"] = o.Description
	}
	if o.ResourceVersion != nil {
		toSerialize["resourceVersion"] = o.ResourceVersion
	}
	if o.CreateTime != nil {
		toSerialize["createTime"] = o.CreateTime
	}
	if o.UpdateTime != nil {
		toSerialize["updateTime"] = o.UpdateTime
	}
	if o.Labels != nil {
		toSerialize["labels"] = o.Labels
	}
	if o.Annotations != nil {
		toSerialize["annotations"] = o.Annotations
	}
	if o.ParentReference != nil {
		toSerialize["parentReference"] = o.ParentReference
	}
	return json.Marshal(toSerialize)
}

type NullableMetadataOfTheAccount struct {
	value *MetadataOfTheAccount
	isSet bool
}

func (v NullableMetadataOfTheAccount) Get() *MetadataOfTheAccount {
	return v.value
}

func (v *NullableMetadataOfTheAccount) Set(val *MetadataOfTheAccount) {
	v.value = val
	v.isSet = true
}

func (v NullableMetadataOfTheAccount) IsSet() bool {
	return v.isSet
}

func (v *NullableMetadataOfTheAccount) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableMetadataOfTheAccount(val *MetadataOfTheAccount) *NullableMetadataOfTheAccount {
	return &NullableMetadataOfTheAccount{value: val, isSet: true}
}

func (v NullableMetadataOfTheAccount) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableMetadataOfTheAccount) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

