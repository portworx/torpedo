/*
public/portworx/pds/tasks/apiv1/tasks.proto

No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)

API version: version not set
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package pdsclient

import (
	"encoding/json"
	"time"
)

// checks if the Deploymenteventsv1Status type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &Deploymenteventsv1Status{}

// Deploymenteventsv1Status Status of the event.
type Deploymenteventsv1Status struct {
	// Action involved in the event.
	Action *string `json:"action,omitempty"`
	// No. of times the event has been generated.
	Count *string `json:"count,omitempty"`
	// Message related to the event.
	Message *string `json:"message,omitempty"`
	// Reason for the event.
	Reason *string `json:"reason,omitempty"`
	// Resource Kind.
	ResourceKind *string `json:"resourceKind,omitempty"`
	// Resource Name.
	ResourceName *string `json:"resourceName,omitempty"`
	// Timestamp of the event.
	TimestampTime *time.Time `json:"timestampTime,omitempty"`
	Type *V1EventType `json:"type,omitempty"`
}

// NewDeploymenteventsv1Status instantiates a new Deploymenteventsv1Status object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewDeploymenteventsv1Status() *Deploymenteventsv1Status {
	this := Deploymenteventsv1Status{}
	var type_ V1EventType = V1EVENTTYPE_EVENT_TYPE_UNSPECIFIED
	this.Type = &type_
	return &this
}

// NewDeploymenteventsv1StatusWithDefaults instantiates a new Deploymenteventsv1Status object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewDeploymenteventsv1StatusWithDefaults() *Deploymenteventsv1Status {
	this := Deploymenteventsv1Status{}
	var type_ V1EventType = V1EVENTTYPE_EVENT_TYPE_UNSPECIFIED
	this.Type = &type_
	return &this
}

// GetAction returns the Action field value if set, zero value otherwise.
func (o *Deploymenteventsv1Status) GetAction() string {
	if o == nil || IsNil(o.Action) {
		var ret string
		return ret
	}
	return *o.Action
}

// GetActionOk returns a tuple with the Action field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Deploymenteventsv1Status) GetActionOk() (*string, bool) {
	if o == nil || IsNil(o.Action) {
		return nil, false
	}
	return o.Action, true
}

// HasAction returns a boolean if a field has been set.
func (o *Deploymenteventsv1Status) HasAction() bool {
	if o != nil && !IsNil(o.Action) {
		return true
	}

	return false
}

// SetAction gets a reference to the given string and assigns it to the Action field.
func (o *Deploymenteventsv1Status) SetAction(v string) {
	o.Action = &v
}

// GetCount returns the Count field value if set, zero value otherwise.
func (o *Deploymenteventsv1Status) GetCount() string {
	if o == nil || IsNil(o.Count) {
		var ret string
		return ret
	}
	return *o.Count
}

// GetCountOk returns a tuple with the Count field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Deploymenteventsv1Status) GetCountOk() (*string, bool) {
	if o == nil || IsNil(o.Count) {
		return nil, false
	}
	return o.Count, true
}

// HasCount returns a boolean if a field has been set.
func (o *Deploymenteventsv1Status) HasCount() bool {
	if o != nil && !IsNil(o.Count) {
		return true
	}

	return false
}

// SetCount gets a reference to the given string and assigns it to the Count field.
func (o *Deploymenteventsv1Status) SetCount(v string) {
	o.Count = &v
}

// GetMessage returns the Message field value if set, zero value otherwise.
func (o *Deploymenteventsv1Status) GetMessage() string {
	if o == nil || IsNil(o.Message) {
		var ret string
		return ret
	}
	return *o.Message
}

// GetMessageOk returns a tuple with the Message field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Deploymenteventsv1Status) GetMessageOk() (*string, bool) {
	if o == nil || IsNil(o.Message) {
		return nil, false
	}
	return o.Message, true
}

// HasMessage returns a boolean if a field has been set.
func (o *Deploymenteventsv1Status) HasMessage() bool {
	if o != nil && !IsNil(o.Message) {
		return true
	}

	return false
}

// SetMessage gets a reference to the given string and assigns it to the Message field.
func (o *Deploymenteventsv1Status) SetMessage(v string) {
	o.Message = &v
}

// GetReason returns the Reason field value if set, zero value otherwise.
func (o *Deploymenteventsv1Status) GetReason() string {
	if o == nil || IsNil(o.Reason) {
		var ret string
		return ret
	}
	return *o.Reason
}

// GetReasonOk returns a tuple with the Reason field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Deploymenteventsv1Status) GetReasonOk() (*string, bool) {
	if o == nil || IsNil(o.Reason) {
		return nil, false
	}
	return o.Reason, true
}

// HasReason returns a boolean if a field has been set.
func (o *Deploymenteventsv1Status) HasReason() bool {
	if o != nil && !IsNil(o.Reason) {
		return true
	}

	return false
}

// SetReason gets a reference to the given string and assigns it to the Reason field.
func (o *Deploymenteventsv1Status) SetReason(v string) {
	o.Reason = &v
}

// GetResourceKind returns the ResourceKind field value if set, zero value otherwise.
func (o *Deploymenteventsv1Status) GetResourceKind() string {
	if o == nil || IsNil(o.ResourceKind) {
		var ret string
		return ret
	}
	return *o.ResourceKind
}

// GetResourceKindOk returns a tuple with the ResourceKind field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Deploymenteventsv1Status) GetResourceKindOk() (*string, bool) {
	if o == nil || IsNil(o.ResourceKind) {
		return nil, false
	}
	return o.ResourceKind, true
}

// HasResourceKind returns a boolean if a field has been set.
func (o *Deploymenteventsv1Status) HasResourceKind() bool {
	if o != nil && !IsNil(o.ResourceKind) {
		return true
	}

	return false
}

// SetResourceKind gets a reference to the given string and assigns it to the ResourceKind field.
func (o *Deploymenteventsv1Status) SetResourceKind(v string) {
	o.ResourceKind = &v
}

// GetResourceName returns the ResourceName field value if set, zero value otherwise.
func (o *Deploymenteventsv1Status) GetResourceName() string {
	if o == nil || IsNil(o.ResourceName) {
		var ret string
		return ret
	}
	return *o.ResourceName
}

// GetResourceNameOk returns a tuple with the ResourceName field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Deploymenteventsv1Status) GetResourceNameOk() (*string, bool) {
	if o == nil || IsNil(o.ResourceName) {
		return nil, false
	}
	return o.ResourceName, true
}

// HasResourceName returns a boolean if a field has been set.
func (o *Deploymenteventsv1Status) HasResourceName() bool {
	if o != nil && !IsNil(o.ResourceName) {
		return true
	}

	return false
}

// SetResourceName gets a reference to the given string and assigns it to the ResourceName field.
func (o *Deploymenteventsv1Status) SetResourceName(v string) {
	o.ResourceName = &v
}

// GetTimestampTime returns the TimestampTime field value if set, zero value otherwise.
func (o *Deploymenteventsv1Status) GetTimestampTime() time.Time {
	if o == nil || IsNil(o.TimestampTime) {
		var ret time.Time
		return ret
	}
	return *o.TimestampTime
}

// GetTimestampTimeOk returns a tuple with the TimestampTime field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Deploymenteventsv1Status) GetTimestampTimeOk() (*time.Time, bool) {
	if o == nil || IsNil(o.TimestampTime) {
		return nil, false
	}
	return o.TimestampTime, true
}

// HasTimestampTime returns a boolean if a field has been set.
func (o *Deploymenteventsv1Status) HasTimestampTime() bool {
	if o != nil && !IsNil(o.TimestampTime) {
		return true
	}

	return false
}

// SetTimestampTime gets a reference to the given time.Time and assigns it to the TimestampTime field.
func (o *Deploymenteventsv1Status) SetTimestampTime(v time.Time) {
	o.TimestampTime = &v
}

// GetType returns the Type field value if set, zero value otherwise.
func (o *Deploymenteventsv1Status) GetType() V1EventType {
	if o == nil || IsNil(o.Type) {
		var ret V1EventType
		return ret
	}
	return *o.Type
}

// GetTypeOk returns a tuple with the Type field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Deploymenteventsv1Status) GetTypeOk() (*V1EventType, bool) {
	if o == nil || IsNil(o.Type) {
		return nil, false
	}
	return o.Type, true
}

// HasType returns a boolean if a field has been set.
func (o *Deploymenteventsv1Status) HasType() bool {
	if o != nil && !IsNil(o.Type) {
		return true
	}

	return false
}

// SetType gets a reference to the given V1EventType and assigns it to the Type field.
func (o *Deploymenteventsv1Status) SetType(v V1EventType) {
	o.Type = &v
}

func (o Deploymenteventsv1Status) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o Deploymenteventsv1Status) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.Action) {
		toSerialize["action"] = o.Action
	}
	if !IsNil(o.Count) {
		toSerialize["count"] = o.Count
	}
	if !IsNil(o.Message) {
		toSerialize["message"] = o.Message
	}
	if !IsNil(o.Reason) {
		toSerialize["reason"] = o.Reason
	}
	if !IsNil(o.ResourceKind) {
		toSerialize["resourceKind"] = o.ResourceKind
	}
	if !IsNil(o.ResourceName) {
		toSerialize["resourceName"] = o.ResourceName
	}
	if !IsNil(o.TimestampTime) {
		toSerialize["timestampTime"] = o.TimestampTime
	}
	if !IsNil(o.Type) {
		toSerialize["type"] = o.Type
	}
	return toSerialize, nil
}

type NullableDeploymenteventsv1Status struct {
	value *Deploymenteventsv1Status
	isSet bool
}

func (v NullableDeploymenteventsv1Status) Get() *Deploymenteventsv1Status {
	return v.value
}

func (v *NullableDeploymenteventsv1Status) Set(val *Deploymenteventsv1Status) {
	v.value = val
	v.isSet = true
}

func (v NullableDeploymenteventsv1Status) IsSet() bool {
	return v.isSet
}

func (v *NullableDeploymenteventsv1Status) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableDeploymenteventsv1Status(val *Deploymenteventsv1Status) *NullableDeploymenteventsv1Status {
	return &NullableDeploymenteventsv1Status{value: val, isSet: true}
}

func (v NullableDeploymenteventsv1Status) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableDeploymenteventsv1Status) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

