package resiliency

import (
	"fmt"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	dslibs "github.com/portworx/torpedo/drivers/unifiedPlatform/pdsLibs"
	"time"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	"github.com/portworx/sched-ops/k8s/apiextensions"
	"github.com/portworx/sched-ops/k8s/apps"
	"github.com/portworx/sched-ops/k8s/core"
	"github.com/portworx/sched-ops/k8s/storage"
	pdsapi "github.com/portworx/torpedo/drivers/pds/api"
	pdscontrolplane "github.com/portworx/torpedo/drivers/pds/controlplane"
	"github.com/portworx/torpedo/pkg/log"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

type PDS_Health_Status string

type Parameter struct {
	DataServiceToTest []struct {
		Name          string `json:"Name"`
		Version       string `json:"Version"`
		Image         string `json:"Image"`
		Replicas      int    `json:"Replicas"`
		ScaleReplicas int    `json:"ScaleReplicas"`
		OldVersion    string `json:"OldVersion"`
		OldImage      string `json:"OldImage"`
	} `json:"DataServiceToTest"`
	InfraToTest struct {
		ControlPlaneURL      string `json:"ControlPlaneURL"`
		AccountName          string `json:"AccountName"`
		TenantName           string `json:"TenantName"`
		ProjectName          string `json:"ProjectName"`
		ClusterType          string `json:"ClusterType"`
		Namespace            string `json:"Namespace"`
		PxNamespace          string `json:"PxNamespace"`
		PDSNamespace         string `json:"PDSNamespace"`
		ServiceIdentityToken bool   `json:"ServiceIdentityToken"`
	} `json:"InfraToTest"`
	PDSHelmVersions struct {
		LatestHelmVersion   string `json:"LatestHelmVersion"`
		PreviousHelmVersion string `json:"PreviousHelmVersion"`
	} `json:"PDSHelmVersions"`
	Users struct {
		AdminUsername    string `json:"AdminUsername"`
		AdminPassword    string `json:"AdminPassword"`
		NonAdminUsername string `json:"NonAdminUsername"`
		NonAdminPassword string `json:"NonAdminPassword"`
	} `json:"Users"`
	ResiliencyTest struct {
		CheckTillReplica int32 `json:"CheckTillReplica"`
	} `json:"ResiliencyTest"`
}

// ResourceSettingTemplate struct used to store template values
type ResourceSettingTemplate struct {
	Resources struct {
		Limits struct {
			CPU    string `json:"cpu"`
			Memory string `json:"memory"`
		} `json:"limits"`
		Requests struct {
			CPU     string `json:"cpu"`
			Memory  string `json:"memory"`
			Storage string `json:"storage"`
		} `json:"requests"`
	} `json:"resources"`
}

// WorkloadGenerationParams has data service creds
type WorkloadGenerationParams struct {
	Host                         string
	User                         string
	Password                     string
	DataServiceName              string
	DeploymentName               string
	DeploymentID                 string
	ScaleFactor                  string
	Iterations                   string
	Namespace                    string
	UseSSL, VerifyCerts, TimeOut string
	Replicas                     int
}

// StorageOptions struct used to store template values
type StorageOptions struct {
	Filesystem  string
	ForceSpread string
	Replicas    int32
	VolumeGroup bool
}

type DBConfig struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		Annotations struct {
			AdDatadoghqComElasticsearchCheckNames  string `yaml:"ad.datadoghq.com/elasticsearch.check_names"`
			AdDatadoghqComElasticsearchInitConfigs string `yaml:"ad.datadoghq.com/elasticsearch.init_configs"`
			AdDatadoghqComElasticsearchInstances   string `yaml:"ad.datadoghq.com/elasticsearch.instances"`
			AdDatadoghqComElasticsearchLogs        string `yaml:"ad.datadoghq.com/elasticsearch.logs"`
			StorkLibopenstorageOrgSkipResource     string `yaml:"stork.libopenstorage.org/skip-resource"`
		} `yaml:"annotations"`
		CreationTimestamp time.Time `yaml:"creationTimestamp"`
		Finalizers        []string  `yaml:"finalizers"`
		Generation        int       `yaml:"generation"`
		Labels            struct {
			Name                           string `yaml:"name"`
			Namespace                      string `yaml:"namespace"`
			PdsMutatorAdmit                string `yaml:"pds.mutator/admit"`
			PdsMutatorInjectCustomRegistry string `yaml:"pds.mutator/injectCustomRegistry"`
			PdsDeploymentID                string `yaml:"pds/deployment-id"`
			PdsDeploymentName              string `yaml:"pds/deployment-name"`
			PdsEnvironment                 string `yaml:"pds/environment"`
			PdsProjectID                   string `yaml:"pds/project-id"`
		} `yaml:"labels"`
		Name            string `yaml:"name"`
		Namespace       string `yaml:"namespace"`
		OwnerReferences []struct {
			APIVersion         string `yaml:"apiVersion"`
			BlockOwnerDeletion bool   `yaml:"blockOwnerDeletion"`
			Controller         bool   `yaml:"controller"`
			Kind               string `yaml:"kind"`
			Name               string `yaml:"name"`
			UID                string `yaml:"uid"`
		} `yaml:"ownerReferences"`
		ResourceVersion string `yaml:"resourceVersion"`
		UID             string `yaml:"uid"`
	} `yaml:"metadata"`
	Spec struct {
		Application      string `yaml:"application"`
		ApplicationShort string `yaml:"applicationShort"`
		Capabilities     struct {
			ParallelPod    string `yaml:"parallel_pod"`
			PdsRestore     string `yaml:"pds_restore"`
			PdsSystemUsers string `yaml:"pds_system_users"`
		} `yaml:"capabilities"`
		ConfigMapData struct {
			CLUSTERNAME        string `yaml:"CLUSTER_NAME"`
			DESIREDREPLICAS    string `yaml:"DESIRED_REPLICAS"`
			DISCOVERYSEEDHOSTS string `yaml:"DISCOVERY_SEED_HOSTS"`
			HEAPSIZE           string `yaml:"HEAP_SIZE"`
		} `yaml:"configMapData"`
		Datastorage struct {
			Name                 string `yaml:"name"`
			NumVolumes           int    `yaml:"numVolumes"`
			PersistentVolumeSpec struct {
				Metadata struct {
					Annotations struct {
						StorkLibopenstorageOrgSkipResource string `yaml:"stork.libopenstorage.org/skip-resource"`
						XPlacementStrategy                 string `yaml:"x-placement_strategy"`
					} `yaml:"annotations"`
					Name string `yaml:"name"`
				} `yaml:"metadata"`
				Spec struct {
					AccessModes []string `yaml:"accessModes"`
					Resources   struct {
						Requests struct {
							Storage string `yaml:"storage"`
						} `yaml:"requests"`
					} `yaml:"resources"`
				} `yaml:"spec"`
				Status struct {
				} `yaml:"status"`
			} `yaml:"persistentVolumeSpec"`
			StorageClass struct {
				AllowVolumeExpansion bool `yaml:"allowVolumeExpansion"`
				Metadata             struct {
					Annotations struct {
						StorkLibopenstorageOrgSkipResource string `yaml:"stork.libopenstorage.org/skip-resource"`
					} `yaml:"annotations"`
					Name string `yaml:"name"`
				} `yaml:"metadata"`
				Parameters struct {
					DisableIoProfileProtection string `yaml:"disable_io_profile_protection"`
					Fg                         string `yaml:"fg"`
					Fs                         string `yaml:"fs"`
					Group                      string `yaml:"group"`
					IoProfile                  string `yaml:"io_profile"`
					PriorityIo                 string `yaml:"priority_io"`
					Repl                       string `yaml:"repl"`
				} `yaml:"parameters"`
				Provisioner       string `yaml:"provisioner"`
				ReclaimPolicy     string `yaml:"reclaimPolicy"`
				VolumeBindingMode string `yaml:"volumeBindingMode"`
			} `yaml:"storageClass"`
		} `yaml:"datastorage"`
		DisruptionBudget struct {
			MaxUnavailable int `yaml:"maxUnavailable"`
		} `yaml:"disruptionBudget"`
		Environment string `yaml:"environment"`
		Initialize  string `yaml:"initialize"`
		RoleRules   []struct {
			APIGroups     []string `yaml:"apiGroups"`
			ResourceNames []string `yaml:"resourceNames"`
			Resources     []string `yaml:"resources"`
			Verbs         []string `yaml:"verbs"`
		} `yaml:"roleRules"`
		Service  string `yaml:"service"`
		Services []struct {
			DNSZone  string `yaml:"dnsZone"`
			Metadata struct {
			} `yaml:"metadata"`
			Name    string `yaml:"name"`
			Publish string `yaml:"publish"`
			Spec    struct {
				ClusterIP string `yaml:"clusterIP"`
				Ports     []struct {
					Name       string `yaml:"name"`
					Port       int    `yaml:"port"`
					Protocol   string `yaml:"protocol"`
					TargetPort int    `yaml:"targetPort"`
				} `yaml:"ports"`
				PublishNotReadyAddresses bool   `yaml:"publishNotReadyAddresses"`
				Type                     string `yaml:"type"`
			} `yaml:"spec,omitempty"`
		} `yaml:"services"`
		SharedStorage struct {
			PersistentVolumeClaim struct {
				Metadata struct {
					Annotations struct {
						StorkLibopenstorageOrgSkipResource         string `yaml:"stork.libopenstorage.org/skip-resource"`
						StorkLibopenstorageOrgSkipSchedulerScoring string `yaml:"stork.libopenstorage.org/skipSchedulerScoring"`
					} `yaml:"annotations"`
					Name string `yaml:"name"`
				} `yaml:"metadata"`
				Spec struct {
					AccessModes []string `yaml:"accessModes"`
					Resources   struct {
						Requests struct {
							Storage string `yaml:"storage"`
						} `yaml:"requests"`
					} `yaml:"resources"`
					StorageClassName string `yaml:"storageClassName"`
				} `yaml:"spec"`
				Status struct {
				} `yaml:"status"`
			} `yaml:"persistentVolumeClaim"`
			StorageClass struct {
				AllowVolumeExpansion bool `yaml:"allowVolumeExpansion"`
				Metadata             struct {
					Annotations struct {
						StorkLibopenstorageOrgSkipResource string `yaml:"stork.libopenstorage.org/skip-resource"`
					} `yaml:"annotations"`
					Name string `yaml:"name"`
				} `yaml:"metadata"`
				Parameters struct {
					Fs       string `yaml:"fs"`
					Repl     string `yaml:"repl"`
					Sharedv4 string `yaml:"sharedv4"`
				} `yaml:"parameters"`
				Provisioner       string `yaml:"provisioner"`
				ReclaimPolicy     string `yaml:"reclaimPolicy"`
				VolumeBindingMode string `yaml:"volumeBindingMode"`
			} `yaml:"storageClass"`
		} `yaml:"sharedStorage"`
		StatefulSet struct {
			PodManagementPolicy string `yaml:"podManagementPolicy"`
			Replicas            int    `yaml:"replicas"`
			Selector            struct {
				MatchLabels struct {
					Name                           string `yaml:"name"`
					Namespace                      string `yaml:"namespace"`
					PdsMutatorAdmit                string `yaml:"pds.mutator/admit"`
					PdsMutatorInjectCustomRegistry string `yaml:"pds.mutator/injectCustomRegistry"`
					PdsDeploymentID                string `yaml:"pds/deployment-id"`
					PdsDeploymentName              string `yaml:"pds/deployment-name"`
					PdsEnvironment                 string `yaml:"pds/environment"`
					PdsProjectID                   string `yaml:"pds/project-id"`
				} `yaml:"matchLabels"`
			} `yaml:"selector"`
			ServiceName string `yaml:"serviceName"`
			Template    struct {
				Metadata struct {
					Annotations struct {
						AdDatadoghqComElasticsearchCheckNames  string `yaml:"ad.datadoghq.com/elasticsearch.check_names"`
						AdDatadoghqComElasticsearchInitConfigs string `yaml:"ad.datadoghq.com/elasticsearch.init_configs"`
						AdDatadoghqComElasticsearchInstances   string `yaml:"ad.datadoghq.com/elasticsearch.instances"`
						AdDatadoghqComElasticsearchLogs        string `yaml:"ad.datadoghq.com/elasticsearch.logs"`
						PdsPortworxComDataService              string `yaml:"pds.portworx.com/data_service"`
						PrometheusIoPort                       string `yaml:"prometheus.io/port"`
						PrometheusIoScrape                     string `yaml:"prometheus.io/scrape"`
						StorkLibopenstorageOrgSkipResource     string `yaml:"stork.libopenstorage.org/skip-resource"`
					} `yaml:"annotations"`
					Labels struct {
						Name                           string `yaml:"name"`
						Namespace                      string `yaml:"namespace"`
						PdsMutatorAdmit                string `yaml:"pds.mutator/admit"`
						PdsMutatorInjectCustomRegistry string `yaml:"pds.mutator/injectCustomRegistry"`
						PdsDeploymentID                string `yaml:"pds/deployment-id"`
						PdsDeploymentName              string `yaml:"pds/deployment-name"`
						PdsEnvironment                 string `yaml:"pds/environment"`
						PdsProjectID                   string `yaml:"pds/project-id"`
					} `yaml:"labels"`
				} `yaml:"metadata"`
				Spec struct {
					Affinity struct {
						NodeAffinity struct {
							RequiredDuringSchedulingIgnoredDuringExecution struct {
								NodeSelectorTerms []struct {
									MatchExpressions []struct {
										Key      string   `yaml:"key"`
										Operator string   `yaml:"operator"`
										Values   []string `yaml:"values"`
									} `yaml:"matchExpressions"`
								} `yaml:"nodeSelectorTerms"`
							} `yaml:"requiredDuringSchedulingIgnoredDuringExecution"`
						} `yaml:"nodeAffinity"`
					} `yaml:"affinity"`
					Containers []struct {
						Env []struct {
							Name  string `yaml:"name"`
							Value string `yaml:"value"`
						} `yaml:"env"`
						EnvFrom []struct {
							ConfigMapRef struct {
								Name string `yaml:"name"`
							} `yaml:"configMapRef"`
						} `yaml:"envFrom,omitempty"`
						Image           string `yaml:"image"`
						ImagePullPolicy string `yaml:"imagePullPolicy,omitempty"`
						Name            string `yaml:"name"`
						Resources       struct {
							Limits struct {
								CPU              string `yaml:"cpu"`
								EphemeralStorage string `yaml:"ephemeral-storage"`
								Memory           string `yaml:"memory"`
							} `yaml:"limits"`
							Requests struct {
								CPU              string `yaml:"cpu"`
								EphemeralStorage string `yaml:"ephemeral-storage"`
								Memory           string `yaml:"memory"`
							} `yaml:"requests"`
						} `yaml:"resources"`
						SecurityContext struct {
							AllowPrivilegeEscalation bool `yaml:"allowPrivilegeEscalation"`
							Capabilities             struct {
								Drop []string `yaml:"drop"`
							} `yaml:"capabilities"`
						} `yaml:"securityContext"`
						StartupProbe struct {
							Exec struct {
								Command []string `yaml:"command"`
							} `yaml:"exec"`
							FailureThreshold int `yaml:"failureThreshold"`
							TimeoutSeconds   int `yaml:"timeoutSeconds"`
						} `yaml:"startupProbe,omitempty"`
						VolumeMounts []struct {
							MountPath string `yaml:"mountPath"`
							Name      string `yaml:"name"`
						} `yaml:"volumeMounts"`
						LivenessProbe struct {
							FailureThreshold int `yaml:"failureThreshold"`
							HTTPGet          struct {
								Path string `yaml:"path"`
								Port int    `yaml:"port"`
							} `yaml:"httpGet"`
							PeriodSeconds    int `yaml:"periodSeconds"`
							SuccessThreshold int `yaml:"successThreshold"`
							TimeoutSeconds   int `yaml:"timeoutSeconds"`
						} `yaml:"livenessProbe,omitempty"`
						Ports []struct {
							ContainerPort int    `yaml:"containerPort"`
							Protocol      string `yaml:"protocol"`
						} `yaml:"ports,omitempty"`
						ReadinessProbe struct {
							FailureThreshold int `yaml:"failureThreshold"`
							HTTPGet          struct {
								Path string `yaml:"path"`
								Port int    `yaml:"port"`
							} `yaml:"httpGet"`
							PeriodSeconds    int `yaml:"periodSeconds"`
							SuccessThreshold int `yaml:"successThreshold"`
							TimeoutSeconds   int `yaml:"timeoutSeconds"`
						} `yaml:"readinessProbe,omitempty"`
					} `yaml:"containers"`
					InitContainers []struct {
						Env []struct {
							Name  string `yaml:"name"`
							Value string `yaml:"value"`
						} `yaml:"env"`
						EnvFrom []struct {
							ConfigMapRef struct {
								Name string `yaml:"name"`
							} `yaml:"configMapRef"`
						} `yaml:"envFrom"`
						Image           string `yaml:"image"`
						ImagePullPolicy string `yaml:"imagePullPolicy"`
						Name            string `yaml:"name"`
						Resources       struct {
							Limits struct {
								CPU              string `yaml:"cpu"`
								EphemeralStorage string `yaml:"ephemeral-storage"`
								Memory           string `yaml:"memory"`
							} `yaml:"limits"`
							Requests struct {
								CPU              string `yaml:"cpu"`
								EphemeralStorage string `yaml:"ephemeral-storage"`
								Memory           string `yaml:"memory"`
							} `yaml:"requests"`
						} `yaml:"resources"`
						SecurityContext struct {
							AllowPrivilegeEscalation bool `yaml:"allowPrivilegeEscalation"`
							Capabilities             struct {
								Drop []string `yaml:"drop"`
							} `yaml:"capabilities"`
						} `yaml:"securityContext"`
						VolumeMounts []struct {
							MountPath string `yaml:"mountPath"`
							Name      string `yaml:"name"`
						} `yaml:"volumeMounts"`
					} `yaml:"initContainers"`
					SchedulerName   string `yaml:"schedulerName"`
					SecurityContext struct {
						FsGroup             int    `yaml:"fsGroup"`
						FsGroupChangePolicy string `yaml:"fsGroupChangePolicy"`
						RunAsGroup          int    `yaml:"runAsGroup"`
						RunAsNonRoot        bool   `yaml:"runAsNonRoot"`
						RunAsUser           int    `yaml:"runAsUser"`
						SeccompProfile      struct {
							Type string `yaml:"type"`
						} `yaml:"seccompProfile"`
					} `yaml:"securityContext"`
					ServiceAccountName            string `yaml:"serviceAccountName"`
					TerminationGracePeriodSeconds int    `yaml:"terminationGracePeriodSeconds"`
					Volumes                       []struct {
						EmptyDir struct {
						} `yaml:"emptyDir,omitempty"`
						Name                  string `yaml:"name"`
						PersistentVolumeClaim struct {
							ClaimName string `yaml:"claimName"`
						} `yaml:"persistentVolumeClaim,omitempty"`
						Secret struct {
							SecretName string `yaml:"secretName"`
						} `yaml:"secret,omitempty"`
					} `yaml:"volumes"`
				} `yaml:"spec"`
			} `yaml:"template"`
			UpdateStrategy struct {
				Type string `yaml:"type"`
			} `yaml:"updateStrategy"`
		} `yaml:"statefulSet"`
		Type string `yaml:"type"`
	} `yaml:"spec"`
	Status struct {
		ConnectionDetails struct {
			Nodes []string `yaml:"nodes"`
			Ports struct {
				Rest      int `yaml:"rest"`
				Transport int `yaml:"transport"`
			} `yaml:"ports"`
		} `yaml:"connectionDetails"`
		Health      string `yaml:"health"`
		Initialized string `yaml:"initialized"`
		Pods        []struct {
			IP         string `yaml:"ip"`
			Name       string `yaml:"name"`
			WorkerNode string `yaml:"workerNode"`
		} `yaml:"pods"`
		ReadyReplicas  int `yaml:"readyReplicas"`
		Replicas       int `yaml:"replicas"`
		ResourceEvents []struct {
			Resource struct {
				APIGroup string `yaml:"apiGroup"`
				Kind     string `yaml:"kind"`
				Name     string `yaml:"name"`
			} `yaml:"resource"`
		} `yaml:"resourceEvents"`
		Resources []struct {
			Conditions []struct {
				LastTransitionTime time.Time `yaml:"lastTransitionTime"`
				Message            string    `yaml:"message"`
				Reason             string    `yaml:"reason"`
				Status             string    `yaml:"status"`
				Type               string    `yaml:"type"`
			} `yaml:"conditions"`
			Resource struct {
				Kind string `yaml:"kind"`
				Name string `yaml:"name"`
			} `yaml:"resource"`
		} `yaml:"resources"`
	} `yaml:"status"`
}

