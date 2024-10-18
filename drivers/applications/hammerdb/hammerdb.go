package hammerdb

import (
	"fmt"
	"github.com/portworx/sched-ops/task"
	k8utils "github.com/pure-px/torpedo/drivers/utilities"
	"github.com/pure-px/torpedo/pkg/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"time"
)

type HammerDB struct {
	Namespace      string
	PodName        string
	TargetHostname string
	TargetUserName string
	TargetPassword string
	TPCCDatabase   string
	DatabaseType   string
	Labels         map[string]string
	ScaleFactor    int
	ScriptName     string
}

func (hammerDB *HammerDB) GetApplicationType() string {
	return "hammerdb"
}

func (hammerDB *HammerDB) GetNamespace() string {
	return hammerDB.Namespace
}

// DeployHammerDBPod deploys HammerDB pod
func (hammerDB *HammerDB) DeployHammerDBPod() error {

	namespace := hammerDB.Namespace

	// Create namespace if it does not exist
	err := k8utils.CreateNamespace(namespace)
	if err != nil {
		return err
	}

	// Create HammerDB Pod
	podDetails := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      hammerDB.PodName,
			Namespace: namespace,
			Labels:    hammerDB.Labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				corev1.Container{
					Name:  "hammerdb",
					Image: "tpcorg/hammerdb:mssqls",
				},
			},
		},
	}

	podDetails, err = k8sCore.CreatePod(podDetails)
	if err != nil {
		return err
	}

	// Wait for HammerDB Pod to be running
	waitForHammerDBToBeRunning := func() (interface{}, bool, error) {
		podList, err := k8sCore.GetPods(namespace, hammerDB.Labels)

		if err != nil || len(podList.Items) == 0 {
			return nil, false, fmt.Errorf("not able to find hammerdb pod under %s - [%s]", namespace, err.Error())
		}

		hammerDBPod := podList.Items[0]

		if hammerDBPod.Status.Phase == corev1.PodRunning && hammerDBPod.Status.ContainerStatuses[0].Ready {
			return nil, true, fmt.Errorf("%s pod is not running state or cotainer status is not ready", hammerDB.PodName)
		}

		time.Sleep(30 * time.Second)
		return nil, false, nil
	}

	_, err = task.DoRetryWithTimeout(waitForHammerDBToBeRunning, hammerDBPodTimeout, hammerDBPodRetryInterval)

	if err != nil {
		return err
	}

	// Copy script to pod
	err = hammerDB.CreateandCopyScriptToPod()

	return err

}

// HammerDBInit initializes HammerDB
func HammerDBInit(namespace string, targetHostname string, targetUserName string,
	targetPassword string, tpccDatabase string, databaseType string, scaleFactor int) *HammerDB {

	podName := "hammerdb" + "-" + databaseType + "-" + k8utils.RandomString(5)

	return &HammerDB{
		Namespace:      namespace,
		PodName:        podName,
		TargetHostname: targetHostname,
		TargetUserName: targetUserName,
		TargetPassword: targetPassword,
		TPCCDatabase:   tpccDatabase,
		DatabaseType:   databaseType,
		Labels:         map[string]string{"app": "hammerdb", "podName": podName, "databaseType": databaseType},
		ScaleFactor:    scaleFactor,
	}
}

// CreateTCLScript creates TCL script
func (hammerDB *HammerDB) createTCLScript() (string, error) {

	scriptName := "tprocch" + "-" + hammerDB.DatabaseType + "-" + k8utils.RandomString(5) + ".tcl"

	scriptContent := fmt.Sprintf(tclScriptMssqlTprocch,
		hammerDB.TargetHostname,
		hammerDB.TargetHostname,
		hammerDB.TargetUserName,
		hammerDB.TargetPassword,
		hammerDB.ScaleFactor,
		hammerDB.TPCCDatabase)

	log.Infof("Script Content: %s", scriptContent)

	err := os.WriteFile(scriptName, []byte(scriptContent), 0755)
	if err != nil {
		return "", err
	}

	return scriptName, nil
}

// CreateandCopyScriptToPod creates TCL script and copies it to HammerDB pod
func (hammerDB *HammerDB) CreateandCopyScriptToPod() error {

	sciptName, err := hammerDB.createTCLScript()
	if err != nil {
		return err
	}

	err = k8utils.CopyFileToPod(hammerDB.Namespace, hammerDB.PodName, sciptName, defaultPath)
	if err != nil {
		return err
	}

	hammerDB.ScriptName = sciptName

	return nil
}

// createTriggerScriptOnPod creates trigger script on HammerDB pod
func (hammerDB *HammerDB) createTriggerScriptOnPod() (string, error) {

	triggerScriptName := "trigger-hammer-" + k8utils.RandomString(5) + ".sh"

	triggerScriptContent := fmt.Sprintf(triggerScript,
		defaultPath,
		defaultPath,
		hammerDB.ScriptName)

	log.Infof("Trigger Script Content: %s", triggerScriptContent)

	err := os.WriteFile(triggerScriptName, []byte(triggerScriptContent), 0755)
	if err != nil {
		return "", err
	}

	err = k8utils.CopyFileToPod(hammerDB.Namespace, hammerDB.PodName, triggerScriptName, defaultPath)
	if err != nil {
		return "", err
	}

	return triggerScriptName, nil

}

// RunHammerDBLoad runs HammerDB load
func (hammerDB *HammerDB) RunHammerDBLoad() (string, error) {

	triggerScriptName, err := hammerDB.createTriggerScriptOnPod()

	if err != nil {
		return "", err
	}

	// Changing the permission of trigger script
	log.Infof("Changing the permission of trigger script [%s%s]", defaultPath, triggerScriptName)
	_, err = k8sCore.RunCommandInPod(
		[]string{"chmod", "+x", fmt.Sprintf("%s%s", defaultPath, hammerDB.ScriptName)},
		hammerDB.PodName,
		"hammerdb",
		hammerDB.Namespace)

	if err != nil {
		return "", err
	}

	// Running trigger script
	log.Infof("Running trigger script [%s%s]", defaultPath, triggerScriptName)
	output, err := k8sCore.RunCommandInPod(
		[]string{"sh", fmt.Sprintf("%s%s", defaultPath, triggerScriptName)},
		hammerDB.PodName,
		"hammerdb",
		hammerDB.Namespace)

	return output, err
}
