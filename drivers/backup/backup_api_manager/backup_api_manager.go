package backup_api_manager

import (
	"fmt"
	"github.com/portworx/torpedo/drivers/backup/backup_api_manager/osutils_api"
	"github.com/portworx/torpedo/drivers/backup/backup_utils"
	"github.com/portworx/torpedo/drivers/node/ssh"
	"github.com/portworx/torpedo/pkg/log"
	"github.com/portworx/torpedo/tests"
	"sync"
)

type (
	// Request represents backup API request
	Request interface{}
	// Response represents backup API response
	Response interface{}
)

var (
	// backupAPIMutex is a mutex used for synchronizing access to the backup API
	backupAPIMutex sync.RWMutex
)

// SwitchClusterContext switches the cluster context to the cluster specified by the configPath
//
// SwitchClusterContext replicates the behaviour of the `tests.SetClusterContext` function in the common.go file of the tests package,
// ensuring that errors encountered during the context switching process, including the case when retrieving the SSH node driver fails,
// are appropriately processed using the backup_utils.ProcessError function and returned
func SwitchClusterContext(configPath string) error {
	if configPath != tests.CurrentClusterConfigPath {
		log.Infof("Switching the cluster context specified by [%s] to [%s]", tests.CurrentClusterConfigPath, configPath)
		err := tests.Inst().S.SetConfig(configPath)
		if err != nil {
			debugStruct := struct {
				ConfigPath string
			}{
				ConfigPath: configPath,
			}
			return backup_utils.ProcessError(err, backup_utils.StructToString(debugStruct))
		}
		err = tests.Inst().S.RefreshNodeRegistry()
		if err != nil {
			return backup_utils.ProcessError(err)
		}
		err = tests.Inst().V.RefreshDriverEndpoints()
		if err != nil {
			return backup_utils.ProcessError(err)
		}
		if sshNodeDriver, ok := tests.Inst().N.(*ssh.SSH); ok {
			err = ssh.RefreshDriver(sshNodeDriver)
			if err != nil {
				debugStruct := struct {
					SSHNodeDriver *ssh.SSH
				}{
					SSHNodeDriver: sshNodeDriver,
				}
				return backup_utils.ProcessError(err, backup_utils.StructToString(debugStruct))
			}
		} else {
			err = fmt.Errorf("failed to get SSH node driver")
			return backup_utils.ProcessError(err)
		}
	}
	log.Infof("Switched the cluster context specified by [%s] to [%s]", tests.CurrentClusterConfigPath, configPath)
	tests.CurrentClusterConfigPath = configPath
	return nil
}

// ProcessOsutilsRequest processes osutils Request
func ProcessOsutilsRequest(request Request) (response Response, err error) {
	switch request.(type) {
	case *osutils_api.ExecShellRequest:
		response, err = osutils_api.ExecShell(request.(*osutils_api.ExecShellRequest))
	}
	if err != nil {
		return nil, err
	}
	return response, nil
}
