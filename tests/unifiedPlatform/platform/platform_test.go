package tests

import (
	. "github.com/onsi/ginkgo/v2"
	dslibs "github.com/portworx/torpedo/drivers/unifiedPlatform/pdsLibs"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/platformLibs"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/stworkflows"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
)

var tc platformLibs.TargetCluster

var _ = Describe("{TenantsCRUD}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("TenantsCRUD", "Create and Get the Tenant", pdsLabels, 0)
	})

	It("Tenants", func() {
		steplog := "Tenants CRUD"
		Step(steplog, func() {
			log.InfoD(steplog)
			var tc platformLibs.TargetCluster
			var tenantId string
			tenantList, err := platformLibs.GetTenantListV1(accID)
			log.FailOnError(err, "error while getting tenant list")
			for _, tenant := range tenantList {
				log.Infof("Available tenant's %s under the account id %s", *tenant.Meta.Name, accID)
				tenantId = *tenant.Meta.Uid
				break
			}
			log.Infof("TenantID [%s]", tenantId)
			clusterId, err := tc.RegisterToControlPlane("1.0.0", tenantId)
			if err != nil {
				log.FailOnError(err, "Failed to register Target Cluster to Control plane")
			}
			log.Infof("Registered Cluster ID is: %v\n", clusterId)
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})

var _ = Describe("{CreateCloudCredentials}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("CreateCloudCredentials", "create cloud credentials", nil, 0)
	})

	It("CreateCloudCredentials", func() {
		Step("create cloud credentials", func() {
			tenantId, err := platformLibs.GetDefaultTenantId(accID)
			log.FailOnError(err, "error occured while fetching tenantID")
			credResp, err := platformLibs.CreateCloudCredentials(tenantId, NewPdsParams.BackUpAndRestore.TargetLocation)
			log.FailOnError(err, "error while creating cloud creds")
			log.Infof("creds response [+%v]", credResp)
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})

var _ = Describe("{DeployDataServicesOnDemand}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("DeployDataService", "Deploy data services", nil, 0)
	})

	It("Deploy and Validate DataService", func() {
		for _, ds := range NewPdsParams.DataServiceToTest {
			_, err := stworkflows.DeployDataService(ds)
			log.FailOnError(err, "Error while deploying ds")
		}
	})

	It("Update DataService", func() {
		for _, ds := range NewPdsParams.DataServiceToTest {
			_, err := stworkflows.UpdateDataService(ds)
			log.FailOnError(err, "Error while updating ds")
		}
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})

var _ = Describe("{BackupConfigCRUD}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("BackupConfigCRUD", "Runs CRUD on backup config", nil, 0)
	})

	It("Create Backup Config", func() {
		Step("Create Backup Config", func() {
			_, err := dslibs.CreateBackupConfig(dslibs.WorkflowBackupInput{
				ProjectId:    "someprojectId",
				DeploymentID: "SomedeploymentId",
			})
			log.Infof("Error while creating backup config - %s", err.Error())
		})

		Step("Update Backup Config", func() {
			_, err := dslibs.UpdateBackupConfig(dslibs.WorkflowBackupInput{
				ProjectId:    "someprojectId2",
				DeploymentID: "SomedeploymentId2",
			})
			log.Infof("Error while updating backup config - %s", err.Error())
		})

		Step("Get Backup Config", func() {
			_, err := dslibs.GetBackupConfig(dslibs.WorkflowBackupInput{})
			log.Infof("Error while fetching backup config - %s", err.Error())
		})

		Step("Delete Backup Config", func() {
			_, err := dslibs.DeleteBackupConfig(dslibs.WorkflowBackupInput{})
			log.Infof("Error while deleting backup config - %s", err.Error())
		})

		Step("List Backup Config", func() {
			_, err := dslibs.ListBackupConfig(dslibs.WorkflowBackupInput{})
			log.Infof("Error while listing backup config - %s", err.Error())
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})

var _ = Describe("{RestoreCRUD}", func() {
	var (
		workflowRestore dslibs.WorkflowRestore
	)
	JustBeforeEach(func() {

		workflowRestore = dslibs.WorkflowRestore{
			DeploymentID: "SomeID",
			NamepsaceID:  "SomeNamespace",
			ProjectId:    "SomeID",
		}
		StartTorpedoTest("RestoreCRUD", "Runs CRUD on restores", nil, 0)
	})

	It("Create Restore", func() {
		Step("Create Backup Config", func() {
			_, err := workflowRestore.CreateRestore()
			log.Infof("Error while creating restores - %s", err.Error())
		})

		Step("Recreate Restore", func() {
			_, err := workflowRestore.ReCreateRestore()
			log.Infof("Error while updating restores - %s", err.Error())
		})

		Step("Get Restore", func() {
			_, err := workflowRestore.GetRestore()
			log.Infof("Error while fetching restores - %s", err.Error())
		})

		Step("Delete Restore", func() {
			_, err := workflowRestore.DeleteRestore()
			log.Infof("Error while deleting restores - %s", err.Error())
		})

		Step("List Restore", func() {
			_, err := workflowRestore.ListRestore()
			log.Infof("Error while listing restores - %s", err.Error())
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})

var _ = Describe("{BackupRD}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("BackupRD", "Runs RD operations on backup", nil, 0)
	})

	It("Create Backup Config", func() {

		Step("Get Backup Config", func() {
			_, err := dslibs.GetBackup(dslibs.WorkflowBackup{})
			log.Infof("Error while fetching backup - %s", err.Error())
		})

		Step("Delete Backup Config", func() {
			_, err := dslibs.DeleteBackup(dslibs.WorkflowBackup{})
			log.Infof("Error while deleting backup - %s", err.Error())
		})

		Step("List Backup Config", func() {
			_, err := dslibs.ListBackup(dslibs.WorkflowBackup{})
			log.Infof("Error while listing backup - %s", err.Error())
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})

var _ = Describe("{ListTenants}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("ListTenants", "List Tenants", nil, 0)
	})

	It("Tenants", func() {
		Step("Create and List Accounts", func() {
			workflowResponse, err := stworkflows.WorkflowListTenants(accID)
			log.FailOnError(err, "Some error occurred while running WorkflowCreateAndListAccounts")
			tenantList := workflowResponse[stworkflows.GetTenantListV1]
			for _, tenant := range tenantList {
				log.Infof("Available Tenant [%s] under account [%s]", *tenant.Meta.Name, accID)
			}
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})

var _ = Describe("{CreateAccount}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("ListAccounts", "Create and List Accounts", nil, 0)
	})

	It("Accounts", func() {
		Step("Create and List Accounts", func() {
			workflowResponse, err := stworkflows.WorkflowCreateAndListAccounts()
			log.FailOnError(err, "Some error occurred while running WorkflowCreateAndListAccounts")
			accountList := workflowResponse[stworkflows.GetAccountListv1]
			for _, account := range accountList {
				log.Infof("Found %s as part of result", account.Meta.Name)
			}
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})

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