type StorageClassConfig struct {
	Parameters struct {
		DisableIoProfileProtection string `yaml:"disable_io_profile_protection"`
		Fg                         string `yaml:"fg"`
		Fs                         string `yaml:"fs"`
		Group                      string `yaml:"group"`
		IoProfile                  string `yaml:"io_profile"`
		PriorityIo                 string `yaml:"priority_io"`
		Repl                       string `yaml:"repl"`
	} `yaml:"parameters"`
	Replicas  int      `yaml:"replicas"`
	Version   string   `yaml:"version"`
	Resources struct { //custom struct
		Limits struct {
			CPU              string `yaml:"cpu"`
			EphemeralStorage string `yaml:"ephemeral-storage"`
			Memory           string `yaml:"memory"`
		} `yaml:"limits"`
		Requests struct {
			CPU              string `yaml:"cpu"`
			EphemeralStorage string `yaml:"ephemeral-storage"`
			Memory           string `yaml:"memory"`
		} `yaml:"requests"`
	} `yaml:"resources"`
}

// PDS const
const (
	PDS_Health_Status_DOWN     PDS_Health_Status = "Partially Available"
	PDS_Health_Status_DEGRADED PDS_Health_Status = "Unavailable"
	PDS_Health_Status_HEALTHY  PDS_Health_Status = "Available"
	PDS_TC_Health_Status_DOWN  PDS_Health_Status = "unhealthy"

	errorChannelSize             = 50
	defaultCommandRetry          = 5 * time.Second
	defaultCommandTimeout        = 1 * time.Minute
	storageTemplateName          = "QaDefault"
	resourceTemplateName         = "Small"
	appConfigTemplateName        = "QaDefault"
	defaultRetryInterval         = 10 * time.Minute
	duration                     = 900
	timeOut                      = 30 * time.Minute
	timeInterval                 = 10 * time.Second
	maxtimeInterval              = 30 * time.Second
	resiliencyInterval           = 1 * time.Second
	defaultTestConnectionTimeout = 15 * time.Minute
	defaultWaitRebootRetry       = 10 * time.Second
	envDsVersion                 = "DS_VERSION"
	envDsBuild                   = "DS_BUILD"
	zookeeper                    = "ZooKeeper"
	redis                        = "Redis"
	consul                       = "Consul"
	cassandraStresImage          = "scylladb/scylla:4.1.11"
	postgresqlStressImage        = "portworx/torpedo-pgbench:pdsloadTest"
	consulBenchImage             = "pwxbuild/consul-bench-0.1.1"
	consulAgentImage             = "pwxbuild/consul-agent-0.1.1"
	esRallyImage                 = "elastic/rally"
	cbloadImage                  = "portworx/pds-loadtests:couchbase-0.0.2"
	pdsTpccImage                 = "portworx/torpedo-tpcc-automation:v1"
	redisStressImage             = "redis:latest"
	rmqStressImage               = "pivotalrabbitmq/perf-test:latest"
	mysqlBenchImage              = "portworx/pds-mysqlbench:v4"
	postgresql                   = "PostgreSQL"
	cassandra                    = "Cassandra"
	elasticSearch                = "Elasticsearch"
	couchbase                    = "Couchbase"
	mongodb                      = "MongoDB Enterprise"
	rabbitmq                     = "RabbitMQ"
	mysql                        = "MySQL"
	mssql                        = "MS SQL Server"
	kafka                        = "Kafka"
	pxLabel                      = "pds.portworx.com/available"
	defaultParams                = "../drivers/pds/parameters/pds_default_parameters.json"
	pdsParamsConfigmap           = "pds-params"
	configmapNamespace           = "default"
)

