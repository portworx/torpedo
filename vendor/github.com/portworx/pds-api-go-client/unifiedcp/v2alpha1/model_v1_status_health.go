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

// V1StatusHealth Enum for Health of the Deployment.   - HEALTH_UNSPECIFIED: Health is unspecified.  - AVAILABLE: Deployment is Available.  - PARTIALLY_AVAILABLE: Deployment is PartiallyAvailable.  - UNAVAILABLE: Deployment is Unavailable.
type V1StatusHealth string

// List of v1StatusHealth
const (
	HEALTH_UNSPECIFIED V1StatusHealth = "HEALTH_UNSPECIFIED"
	AVAILABLE V1StatusHealth = "AVAILABLE"
	PARTIALLY_AVAILABLE V1StatusHealth = "PARTIALLY_AVAILABLE"
	UNAVAILABLE V1StatusHealth = "UNAVAILABLE"
)

// All allowed values of V1StatusHealth enum
var AllowedV1StatusHealthEnumValues = []V1StatusHealth{
	"HEALTH_UNSPECIFIED",
	"AVAILABLE",
	"PARTIALLY_AVAILABLE",
	"UNAVAILABLE",
}

func (v *V1StatusHealth) UnmarshalJSON(src []byte) error {
	var value string
	err := json.Unmarshal(src, &value)
	if err != nil {
		return err
	}
	enumTypeValue := V1StatusHealth(value)
	for _, existing := range AllowedV1StatusHealthEnumValues {
		if existing == enumTypeValue {
			*v = enumTypeValue
			return nil
		}
	}

	return fmt.Errorf("%+v is not a valid V1StatusHealth", value)
}

// NewV1StatusHealthFromValue returns a pointer to a valid V1StatusHealth
// for the value passed as argument, or an error if the value passed is not allowed by the enum
func NewV1StatusHealthFromValue(v string) (*V1StatusHealth, error) {
	ev := V1StatusHealth(v)
	if ev.IsValid() {
		return &ev, nil
	} else {
		return nil, fmt.Errorf("invalid value '%v' for V1StatusHealth: valid values are %v", v, AllowedV1StatusHealthEnumValues)
	}
}

// IsValid return true if the value is valid for the enum, false otherwise
func (v V1StatusHealth) IsValid() bool {
	for _, existing := range AllowedV1StatusHealthEnumValues {
		if existing == v {
			return true
		}
	}
	return false
}

// Ptr returns reference to v1StatusHealth value
func (v V1StatusHealth) Ptr() *V1StatusHealth {
	return &v
}

type NullableV1StatusHealth struct {
	value *V1StatusHealth
	isSet bool
}

func (v NullableV1StatusHealth) Get() *V1StatusHealth {
	return v.value
}

func (v *NullableV1StatusHealth) Set(val *V1StatusHealth) {
	v.value = val
	v.isSet = true
}

func (v NullableV1StatusHealth) IsSet() bool {
	return v.isSet
}

func (v *NullableV1StatusHealth) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableV1StatusHealth(val *V1StatusHealth) *NullableV1StatusHealth {
	return &NullableV1StatusHealth{value: val, isSet: true}
}

func (v NullableV1StatusHealth) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableV1StatusHealth) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
