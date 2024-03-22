package pdslibs

import (
	"context"
	"github.com/portworx/sched-ops/k8s/apps"
	"github.com/portworx/sched-ops/k8s/core"
	"github.com/portworx/sched-ops/k8s/rbac"
	"github.com/portworx/sched-ops/k8s/storage"
	pdsdriver "github.com/portworx/torpedo/drivers/pds"
	"github.com/portworx/torpedo/drivers/unifiedPlatform"
	"github.com/portworx/torpedo/pkg/aetosutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"strconv"
	"time"
)

var (
	v2Components *unifiedPlatform.UnifiedPlatformComponents
	namespaceId  string
	err          error
)

var (
	k8sCore    = core.Instance()
	k8sApps    = apps.Instance()
	k8sRbac    = rbac.Instance()
	k8sStorage = storage.Instance()
	dash       *aetosutil.Dashboard
)

const (
	validateDeploymentTimeOut      = 50 * time.Minute
	validateDeploymentTimeInterval = 60 * time.Second
	timeOut                        = 30 * time.Minute
	maxtimeInterval                = 30 * time.Second
	pdsWorkloadImage               = "portworx/pds-loadtests:sample-load-pds-qa"
)
const (
	postgresql    = "PostgreSQL"
	cassandra     = "Cassandra"
	elasticSearch = "Elasticsearch"
	couchbase     = "Couchbase"
	mongodb       = "MongoDB Enterprise"
	rabbitmq      = "RabbitMQ"
	mysql         = "MySQL"
	mssql         = "MS SQL Server"
	kafka         = "Kafka"
	zookeeper     = "ZooKeeper"
	redis         = "Redis"
	consul        = "Consul"
)

type PDSDataService struct {
	DeploymentName        string "json:\"DeploymentName\""
	Name                  string "json:\"Name\""
	Version               string "json:\"Version\""
	Image                 string "json:\"Image\""
	Replicas              int    "json:\"Replicas\""
	ScaleReplicas         int    "json:\"ScaleReplicas\""
	OldVersion            string "json:\"OldVersion\""
	OldImage              string "json:\"OldImage\""
	DataServiceEnabledTLS bool   "json:\"DataServiceEnabledTLS\""
	ServiceType           string "json:\"ServiceType\""
}

type LoadGenParams struct {
	LoadGenDepName    string "json:\"LoadGenDepName\""
	PdsDeploymentName string "json:\"PdsDeploymentName\""
	Namespace         string "json:\"Namespace\""
	FailOnError       string "json:\"FailOnError\""
	Mode              string "json:\"Mode\""
	TableName         string "json:\"TableName\""
	NumOfRows         string "json:\"NumOfRows\""
	Iterations        string "json:\"Iterations\""
	Timeout           string "json:\"Timeout\""
	ReplacePassword   string "json:\"ReplacePassword\""
	ClusterMode       string "json:\"ClusterMode\""
	Replicas          int32  "json:\"Replicas\""
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

// StorageOptions struct used to store template values
type StorageOptions struct {
	Filesystem  string
	ForceSpread string
	Replicas    int32
	VolumeGroup bool
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

// GetCRObject
func GetCRObject(namespace, group, version, resource string) (*unstructured.UnstructuredList, error) {
	_, config, err := pdsdriver.GetK8sContext()
	if err != nil {
		return nil, err
	}

	dynamicClient := dynamic.NewForConfigOrDie(config)

	// Get the GVR of the CRD.
	gvr := metav1.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: resource,
	}
	objects, err := dynamicClient.Resource(schema.GroupVersionResource(gvr)).Namespace(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return objects, nil
}

func GetDeploymentNameAndId(deployment map[string]string) (string, string) {
	var (
		deploymentName string
		deploymentId   string
	)

	for key, value := range deployment {
		deploymentName = key
		deploymentId = value
	}

	return deploymentName, deploymentId

}

func intToPointerString(n int) *string {
	// Convert the integer to a string
	str := strconv.Itoa(n)
	// Create a pointer to the string
	ptr := &str
	// Return the pointer to the string
	return ptr
}

func int64Ptr(i int64) *int64 {
	return &i
}

func boolPtr(b bool) *bool {
	return &b
}
