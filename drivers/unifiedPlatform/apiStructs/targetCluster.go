package apiStructs

type PlatformTargetClusterOutput struct {
	Manifest PlatformManifestOutput `copier:"must,nopanic"`
}

type PlatformTargetCluster struct {
	GetManifest         PlatformGetTargetClusterManifest `copier:"must,nopanic"`
	ListTargetClusters  PlatformListTargetCluster        `copier:"must,nopanic"`
	DeleteTargetCluster PlatformDeleteTargetCluster      `copier:"must,nopanic"`
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
