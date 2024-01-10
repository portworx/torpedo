package postmanApiLoadDriver

import (
	"fmt"
	"github.com/portworx/sched-ops/k8s/apps"
	"github.com/portworx/sched-ops/k8s/core"
	"github.com/portworx/sched-ops/k8s/rbac"
	"github.com/portworx/torpedo/pkg/log"
	"github.com/portworx/torpedo/pkg/osutils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	postmanCentosImage          = "portworx/torpedo-postman:v1"
	deploymentName              = "postman-newman"
	defaultCollectionPath       = "../drivers/postmanApiLoadDriver/collections/collection.json"
	timeOut                     = 30 * time.Minute
	timeInterval                = 10 * time.Second
	PxDataServices              = "pds"
	PdsControlPlaneApiKeyExpiry = "2025-01-02T15:04:05Z"
)

var k8sRbac = rbac.Instance()
var k8sCore = core.Instance()

type PostmanDriver struct {
	ResultsFileName string
	ResultType      string
	Namespace       string
	Replicas        int32
	Iteration       string
	Kubeconfig      string
}

// ConfigMapData represents the structure of the ConfigMap data field
type ConfigMapData struct {
	YourField string `yaml:"postman-configmap"`
}

var k8sApps = apps.Instance()

func GetProjectNameToExecutePostman(projectName string, driver *PostmanDriver) {
	if projectName == PxDataServices {
		_, err := ExecutePostmanCommandInTorpedo()
		if err != nil {
			log.FailOnError(err, "Postman execution failed.. Please check the logs manually.")
		}
	}
	//ToDo: Add cases for other PX Projects
}

func GetPostmanCollectionPath() (string, error) {
	postmanCollectionFile, err := filepath.Abs(defaultCollectionPath)
	if err != nil {
		return "", fmt.Errorf("postman Collection Json not found, Please create a Collection json manually and export to {%v} folder", defaultCollectionPath)
	}
	flag, _ := CheckIfPostmanCollectionIsAvailable(postmanCollectionFile)
	if flag != true {
		return "", fmt.Errorf("postman Collection Json not found, Please create a Collection json manually and export to {%v} folder", defaultCollectionPath)
	}
	log.InfoD("PostmanCollectionFile found is- [%v]", postmanCollectionFile)
	return postmanCollectionFile, nil
}

func CheckIfPostmanCollectionIsAvailable(postmanCollectionFile string) (bool, error) {
	_, err := os.Stat(postmanCollectionFile)
	if os.IsNotExist(err) {
		return false, err
	}
	return true, nil
}

// Create Newman k8s Pod
func (postman *PostmanDriver) RunNewmanWorkload(postmanParams *PostmanDriver) (*corev1.Pod, error) {
	podSpec := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: deploymentName + "-",
			Namespace:    postmanParams.Namespace,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    deploymentName,
					Image:   postmanCentosImage,
					Command: []string{"sleep", "infinity"},
				},
			},
			RestartPolicy: corev1.RestartPolicyOnFailure,
		},
	}

	pod, err := k8sCore.CreatePod(podSpec)
	if err != nil {
		return nil, fmt.Errorf("failed to create pod [%s], Err: %v", podSpec.Name, err)
	}

	err = k8sCore.ValidatePod(pod, timeOut, timeInterval)
	if err != nil {
		return nil, fmt.Errorf("failed to validate pod [%s], Err: %v", pod.Name, err)
	}
	time.Sleep(1 * time.Minute)

	return pod, nil
}

func RunKubectlCommandInPod(podName string, namespace string, kubeconfig string, command string) (string, error) {
	cmd := fmt.Sprintf("kubectl --kubeconfig %v -n %v exec -it %v -- %v", kubeconfig, namespace, podName, command)
	log.Infof("Command: ", cmd)
	output, _, err := osutils.ExecShell(cmd)
	if err != nil {
		return "", err
	}
	log.Infof("Terminal output: %v", output)
	return output, nil
}

func ExecuteCommandInShell(command string) (string, string, error) {
	out, res, err := osutils.ExecShell(command)
	if err != nil {
		return "", "", err
	}
	return out, res, nil
}

func ExecutePostmanForPdsInAK8sPod(postmanParams *PostmanDriver) (*corev1.Pod, bool, error) {
	collectionPath, err := GetPostmanCollectionPath()
	if err != nil {
		log.FailOnError(err, "Postman Collection Json not found, Please create a Collection json manually and export to {%v} folder", defaultCollectionPath)
	}
	log.InfoD("Postman Collection found is- %v", collectionPath)

	postDep, err := postmanParams.RunNewmanWorkload(postmanParams)
	copyCmd := "kubectl cp " + collectionPath + " " + postDep.Name + ":collection.json" + " -n " + postmanParams.Namespace
	out, res, err := ExecuteCommandInShell(copyCmd)
	if err != nil {
		return nil, false, fmt.Errorf("there was some problem in executing Postman Newman container due to- [%v]", err)
	}
	log.Info("output of copy command is - %v, %v", out, res)
	newmanCmd := "newman run collection.json --verbose"
	output, err := RunKubectlCommandInPod(postDep.Name, postmanParams.Namespace, postmanParams.Kubeconfig, newmanCmd)
	if err != nil {
		return nil, false, fmt.Errorf("there was some problem in executing Postman Newman container due to- [%v]", err)
	}
	log.InfoD("output from the newman execution is- %v", output)
	if strings.Contains(output, "failure") {
		log.FailOnError(err, "newman exited with a failure.. Please check logs for more details")
	}
	k8sCore.DeletePod(postDep.Name, postDep.Namespace, true)
	return postDep, true, nil
}

func ExecutePostmanCommandInTorpedo() (bool, error) {
	collectionPath, err := GetPostmanCollectionPath()
	if err != nil {
		log.FailOnError(err, "Postman Collection Json not found, Please create a Collection json manually and export to {%v} folder", defaultCollectionPath)
	}
	log.InfoD("Postman Collection found is- %v", collectionPath)

	newmanCmd := "newman run " + collectionPath + " --verbose"
	log.InfoD("Newman command formed is- [%v]", newmanCmd)
	output, _, err := ExecuteCommandInShell(newmanCmd)
	if err != nil {
		return false, fmt.Errorf("there was some problem in executing Postman Newman container due to- [%v]", output)
	}
	log.InfoD("output from the newman execution is- %v", output)
	if strings.Contains(output, "failure") {
		log.FailOnError(err, "newman exited with a failure.. Please check logs for more details")
	}
	return true, nil
}
