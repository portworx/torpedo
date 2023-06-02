package tests

import (
	"fmt"
	"time"

	"github.com/portworx/torpedo/drivers/pds/parameters"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/tests"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"

	"github.com/portworx/sched-ops/k8s/core"
	pdslib "github.com/portworx/torpedo/drivers/pds/lib"
	"github.com/portworx/torpedo/pkg/aetosutil"
	"github.com/portworx/torpedo/pkg/log"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

type PDSDataService struct {
	Name          string "json:\"Name\""
	Version       string "json:\"Version\""
	Image         string "json:\"Image\""
	Replicas      int    "json:\"Replicas\""
	ScaleReplicas int    "json:\"ScaleReplicas\""
	OldVersion    string "json:\"OldVersion\""
	OldImage      string "json:\"OldImage\""
}

const (
	pdsNamespace                     = "pds-system"
	deploymentName                   = "qa"
	envDeployAllDataService          = "DEPLOY_ALL_DATASERVICE"
	postgresql                       = "PostgreSQL"
	cassandra                        = "Cassandra"
	elasticSearch                    = "Elasticsearch"
	couchbase                        = "Couchbase"
	redis                            = "Redis"
	rabbitmq                         = "RabbitMQ"
	mongodb                          = "MongoDB"
	mysql                            = "MySQL"
	kafka                            = "Kafka"
	zookeeper                        = "ZooKeeper"
	consul                           = "Consul"
	pdsNamespaceLabel                = "pds.portworx.com/available"
	timeOut                          = 30 * time.Minute
	maxtimeInterval                  = 30 * time.Second
	timeInterval                     = 1 * time.Second
	ActiveNodeRebootDuringDeployment = "active-node-reboot-during-deployment"
	RebootNodeDuringAppVersionUpdate = "reboot-node-during-app-version-update"
	KillDeploymentControllerPod      = "kill-deployment-controller-pod-during-deployment"
	RestartPxDuringDSScaleUp         = "restart-portworx-during-ds-scaleup"
	RestartAppDuringResourceUpdate   = "restart-app-during-resource-update"
	BackUpCRD                        = "backups.pds.io"
	DeploymentCRD                    = "deployments.pds.io"
	RebootNodesDuringDeployment      = "reboot-multiple-nodes-during-deployment"
	KillAgentPodDuringDeployment     = "kill-agent-pod-during-deployment"
	KillTeleportPodDuringDeployment  = "kill-teleport-pod-during-deployment"
)

var (
	namespace                               string
	pxnamespace                             string
	tenantID                                string
	dnsZone                                 string
	clusterID                               string
	projectID                               string
	serviceType                             string
	deploymentTargetID                      string
	replicas                                int32
	err                                     error
	supportedDataServices                   []string
	dataServiceNameDefaultAppConfigMap      map[string]string
	namespaceID                             string
	storageTemplateID                       string
	dataServiceDefaultResourceTemplateIDMap map[string]string
	dataServiceNameIDMap                    map[string]string
	supportedDataServicesNameIDMap          map[string]string
	DeployAllVersions                       bool
	DataService                             string
	DeployAllImages                         bool
	dataServiceDefaultResourceTemplateID    string
	dataServiceDefaultAppConfigID           string
	dataServiceVersionBuildMap              map[string][]string
	dataServiceImageMap                     map[string][]string
	dep                                     *v1.Deployment
	pod                                     *corev1.Pod
	params                                  *parameters.Parameter
	podList                                 *corev1.PodList
	isDeploymentsDeleted                    bool
	isNamespacesDeleted                     bool
	dash                                    *aetosutil.Dashboard
	deployment                              *pds.ModelsDeployment
	k8sCore                                 = core.Instance()
	pdsLabels                               = make(map[string]string)
	accountID                               string
)

var dataServiceDeploymentWorkloads = []string{cassandra, elasticSearch, postgresql, consul, mysql}
var dataServicePodWorkloads = []string{redis, rabbitmq, couchbase}

func RunWorkloads(params pdslib.WorkloadGenerationParams, ds PDSDataService, deployment *pds.ModelsDeployment, namespace string) (*corev1.Pod, *v1.Deployment, error) {
	params.DataServiceName = ds.Name
	params.DeploymentID = deployment.GetId()
	params.Namespace = namespace
	log.Infof("Dataservice Name : %s", ds.Name)

	if ds.Name == postgresql {
		params.DeploymentName = "pgload"
		params.ScaleFactor = "100"
		params.Iterations = "1"

		log.Infof("Running Workloads on DataService %v ", ds.Name)
		pod, dep, err = pdslib.CreateDataServiceWorkloads(params)

	}
	if ds.Name == rabbitmq {
		params.DeploymentName = "rmq"
		log.Infof("Running Workloads on DataService %v ", ds.Name)
		pod, dep, err = pdslib.CreateDataServiceWorkloads(params)

	}
	if ds.Name == redis {
		params.DeploymentName = "redisbench"
		params.Replicas = ds.Replicas
		log.Infof("Running Workloads on DataService %v ", ds.Name)

		pod, dep, err = pdslib.CreateDataServiceWorkloads(params)

	}
	if ds.Name == cassandra {
		params.DeploymentName = "cassandra-stress"
		log.Infof("Running Workloads on DataService %v ", ds.Name)
		pod, dep, err = pdslib.CreateDataServiceWorkloads(params)

	}
	if ds.Name == elasticSearch {
		params.DeploymentName = "es-rally"
		params.User = "elastic"
		params.UseSSL = "false"
		params.VerifyCerts = "false"
		params.TimeOut = "60"
		log.Infof("Running Workloads on DataService %v ", ds.Name)
		pod, dep, err = pdslib.CreateDataServiceWorkloads(params)

	}
	if ds.Name == couchbase {
		params.DeploymentName = "cb-load"
		log.Infof("Running Workloads on DataService %v ", ds.Name)
		pod, dep, err = pdslib.CreateDataServiceWorkloads(params)

	}
	if ds.Name == consul {
		params.DeploymentName = *deployment.ClusterResourceName
		log.Infof("Running Workloads on DataService %v ", ds.Name)
		pod, dep, err = pdslib.CreateDataServiceWorkloads(params)
	}
	if ds.Name == mysql {
		params.DeploymentName = *deployment.ClusterResourceName
		log.Infof("Running Workloads on DataService %v ", ds.Name)
		pod, dep, err = pdslib.CreateDataServiceWorkloads(params)
	}

	return pod, dep, err

}

// Check the DS related PV usage and resize in case of 90% full
func GetPVCtoFullConditionAndResize(deploymentName string, namespace string, context []*scheduler.Context) error {
	log.Debugf("Start polling the pvc consumption for the DS %v", deploymentName)
	threshold := 90
	isthresholdmet := false
	for _, ctx := range context {
		vols, err := tests.Inst().S.GetVolumes(ctx)
		if err != nil {
			return fmt.Errorf("persistant volumes Not Found due to : %v", err)
		}
		waitErr := wait.Poll(timeOut, timeInterval, func() (bool, error) {
			for _, vol := range vols {
				log.Debugf("VOLUME TO BE INSPECTED IS : %v", vol)
				appVol, err := tests.Inst().V.InspectVolume(vol.ID)
				log.Debugf("THE VOL DESC IS ----- %v", appVol)
				if err != nil {
					return true, err
				}
				usedBytes := appVol.GetUsage()
				log.Debugf("Capacity in bytes is %v", appVol.Spec.Size)
				log.Debugf("USED IN BYTES IS ---- %v", usedBytes)
				pvcCapacity := appVol.Spec.Size
				pvcUsed := int((usedBytes / pvcCapacity) * 100)
				log.Debugf("Threshold achieved ---- %v", pvcUsed)
				if pvcUsed >= threshold {
					isthresholdmet = true
				}
			}
			if isthresholdmet {
				return true, nil
			}
			return false, nil
		})
		if !isthresholdmet {
			return fmt.Errorf("threshold not met due to : %v", waitErr)
		}
	}

	return err
}