// K8s/PDS Instances
var (
	k8sCore       = core.Instance()
	k8sApps       = apps.Instance()
	k8sStorage    = storage.Instance()
	apiExtentions = apiextensions.Instance()
	serviceType   = "LoadBalancer"
)

// PDS vars
var (
	components    *pdsapi.Components
	deployment    *pds.ModelsDeployment
	controlplane  *pdscontrolplane.ControlPlane
	apiClient     *pds.APIClient
	ns            *corev1.Namespace
	pdsAgentpod   corev1.Pod
	ApiComponents *pdsapi.Components

	err                                   error
	isavailable                           bool
	isTemplateavailable                   bool
	isVersionAvailable                    bool
	isBuildAvailable                      bool
	currentReplicas                       int32
	deploymentTargetID, storageTemplateID string
	resourceTemplateID                    string
	appConfigTemplateID                   string
	versionID                             string
	imageID                               string
	serviceAccId                          string
	AccountID                             string
	projectID                             string
	tenantID                              string
	istargetclusterAvailable              bool
	isAccountAvailable                    bool
	isStorageTemplateAvailable            bool

	dataServiceDefaultResourceTemplateIDMap = make(map[string]string)
	dataServiceNameIDMap                    = make(map[string]string)
	dataServiceNameVersionMap               = make(map[string][]string)
	dataServiceIDImagesMap                  = make(map[string][]string)
	dataServiceNameDefaultAppConfigMap      = make(map[string]string)
	deploymentsMap                          = make(map[string][]*pds.ModelsDeployment)
	namespaceNameIDMap                      = make(map[string]string)
	dataServiceVersionBuildMap              = make(map[string][]string)
	dataServiceImageMap                     = make(map[string][]string)
)

