/*
public/portworx/platform/project/apiv1/project.proto

No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)

API version: version not set
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package project

import (
	"encoding/json"
	"time"
)

// checks if the MetadataOfTheProject type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &MetadataOfTheProject{}

// MetadataOfTheProject Metadata of the project.
type MetadataOfTheProject struct {
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
	// Resource names holds the mapping between the resource IDs and its display name which will be consumed by the frontend.
	ResourceNames *map[string]string `json:"resourceNames,omitempty"`
}

// NewMetadataOfTheProject instantiates a new MetadataOfTheProject object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewMetadataOfTheProject() *MetadataOfTheProject {
	this := MetadataOfTheProject{}
	return &this
}

// NewMetadataOfTheProjectWithDefaults instantiates a new MetadataOfTheProject object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewMetadataOfTheProjectWithDefaults() *MetadataOfTheProject {
	this := MetadataOfTheProject{}
	return &this
}

// GetName returns the Name field value if set, zero value otherwise.
func (o *MetadataOfTheProject) GetName() string {
	if o == nil || IsNil(o.Name) {
		var ret string
		return ret
	}
	return *o.Name
}

// GetNameOk returns a tuple with the Name field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *MetadataOfTheProject) GetNameOk() (*string, bool) {
	if o == nil || IsNil(o.Name) {
		return nil, false
	}
	return o.Name, true
}

// HasName returns a boolean if a field has been set.
func (o *MetadataOfTheProject) HasName() bool {
	if o != nil && !IsNil(o.Name) {
		return true
	}

	return false
}

// SetName gets a reference to the given string and assigns it to the Name field.
func (o *MetadataOfTheProject) SetName(v string) {
	o.Name = &v
}

// GetDescription returns the Description field value if set, zero value otherwise.
func (o *MetadataOfTheProject) GetDescription() string {
	if o == nil || IsNil(o.Description) {
		var ret string
		return ret
	}
	return *o.Description
}

// GetDescriptionOk returns a tuple with the Description field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *MetadataOfTheProject) GetDescriptionOk() (*string, bool) {
	if o == nil || IsNil(o.Description) {
		return nil, false
	}
	return o.Description, true
}

// HasDescription returns a boolean if a field has been set.
func (o *MetadataOfTheProject) HasDescription() bool {
	if o != nil && !IsNil(o.Description) {
		return true
	}

	return false
}

// SetDescription gets a reference to the given string and assigns it to the Description field.
func (o *MetadataOfTheProject) SetDescription(v string) {
	o.Description = &v
}

// GetResourceVersion returns the ResourceVersion field value if set, zero value otherwise.
func (o *MetadataOfTheProject) GetResourceVersion() string {
	if o == nil || IsNil(o.ResourceVersion) {
		var ret string
		return ret
	}
	return *o.ResourceVersion
}

// GetResourceVersionOk returns a tuple with the ResourceVersion field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *MetadataOfTheProject) GetResourceVersionOk() (*string, bool) {
	if o == nil || IsNil(o.ResourceVersion) {
		return nil, false
	}
	return o.ResourceVersion, true
}

// HasResourceVersion returns a boolean if a field has been set.
func (o *MetadataOfTheProject) HasResourceVersion() bool {
	if o != nil && !IsNil(o.ResourceVersion) {
		return true
	}

	return false
}

// SetResourceVersion gets a reference to the given string and assigns it to the ResourceVersion field.
func (o *MetadataOfTheProject) SetResourceVersion(v string) {
	o.ResourceVersion = &v
}

// GetCreateTime returns the CreateTime field value if set, zero value otherwise.
func (o *MetadataOfTheProject) GetCreateTime() time.Time {
	if o == nil || IsNil(o.CreateTime) {
		var ret time.Time
		return ret
	}
	return *o.CreateTime
}

// GetCreateTimeOk returns a tuple with the CreateTime field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *MetadataOfTheProject) GetCreateTimeOk() (*time.Time, bool) {
	if o == nil || IsNil(o.CreateTime) {
		return nil, false
	}
	return o.CreateTime, true
}

// HasCreateTime returns a boolean if a field has been set.
func (o *MetadataOfTheProject) HasCreateTime() bool {
	if o != nil && !IsNil(o.CreateTime) {
		return true
	}

	return false
}

// SetCreateTime gets a reference to the given time.Time and assigns it to the CreateTime field.
func (o *MetadataOfTheProject) SetCreateTime(v time.Time) {
	o.CreateTime = &v
}

// GetUpdateTime returns the UpdateTime field value if set, zero value otherwise.
func (o *MetadataOfTheProject) GetUpdateTime() time.Time {
	if o == nil || IsNil(o.UpdateTime) {
		var ret time.Time
		return ret
	}
	return *o.UpdateTime
}

// GetUpdateTimeOk returns a tuple with the UpdateTime field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *MetadataOfTheProject) GetUpdateTimeOk() (*time.Time, bool) {
	if o == nil || IsNil(o.UpdateTime) {
		return nil, false
	}
	return o.UpdateTime, true
}

// HasUpdateTime returns a boolean if a field has been set.
func (o *MetadataOfTheProject) HasUpdateTime() bool {
	if o != nil && !IsNil(o.UpdateTime) {
		return true
	}

	return false
}

// SetUpdateTime gets a reference to the given time.Time and assigns it to the UpdateTime field.
func (o *MetadataOfTheProject) SetUpdateTime(v time.Time) {
	o.UpdateTime = &v
}

// GetLabels returns the Labels field value if set, zero value otherwise.
func (o *MetadataOfTheProject) GetLabels() map[string]string {
	if o == nil || IsNil(o.Labels) {
		var ret map[string]string
		return ret
	}
	return *o.Labels
}

// GetLabelsOk returns a tuple with the Labels field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *MetadataOfTheProject) GetLabelsOk() (*map[string]string, bool) {
	if o == nil || IsNil(o.Labels) {
		return nil, false
	}
	return o.Labels, true
}

// HasLabels returns a boolean if a field has been set.
func (o *MetadataOfTheProject) HasLabels() bool {
	if o != nil && !IsNil(o.Labels) {
		return true
	}

	return false
}

// SetLabels gets a reference to the given map[string]string and assigns it to the Labels field.
func (o *MetadataOfTheProject) SetLabels(v map[string]string) {
	o.Labels = &v
}

// GetAnnotations returns the Annotations field value if set, zero value otherwise.
func (o *MetadataOfTheProject) GetAnnotations() map[string]string {
	if o == nil || IsNil(o.Annotations) {
		var ret map[string]string
		return ret
	}
	return *o.Annotations
}

// GetAnnotationsOk returns a tuple with the Annotations field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *MetadataOfTheProject) GetAnnotationsOk() (*map[string]string, bool) {
	if o == nil || IsNil(o.Annotations) {
		return nil, false
	}
	return o.Annotations, true
}

// HasAnnotations returns a boolean if a field has been set.
func (o *MetadataOfTheProject) HasAnnotations() bool {
	if o != nil && !IsNil(o.Annotations) {
		return true
	}

	return false
}

// SetAnnotations gets a reference to the given map[string]string and assigns it to the Annotations field.
func (o *MetadataOfTheProject) SetAnnotations(v map[string]string) {
	o.Annotations = &v
}

// GetParentReference returns the ParentReference field value if set, zero value otherwise.
func (o *MetadataOfTheProject) GetParentReference() V1Reference {
	if o == nil || IsNil(o.ParentReference) {
		var ret V1Reference
		return ret
	}
	return *o.ParentReference
}

// GetParentReferenceOk returns a tuple with the ParentReference field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *MetadataOfTheProject) GetParentReferenceOk() (*V1Reference, bool) {
	if o == nil || IsNil(o.ParentReference) {
		return nil, false
	}
	return o.ParentReference, true
}

// HasParentReference returns a boolean if a field has been set.
func (o *MetadataOfTheProject) HasParentReference() bool {
	if o != nil && !IsNil(o.ParentReference) {
		return true
	}

	return false
}

// SetParentReference gets a reference to the given V1Reference and assigns it to the ParentReference field.
func (o *MetadataOfTheProject) SetParentReference(v V1Reference) {
	o.ParentReference = &v
}

// GetResourceNames returns the ResourceNames field value if set, zero value otherwise.
func (o *MetadataOfTheProject) GetResourceNames() map[string]string {
	if o == nil || IsNil(o.ResourceNames) {
		var ret map[string]string
		return ret
	}
	return *o.ResourceNames
}

// GetResourceNamesOk returns a tuple with the ResourceNames field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *MetadataOfTheProject) GetResourceNamesOk() (*map[string]string, bool) {
	if o == nil || IsNil(o.ResourceNames) {
		return nil, false
	}
	return o.ResourceNames, true
}

// HasResourceNames returns a boolean if a field has been set.
func (o *MetadataOfTheProject) HasResourceNames() bool {
	if o != nil && !IsNil(o.ResourceNames) {
		return true
	}

	return false
}

// SetResourceNames gets a reference to the given map[string]string and assigns it to the ResourceNames field.
func (o *MetadataOfTheProject) SetResourceNames(v map[string]string) {
	o.ResourceNames = &v
}

func (o MetadataOfTheProject) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o MetadataOfTheProject) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.Name) {
		toSerialize["name"] = o.Name
	}
	if !IsNil(o.Description) {
		toSerialize["description"] = o.Description
	}
	if !IsNil(o.ResourceVersion) {
		toSerialize["resourceVersion"] = o.ResourceVersion
	}
	if !IsNil(o.CreateTime) {
		toSerialize["createTime"] = o.CreateTime
	}
	if !IsNil(o.UpdateTime) {
		toSerialize["updateTime"] = o.UpdateTime
	}
	if !IsNil(o.Labels) {
		toSerialize["labels"] = o.Labels
	}
	if !IsNil(o.Annotations) {
		toSerialize["annotations"] = o.Annotations
	}
	if !IsNil(o.ParentReference) {
		toSerialize["parentReference"] = o.ParentReference
	}
	if !IsNil(o.ResourceNames) {
		toSerialize["resourceNames"] = o.ResourceNames
	}
	return toSerialize, nil
}

type NullableMetadataOfTheProject struct {
	value *MetadataOfTheProject
	isSet bool
}

func (v NullableMetadataOfTheProject) Get() *MetadataOfTheProject {
	return v.value
}

func (v *NullableMetadataOfTheProject) Set(val *MetadataOfTheProject) {
	v.value = val
	v.isSet = true
}

func (v NullableMetadataOfTheProject) IsSet() bool {
	return v.isSet
}

func (v *NullableMetadataOfTheProject) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableMetadataOfTheProject(val *MetadataOfTheProject) *NullableMetadataOfTheProject {
	return &NullableMetadataOfTheProject{value: val, isSet: true}
}

func (v NullableMetadataOfTheProject) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableMetadataOfTheProject) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

