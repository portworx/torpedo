package automationModels

type DeploymentTopology struct {
	Name *string `copier:"must,nopanic"`
	// Description of the deployment topology.
	Description *string `copier:"must,nopanic"`
	// Number of replicas of data services.
	Replicas *string `copier:"must,nopanic"`
	// Service type are standard Kubernetes service types such as clusterIP, NodePort, load balancers, etc.
	ServiceType *string `copier:"must,nopanic"`
	// Service name is the name of service as provided by user.
	ServiceName *string `copier:"must,nopanic"`
	// Source IP ranges to use for the deployed Load Balancer.
	LoadBalancerSourceRanges []string      `copier:"must,nopanic"`
	ResourceSettings         *PdsTemplates `copier:"must,nopanic"`
	ServiceConfigurations    *PdsTemplates `copier:"must,nopanic"`
	StorageOptions           *PdsTemplates `copier:"must,nopanic"`
}

type PdsTemplates struct {
	// UID of the Template.
	Id *string `copier:"must,nopanic"`
	// Resource version of the template.
	ResourceVersion *string `copier:"must,nopanic"`
	// Values required for template.
	Values *map[string]ProtobufAny4 `copier:"must,nopanic"`
}

type Reference struct {
	// UID of the target cluster in which Data Service will be deployed.
	TargetClusterId string `copier:"must,nopanic"`
	// UID of the image to be used for the Data Service Deployment.
	ImageId *string `copier:"must,nopanic"`
	// UID of the project to which DataService Deployment associated.
	ProjectId *string `copier:"must,nopanic"`
	// UID of the restore id for the Deployment.
	RestoreId *string `copier:"must,nopanic"`
}

type PDSDeployment struct {
	NamespaceID  string       `copier:"must,nopanic"`
	ProjectID    string       `copier:"must,nopanic"`
	V1Deployment V1Deployment `copier:"must,nopanic"`
}

type PDSDeploymentRequest struct {
	Create PDSDeployment
	Update PDSDeployment
}
