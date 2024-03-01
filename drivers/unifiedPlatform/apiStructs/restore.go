package apiStructs

import (
	"time"
)

type PDSRestore struct {
	V1   PDSRestoreV1   `copier:"must,nopanic"`
	GRPC PDSRestoreGRPC `copier:"must,nopanic"`
}

type PDSRestoreV1 struct {
	Create   CreatePDSRestore   `copier:"must,nopanic"`
	ReCreate ReCreatePDSRestore `copier:"must,nopanic"`
	Get      GetPDSRestore      `copier:"must,nopanic"`
	List     ListPDSRestore     `copier:"must,nopanic"`
	Delete   DeletePDSRestore   `copier:"must,nopanic"`
}

type PDSRestoreGRPC struct {
	Create   CreatePDSRestore   `copier:"must,nopanic"`
	ReCreate ReCreatePDSRestore `copier:"must,nopanic"`
	Get      GetPDSRestore      `copier:"must,nopanic"`
	List     ListPDSRestore     `copier:"must,nopanic"`
	Delete   DeletePDSRestore   `copier:"must,nopanic"`
}

type CreatePDSRestore struct {
	NamespaceId string     `copier:"must,nopanic"`
	V1Restore   *V1Restore `copier:"must,nopanic"`
}

type ReCreatePDSRestore struct {
	Id                                string                             `copier:"must,nopanic"`
	RestoreServiceRecreateRestoreBody *RestoreServiceRecreateRestoreBody `copier:"must,nopanic"`
}

type GetPDSRestore struct {
	Id string `copier:"must,nopanic"`
}

type ListPDSRestore struct {
	accountId            *string `copier:"must,nopanic"`
	tenantId             *string `copier:"must,nopanic"`
	projectId            *string `copier:"must,nopanic"`
	deploymentId         *string `copier:"must,nopanic"`
	backupId             *string `copier:"must,nopanic"`
	sortSortBy           *string `copier:"must,nopanic"`
	sortSortOrder        *string `copier:"must,nopanic"`
	paginationPageNumber *string `copier:"must,nopanic"`
	paginationPageSize   *string `copier:"must,nopanic"`
}

type DeletePDSRestore struct {
	Id string `copier:"must,nopanic"`
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
