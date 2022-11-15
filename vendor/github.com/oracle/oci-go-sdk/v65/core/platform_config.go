// Copyright (c) 2016, 2018, 2022, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

// Core Services API
//
// Use the Core Services API to manage resources such as virtual cloud networks (VCNs),
// compute instances, and block storage volumes. For more information, see the console
// documentation for the Networking (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/overview.htm),
// Compute (https://docs.cloud.oracle.com/iaas/Content/Compute/Concepts/computeoverview.htm), and
// Block Volume (https://docs.cloud.oracle.com/iaas/Content/Block/Concepts/overview.htm) services.
//

package core

import (
	"encoding/json"
	"fmt"
	"github.com/oracle/oci-go-sdk/v65/common"
	"strings"
)

// PlatformConfig The platform configuration for the instance.
type PlatformConfig interface {

	// Whether Secure Boot is enabled on the instance.
	GetIsSecureBootEnabled() *bool

	// Whether the Trusted Platform Module (TPM) is enabled on the instance.
	GetIsTrustedPlatformModuleEnabled() *bool

	// Whether the Measured Boot feature is enabled on the instance.
	GetIsMeasuredBootEnabled() *bool
}

type platformconfig struct {
	JsonData                       []byte
	IsSecureBootEnabled            *bool  `mandatory:"false" json:"isSecureBootEnabled"`
	IsTrustedPlatformModuleEnabled *bool  `mandatory:"false" json:"isTrustedPlatformModuleEnabled"`
	IsMeasuredBootEnabled          *bool  `mandatory:"false" json:"isMeasuredBootEnabled"`
	Type                           string `json:"type"`
}

// UnmarshalJSON unmarshals json
func (m *platformconfig) UnmarshalJSON(data []byte) error {
	m.JsonData = data
	type Unmarshalerplatformconfig platformconfig
	s := struct {
		Model Unmarshalerplatformconfig
	}{}
	err := json.Unmarshal(data, &s.Model)
	if err != nil {
		return err
	}
	m.IsSecureBootEnabled = s.Model.IsSecureBootEnabled
	m.IsTrustedPlatformModuleEnabled = s.Model.IsTrustedPlatformModuleEnabled
	m.IsMeasuredBootEnabled = s.Model.IsMeasuredBootEnabled
	m.Type = s.Model.Type

	return err
}

// UnmarshalPolymorphicJSON unmarshals polymorphic json
func (m *platformconfig) UnmarshalPolymorphicJSON(data []byte) (interface{}, error) {

	if data == nil || string(data) == "null" {
		return nil, nil
	}

	var err error
	switch m.Type {
	case "AMD_MILAN_BM":
		mm := AmdMilanBmPlatformConfig{}
		err = json.Unmarshal(data, &mm)
		return mm, err
	case "AMD_ROME_BM":
		mm := AmdRomeBmPlatformConfig{}
		err = json.Unmarshal(data, &mm)
		return mm, err
	case "INTEL_SKYLAKE_BM":
		mm := IntelSkylakeBmPlatformConfig{}
		err = json.Unmarshal(data, &mm)
		return mm, err
	case "AMD_ROME_BM_GPU":
		mm := AmdRomeBmGpuPlatformConfig{}
		err = json.Unmarshal(data, &mm)
		return mm, err
	case "INTEL_ICELAKE_BM":
		mm := IntelIcelakeBmPlatformConfig{}
		err = json.Unmarshal(data, &mm)
		return mm, err
	case "AMD_VM":
		mm := AmdVmPlatformConfig{}
		err = json.Unmarshal(data, &mm)
		return mm, err
	case "INTEL_VM":
		mm := IntelVmPlatformConfig{}
		err = json.Unmarshal(data, &mm)
		return mm, err
	default:
		return *m, nil
	}
}

//GetIsSecureBootEnabled returns IsSecureBootEnabled
func (m platformconfig) GetIsSecureBootEnabled() *bool {
	return m.IsSecureBootEnabled
}

