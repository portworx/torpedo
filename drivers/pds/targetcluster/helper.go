package targetcluster

import (
	"fmt"
	"github.com/portworx/sched-ops/k8s/core"
	"os"
)

func createTargetKubeconfigFile() error {
	path := "/tmp/targetkuebconfig"
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error while creating the file -> %v", err)
	}
	defer f.Close()
	err = f.Truncate(0)
	if err != nil {
		return fmt.Errorf("error truncating file. Err: %v", err)
	}
	cm, err := core.Instance().GetConfigMap("target-kubeconfigs", "default")
	if err != nil {
		return err
	}
	if _, ok := cm.Data["kubeconfig"]; ok {
		kubeconfig := cm.Data["kubeconfig"]
		_, err = f.WriteString(kubeconfig)
		if err != nil {
			return fmt.Errorf("error while writing the data to file -> %v", err)
		}
		os.Setenv("TARGET_KUBECONFIG", path)
		return nil
	}
	return fmt.Errorf("key: kubeconfigs doesn't exists in the target-kubeconfigs configmap")

}
