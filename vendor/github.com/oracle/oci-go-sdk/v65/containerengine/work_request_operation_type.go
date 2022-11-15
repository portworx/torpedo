// Copyright (c) 2016, 2018, 2022, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

// Container Engine for Kubernetes API
//
// API for the Container Engine for Kubernetes service. Use this API to build, deploy,
// and manage cloud-native applications. For more information, see
// Overview of Container Engine for Kubernetes (https://docs.cloud.oracle.com/iaas/Content/ContEng/Concepts/contengoverview.htm).
//

package containerengine

import (
	"strings"
)

// WorkRequestOperationTypeEnum Enum with underlying type: string
type WorkRequestOperationTypeEnum string

// Set of constants representing the allowable values for WorkRequestOperationTypeEnum
const (
	WorkRequestOperationTypeClusterCreate     WorkRequestOperationTypeEnum = "CLUSTER_CREATE"
	WorkRequestOperationTypeClusterUpdate     WorkRequestOperationTypeEnum = "CLUSTER_UPDATE"
	WorkRequestOperationTypeClusterDelete     WorkRequestOperationTypeEnum = "CLUSTER_DELETE"
	WorkRequestOperationTypeNodepoolCreate    WorkRequestOperationTypeEnum = "NODEPOOL_CREATE"
	WorkRequestOperationTypeNodepoolUpdate    WorkRequestOperationTypeEnum = "NODEPOOL_UPDATE"
	WorkRequestOperationTypeNodepoolDelete    WorkRequestOperationTypeEnum = "NODEPOOL_DELETE"
	WorkRequestOperationTypeNodepoolReconcile WorkRequestOperationTypeEnum = "NODEPOOL_RECONCILE"
	WorkRequestOperationTypeWorkrequestCancel WorkRequestOperationTypeEnum = "WORKREQUEST_CANCEL"
)

var mappingWorkRequestOperationTypeEnum = map[string]WorkRequestOperationTypeEnum{
	"CLUSTER_CREATE":     WorkRequestOperationTypeClusterCreate,
	"CLUSTER_UPDATE":     WorkRequestOperationTypeClusterUpdate,
	"CLUSTER_DELETE":     WorkRequestOperationTypeClusterDelete,
	"NODEPOOL_CREATE":    WorkRequestOperationTypeNodepoolCreate,
	"NODEPOOL_UPDATE":    WorkRequestOperationTypeNodepoolUpdate,
	"NODEPOOL_DELETE":    WorkRequestOperationTypeNodepoolDelete,
	"NODEPOOL_RECONCILE": WorkRequestOperationTypeNodepoolReconcile,
	"WORKREQUEST_CANCEL": WorkRequestOperationTypeWorkrequestCancel,
}

var mappingWorkRequestOperationTypeEnumLowerCase = map[string]WorkRequestOperationTypeEnum{
	"cluster_create":     WorkRequestOperationTypeClusterCreate,
	"cluster_update":     WorkRequestOperationTypeClusterUpdate,
	"cluster_delete":     WorkRequestOperationTypeClusterDelete,
	"nodepool_create":    WorkRequestOperationTypeNodepoolCreate,
	"nodepool_update":    WorkRequestOperationTypeNodepoolUpdate,
	"nodepool_delete":    WorkRequestOperationTypeNodepoolDelete,
	"nodepool_reconcile": WorkRequestOperationTypeNodepoolReconcile,
	"workrequest_cancel": WorkRequestOperationTypeWorkrequestCancel,
}

// GetWorkRequestOperationTypeEnumValues Enumerates the set of values for WorkRequestOperationTypeEnum
func GetWorkRequestOperationTypeEnumValues() []WorkRequestOperationTypeEnum {
	values := make([]WorkRequestOperationTypeEnum, 0)
	for _, v := range mappingWorkRequestOperationTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetWorkRequestOperationTypeEnumStringValues Enumerates the set of values in String for WorkRequestOperationTypeEnum
func GetWorkRequestOperationTypeEnumStringValues() []string {
	return []string{
		"CLUSTER_CREATE",
		"CLUSTER_UPDATE",
		"CLUSTER_DELETE",
		"NODEPOOL_CREATE",
		"NODEPOOL_UPDATE",
		"NODEPOOL_DELETE",
		"NODEPOOL_RECONCILE",
		"WORKREQUEST_CANCEL",
	}
}

// GetMappingWorkRequestOperationTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingWorkRequestOperationTypeEnum(val string) (WorkRequestOperationTypeEnum, bool) {
	enum, ok := mappingWorkRequestOperationTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
