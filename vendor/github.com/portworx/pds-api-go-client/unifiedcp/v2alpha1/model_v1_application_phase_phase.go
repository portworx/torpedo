/*
public/portworx/pds/backupconfig/apiv1/backupconfig.proto

No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)

API version: version not set
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package openapi

import (
	"encoding/json"
	"fmt"
)

// V1ApplicationPhasePhase - PHASE_UNSPECIFIED: Must be set in the proto file; ignore.  - PENDING: Application yet to be installed  - DEPLOYING: Application deployment on the target cluster in progress  - SUCCEEDED: Installed successfully  - FAILED: Application failed to install  - DELETING: Application is being deleted
type V1ApplicationPhasePhase string

// List of v1ApplicationPhasePhase
const (
	PHASE_UNSPECIFIED V1ApplicationPhasePhase = "PHASE_UNSPECIFIED"
	PENDING V1ApplicationPhasePhase = "PENDING"
	DEPLOYING V1ApplicationPhasePhase = "DEPLOYING"
	SUCCEEDED V1ApplicationPhasePhase = "SUCCEEDED"
	FAILED V1ApplicationPhasePhase = "FAILED"
	DELETING V1ApplicationPhasePhase = "DELETING"
)

// All allowed values of V1ApplicationPhasePhase enum
var AllowedV1ApplicationPhasePhaseEnumValues = []V1ApplicationPhasePhase{
	"PHASE_UNSPECIFIED",
	"PENDING",
	"DEPLOYING",
	"SUCCEEDED",
	"FAILED",
	"DELETING",
}

func (v *V1ApplicationPhasePhase) UnmarshalJSON(src []byte) error {
	var value string
	err := json.Unmarshal(src, &value)
	if err != nil {
		return err
	}
	enumTypeValue := V1ApplicationPhasePhase(value)
	for _, existing := range AllowedV1ApplicationPhasePhaseEnumValues {
		if existing == enumTypeValue {
			*v = enumTypeValue
			return nil
		}
	}

	return fmt.Errorf("%+v is not a valid V1ApplicationPhasePhase", value)
}

// NewV1ApplicationPhasePhaseFromValue returns a pointer to a valid V1ApplicationPhasePhase
// for the value passed as argument, or an error if the value passed is not allowed by the enum
func NewV1ApplicationPhasePhaseFromValue(v string) (*V1ApplicationPhasePhase, error) {
	ev := V1ApplicationPhasePhase(v)
	if ev.IsValid() {
		return &ev, nil
	} else {
		return nil, fmt.Errorf("invalid value '%v' for V1ApplicationPhasePhase: valid values are %v", v, AllowedV1ApplicationPhasePhaseEnumValues)
	}
}

// IsValid return true if the value is valid for the enum, false otherwise
func (v V1ApplicationPhasePhase) IsValid() bool {
	for _, existing := range AllowedV1ApplicationPhasePhaseEnumValues {
		if existing == v {
			return true
		}
	}
	return false
}

// Ptr returns reference to v1ApplicationPhasePhase value
func (v V1ApplicationPhasePhase) Ptr() *V1ApplicationPhasePhase {
	return &v
}

type NullableV1ApplicationPhasePhase struct {
	value *V1ApplicationPhasePhase
	isSet bool
}

func (v NullableV1ApplicationPhasePhase) Get() *V1ApplicationPhasePhase {
	return v.value
}

func (v *NullableV1ApplicationPhasePhase) Set(val *V1ApplicationPhasePhase) {
	v.value = val
	v.isSet = true
}

func (v NullableV1ApplicationPhasePhase) IsSet() bool {
	return v.isSet
}

func (v *NullableV1ApplicationPhasePhase) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableV1ApplicationPhasePhase(val *V1ApplicationPhasePhase) *NullableV1ApplicationPhasePhase {
	return &NullableV1ApplicationPhasePhase{value: val, isSet: true}
}

func (v NullableV1ApplicationPhasePhase) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableV1ApplicationPhasePhase) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
