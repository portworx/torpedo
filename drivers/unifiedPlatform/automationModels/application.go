package automationModels

type PDSApplication struct {
	Install                     PDSApplicationInstall               `copier:"must,nopanic"`
	ListAvailableAppsForTenant  PlatformListAvailableAppsForTenant  `copier:"must,nopanic"`
	ListAvailableAppsForCluster PlatformListAvailableAppsForCluster `copier:"must,nopanic"`
	Update                      PDSApplicationUpdate                `copier:"must,nopanic"`
}

type PDSApplicationResponse struct {
	List []V1Application
	Get  V1Application
}

type PDSApplicationUpdate struct {
	Application V1Application
}

type PDSApplicationInstall struct {
	ClusterId      string         `copier:"must,nopanic"`
	V1Application1 *V1Application `copier:"must,nopanic"`
}

type V1Application struct {
	Meta   *V1Meta              `copier:"must,nopanic"`
	Config *AppConfig           `copier:"must,nopanic"`
	Status *Applicationv1Status `copier:"must,nopanic"`
}

type Applicationv1Status struct {
	Version *string                `copier:"must,nopanic"`
	Phase   *ApplicationPhasePhase `copier:"must,nopanic"`
}

type ApplicationPhasePhase string

type PlatformListAvailableAppsForCluster struct {
	ClusterId string `copier:"must,nopanic"`
}

type PlatformListAvailableAppsForTenant struct {
	TenantId  string `copier:"must,nopanic"`
	ClusterId string `copier:"must,nopanic"`
}

type AppConfig struct {
	Namespace string           `copier:"must,nopanic"`
	Version   string           `copier:"must,nopanic"`
	Pds       *V1PDSProperties `copier:"must,nopanic"`
}

// V1PDSProperties PDSProperties are the properties available for PDS.
type V1PDSProperties struct {
	Global *PDSPropertiesGlobal `copier:"must,nopanic"`
}

// PDSPropertiesGlobal Global is the global property block for PDS.
type PDSPropertiesGlobal struct {
	// data_service_tls_enabled enables TLS for dataservices. This requires cert-manager to be pre-installed.
	DataServiceTlsEnabled *bool `copier:"must,nopanic"`
}
