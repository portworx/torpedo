package apiStructs

type TargetClusterManifest struct {
	ClusterName string `copier:"must,nopanic"`
	TenantId    string `copier:"must,nopanic"`
	Config      Config `copier:"must,nopanic"`
}