//GetIsTrustedPlatformModuleEnabled returns IsTrustedPlatformModuleEnabled
func (m platformconfig) GetIsTrustedPlatformModuleEnabled() *bool {
	return m.IsTrustedPlatformModuleEnabled
}

//GetIsMeasuredBootEnabled returns IsMeasuredBootEnabled
func (m platformconfig) GetIsMeasuredBootEnabled() *bool {
	return m.IsMeasuredBootEnabled
}

func (m platformconfig) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m platformconfig) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// PlatformConfigTypeEnum Enum with underlying type: string
type PlatformConfigTypeEnum string

// Set of constants representing the allowable values for PlatformConfigTypeEnum
const (
	PlatformConfigTypeAmdMilanBm     PlatformConfigTypeEnum = "AMD_MILAN_BM"
	PlatformConfigTypeAmdRomeBm      PlatformConfigTypeEnum = "AMD_ROME_BM"
	PlatformConfigTypeAmdRomeBmGpu   PlatformConfigTypeEnum = "AMD_ROME_BM_GPU"
	PlatformConfigTypeIntelIcelakeBm PlatformConfigTypeEnum = "INTEL_ICELAKE_BM"
	PlatformConfigTypeIntelSkylakeBm PlatformConfigTypeEnum = "INTEL_SKYLAKE_BM"
	PlatformConfigTypeAmdVm          PlatformConfigTypeEnum = "AMD_VM"
	PlatformConfigTypeIntelVm        PlatformConfigTypeEnum = "INTEL_VM"
)

var mappingPlatformConfigTypeEnum = map[string]PlatformConfigTypeEnum{
	"AMD_MILAN_BM":     PlatformConfigTypeAmdMilanBm,
	"AMD_ROME_BM":      PlatformConfigTypeAmdRomeBm,
	"AMD_ROME_BM_GPU":  PlatformConfigTypeAmdRomeBmGpu,
	"INTEL_ICELAKE_BM": PlatformConfigTypeIntelIcelakeBm,
	"INTEL_SKYLAKE_BM": PlatformConfigTypeIntelSkylakeBm,
	"AMD_VM":           PlatformConfigTypeAmdVm,
	"INTEL_VM":         PlatformConfigTypeIntelVm,
}

var mappingPlatformConfigTypeEnumLowerCase = map[string]PlatformConfigTypeEnum{
	"amd_milan_bm":     PlatformConfigTypeAmdMilanBm,
	"amd_rome_bm":      PlatformConfigTypeAmdRomeBm,
	"amd_rome_bm_gpu":  PlatformConfigTypeAmdRomeBmGpu,
	"intel_icelake_bm": PlatformConfigTypeIntelIcelakeBm,
	"intel_skylake_bm": PlatformConfigTypeIntelSkylakeBm,
	"amd_vm":           PlatformConfigTypeAmdVm,
	"intel_vm":         PlatformConfigTypeIntelVm,
}

// GetPlatformConfigTypeEnumValues Enumerates the set of values for PlatformConfigTypeEnum
func GetPlatformConfigTypeEnumValues() []PlatformConfigTypeEnum {
	values := make([]PlatformConfigTypeEnum, 0)
	for _, v := range mappingPlatformConfigTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetPlatformConfigTypeEnumStringValues Enumerates the set of values in String for PlatformConfigTypeEnum
func GetPlatformConfigTypeEnumStringValues() []string {
	return []string{
		"AMD_MILAN_BM",
		"AMD_ROME_BM",
		"AMD_ROME_BM_GPU",
		"INTEL_ICELAKE_BM",
		"INTEL_SKYLAKE_BM",
		"AMD_VM",
		"INTEL_VM",
	}
}

// GetMappingPlatformConfigTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingPlatformConfigTypeEnum(val string) (PlatformConfigTypeEnum, bool) {
	enum, ok := mappingPlatformConfigTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
