package storkctlcli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"time"

	storkv1 "github.com/libopenstorage/stork/pkg/apis/stork/v1alpha1"
	"github.com/libopenstorage/stork/pkg/storkctl"
	storkops "github.com/portworx/sched-ops/k8s/stork"
	"github.com/portworx/sched-ops/task"
	"github.com/sirupsen/logrus"

	"github.com/portworx/torpedo/pkg/aetosutil"
	"github.com/portworx/torpedo/pkg/log"
)

var dash *aetosutil.Dashboard

const (
	drPrefix            = "automation-"
	actionRetryTimeout  = 10 * time.Minute
	actionRetryInterval = 10 * time.Second
)

var (
	migSchedNs = "kube-system"
)

func ScheduleStorkctlMigrationSched(schedName, clusterPair, namespace string, extraArgs map[string]string) error {
	if namespace != "" {
		migSchedNs = namespace
	}
	cmdArgs := map[string]string{
		"cluster-pair": clusterPair,
		"namespace":    migSchedNs,
	}
	err := createMigrationScheduleCli(schedName, cmdArgs, extraArgs)
	return err
}

func createMigrationScheduleCli(schedName string, cmdArgs map[string]string, extraArgs map[string]string) error {
	factory := storkctl.NewFactory()
	var outputBuffer bytes.Buffer
	cmd := storkctl.NewCommand(factory, os.Stdin, &outputBuffer, os.Stderr)
	migCmdArgs := []string{"create", "migrationschedule", schedName}
	// add the custom args to the command
	for key, value := range cmdArgs {
		migCmdArgs = append(migCmdArgs, "--"+key)
		if value != "" {
			migCmdArgs = append(migCmdArgs, value)
		}
	}
	if extraArgs != nil {
		for key, value := range extraArgs {
			migCmdArgs = append(migCmdArgs, "--"+key)
			if value != "" {
				migCmdArgs = append(migCmdArgs, value)
			}
		}
	}
	cmd.SetArgs(migCmdArgs)
	// execute the command
	logrus.Infof("The storkctl command being executed is %v", migCmdArgs)
	if err := cmd.Execute(); err != nil {
		if err != nil {
			return fmt.Errorf("Error in executing create migration schedule command: %v", err)
		}
	}
	return nil
}

func PerformFailoverOrFailback(action, namespace, migSchdRef string, skipSourceOp bool, extraArgs map[string]string) (error, string) {
	failoverFailbackCmdArgs := []string{"perform", action, "--migration-reference", migSchdRef, "--namespace", migSchedNs}
	if namespace != "" {
		migSchedNs = namespace
	}

	factory := storkctl.NewFactory()
	var outputBuffer bytes.Buffer
	cmd := storkctl.NewCommand(factory, os.Stdin, &outputBuffer, os.Stderr)
	if skipSourceOp && action == "failover" {
		failoverFailbackCmdArgs = append(failoverFailbackCmdArgs, "--skip-source-operations")
	}
	if extraArgs != nil {
		for key, value := range extraArgs {
			failoverFailbackCmdArgs = append(failoverFailbackCmdArgs, "--"+key)
			if value != "" {
				failoverFailbackCmdArgs = append(failoverFailbackCmdArgs, value)
			}
		}
	}
	cmd.SetArgs(failoverFailbackCmdArgs)
	// execute the command
	logrus.Infof("The storkctl command being executed is %v", failoverFailbackCmdArgs)
	if err := cmd.Execute(); err != nil {
		if err != nil {
			return fmt.Errorf("Error in executing perform %v command: %v", action, err), ""
		}
	}
	// Get the captured output as a string
	actualOutput := outputBuffer.String()
	logrus.Infof("Actual output is: %s", actualOutput)
	return nil, actualOutput
}

func GetDRActionStatus(actionName, actionNamespace string) (string, string, error) {
	var action *storkv1.Action
	action, err := storkops.Instance().GetAction(actionName, actionNamespace)
	if err != nil {
		return "", "", err
	}
	return string(action.Status.Status), string(action.Status.Stage), nil
}

