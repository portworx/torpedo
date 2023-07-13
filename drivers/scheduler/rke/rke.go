package rke

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/portworx/sched-ops/k8s/core"
	"github.com/portworx/torpedo/drivers/node"
	"github.com/portworx/torpedo/drivers/scheduler"
	kube "github.com/portworx/torpedo/drivers/scheduler/k8s"
	"github.com/portworx/torpedo/drivers/scheduler/spec"
	"github.com/portworx/torpedo/pkg/log"
	_ "github.com/rancher/norman/clientbase"
	rancherClientBase "github.com/rancher/norman/clientbase"
	rancherClient "github.com/rancher/rancher/pkg/client/generated/management/v3"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
)

const (
	// scheduleName is the name of the kubernetes scheduler driver implementation
	scheduleName = "rke"
	// SystemdScheduleServiceName is the name of the system service responsible for scheduling
	SystemdScheduleServiceName = "kubelet"
	// PxLabelNameKey is key for map
	PxLabelNameKey = "name"
	// PxLabelValue portworx pod label
	PxLabelValue = "portworx"
)

var RancherClusterParametersValue *RancherClusterParameters

type Rancher struct {
	kube.K8s
	client *rancherClient.Client
}

type RancherClusterParameters struct {
	Token     string
	Endpoint  string
	AccessKey string
	SecretKey string
}

// String returns the string name of this driver.
func (k *Rancher) String() string {
	return scheduleName
}

// Init Initialize the driver
func (k *Rancher) Init(scheduleOpts scheduler.InitOptions) error {
	var err error
	nodes, err := core.Instance().GetNodes()
	if err != nil {
		return err
	}
	for _, n := range nodes.Items {
		if err = k.AddNewNode(n); err != nil {
			return err
		}
	}
	// Update node PxPodRestartCount during init
	namespace, err := k.GetAutopilotNamespace()
	if err != nil {
		log.Fatalf(fmt.Sprintf("%v", err))
	}
	pxLabel := make(map[string]string)
	pxLabel[PxLabelNameKey] = PxLabelValue
	pxPodRestartCountMap, err := k.GetPodsRestartCount(namespace, pxLabel)
	if err != nil {
		log.Fatalf(fmt.Sprintf("%v", err))
	}
	for pod, value := range pxPodRestartCountMap {
		n, err := node.GetNodeByIP(pod.Status.HostIP)
		if err != nil {
			log.Fatalf(fmt.Sprintf("%v", err))
		}
		n.PxPodRestartCount = value
	}
	k.SpecFactory, err = spec.NewFactory(scheduleOpts.SpecDir, scheduleOpts.VolDriverName, k)
	if err != nil {
		return err
	}
	err, RancherClusterParametersValue = k.GetRancherClusterParametersValue()
	if err != nil {
		return err
	}
	rancherClientOpts := rancherClientBase.ClientOpts{
		URL:       RancherClusterParametersValue.Endpoint,
		TokenKey:  RancherClusterParametersValue.Token,
		AccessKey: RancherClusterParametersValue.AccessKey,
		SecretKey: RancherClusterParametersValue.SecretKey,
		Insecure:  true,
	}
	k.client, err = rancherClient.NewClient(&rancherClientOpts)
	if err != nil {
		return err
	}
	return nil
}

// GetRancherClusterParametersValue returns the rancher token,endpoint,secret key, access key
func (k *Rancher) GetRancherClusterParametersValue() (error, *RancherClusterParameters) {
	var data map[string]interface{}
	var var1 RancherClusterParameters
	// To Do: Rancher URL for cloud cluster will not be fetched from master node IP
	masterNodeName := node.GetMasterNodes()[0].Name
	endpoint := "https://" + masterNodeName + "/v3"
	rancherURL := "https://" + masterNodeName + "/v3-public/localProviders/local?action=login"
	//To get from config map after https://github.com/portworx/torpedo/pull/1517 is merged
	username := "admin"
	password := "1JVWD8juKPB8fVgK"
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	body := strings.NewReader(fmt.Sprintf(`{"username":"%s","password":"%s"}`, username, password))
	req, err := http.NewRequest("POST", rancherURL, body)
	if err != nil {
		log.Errorf("Failed to create rancher POST request:", err)
		return err, &var1
	}
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Failed to send request:", err)
		return err, &var1
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Failed to read response:", err)
		return err, &var1
	}
	response := string(respBody)
	err = json.Unmarshal([]byte(response), &data)
	if err != nil {
		fmt.Println("Error:", err)
		return err, &var1
	}
	var1.Endpoint = endpoint
	var1.Token = data["token"].(string)
	var1.AccessKey = strings.Split(data["token"].(string), ":")[0]
	var1.SecretKey = strings.Split(data["token"].(string), ":")[1]
	return nil, &var1
}

