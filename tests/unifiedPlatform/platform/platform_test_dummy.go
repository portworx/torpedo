package tests

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/platformLibs"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/stworkflows/pds"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/stworkflows/platform"
	"github.com/portworx/torpedo/drivers/utilities"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
	. "github.com/portworx/torpedo/tests/unifiedPlatform"
)

var _ = Describe("{CreateAndGeBackupLocation}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("CreateAndGeBackupLocation", "create backup locations", nil, 0)
	})

	It("CreateAndGeBackupLocation", func() {
		Step("create credentials and backup location", func() {
			var (
				workflowCc     platform.WorkflowCloudCredentials
				workflowbkpLoc platform.WorkflowBackupLocation
			)

			tenantId, err := platformLibs.GetDefaultTenantId(AccID)
			log.FailOnError(err, "error occured while fetching tenantID")

			workflowCc.Platform.TenantId = tenantId
			workflowCc.CloudCredentials = make(map[string]platform.CloudCredentialsType)
			cc, err := workflowCc.CreateCloudCredentials(NewPdsParams.BackUpAndRestore.TargetLocation)
			log.FailOnError(err, "error occured while creating cloud credentials")

			for _, value := range cc.CloudCredentials {
				log.Infof("cloud credentials name: [%s]", value.Name)
				log.Infof("cloud credentials id: [%s]", value.ID)
				log.Infof("cloud provider type: [%s]", value.CloudProviderType)
			}

			workflowbkpLoc.WfCloudCredentials = workflowCc

			wfbkpLoc, err := workflowbkpLoc.CreateBackupLocation(PDSBucketName, NewPdsParams.BackUpAndRestore.TargetLocation)
			log.FailOnError(err, "error while creating backup location")
			log.Infof("wfBkpLoc id: [%s]", wfbkpLoc.BkpLocation.BkpLocationId)
			log.Infof("wfBkpLoc name: [%s]", wfbkpLoc.BkpLocation.Name)

			// Listing backuplocation after creation
			bkpLocations, err := workflowbkpLoc.ListBackupLocation()
			log.FailOnError(err, "error while listing backup location")

			for _, bkpLocation := range bkpLocations {
				log.Infof("wfBkpLoc Name: [%s]", bkpLocation.BkpLocation.Name)
				log.Infof("wfBkpLoc Id: [%s]", bkpLocation.BkpLocation.BkpLocationId)
				for _, cred := range bkpLocation.WfCloudCredentials.CloudCredentials {
					log.Infof("credentials Id: [%s]", cred.ID)
				}
			}

			//Deleting the backuplocation ID
			log.Infof("Deleting backuplocation id [%s]", wfbkpLoc.BkpLocation.BkpLocationId)
			err = wfbkpLoc.DeleteBackupLocation(wfbkpLoc.BkpLocation.BkpLocationId)
			log.FailOnError(err, "error while deleting backup location")

			// Listing backup location after deletion
			bkpLocations, err = workflowbkpLoc.ListBackupLocation()
			log.FailOnError(err, "error while listing backup location")

			for _, bkpLocation := range bkpLocations {
				log.Infof("wfBkpLoc Name: [%s]", bkpLocation.BkpLocation.Name)
				log.Infof("wfBkpLoc Id: [%s]", bkpLocation.BkpLocation.BkpLocationId)
				for _, cred := range bkpLocation.WfCloudCredentials.CloudCredentials {
					log.Infof("credentials Id: [%s]", cred.ID)
				}
			}
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})

var _ = Describe("{CreateAndGetCloudCredentials}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("CreateCloudCredentials", "create cloud credentials", nil, 0)
	})

	It("CreateCloudCredentials", func() {
		Step("create and cloud credentials", func() {
			tenantId, err := platformLibs.GetDefaultTenantId(AccID)
			log.FailOnError(err, "error occured while fetching tenantID")
			credResp, err := platformLibs.CreateCloudCredentials(tenantId, NewPdsParams.BackUpAndRestore.TargetLocation)
			log.FailOnError(err, "error while creating cloud creds")
			log.Infof("creds resp [%+v]", credResp.Create.Config.S3Credentials.AccessKey)
			log.Infof("creds id [%+v]", *credResp.Create.Meta.Uid)

			isconfigRequiredTrue, err := platformLibs.GetCloudCredentials(*credResp.Create.Meta.Uid, NewPdsParams.BackUpAndRestore.TargetLocation, true)
			log.FailOnError(err, "error occured while getting cloud required with false flag")
			log.Debugf("Cred Name [%+v]", *isconfigRequiredTrue.Create.Meta.Name)
			log.Debugf("Cred Id [%+v]", *isconfigRequiredTrue.Create.Meta.Uid)
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})

//var _ = Describe("{RestoreCRUD}", func() {
//	var (
//		workflowRestore dslibs.WorkflowRestore
//	)
//	JustBeforeEach(func() {
//
//		workflowRestore = dslibs.WorkflowRestore{
//			DeploymentID: "SomeID",
//			NamepsaceID:  "SomeNamespace",
//			ProjectId:    "SomeID",
//		}
//		StartTorpedoTest("RestoreCRUD", "Runs CRUD on restores", nil, 0)
//	})
//
//	It("Create Restore", func() {
//		Step("Create Backup Config", func() {
//			_, err := workflowRestore.CreateRestore()
//			log.Infof("Error while creating restores - %s", err.Error())
//		})
//
//		Step("Recreate Restore", func() {
//			_, err := workflowRestore.ReCreateRestore()
//			log.Infof("Error while updating restores - %s", err.Error())
//		})
//
//		Step("Get Restore", func() {
//			_, err := workflowRestore.GetRestore()
//			log.Infof("Error while fetching restores - %s", err.Error())
//		})
//
//		Step("Delete Restore", func() {
//			_, err := workflowRestore.DeleteRestore()
//			log.Infof("Error while deleting restores - %s", err.Error())
//		})
//
//		Step("List Restore", func() {
//			_, err := workflowRestore.ListRestore()
//			log.Infof("Error while listing restores - %s", err.Error())
//		})
//	})
//
//	JustAfterEach(func() {
//		defer EndTorpedoTest()
//	})
//})

//
//var _ = Describe("{ListTenants}", func() {
//	JustBeforeEach(func() {
//		StartTorpedoTest("ListTenants", "List Tenants", nil, 0)
//	})
//
//	It("Tenants", func() {
//		Step("Create and List Accounts", func() {
//			workflowResponse, err := stworkflows.WorkflowListTenants(accID)
//			log.FailOnError(err, "Some error occurred while running WorkflowCreateAndListAccounts")
//			tenantList := workflowResponse[stworkflows.GetTenantListV1]
//			for _, tenant := range tenantList {
//				log.Infof("Available Tenant [%s] under account [%s]", *tenant.Meta.Name, accID)
//			}
//		})
//	})
//
//	JustAfterEach(func() {
//		defer EndTorpedoTest()
//	})
//})
//
//var _ = Describe("{CreateAccount}", func() {
//	JustBeforeEach(func() {
//		StartTorpedoTest("ListAccounts", "Create and List Accounts", nil, 0)
//	})
//
//	It("Accounts", func() {
//		Step("Create and List Accounts", func() {
//			workflowResponse, err := stworkflows.WorkflowCreateAndListAccounts()
//			log.FailOnError(err, "Some error occurred while running WorkflowCreateAndListAccounts")
//			accountList := workflowResponse[stworkflows.GetAccountListv1]
//			for _, account := range accountList {
//				log.Infof("Found %s as part of result", account.Meta.Name)
//			}
//		})
//	})
//
//	JustAfterEach(func() {
//		defer EndTorpedoTest()
//	})
//})

//var _ = Describe("{ListAccounts}", func() {
//	steplog := "List Accounts"
//	JustBeforeEach(func() {
//		StartTorpedoTest("ListAccounts", "Create and List Accounts", nil, 0)
//	})
//
//	log.InfoD(steplog)
//	It("Accounts", func() {
//		var accountID string
//		steplog = "ListAccounts"
//		Step(steplog, func() {
//			log.InfoD(steplog)
//			accList, err := platformLibs.GetAccountListv1()
//			log.FailOnError(err, "error while getting account list")
//			for _, acc := range accList {
//				if *acc.Meta.Name == defaultTestAccount {
//					log.Infof("Available account %s", *acc.Meta.Name)
//					log.Infof("Available account ID %s", *acc.Meta.Uid)
//					accountID = *acc.Meta.Uid
//				}
//			}
//		})
//		steplog = "GetAccounts"
//		Step(steplog, func() {
//			log.InfoD(steplog)
//			acc, err := platformLibs.GetAccount(accountID)
//			log.FailOnError(err, "error while getting account info")
//			log.Infof("account name is %s", *acc.Meta.Name)
//		})
//	})
//
//	JustAfterEach(func() {
//		defer EndTorpedoTest()
//	})
//})

var _ = Describe("{TestRbacForPds}", func() {
	var (
		pdsRbac         platform.WorkflowServiceAccount
		workflowProject platform.WorkflowProject
		saName          string
	)
	JustBeforeEach(func() {
		log.Infof("Initialising values for tenant")
		WorkflowPlatform.AdminAccountId = AccID
		WorkflowPlatform.TenantInit()
		saName = "pdsUserSA-" + RandomString(5)

		StartTorpedoTest("ListAccounts", "Create and List Accounts", nil, 0)
	})

	It("Accounts", func() {
		Step("Create a project for SVC Account", func() {
			workflowProject.Platform = WorkflowPlatform
			workflowProject.ProjectName = fmt.Sprintf("project-%s", utilities.RandomString(5))
			workflowProject, err := workflowProject.CreateProject()
			log.FailOnError(err, "Unable to create project")
			log.Infof("Project created with ID - [%s]", workflowProject.ProjectId)
			log.Infof("Tenant ID - [%s]", workflowProject.Platform.TenantId)
		})

		Step("Create SVC Account", func() {
			pdsRbac.WorkflowProject = workflowProject
			pdsRbac.UserRoles = make(map[string]platform.SeviceAccount)
			log.Infof("Tenant ID - [%s]", pdsRbac.WorkflowProject.Platform.TenantId)
			_, err := pdsRbac.CreateServiceAccount(saName, []string{platform.ProjectAdmin})
			log.FailOnError(err, "Unable to create service account")
			log.Infof("Service Account created")
		})

		Step("Switch to SVC Account", func() {
			pdsRbac.SwitchToServiceAccount(saName)
			log.Infof("Switched to service account")
			list, err := workflowProject.GetProjectList()
			if err != nil {
				log.Infof("error - [%s]", err.Error())
			} else {
				log.Infof("List of projects - [%+v]", list)
			}
		})

		Step("Switch to ADMIN", func() {
			pdsRbac.SwitchToAdmin()
			log.Infof("Switched to service account")
			list, err := workflowProject.GetProjectList()
			if err != nil {
				log.Infof("error - [%s]", err.Error())
			} else {
				log.Infof("List of projects - [%+v]", list)
			}
		})

	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})

var _ = Describe("{TestPlatformTemplates}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("TestPlatformTemplates", "create custom templates for PDS", nil, 0)
		Step("Get Default Tenant", func() {
			log.Infof("Initialising values for tenant")
			WorkflowPlatform.AdminAccountId = AccID
			WorkflowPlatform.TenantInit()
		})
	})

	var (
		workFlowTemplates pds.WorkflowPDSTemplates
		tempList          []string
	)
	It("TestPlatformTemplates", func() {
		Step("create custom templates for PDS", func() {
			workFlowTemplates.Platform = WorkflowPlatform
			for _, ds := range NewPdsParams.DataServiceToTest {
				serviceConfigId, stConfigId, resConfigId, err := workFlowTemplates.CreatePdsCustomTemplatesAndFetchIds(NewPdsParams, ds.Name)
				log.FailOnError(err, "Unable to create Custom Templates for PDS")
				log.InfoD("Created serviceConfig Template ID- [serviceConfigId- %v]", serviceConfigId)
				log.InfoD("Created stConfig Template ID- [stConfigId- %v]", stConfigId)
				log.InfoD("Created resConfig Template ID- [resConfigId- %v]", resConfigId)
				tempList = append(tempList, serviceConfigId, stConfigId, resConfigId)
			}
		})
		Step("Cleanup Created Templates after dissociating linked resources", func() {
			err := workFlowTemplates.DeleteCreatedCustomPdsTemplates(tempList)
			log.FailOnError(err, "Unable to delete Custom Templates for PDS")
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})