// WaitForMigration - waits until all migrations in the given list are successful
func WaitForActionSuccessful(actionName string, actionNamespace string, timeoutScale int) error {
	checkMigrations := func() (interface{}, bool, error) {
		isComplete := true
		status, stage, err := GetDRActionStatus(actionName, actionNamespace)
		if err != nil {
			return "", false, err
		}
		if status != "Successful" || stage != "Final" {
			isComplete = false
		}
		if isComplete {
			return "", false, nil
		}
		return "", true, fmt.Errorf("Action status is %v waiting for successful status", status)
	}
	actionTimeout := actionRetryTimeout * time.Duration(timeoutScale)
	_, err := task.DoRetryWithTimeout(checkMigrations, actionTimeout, actionRetryInterval)
	return err
}

func GetActualClusterDomainStatus() (string, string, string, error) {
	type ClusterDomainStatusOutput struct {
		Kind       string `json:"kind"`
		APIVersion string `json:"apiVersion"`
		Metadata   struct {
			ResourceVersion string `json:"resourceVersion"`
		} `json:"metadata"`
		Items []struct {
			Kind       string `json:"kind"`
			APIVersion string `json:"apiVersion"`
			Metadata   struct {
				Name              string    `json:"name"`
				UID               string    `json:"uid"`
				ResourceVersion   string    `json:"resourceVersion"`
				Generation        int       `json:"generation"`
				CreationTimestamp time.Time `json:"creationTimestamp"`
			} `json:"metadata"`
			Status struct {
				LocalDomain        string `json:"localDomain"`
				ClusterDomainInfos []struct {
					Name       string `json:"name"`
					State      string `json:"state"`
					SyncStatus string `json:"syncStatus"`
				} `json:"clusterDomainInfos"`
			} `json:"status"`
		} `json:"items"`
	}

	// Get the actual clusterdomain status.
	factory := storkctl.NewFactory()
	var outputBuffer bytes.Buffer
	cmd := storkctl.NewCommand(factory, os.Stdin, &outputBuffer, os.Stderr)
	cmdArgs := []string{"get", "clusterdomainsstatus", "-o", "json"}
	cmd.SetArgs(cmdArgs)
	// execute the command
	log.Infof("The storkctl command being executed is %v", cmdArgs)
	if err := cmd.Execute(); err != nil {
		return "", "", "", fmt.Errorf("Error in executing get cluster domain status command: %v", err)
	}
	actualOutput := outputBuffer.String()
	log.InfoD("Actual output is: %s\n", actualOutput)

	// Parse the JSON output to check the status of the witness node.
	domainStatus := ClusterDomainStatusOutput{}
	err := json.Unmarshal([]byte(actualOutput), &domainStatus)
	if err != nil {
		return "", "", "", fmt.Errorf("Error in unmarshalling the cluster domain status output: %v", err)
	}

	testClusterDomain := os.Getenv("TEST_CLUSTER_DOMAIN")

	var actualSrcDomainStatus, actualDestDomainStatus, witnessNodeDomainStatus string
	var sourceClusterDomain = fmt.Sprintf("%s1", testClusterDomain)
	var destClusterDomain = fmt.Sprintf("%s2", testClusterDomain)

	log.Infof("*******************domainInfos: %v*******************", domainStatus.Items[0].Status.ClusterDomainInfos)

	for _, info := range domainStatus.Items[0].Status.ClusterDomainInfos {
		switch info.Name {
		case "witness":
			witnessNodeDomainStatus = info.State
			log.Infof("*******************Witness node domain status: %v*******************", witnessNodeDomainStatus)
		case sourceClusterDomain:
			actualSrcDomainStatus = info.State
			log.Infof("*******************Source cluster domain status: %v*******************", actualSrcDomainStatus)
		case destClusterDomain:
			actualDestDomainStatus = info.State
			log.Infof("*******************Destination cluster domain status: %v*******************", actualDestDomainStatus)
		}
	}
	return actualSrcDomainStatus, actualDestDomainStatus, witnessNodeDomainStatus, nil
}
