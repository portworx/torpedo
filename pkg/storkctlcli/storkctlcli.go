package storkctlcli

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"time"

	storkv1 "github.com/libopenstorage/stork/pkg/apis/stork/v1alpha1"
	"github.com/libopenstorage/stork/pkg/storkctl"
	"github.com/portworx/sched-ops/task"
	"github.com/sirupsen/logrus"

	"github.com/portworx/torpedo/pkg/aetosutil"
	"github.com/portworx/torpedo/pkg/log"
)

var dash *aetosutil.Dashboard

const (
	drPrefix = "automation-"
	actionRetryTimeout = 10 * time.Minute
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
		"namespace":   migSchedNs,
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
	// Get the captured output as a string
	actualOutput := outputBuffer.String()
	logrus.Infof("Actual output is: %s", actualOutput)
	return nil
}

func PerformFailoverOrFailbackStlCli(action, namespace, migSchdRef string, skipSourceOp bool, extraArgs map[string]string) (error, string) {
	var pfCmdArgs []string
	if namespace != "" {
		migSchedNs = namespace 
	}

	factory := storkctl.NewFactory()
	var outputBuffer bytes.Buffer
	cmd := storkctl.NewCommand(factory, os.Stdin, &outputBuffer, os.Stderr)
	if action == "failover" {
		if skipSourceOp {
			pfCmdArgs = []string{"perform", "failover", "--migration-reference", migSchdRef, "--namespace", migSchedNs, "--skip-source-operations"}
		} else {
			pfCmdArgs = []string{"perform", "failover", "--migration-reference", migSchdRef, "--namespace", migSchedNs}
		}
	} else {
		pfCmdArgs = []string{"perform", "failback", "--migration-reference", migSchdRef, "--namespace", migSchedNs}
	}
	if extraArgs != nil {
		for key, value := range extraArgs {
			pfCmdArgs = append(pfCmdArgs, "--"+key)
			if value != "" {
				pfCmdArgs = append(pfCmdArgs, value)
			}
		}
	}
	cmd.SetArgs(pfCmdArgs)
	// execute the command
	logrus.Infof("The storkctl command being executed is %v", pfCmdArgs)
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

func GetDRActionStatus(actionType storkv1.ActionType, actionName string, actionNamespace string, configPath string) (string, error) {
	factory := storkctl.NewFactory()
	var outputBuffer bytes.Buffer
	cmd := storkctl.NewCommand(factory, os.Stdin, &outputBuffer, os.Stderr)
	cmdArgs := []string{"get", string(actionType), actionName, "-n", actionNamespace, "--kubeconfig", configPath}
	cmd.SetArgs(cmdArgs)
	if err := cmd.Execute(); err != nil {
		if err != nil {
			return "", fmt.Errorf("Error in executing perform %v command: %v", string(actionType), err)
		}
	}
	// Get the captured output as a string
	actualOutput := outputBuffer.String()
	log.InfoD("Actual output is: %s", actualOutput)
	output := strings.Split(actualOutput, "\n")
	startIndex := make(map[string]int)
	columns := []string{"NAME", "CREATED", "STAGE", "STATUS", "MORE INFO"}
	for _, column := range columns {
		startIndex[column] = strings.Index(output[0], column)
	}
	currentStatus := strings.TrimSpace(output[1][startIndex["STATUS"]:startIndex["MORE INFO"]])
	return currentStatus, nil
}

// WaitForMigration - waits until all migrations in the given list are successful
func WaitForActionSuccessful(actionType storkv1.ActionType, actionName string, actionNamespace string, configPath string) error {
	checkMigrations := func() (interface{}, bool, error) {
		isComplete := true
		status, err := GetDRActionStatus(actionType, actionName, actionNamespace, configPath)
		if err != nil {
			return "", false, err
		}
		if status != "Successful" {
			isComplete = false
		}
		if isComplete {
			return "", false, nil
		}
		return "", true, fmt.Errorf("Action status is %v waiting for successful status", status)
	}
	_, err := task.DoRetryWithTimeout(checkMigrations, actionRetryTimeout, actionRetryInterval)
	return err
}
