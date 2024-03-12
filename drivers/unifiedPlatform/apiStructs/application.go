package apiStructs

type PDSApplicaition struct {
	Install                    PDSApplicationInstall              `copier:"must,nopanic"`
	ListAvailableAppsForTenant PlatformListAvailableAppsForTenant `copier:"must,nopanic"`
}

type PDSApplicationInstall struct {
	ClusterId      string          `copier:"must,nopanic"`
	V1Application1 *V1Application1 `copier:"must,nopanic"`
}

type V1Application1 struct {
	Meta   *V1Meta              `copier:"must,nopanic"`
	Config *AppConfig           `copier:"must,nopanic"`
	Status *Applicationv1Status `copier:"must,nopanic"`
}

type Applicationv1Status struct {
	Version *string                `copier:"must,nopanic"`
	Phase   *ApplicationPhasePhase `copier:"must,nopanic"`
}

type ApplicationPhasePhase string

type PlatformListAvailableAppsForTenant struct {
	TenantId  string `copier:"must,nopanic"`
	ClusterId string `copier:"must,nopanic"`
}

type AppConfig struct {
	Namespace string `copier:"must,nopanic"`
	Version   string `copier:"must,nopanic"`
}
