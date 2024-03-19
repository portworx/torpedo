package automationModels

import "time"

type PlatformTargetClusterOutput struct {
	Manifest PlatformManifestOutput `copier:"must,nopanic"`
}

type PlatformTargetClusterRequest struct {
	GetManifest            PlatformGetTargetClusterManifest `copier:"must,nopanic"`
	ListTargetClusters     PlatformListTargetCluster        `copier:"must,nopanic"`
	DeleteTargetCluster    PlatformDeleteTargetCluster      `copier:"must,nopanic"`
	GetTargetCluster       PlatformGetTargetCluster         `copier:"must,nopanic"`
	GetTargetClusterHealth PlatformGetTargetClusterHealth   `copier:"must,nopanic"`
}

type PlatformTargetClusterResponse struct {
	GetManifest            V1TargetClusterRegistrationManifest `copier:"must,nopanic"`
	ListTargetClusters     V1ListTargetClustersResponse        `copier:"must,nopanic"`
	GetTargetCluster       V1TargetCluster                     `copier:"must,nopanic"`
	GetTargetClusterHealth V1TargetCluster                     `copier:"must,nopanic"`
}

type PlatformGetTargetCluster struct {
	Id string `copier:"must,nopanic"`
}

type PlatformGetTargetClusterHealth struct {
	Id string `copier:"must,nopanic"`
}

type PlatformManifestOutput struct {
	Manifest string `copier:"must,nopanic"`
}

type PlatformGetTargetClusterManifest struct {
	ClusterName string `copier:"must,nopanic"`
	TenantId    string `copier:"must,nopanic"`
	Config      Config `copier:"must,nopanic"`
}

type PlatformListTargetCluster struct {
	TenantId string `copier:"must,nopanic"`
}

type PlatformDeleteTargetCluster struct {
	Id string `copier:"must,nopanic"`
}

type V1ListTargetClustersResponse struct {
	Clusters []V1TargetCluster `copier:"must,nopanic"`
}

type V1TargetCluster struct {
	Meta   *V1Meta                        `copier:"must,nopanic"`
	Status *PlatformTargetClusterv1Status `copier:"must,nopanic"`
}

type PlatformTargetClusterv1Status struct {
	Metadata             V1Metadata                         `copier:"must,nopanic"`
	Phase                V1TargetClusterPhasePhase          `copier:"must,nopanic"`
	LastStatusUpdateTime *time.Time                         `copier:"must,nopanic"`
	PlatformAgent        V1ApplicationPhasePhase            `copier:"must,nopanic"`
	Applications         map[string]V1ApplicationPhasePhase `copier:"must,nopanic"`
}

type V1TargetClusterPhasePhase string

type V1ApplicationPhasePhase string

type V1TargetClusterRegistrationManifest struct {
	Manifest *string `copier:"must,nopanic"`
}
