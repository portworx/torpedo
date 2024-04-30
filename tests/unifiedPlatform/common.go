package tests

import (
	"encoding/json"
	"fmt"
	"github.com/onsi/ginkgo/v2"
	"github.com/portworx/sched-ops/k8s/core"
	"github.com/portworx/torpedo/drivers/pds/parameters"
	"github.com/portworx/torpedo/drivers/scheduler"
	platform2 "github.com/portworx/torpedo/drivers/unifiedPlatform/stworkflows/platform"
	"github.com/portworx/torpedo/pkg/aetosutil"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"time"
)

var dash *aetosutil.Dashboard

const (
	EnvControlPlaneUrl     = "CONTROL_PLANE_URL"
	DefaultTestAccount     = "pds-qa"
	DefaultProject         = "PDS-Project"
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
	WorkflowPlatform      platform2.WorkflowPlatform
	WorkflowTargetCluster platform2.WorkflowTargetCluster
	WorkflowProject       platform2.WorkflowProject
	WorkflowNamespace     platform2.WorkflowNamespace
	WorkflowCc            platform2.WorkflowCloudCredentials
	WorkflowbkpLoc        platform2.WorkflowBackupLocation
	NewPdsParams          *parameters.NewPDSParams
	PdsLabels             = make(map[string]string)
	PDS_DEFAULT_NAMESPACE string
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

// EndPDSTorpedoTest ends the logging for PDS torpedo test and updates results in testrail
func EndPDSTorpedoTest() {

	// Creating empty contexts as no contexts are created during PDS test
	contexts := make([]*scheduler.Context, 0)

	defer func() {
		err := SetSourceKubeConfig()
		log.FailOnError(err, "failed to switch context to source cluster")
	}()
	CloseLogger(TestLogger)
	Inst().Dash.TestCaseEnd()
	if TestRailSetupSuccessful && CurrentTestRailTestCaseId != 0 && RunIdForSuite != 0 {
		AfterEachTest(contexts, CurrentTestRailTestCaseId, RunIdForSuite)
	}

	currentSpecReport := ginkgo.CurrentSpecReport()
	if currentSpecReport.Failed() {
		log.Infof(">>>> FAILED TEST: %s", currentSpecReport.FullText())
	}
}

// StartPDSTorpedoTest starts the logging for PDS torpedo test
func StartPDSTorpedoTest(testName string, testDescription string, tags map[string]string, testRepoID int) {
	instanceIDString := strconv.Itoa(testRepoID)
	timestamp := time.Now().Format("01-02-15h04m05s")
	Inst().InstanceID = fmt.Sprintf("%s-%s", instanceIDString, timestamp)
	StartTorpedoTest(testName, testDescription, tags, testRepoID)
}