// Function to check for set amount of Replica Pods
func GetPdsSs(depName string, ns string, checkTillReplica int32) error {
	var ss *v1.StatefulSet
	log.Debugf("expected replica %v", checkTillReplica)
	conditionError := wait.Poll(resiliencyInterval, timeOut, func() (bool, error) {
		ss, err = k8sApps.GetStatefulSet(depName, ns)
		if err != nil {
			log.Warnf("An Error Occured while getting statefulsets %v", err)
			return false, nil
		}
		log.Debugf("pods current replica %v", ss.Status.Replicas)
		if ss.Status.Replicas >= checkTillReplica {
			// Checking If this is a resiliency test case
			if ResiliencyFlag {
				ResiliencyCondition <- true
			}
			log.InfoD("Resiliency Condition Met. Will go ahead and try to induce failure now")
			return true, nil
		}
		log.Infof("Resiliency Condition still not met. Will retry to see if it has met now.....")
		return false, nil
	})
	if conditionError != nil {
		if ResiliencyFlag {
			ResiliencyCondition <- false
			CapturedErrors <- conditionError
		}
	}
	return conditionError
}

func ResizeDataServiceStorage(deployment *automationModels.V1Deployment, ds dslibs.PDSDataService, namespaceId, newResConfigId string) (bool, error) {
	log.Debugf("Starting to resize the storage and UpdateDeploymentResourceConfig")

	//Get required Id's
	stConfigId := *deployment.Config.DeploymentTopologies[0].StorageOptions.Id
	appConfigId := *deployment.Config.DeploymentTopologies[0].ServiceConfigurations.Id
	oldResConfigId := *deployment.Config.DeploymentTopologies[0].ResourceSettings.Id
	projectId := *deployment.Config.References.ProjectId
	imageId := *deployment.Config.References.ImageId
	deploymentId := *deployment.Meta.Uid

	resourceTemp, err := dslibs.GetResourceTemplateConfigs(oldResConfigId)
	if err != nil {
		if ResiliencyFlag {
			ResiliencyCondition <- false
			CapturedErrors <- err
		}
		return false, err
	}

	// Get the initial capacity of the DataService
	initialCapacity := resourceTemp.Resources.Requests.Storage
	log.Debugf("Initial Capacity of the dataservice is [%s]", initialCapacity)

	newDeployment, err := dslibs.UpdateDataService(ds, deploymentId, namespaceId, projectId, imageId, appConfigId, newResConfigId, stConfigId)
	if err != nil {
		if ResiliencyFlag {
			ResiliencyCondition <- false
			CapturedErrors <- err
		}
		return false, err
	}

	if ResiliencyFlag {
		ResiliencyCondition <- true
	}
	log.InfoD("Resiliency Condition is met, now proceeding to validate if storage size is increased.")
	err = dslibs.ValidateDeploymentConfigUpdate(*newDeployment.Update.Meta.Uid, "COMPLETED")
	if err != nil {
		if ResiliencyFlag {
			ResiliencyCondition <- false
			CapturedErrors <- err
		}
		return false, err
	}

	err = dslibs.ValidateDataServiceDeploymentHealth(deploymentId, "AVAILABLE")
	if err != nil {
		if ResiliencyFlag {
			ResiliencyCondition <- false
			CapturedErrors <- err
		}
		return false, err
	}

	UpdatedDeployment, _, err := dslibs.GetDeploymentAndPodDetails(deploymentId)
	if err != nil {
		if ResiliencyFlag {
			ResiliencyCondition <- false
			CapturedErrors <- err
		}
		return false, err
	}

	newResourceTemp, err := dslibs.GetResourceTemplateConfigs(*UpdatedDeployment.Get.Config.DeploymentTopologies[0].ResourceSettings.Id)
	if err != nil {
		if ResiliencyFlag {
			ResiliencyCondition <- false
			CapturedErrors <- err
		}
		return false, err
	}

	// Get the updated capacity of the DataService
	updatedCapacity := newResourceTemp.Resources.Requests.Storage
	log.Debugf("Updated Capacity of the dataservice is [%s]", updatedCapacity)

	if updatedCapacity > initialCapacity {
		log.InfoD("Initial PVC Capacity is- %v and Updated PVC Capacity is- %v", initialCapacity, updatedCapacity)
		log.InfoD("Storage is Successfully increased to  [%v]", updatedCapacity)
	} else {
		log.FailOnError(fmt.Errorf("Failed to verify Storage Resize at PV/PVC level \n"), "updatedCapacity should be higher than the initial capacity")
	}
	return true, nil
}
