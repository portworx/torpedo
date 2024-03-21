package tests

import (
	"encoding/json"
	"github.com/portworx/sched-ops/k8s/core"
	"github.com/portworx/torpedo/drivers/pds/parameters"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/stworkflows"
	"github.com/portworx/torpedo/pkg/log"
	"io/ioutil"
	"path/filepath"
)

const (
	EnvControlPlaneUrl     = "CONTROL_PLANE_URL"
	DefaultTestAccount     = "pds-qa"
	DefaultProject         = "px-system-project"
	DefaultTenant          = "px-system-tenant"
	EnvPlatformAccountName = "PLATFORM_ACCOUNT_NAME"
	EnvAccountDisplayName  = "PLATFORM_ACCOUNT_DISPLAY_NAME"
	EnvUserMailId          = "USER_MAIL_ID"
	defaultParams          = "../drivers/pds/parameters/pds_default_parameters.json"
	pdsParamsConfigmap     = "pds-params"
	configmapNamespace     = "default"
)

var (
	AccID         string
	PDSBucketName string
	Namespace     string
	ProjectId     string
)

var (
	WorkflowPlatform      stworkflows.WorkflowPlatform
	WorkflowTargetCluster stworkflows.WorkflowTargetCluster
	WorkflowProject       stworkflows.WorkflowProject
	WorkflowNamespace     stworkflows.WorkflowNamespace
	NewPdsParams          *parameters.NewPDSParams
	CustomParams          *parameters.Customparams
	PdsLabels             = make(map[string]string)
)

// ReadParams reads the params from given or default json
func ReadNewParams(filename string) (*parameters.NewPDSParams, error) {
	var jsonPara parameters.NewPDSParams
	var err error

	if filename == "" {
		filename, err = filepath.Abs(defaultParams)
		log.Infof("filename %v", filename)
		if err != nil {
			return nil, err
		}
		log.Infof("Parameter json file is not used, use initial parameters value.")
		log.InfoD("Reading params from %v ", filename)
		file, err := ioutil.ReadFile(filename)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(file, &jsonPara)
		if err != nil {
			return nil, err
		}
	} else {
		cm, err := core.Instance().GetConfigMap(pdsParamsConfigmap, configmapNamespace)
		if err != nil {
			return nil, err
		}
		if len(cm.Data) > 0 {
			configmap := &cm.Data
			for key, data := range *configmap {
				log.InfoD("key %v \n value %v", key, data)
				json_data := []byte(data)
				err = json.Unmarshal(json_data, &jsonPara)
				if err != nil {
					log.FailOnError(err, "Error while unmarshalling json:")
				}
			}
		}
	}
	return &jsonPara, nil
}
