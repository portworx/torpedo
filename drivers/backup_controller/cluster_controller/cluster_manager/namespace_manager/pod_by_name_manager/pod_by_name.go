package pod_by_name_manager

import (
	"fmt"
	"github.com/portworx/torpedo/drivers/backup_controller/backup_utils"
	"github.com/portworx/torpedo/drivers/backup_controller/backup_utils/osutils_api"
	"github.com/portworx/torpedo/drivers/scheduler/k8s"
	"github.com/portworx/torpedo/tests"
)

// CollectLogs saves the logs of a pod at the specified filePath
func (c *PodByNameConfig) CollectLogs(filePath string) error {
	if c == nil {
		err := fmt.Errorf("nil pod-by-name config")
		return backup_utils.ProcessError(err)
	}
	switch tests.Inst().S.(type) {
	case *k8s.K8s:
		err := backup_utils.CreateFileWithNestedDirectories(filePath)
		if err != nil {
			debugStruct := struct {
				FilePath string
			}{
				FilePath: filePath,
			}
			return backup_utils.ProcessError(err, backup_utils.StructToString(debugStruct))
		}
		podName := c.GetPodByNameMetaData().GetPodName()
		namespaceName := c.GetPodByNameMetaData().GetNamespaceMetaData().GetNamespaceName()
		command := fmt.Sprintf("kubectl logs %s -n %s > %s", podName, namespaceName, filePath)
		execShellRequest := osutils_api.NewExecShellRequest(command)
		response, err := backup_utils.ProcessOsutilsRequest(execShellRequest)
		if err != nil {
			return backup_utils.ProcessError(err, backup_utils.StructToString(execShellRequest))
		}
		execShellResponse, ok := response.(*osutils_api.ExecShellResponse)
		if ok {
			if execShellResponse.GetStdErr() != "" {
				err = fmt.Errorf("%s", execShellResponse.GetStdErr())
				return backup_utils.ProcessError(err, backup_utils.StructToString(execShellRequest))
			}
		}
	}
	return nil
}
