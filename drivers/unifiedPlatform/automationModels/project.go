package automationModels

type PlaformProjectRequest struct {
	Create     PlatformCreateProject    `copier:"must,nopanic"`
	Get        PlatformGetProject       `copier:"must,nopanic"`
	Delete     PlatformDeleteProject    `copier:"must,nopanic"`
	List       PlatformListProject      `copier:"must,nopanic"`
	Associate  PlatformAssociateProject `copier:"must,nopanic"`
	Dissociate PlatformAssociateProject `copier:"must,nopanic"`
}

type PlaformProjectResponse struct {
	Create     V1Project              `copier:"must,nopanic"`
	Get        V1Project              `copier:"must,nopanic"`
	Associate  V1Project              `copier:"must,nopanic"`
	Dissociate V1Project              `copier:"must,nopanic"`
	List       V1ListProjectsResponse `copier:"must,nopanic"`
}

type PlatformListProject struct {
	TenantId string `copier:"must,nopanic"`
}

type PlatformCreateProject struct {
	Project *V1Project `copier:"must,nopanic"`
}

type PlatformGetProject struct {
	ProjectId string `copier:"must,nopanic"`
}

type PlatformDeleteProject struct {
	ProjectId string `copier:"must,nopanic"`
}

type V1Project struct {
	Meta   *V1Meta          `copier:"must,nopanic"`
	Config *V1ProjectConfig `copier:"must,nopanic"`
	Status *Projectv1Status `copier:"must,nopanic"`
}

type V1ProjectConfig struct {
	InfraResources        *V1Resources            `copier:"must,nopanic"`
	ApplicationsResources *V1ApplicationResources `copier:"must,nopanic"`
}

type V1Resources struct {
	Clusters        []string `copier:"must,nopanic"`
	Namespaces      []string `copier:"must,nopanic"`
	Credentials     []string `copier:"must,nopanic"`
	BackupLocations []string `copier:"must,nopanic"`
	Templates       []string `copier:"must,nopanic"`
	BackupPolicies  []string `copier:"must,nopanic"`
}

type V1ApplicationResources struct {
	PdsResources *V1PDSResources `copier:"must,nopanic"`
}

type V1PDSResources struct {
	Deployments   []string `copier:"must,nopanic"`
	BackupConfigs []string `copier:"must,nopanic"`
	Restores      []string `copier:"must,nopanic"`
}

type Projectv1Status struct {
	Reason *string      `copier:"must,nopanic"`
	Phase  *V1PhaseType `copier:"must,nopanic"`
}

type PlatformAssociateProject struct {
	ProjectId                            string                                `copier:"must,nopanic"`
	ProjectServiceAssociateResourcesBody *ProjectServiceAssociateResourcesBody `copier:"must,nopanic"`
}

type PlatformDissociateProject struct {
	ProjectId                            string                                   `copier:"must,nopanic"`
	ProjectServiceAssociateResourcesBody *ProjectServiceDisassociateResourcesBody `copier:"must,nopanic"`
}

type ProjectServiceAssociateResourcesBody struct {
	InfraResource *V1Resources `copier:"must,nopanic"`
}

type ProjectServiceDisassociateResourcesBody struct {
	InfraResource *V1Resources `copier:"must,nopanic"`
}

type V1ListProjectsResponse struct {
	Projects []V1Project `copier:"must,nopanic"`
}
