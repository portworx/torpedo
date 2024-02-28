package apiStructs

import (
	"time"
)

type PDSRestore struct {
	V1   PDSRestoreV1
	GRPC PDSRestoreGRPC
}

type PDSRestoreV1 struct {
	Create   CreatePDSRestore
	ReCreate ReCreatePDSRestore
	Get      GetPDSRestore
	List     ListPDSRestore
	Delete   DeletePDSRestore
}

type PDSRestoreGRPC struct {
	Create   CreatePDSRestore
	ReCreate ReCreatePDSRestore
	Get      GetPDSRestore
	List     ListPDSRestore
	Delete   DeletePDSRestore
}

type CreatePDSRestore struct {
	NamespaceId string
	V1Restore   *V1Restore
}

type ReCreatePDSRestore struct {
	Id                                string
	RestoreServiceRecreateRestoreBody *RestoreServiceRecreateRestoreBody
}

type GetPDSRestore struct {
	Id string
}

type ListPDSRestore struct {
	accountId            *string
	tenantId             *string
	projectId            *string
	deploymentId         *string
	backupId             *string
	sortSortBy           *string
	sortSortOrder        *string
	paginationPageNumber *string
	paginationPageSize   *string
}

type DeletePDSRestore struct {
	Id string
}

type V1Restore struct {
	Meta   *V1Meta          `json:"meta,omitempty"`
	Config *V1Config3       `json:"config,omitempty"`
	Status *Restorev1Status `json:"status,omitempty"`
}

type V1Config3 struct {
	SourceReferences      *V1SourceReferences      `json:"sourceReferences,omitempty"`
	DestinationReferences *V1DestinationReferences `json:"destinationReferences,omitempty"`
	CustomResourceName    *string                  `json:"customResourceName,omitempty"`
}

type V1SourceReferences struct {
	DeploymentId     *string `json:"deploymentId,omitempty"`
	BackupId         *string `json:"backupId,omitempty"`
	BackupLocationId *string `json:"backupLocationId,omitempty"`
	CloudsnapId      *string `json:"cloudsnapId,omitempty"`
}

type V1DestinationReferences struct {
	TargetClusterId *string `json:"targetClusterId,omitempty"`
	DeploymentId    *string `json:"deploymentId,omitempty"`
}

type Restorev1Status struct {
	StartedAt    *time.Time   `json:"startedAt,omitempty"`
	CompletedAt  *time.Time   `json:"completedAt,omitempty"`
	ErrorCode    *V1ErrorCode `json:"errorCode,omitempty"`
	ErrorMessage *string      `json:"errorMessage,omitempty"`
	Phase        *V1Phase     `json:"phase,omitempty"`
}

type RestoreServiceRecreateRestoreBody struct {
	TargetClusterId *string `json:"targetClusterId,omitempty"`
	Name            *string `json:"name,omitempty"`
	NamespaceId     *string `json:"namespaceId,omitempty"`
}
