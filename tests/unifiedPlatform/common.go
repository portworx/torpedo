package tests

import (
	"encoding/json"
	"fmt"
	"github.com/onsi/ginkgo/v2"
	"github.com/portworx/sched-ops/k8s/core"
	"github.com/portworx/torpedo/drivers/pds/parameters"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/stworkflows/pds"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/stworkflows/platform"
	"github.com/portworx/torpedo/pkg/aetosutil"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"time"
)

var dash *aetosutil.Dashboard

var (
	WorkflowDataService     pds.WorkflowDataService
	WorkflowPDSBackupConfig pds.WorkflowPDSBackupConfig
	WorkflowPDSBackup       pds.WorkflowPDSBackup
	WorkflowPDSRestore      pds.WorkflowPDSRestore
	WorkflowPDSTemplate     pds.WorkflowPDSTemplates
)

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
	WorkflowPlatform      platform.WorkflowPlatform
	WorkflowTargetCluster platform.WorkflowTargetCluster
	WorkflowProject       platform.WorkflowProject
	WorkflowNamespace     platform.WorkflowNamespace
	WorkflowCc            platform.WorkflowCloudCredentials
	WorkflowbkpLoc        platform.WorkflowBackupLocation
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

	Step("Purging all PDS related objects", func() {

		PurgePDS()

	})

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

	PDS_DEFAULT_NAMESPACE = "pds-namespace-" + RandomString(5)

	Step("Create a namespace for PDS", func() {
		WorkflowNamespace.TargetCluster = WorkflowTargetCluster
		WorkflowNamespace.Namespaces = make(map[string]string)
		_, err := WorkflowNamespace.CreateNamespaces(PDS_DEFAULT_NAMESPACE)
		log.FailOnError(err, "Unable to create namespace")
		log.Infof("Namespaces created - [%s]", WorkflowNamespace.Namespaces)
	})

	Step("Associate namespace to Project", func() {
		err := WorkflowProject.Associate(
			[]string{},
			[]string{WorkflowNamespace.Namespaces[PDS_DEFAULT_NAMESPACE]},
			[]string{},
			[]string{},
			[]string{},
			[]string{},
		)
		log.FailOnError(err, "Unable to associate Cluster to Project")
		log.Infof("Associated Resources - [%+v]", WorkflowProject.AssociatedResources)
	})

	Step("Creating all PDS related structs", func() {

		log.Infof("Creating data service struct")
		WorkflowDataService.NamespaceName = PDS_DEFAULT_NAMESPACE
		WorkflowDataService.Namespace = WorkflowNamespace
		WorkflowDataService.DataServiceDeployment = make(map[string]string)
		WorkflowDataService.Dash = Inst().Dash

		log.Infof("Creating backup config struct")
		WorkflowPDSBackupConfig.WorkflowBackupLocation = WorkflowbkpLoc
		WorkflowPDSBackupConfig.WorkflowDataService = WorkflowDataService
		WorkflowPDSBackupConfig.Backups = make(map[string]automationModels.V1BackupConfig)

		log.Infof("Creating Backup struct")
		WorkflowPDSBackup.WorkflowDataService = WorkflowDataService

		log.Infof("Creating restore object for same cluster and same project")
		WorkflowPDSRestore.Source = WorkflowNamespace
		WorkflowPDSRestore.Destination = WorkflowNamespace
		WorkflowPDSRestore.RestoredDeployments = pds.WorkflowDataService{}
		WorkflowPDSRestore.RestoredDeployments.DataServiceDeployment = make(map[string]string)
	})

	instanceIDString := strconv.Itoa(testRepoID)
	timestamp := time.Now().Format("01-02-15h04m05s")
	Inst().InstanceID = fmt.Sprintf("%s-%s", instanceIDString, timestamp)
	StartTorpedoTest(testName, testDescription, tags, testRepoID)
}

// PurgePDS purges all default PDS related resources created during testcase run
func PurgePDS() {
	log.InfoD("Purging all dataservice objects")
	err := WorkflowDataService.Purge()
	log.FailOnError(err, "some error occurred while purging data service objects")

	log.InfoD("Purging all backup objects")
	err = WorkflowPDSBackupConfig.Purge(true)
	log.FailOnError(err, "some error occurred while purging backup config objects")

	log.InfoD("Purging all restore objects")
	err = WorkflowPDSRestore.Purge()
	log.FailOnError(err, "some error occurred while purging restore objects")

	log.InfoD("Purging all restore source namespaces")
	err = WorkflowPDSRestore.Source.Purge()
	log.FailOnError(err, "some error occurred while purging restore source namespaces")

	log.InfoD("Purging all restore destination namespaces")
	err = WorkflowPDSRestore.Destination.Purge()
	log.FailOnError(err, "some error occurred while purging restore destination namespaces")

	log.InfoD("Purging all namespace objects")
	err = WorkflowNamespace.Purge()
	log.FailOnError(err, "some error occurred while purging namespace objects")
}