// GetActiveRancherClusterID returns the ID of active rancher cluster
func (k *Rancher) GetActiveRancherClusterID() (string, error) {
	var clusterId string
	clusterCollection, err := k.client.Cluster.List(nil)
	if err != nil {
		return "", err
	}
	for _, cluster := range clusterCollection.Data {
		if cluster.State == "active" {
			clusterId = cluster.ID
		}
	}
	return clusterId, nil
}

// CreateRancherProject creates new project in rancher cluster
func (k *Rancher) CreateRancherProject(projectName string, projectDescription string) (error, *rancherClient.Project) {
	var clusterId string
	var newProject *rancherClient.Project
	clusterId, err := k.GetActiveRancherClusterID()
	if err != nil {
		return err, newProject
	}
	projectRequest := &rancherClient.Project{
		Name:        projectName,
		Description: projectDescription,
		ClusterID:   clusterId,
	}
	newProject, err = k.client.Project.Create(projectRequest)
	if err != nil {
		return err, newProject
	}
	return nil, newProject
}

// GetProjectID return the project ID
func (k *Rancher) GetProjectID(projectName string) (error, string) {
	var projectId string
	projectList, err := k.client.Project.List(nil)
	if err != nil {
		return err, projectId
	}
	for _, project := range projectList.Data {
		if project.Name == projectName {
			projectId = project.ID
			break
		}
	}
	return nil, projectId
}

// AddNamespacesToProject adds namespace to the given project
func (k *Rancher) AddNamespacesToProject(projectName string, nsList []string) error {
	var projectId string
	var err error
	namespaceAnnotation := make(map[string]string)
	namespaceLabel := make(map[string]string)
	err, projectId = k.GetProjectID(projectName)
	if err != nil {
		return err
	}
	namespaceAnnotation["field.cattle.io/projectId"] = projectId
	namespaceLabel["field.cattle.io/projectId"] = strings.Split(projectId, ":")[1]
	for _, ns := range nsList {
		ns, err := core.Instance().GetNamespace(ns)
		if err != nil {
			return err
		}
		newLabels := kube.MergeMaps(ns.Labels, namespaceLabel)
		newAnnotation := kube.MergeMaps(ns.Annotations, namespaceAnnotation)
		ns.SetLabels(newLabels)
		ns.SetAnnotations(newAnnotation)
		_, err = core.Instance().UpdateNamespace(ns)
		if err != nil {
			return err
		}
	}
	return nil
}

// VerifyProjectOfNamespace verifies if the namespace belongs to a particular project
func (k *Rancher) VerifyProjectOfNamespace(projectName string, nsList []string) error {
	//var projectId string
	//var err error

	err, _ := k.GetProjectID(projectName)
	if err != nil {
		return err
	}

	ns, err := core.Instance().GetNamespace(nsList[0])
	if err != nil {
		return err
	}
	nsLabel := ns.GetLabels()
	nsAnnotation := ns.GetAnnotations()
	log.Infof(" The ns label is ", nsLabel)
	log.Infof(" The ns annotation is", nsAnnotation)

	log.Infof(" The ns label type is ", reflect.TypeOf(nsLabel))
	log.Infof(" The ns annotation type is", reflect.TypeOf(nsAnnotation))

	return nil
}

// SaveSchedulerLogsToFile gathers all scheduler logs into a file
func (k *Rancher) SaveSchedulerLogsToFile(n node.Node, location string) error {
	driver, _ := node.Get(k.K8s.NodeDriverName)
	// requires 2>&1 since docker logs command send the logs to stdrr instead of sdout
	cmd := fmt.Sprintf("docker logs %s > %s/kubelet.log 2>&1", SystemdScheduleServiceName, location)
	_, err := driver.RunCommand(n, cmd, node.ConnectionOpts{
		Timeout:         kube.DefaultTimeout,
		TimeBeforeRetry: kube.DefaultRetryInterval,
		Sudo:            true,
	})
	return err
}

func init() {
	k := &Rancher{}
	scheduler.Register(scheduleName, k)
}
