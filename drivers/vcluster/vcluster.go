package vcluster

import (
	"fmt"
	"github.com/portworx/torpedo/pkg/log"
	"os/exec"
	"strings"
	"time"
)

var (
	UpdatedClusterContext string
	CurrentClusterContext string
	ContextChange         = false
)

// This method switches kube context between host and any vcluster
func SwitchKubeContext(target string) error {
	cmd := exec.Command("kubectl", "config", "get-contexts", "-o", "name")
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("Failed to fetch kube contexts: %v", err)
	}
	contexts := strings.Split(string(out), "\n")
	var desiredContext string
	if target == "host" {
		for _, ctx := range contexts {
			if ctx == "kubernetes-admin@cluster.local" {
				desiredContext = ctx
				break
			}
		}
	} else {
		prefix := fmt.Sprintf("vcluster_%s_", target)
		for _, ctx := range contexts {
			if strings.HasPrefix(ctx, prefix) {
				desiredContext = ctx
				break
			}
		}
	}
	if desiredContext == "" {
		return fmt.Errorf("Context for %s not found", target)
	}
	log.Infof("Desired Context is : %v", desiredContext)
	cmd = exec.Command("kubectl", "config", "use-context", desiredContext)
	_, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Failed to switch to context %s: %v", desiredContext, err)
	}
	cmd = exec.Command("kubectl", "config", "current-context")
	out, err = cmd.Output()
	if err != nil {
		return fmt.Errorf("Failed to get current context: %v", err)
	}
	if strings.TrimSpace(string(out)) != desiredContext {
		return fmt.Errorf("Failed to switch to the desired context: %s", desiredContext)
	}
	return nil
}

// This method deletes a vcluster
func DeleteVCluster(vclusterName string) error {
	cmd := exec.Command("vcluster", "delete", vclusterName)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("Failed to delete vcluster %v", vclusterName)
	}
	return nil
}

// This method creates a vcluster. This requires vcluster.yaml saved in a specific location.
func CreateVCluster(vclusterName string) error {
	cmd := exec.Command("vcluster", "create", vclusterName, "-f", "/Users/dbhatnagar/PxOne/vcluster/torpedo/vcluster.yaml", "--connect=false")
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to create vcluster %s: %v", vclusterName, err)
	}
	return nil
}

// This method waits for vcluster to come up in Running state and waits for a specific timeout to throw an error
func WaitForVClusterRunning(vclusterName string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		cmd := exec.Command("vcluster", "list")
		output, err := cmd.Output()
		if err != nil {
			return err
		}
		if strings.Contains(string(output), vclusterName) && strings.Contains(string(output), "Running") {
			return nil
		}
		time.Sleep(10 * time.Second)
	}
	return fmt.Errorf("vcluster %s did not reach Running status within the timeout", vclusterName)
}
