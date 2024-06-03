package automationModels

type V1StatusHealth string
type V1StatusPhase string
type V1DeploymentTopologyStatusHealth string
type V1DeploymentTopologyStatusPhase string
type Pdsdeploymentconfigupdatev1StatusPhase string

type PDSDeploymentRequest struct {
	Create PDSDeployment
	Update PDSDeploymentUpdate
}

type PDSDeploymentResponse struct {
	Create V1Deployment
	Update V1DeploymentUpdate
	Get    V1Deployment
	List   []V1Deployment
}

type V1DeploymentUpdate struct {
	Meta   Meta
	Config DataServiceDeploymentUpdateConfig
	Status DeploymentUpdateStatus
}

// Pdsdeploymentconfigupdatev1Status Status of the deployment config update.
type DeploymentUpdateStatus struct {
	// Error Code is a short string that represents the error.
	ErrorCode *string `copier:"must,nopanic"`
	// Error Message is a description of the error.
	ErrorMessage *string `copier:"must,nopanic"`
	// Number of times the deployment config update has been retried.
	RetryCount *int32                                  `copier:"must,nopanic"`
	Phase      *Pdsdeploymentconfigupdatev1StatusPhase `copier:"must,nopanic"`
}

type DataServiceDeploymentUpdateConfig struct {
	DataServiceDeploymentMeta   Meta      `copier:"must,nopanic"`
	DataServiceDeploymentConfig V1Config1 `copier:"must,nopanic"`
}

type V1Deployment struct {
	Meta   Meta               `copier:"must,nopanic"`
	Config V1Config1          `copier:"must,nopanic"`
	Status Deploymentv1Status `copier:"must,nopanic"`
}

// Deploymentv1Status Status of the Deployment.
type Deploymentv1Status struct {
	Health *V1StatusHealth `copier:"must,nopanic"`
	Phase  *V1StatusPhase  `copier:"must,nopanic"`
	// ConnectionDetails urls, ports, credentials, etc for connecting to the data service.
	ConnectionInfo map[string]interface{} `copier:"must,nopanic"`
	// Initialize used to control startup scripts.
	Initialized *string `copier:"must,nopanic"`
	// Status of the deployment topology.
	DataServiceDeploymentTopologyStatus []V1DataServiceDeploymentTopologyStatus `copier:"must,nopanic"`
	CustomResourceName                  *string                                 `copier:"must,nopanic"`
}

// V1DeploymentTopologyStatus Status of the deployment topology. It is consumed in Deployment.
type V1DataServiceDeploymentTopologyStatus struct {
	Health *V1DeploymentTopologyStatusHealth `copier:"must,nopanic"`
	Phase  *V1DeploymentTopologyStatusPhase  `copier:"must,nopanic"`
	// Number of replicas reported by Target Cluster that are up and running.
	ReadyInstances *string           `copier:"must,nopanic"`
	ConnectionInfo *V1ConnectionInfo `copier:"must,nopanic"`
	//CustomResourceName *string
}

// V1ConnectionInfo Connection Information for the Deployment Topology.
type V1ConnectionInfo struct {
	// Ready pods.
	ReadyInstances []V1InstanceInfo `copier:"must,nopanic"`
	// Pods that are not ready.
	NotReadyPods      []V1InstanceInfo     `copier:"must,nopanic"`
	ConnectionDetails *V1ConnectionDetails `copier:"must,nopanic"`
	// Stores details about the cluster.
	ClusterDetails map[string]interface{} `copier:"must,nopanic"`
}

// V1PodInfo PodInfo contains information about a pod.
type V1InstanceInfo struct {
	// The IP of a pod.
	Ip *string `copier:"must,nopanic"`
	// Name is the Hostname of a pod.
	Name *string `copier:"must,nopanic"`
	// Node that hosts a particular pod.
	WorkerNode *string `copier:"must,nopanic"`
}

// V1ConnectionDetails ConnectionDetails of data service.
type V1ConnectionDetails struct {
	// Nodes of the data service.
	Instances []string `copier:"must,nopanic"`
	// Ports provided by the data service (name and number).
	Ports *map[string]int32 `copier:"must,nopanic"`
}

type V1Config1 struct {
	References Reference `copier:"must,nopanic"`
	// Flag to enable TLS for the Data Service.
	TlsEnabled *V1TLSConfig `copier:"must,nopanic"`
	// A deployment topology contains a number of nodes that have various attributes as a collective group.
	DataServiceDeploymentTopologies []V1DataServiceDeploymentTopology `copier:"must,nopanic"`
}

// V1TLSConfig TLS configuration for the Data Service.
type V1TLSConfig struct {
	// Flag to enable TLS for the Data Service.
	Enabled *bool `json:"enabled,omitempty"`
	// Issuer (Certificate Authority) name for the TLS certificates.
	IssuerName *string `json:"issuerName,omitempty"`
}

type V1DataServiceDeploymentTopology struct {
	Name *string `copier:"must,nopanic"`
	// Description of the deployment topology.
	Description *string `copier:"must,nopanic"`
	// Number of replicas of data services.
	Instances *string `copier:"must,nopanic"`
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

type PDSDeploymentUpdate struct {
	NamespaceID        string             `copier:"must,nopanic"`
	ProjectID          string             `copier:"must,nopanic"`
	DeploymentID       string             `copier:"must,nopanic"`
	DeploymentConfigId string             `copier:"must,nopanic"`
	V1Deployment       V1DeploymentUpdate `copier:"must,nopanic"`
}
