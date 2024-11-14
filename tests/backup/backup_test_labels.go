package tests

type TestCaseLabel = string
type TestCaseName = string

// Test case names
const (
	CreateMultipleUsersAndGroups                                                       TestCaseName = "CreateMultipleUsersAndGroups"
	DuplicateSharedBackup                                                              TestCaseName = "DuplicateSharedBackup"
	DifferentAccessSameUser                                                            TestCaseName = "DifferentAccessSameUser"
	ShareBackupWithUsersAndGroups                                                      TestCaseName = "ShareBackupWithUsersAndGroups"
	ShareLargeNumberOfBackupsWithLargeNumberOfUsers                                    TestCaseName = "ShareLargeNumberOfBackupsWithLargeNumberOfUsers"
	CancelClusterBackupShare                                                           TestCaseName = "CancelClusterBackupShare"
	ShareBackupAndEdit                                                                 TestCaseName = "ShareBackupAndEdit"
	SharedBackupDelete                                                                 TestCaseName = "SharedBackupDelete"
	ClusterBackupShareToggle                                                           TestCaseName = "ClusterBackupShareToggle"
	ShareBackupsAndClusterWithUser                                                     TestCaseName = "ShareBackupsAndClusterWithUser"
	ShareBackupWithDifferentRoleUsers                                                  TestCaseName = "ShareBackupWithDifferentRoleUsers"
	DeleteSharedBackup                                                                 TestCaseName = "DeleteSharedBackup"
	ShareAndRemoveBackupLocation                                                       TestCaseName = "ShareAndRemoveBackupLocation"
	ViewOnlyFullBackupRestoreIncrementalBackup                                         TestCaseName = "ViewOnlyFullBackupRestoreIncrementalBackup"
	IssueMultipleRestoresWithNamespaceAndStorageClassMapping                           TestCaseName = "IssueMultipleRestoresWithNamespaceAndStorageClassMapping"
	DeleteUsersRole                                                                    TestCaseName = "DeleteUsersRole"
	IssueMultipleDeletesForSharedBackup                                                TestCaseName = "IssueMultipleDeletesForSharedBackup"
	SwapShareBackup                                                                    TestCaseName = "SwapShareBackup"
	NamespaceLabelledBackupSharedWithDifferentAccessMode                               TestCaseName = "NamespaceLabelledBackupSharedWithDifferentAccessMode"
	BackupScheduleForOldAndNewNS                                                       TestCaseName = "BackupScheduleForOldAndNewNS"
	ManualAndScheduledBackupUsingNamespaceAndResourceLabel                             TestCaseName = "ManualAndScheduledBackupUsingNamespaceAndResourceLabel"
	ScheduleBackupWithAdditionAndRemovalOfNS                                           TestCaseName = "ScheduleBackupWithAdditionAndRemovalOfNS"
	ManualAndScheduleBackupUsingNSLabelWithMaxCharLimit                                TestCaseName = "ManualAndScheduleBackupUsingNSLabelWithMaxCharLimit"
	NamespaceLabelledBackupOfEmptyNamespace                                            TestCaseName = "NamespaceLabelledBackupOfEmptyNamespace"
	DeleteNfsExecutorPodWhileBackupAndRestoreInProgress                                TestCaseName = "DeleteNfsExecutorPodWhileBackupAndRestoreInProgress"
	SingleNamespaceBackupRestoreToNamespaceInSameAndDifferentProject                   TestCaseName = "SingleNamespaceBackupRestoreToNamespaceInSameAndDifferentProject"
	NamespaceMoveFromProjectToProjectToNoProjectWhileRestore                           TestCaseName = "NamespaceMoveFromProjectToProjectToNoProjectWhileRestore"
	MultipleProjectsAndNamespacesBackupAndRestore                                      TestCaseName = "MultipleProjectsAndNamespacesBackupAndRestore"
	BackupRestartPX                                                                    TestCaseName = "BackupRestartPX"
	KillStorkWithBackupsAndRestoresInProgress                                          TestCaseName = "KillStorkWithBackupsAndRestoresInProgress"
	RestartBackupPodDuringBackupSharing                                                TestCaseName = "RestartBackupPodDuringBackupSharing"
	CancelAllRunningBackupJobs                                                         TestCaseName = "CancelAllRunningBackupJobs"
	ScaleMongoDBWhileBackupAndRestore                                                  TestCaseName = "ScaleMongoDBWhileBackupAndRestore"
	RebootNodesWhenBackupsAreInProgress                                                TestCaseName = "RebootNodesWhenBackupsAreInProgress"
	ScaleDownPxBackupPodWhileBackupAndRestoreIsInProgress                              TestCaseName = "ScaleDownPxBackupPodWhileBackupAndRestoreIsInProgress"
	CancelAllRunningRestoreJobs                                                        TestCaseName = "CancelAllRunningRestoreJobs"
	DeleteSameNameObjectsByMultipleUsersFromAdmin                                      TestCaseName = "DeleteSameNameObjectsByMultipleUsersFromAdmin"
	DeleteUserBackupsAndRestoresOfDeletedAndInActiveClusterFromAdmin                   TestCaseName = "DeleteUserBackupsAndRestoresOfDeletedAndInActiveClusterFromAdmin"
	DeleteObjectsByMultipleUsersFromNewAdmin                                           TestCaseName = "DeleteObjectsByMultipleUsersFromNewAdmin"
	DeleteFailedInProgressBackupAndRestoreOfUserFromAdmin                              TestCaseName = "DeleteFailedInProgressBackupAndRestoreOfUserFromAdmin"
	DeleteSharedBackupOfUserFromAdmin                                                  TestCaseName = "DeleteSharedBackupOfUserFromAdmin"
	DeleteBackupOfUserNonSharedRBAC                                                    TestCaseName = "DeleteBackupOfUserNonSharedRBAC"
	DeleteBackupOfUserSharedRBAC                                                       TestCaseName = "DeleteBackupOfUserSharedRBAC"
	UpdatesBackupOfUserFromAdmin                                                       TestCaseName = "UpdatesBackupOfUserFromAdmin"
	DeleteBackupSharedByMultipleUsersFromAdmin                                         TestCaseName = "DeleteBackupSharedByMultipleUsersFromAdmin"
	NodeCountForLicensing                                                              TestCaseName = "NodeCountForLicensing"
	LicensingCountWithNodeLabelledBeforeClusterAddition                                TestCaseName = "LicensingCountWithNodeLabelledBeforeClusterAddition"
	LicensingCountBeforeAndAfterBackupPodRestart                                       TestCaseName = "LicensingCountBeforeAndAfterBackupPodRestart"
	BackupLocationWithEncryptionKey                                                    TestCaseName = "BackupLocationWithEncryptionKey"
	ReplicaChangeWhileRestore                                                          TestCaseName = "ReplicaChangeWhileRestore"
	ResizeOnRestoredVolume                                                             TestCaseName = "ResizeOnRestoredVolume"
	RestoreEncryptedAndNonEncryptedBackups                                             TestCaseName = "RestoreEncryptedAndNonEncryptedBackups"
	ResizeVolumeOnScheduleBackup                                                       TestCaseName = "ResizeVolumeOnScheduleBackup"
	BackupClusterVerification                                                          TestCaseName = "BackupClusterVerification"
	UserGroupManagement                                                                TestCaseName = "UserGroupManagement"
	BasicBackupCreation                                                                TestCaseName = "BasicBackupCreation"
	CreateBackupAndRestoreForAllCombinationsOfSSES3AndDenyPolicy                       TestCaseName = "CreateBackupAndRestoreForAllCombinationsOfSSES3AndDenyPolicy"
	BasicSelectiveRestore                                                              TestCaseName = "BasicSelectiveRestore"
	CustomResourceBackupAndRestore                                                     TestCaseName = "CustomResourceBackupAndRestore"
	DeleteAllBackupObjects                                                             TestCaseName = "DeleteAllBackupObjects"
	ScheduleBackupCreationAllNS                                                        TestCaseName = "ScheduleBackupCreationAllNS"
	CustomResourceRestore                                                              TestCaseName = "CustomResourceRestore"
	AllNSBackupWithIncludeNewNSOption                                                  TestCaseName = "AllNSBackupWithIncludeNewNSOption"
	BackupSyncBasicTest                                                                TestCaseName = "BackupSyncBasicTest"
	BackupMultipleNsWithSameLabel                                                      TestCaseName = "BackupMultipleNsWithSameLabel"
	MultipleCustomRestoreSameTimeDiffStorageClassMapping                               TestCaseName = "MultipleCustomRestoreSameTimeDiffStorageClassMapping"
	AddMultipleNamespaceLabels                                                         TestCaseName = "AddMultipleNamespaceLabels"
	MultipleInPlaceRestoreSameTime                                                     TestCaseName = "MultipleInPlaceRestoreSameTime"
	CloudSnapsSafeWhenBackupLocationDeleteTest                                         TestCaseName = "CloudSnapsSafeWhenBackupLocationDeleteTest"
	SetUnsetNSLabelDuringScheduleBackup                                                TestCaseName = "SetUnsetNSLabelDuringScheduleBackup"
	BackupRestoreOnDifferentK8sVersions                                                TestCaseName = "BackupRestoreOnDifferentK8sVersions"
	BackupCRsThenMultipleRestoresOnHigherK8sVersion                                    TestCaseName = "BackupCRsThenMultipleRestoresOnHigherK8sVersion"
	ScheduleBackupDeleteAndRecreateNS                                                  TestCaseName = "ScheduleBackupDeleteAndRecreateNS"
	DeleteNSDeleteClusterRestore                                                       TestCaseName = "DeleteNSDeleteClusterRestore"
	AlternateBackupBetweenNfsAndS3                                                     TestCaseName = "AlternateBackupBetweenNfsAndS3"
	BackupNamespaceInNfsRestoredFromS3                                                 TestCaseName = "BackupNamespaceInNfsRestoredFromS3"
	DeleteS3ScheduleAndCreateNfsSchedule                                               TestCaseName = "DeleteS3ScheduleAndCreateNfsSchedule"
	KubeAndPxNamespacesSkipOnAllNSBackup                                               TestCaseName = "KubeAndPxNamespacesSkipOnAllNSBackup"
	MultipleBackupLocationWithSameEndpoint                                             TestCaseName = "MultipleBackupLocationWithSameEndpoint"
	UpgradePxBackup                                                                    TestCaseName = "UpgradePxBackup"
	StorkUpgradeWithBackup                                                             TestCaseName = "StorkUpgradeWithBackup"
	PXBackupEndToEndBackupAndRestoreWithUpgrade                                        TestCaseName = "PXBackupEndToEndBackupAndRestoreWithUpgrade"
	IssueDeleteOfIncrementalBackupsAndRestore                                          TestCaseName = "IssueDeleteOfIncrementalBackupsAndRestore"
	DeleteIncrementalBackupsAndRecreateNew                                             TestCaseName = "DeleteIncrementalBackupsAndRecreateNew"
	DeleteBucketVerifyCloudBackupMissing                                               TestCaseName = "DeleteBucketVerifyCloudBackupMissing"
	DeleteBackupAndCheckIfBucketIsEmpty                                                TestCaseName = "DeleteBackupAndCheckIfBucketIsEmpty"
	KubevirtVMBackupRestoreWithDifferentStates                                         TestCaseName = "KubevirtVMBackupRestoreWithDifferentStates"
	KubevirtUpgradeTest                                                                TestCaseName = "KubevirtUpgradeTest"
	KubevirtVMBackupOrDeletionInProgress                                               TestCaseName = "KubevirtVMBackupOrDeletionInProgress"
	KubevirtVMBackupRestoreWithNodeSelector                                            TestCaseName = "KubevirtVMBackupRestoreWithNodeSelector"
	KubevirtVMWithFreezeUnfreeze                                                       TestCaseName = "KubevirtVMWithFreezeUnfreeze"
	KubevirtInPlaceRestoreWithReplaceAndRetain                                         TestCaseName = "KubevirtInPlaceRestoreWithReplaceAndRetain"
	KubevirtVMRestoreWithAfterChangingVMConfig                                         TestCaseName = "KubevirtVMRestoreWithAfterChangingVMConfig"
	BackupAlternatingBetweenLockedAndUnlockedBuckets                                   TestCaseName = "BackupAlternatingBetweenLockedAndUnlockedBuckets"
	LockedBucketResizeOnRestoredVolume                                                 TestCaseName = "LockedBucketResizeOnRestoredVolume"
	LockedBucketResizeVolumeOnScheduleBackup                                           TestCaseName = "LockedBucketResizeVolumeOnScheduleBackup"
	DeleteLockedBucketUserObjectsFromAdmin                                             TestCaseName = "DeleteLockedBucketUserObjectsFromAdmin"
	VerifyRBACForInfraAdmin                                                            TestCaseName = "VerifyRBACForInfraAdmin"
	VerifyRBACForPxAdmin                                                               TestCaseName = "VerifyRBACForPxAdmin"
	VerifyRBACForAppAdmin                                                              TestCaseName = "VerifyRBACForAppAdmin"
	VerifyRBACForAppUser                                                               TestCaseName = "VerifyRBACForAppUser"
	DefaultBackupRestoreWithKubevirtAndNonKubevirtNS                                   TestCaseName = "DefaultBackupRestoreWithKubevirtAndNonKubevirtNS"
	KubevirtScheduledVMDelete                                                          TestCaseName = "KubevirtScheduledVMDelete"
	CustomBackupRestoreWithKubevirtAndNonKubevirtNS                                    TestCaseName = "CustomBackupRestoreWithKubevirtAndNonKubevirtNS"
	BackupAndRestoreSyncDR                                                             TestCaseName = "BackupAndRestoreSyncDR"
	ExcludeInvalidDirectoryFileBackup                                                  TestCaseName = "ExcludeInvalidDirectoryFileBackup"
	ExcludeDirectoryFileBackup                                                         TestCaseName = "ExcludeDirectoryFileBackup"
	MultipleProvisionerCsiSnapshotDeleteBackupAndRestore                               TestCaseName = "MultipleProvisionerCsiSnapshotDeleteBackupAndRestore"
	MultipleMemberProjectBackupAndRestoreForSingleNamespace                            TestCaseName = "MultipleMemberProjectBackupAndRestoreForSingleNamespace"
	BackupNetworkErrorTest                                                             TestCaseName = "BackupNetworkErrorTest"
	IssueMultipleBackupsAndRestoreInterleavedCopies                                    TestCaseName = "IssueMultipleBackupsAndRestoreInterleavedCopies"
	ValidateFiftyVolumeBackups                                                         TestCaseName = "ValidateFiftyVolumeBackups"
	BackupAndRestoreWithNonExistingAdminNamespaceAndUpdatedResumeSuspendBackupPolicies TestCaseName = "BackupAndRestoreWithNonExistingAdminNamespaceAndUpdatedResumeSuspendBackupPolicies"
	PXBackupClusterUpgradeTest                                                         TestCaseName = "PXBackupClusterUpgradeTest"
	BackupToLockedBucketWithSharedObjects                                              TestCaseName = "BackupToLockedBucketWithSharedObjects"
	RemoveJSONFilesFromNFSBackupLocation                                               TestCaseName = "RemoveJSONFilesFromNFSBackupLocation"
	CloudSnapshotMissingValidationForNFSLocation                                       TestCaseName = "CloudSnapshotMissingValidationForNFSLocation"
	MultipleProvisionerCsiKdmpBackupAndRestore                                         TestCaseName = "MultipleProvisionerCsiKdmpBackupAndRestore"
	KubevirtVMMigrationTest                                                            TestCaseName = "KubevirtVMMigrationTest"
	EnableNsAndClusterLevelPSAWithBackupAndRestore                                     TestCaseName = "EnableNsAndClusterLevelPSAWithBackupAndRestore"
	DummyPSATestcase                                                                   TestCaseName = "DummyPSATestcase"
	BackupCSIVolumesWithPartialSuccess                                                 TestCaseName = "BackupCSIVolumesWithPartialSuccess"
	RestoreFromHigherPrivilegedNamespaceToLower                                        TestCaseName = "RestoreFromHigherPrivilegedNamespaceToLower"
	BackupStateTransitionForScheduledBackups                                           TestCaseName = "BackupStateTransitionForScheduledBackups"
	PartialBackupSuccessWithPxVolumes                                                  TestCaseName = "PartialBackupSuccessWithPxVolumes"
	PartialBackupSuccessWithPxAndKDMPVolumes                                           TestCaseName = "PartialBackupSuccessWithPxAndKDMPVolumes"
	PartialBackupWithLowerStorkVersion                                                 TestCaseName = "PartialBackupWithLowerStorkVersion"
	PartialBackupSuccessWithAzureEndpoint                                              TestCaseName = "PartialBackupSuccessWithAzureEndpoint"
	PSALowerPrivilegeToHigherPrivilegeWithProjectMapping                               TestCaseName = "PSALowerPrivilegeToHigherPrivilegeWithProjectMapping"
	AzureCloudAccountCreationWithMandatoryAndNonMandatoryFields                        TestCaseName = "AzureCloudAccountCreationWithMandatoryAndNonMandatoryFields"
	AzureCloudAccountForLockedBucket                                                   TestCaseName = "AzureCloudAccountForLockedBucket"
	PXBackupUpgradeWithAzureCredChange                                                 TestCaseName = "PXBackupUpgradeWithAzureCredChange"
	ClusterShare                                                                       TestCaseName = "ClusterShare"
	VerifyBackupDeletionWhenRetentionIsMet                                             TestCaseName = "VerifyBackupDeletionWhenRetentionIsMet"
	DeleteVerifyBackupDeletionWhenRetentionIsMet                                       TestCaseName = "DeleteVerifyBackupDeletionWhenRetentionIsMet"
	VerifyBackupAutoDeletionWhenNewPVCsAreAddedBetweenSchedules                        TestCaseName = "VerifyBackupAutoDeletionWhenNewPVCsAreAddedBetweenSchedules"
	DeleteVerifyBackupAutoDeletionWhenNewPVCsAreAddedBetweenSchedules                  TestCaseName = "DeleteVerifyBackupAutoDeletionWhenNewPVCsAreAddedBetweenSchedules"
	PsaTakeBackupInLowerPrivilegeRestoreInHigherPrivilege                              TestCaseName = "PsaTakeBackupInLowerPrivilegeRestoreInHigherPrivilege"
	SuperAdmin                                                                         TestCaseName = "SuperAdmin"
	SuperAdminBackupShare                                                              TestCaseName = "SuperAdminBackupShare"
	TimeTakenToDeleteBackupWithHeavyLoad                                               TestCaseName = "TimeTakenToDeleteBackupWithHeavyLoad"
	BackupShare                                                                        TestCaseName = "BackupShare"
	TimeTakenToDeleteScheduleBackupWithHeavyLoadAfterSuspendingTheSchedule             TestCaseName = "TimeTakenToDeleteScheduleBackupWithHeavyLoadAfterSuspendingTheSchedule"
	BackupSuperAdminRoleForLocalUser                                                   TestCaseName = "BackupSuperAdminRoleForLocalUser"
	ClusterShareWithLargeNumberOfUsersAndClusters                                      TestCaseName = "ClusterShareWithLargeNumberOfUsersAndClusters"
	ValidateClusterShareWhileBringDownPxBackupPods                                     TestCaseName = "ValidateClusterShareWhileBringDownPxBackupPods"
	ValidateBackupShareUsingBackupsFromSharedCluster                                   TestCaseName = "ValidateBackupShareUsingBackupsFromSharedCluster"
	ValidateClusterShareWithConcurrentBackupOperations                                 TestCaseName = "ValidateClusterShareWithConcurrentBackupOperations"
	ValidateUserAccessLevel                                                            TestCaseName = "ValidateUserAccessLevel"
	SoftDeleteAndRecoverBackupOnContainerAndBlobLevel                                  TestCaseName = "SoftDeleteAndRecoverBackupOnContainerAndBlobLevel"
	DeleteSoftDeleteAndRecoverBackupOnContainerAndBlobLevel                            TestCaseName = "DeleteSoftDeleteAndRecoverBackupOnContainerAndBlobLevel"
	ValidateMetrics                                                                    TestCaseName = "ValidateMetrics"
	KDMPBackup                                                                         TestCaseName = "ValidateGenericBackupDeletionWithMissingS3Bucket"
	StorkControllerConfigCM                                                            TestCaseName = "StorkControllerConfigCM"
	BackupDeletionWithDynamicPVCGeneration                                             TestCaseName = "BackupDeletionWithDynamicPVCGeneration"
)

// Test case labels
const (
	CreateMultipleUsersAndGroupsLabel                                                       TestCaseLabel = "CreateMultipleUsersAndGroups"
	DuplicateSharedBackupLabel                                                              TestCaseLabel = "DuplicateSharedBackup"
	DifferentAccessSameUserLabel                                                            TestCaseLabel = "DifferentAccessSameUser"
	ShareBackupWithUsersAndGroupsLabel                                                      TestCaseLabel = "ShareBackupWithUsersAndGroups"
	ShareLargeNumberOfBackupsWithLargeNumberOfUsersLabel                                    TestCaseLabel = "ShareLargeNumberOfBackupsWithLargeNumberOfUsers"
	CancelClusterBackupShareLabel                                                           TestCaseLabel = "CancelClusterBackupShare"
	ShareBackupAndEditLabel                                                                 TestCaseLabel = "ShareBackupAndEdit"
	SharedBackupDeleteLabel                                                                 TestCaseLabel = "SharedBackupDelete"
	ClusterBackupShareToggleLabel                                                           TestCaseLabel = "ClusterBackupShareToggle"
	ShareBackupsAndClusterWithUserLabel                                                     TestCaseLabel = "ShareBackupsAndClusterWithUser"
	ShareBackupWithDifferentRoleUsersLabel                                                  TestCaseLabel = "ShareBackupWithDifferentRoleUsers"
	DeleteSharedBackupLabel                                                                 TestCaseLabel = "DeleteSharedBackup"
	ShareAndRemoveBackupLocationLabel                                                       TestCaseLabel = "ShareAndRemoveBackupLocation"
	ViewOnlyFullBackupRestoreIncrementalBackupLabel                                         TestCaseLabel = "ViewOnlyFullBackupRestoreIncrementalBackup"
	IssueMultipleRestoresWithNamespaceAndStorageClassMappingLabel                           TestCaseLabel = "IssueMultipleRestoresWithNamespaceAndStorageClassMapping"
	DeleteUsersRoleLabel                                                                    TestCaseLabel = "DeleteUsersRole"
	IssueMultipleDeletesForSharedBackupLabel                                                TestCaseLabel = "IssueMultipleDeletesForSharedBackup"
	SwapShareBackupLabel                                                                    TestCaseLabel = "SwapShareBackup"
	NamespaceLabelledBackupSharedWithDifferentAccessModeLabel                               TestCaseLabel = "NamespaceLabelledBackupSharedWithDifferentAccessMode"
	BackupScheduleForOldAndNewNSLabel                                                       TestCaseLabel = "BackupScheduleForOldAndNewNS"
	ManualAndScheduledBackupUsingNamespaceAndResourceLabelLabel                             TestCaseLabel = "ManualAndScheduledBackupUsingNamespaceAndResourceLabel"
	ScheduleBackupWithAdditionAndRemovalOfNSLabel                                           TestCaseLabel = "ScheduleBackupWithAdditionAndRemovalOfNS"
	ManualAndScheduleBackupUsingNSLabelWithMaxCharLimitLabel                                TestCaseLabel = "ManualAndScheduleBackupUsingNSLabelWithMaxCharLimit"
	NamespaceLabelledBackupOfEmptyNamespaceLabel                                            TestCaseLabel = "NamespaceLabelledBackupOfEmptyNamespace"
	DeleteNfsExecutorPodWhileBackupAndRestoreInProgressLabel                                TestCaseLabel = "DeleteNfsExecutorPodWhileBackupAndRestoreInProgress"
	SingleNamespaceBackupRestoreToNamespaceInSameAndDifferentProjectLabel                   TestCaseLabel = "SingleNamespaceBackupRestoreToNamespaceInSameAndDifferentProject"
	NamespaceMoveFromProjectToProjectToNoProjectWhileRestoreLabel                           TestCaseLabel = "NamespaceMoveFromProjectToProjectToNoProjectWhileRestore"
	MultipleProjectsAndNamespacesBackupAndRestoreLabel                                      TestCaseLabel = "MultipleProjectsAndNamespacesBackupAndRestore"
	BackupRestartPXLabel                                                                    TestCaseLabel = "BackupRestartPX"
	KillStorkWithBackupsAndRestoresInProgressLabel                                          TestCaseLabel = "KillStorkWithBackupsAndRestoresInProgress"
	RestartBackupPodDuringBackupSharingLabel                                                TestCaseLabel = "RestartBackupPodDuringBackupSharing"
	CancelAllRunningBackupJobsLabel                                                         TestCaseLabel = "CancelAllRunningBackupJobs"
	ScaleMongoDBWhileBackupAndRestoreLabel                                                  TestCaseLabel = "ScaleMongoDBWhileBackupAndRestore"
	RebootNodesWhenBackupsAreInProgressLabel                                                TestCaseLabel = "RebootNodesWhenBackupsAreInProgress"
	ScaleDownPxBackupPodWhileBackupAndRestoreIsInProgressLabel                              TestCaseLabel = "ScaleDownPxBackupPodWhileBackupAndRestoreIsInProgress"
	CancelAllRunningRestoreJobsLabel                                                        TestCaseLabel = "CancelAllRunningRestoreJobs"
	DeleteSameNameObjectsByMultipleUsersFromAdminLabel                                      TestCaseLabel = "DeleteSameNameObjectsByMultipleUsersFromAdmin"
	DeleteUserBackupsAndRestoresOfDeletedAndInActiveClusterFromAdminLabel                   TestCaseLabel = "DeleteUserBackupsAndRestoresOfDeletedAndInActiveClusterFromAdmin"
	DeleteObjectsByMultipleUsersFromNewAdminLabel                                           TestCaseLabel = "DeleteObjectsByMultipleUsersFromNewAdmin"
	DeleteFailedInProgressBackupAndRestoreOfUserFromAdminLabel                              TestCaseLabel = "DeleteFailedInProgressBackupAndRestoreOfUserFromAdmin"
	DeleteSharedBackupOfUserFromAdminLabel                                                  TestCaseLabel = "DeleteSharedBackupOfUserFromAdmin"
	DeleteBackupOfUserNonSharedRBACLabel                                                    TestCaseLabel = "DeleteBackupOfUserNonSharedRBAC"
	DeleteBackupOfUserSharedRBACLabel                                                       TestCaseLabel = "DeleteBackupOfUserSharedRBAC"
	UpdatesBackupOfUserFromAdminLabel                                                       TestCaseLabel = "UpdatesBackupOfUserFromAdmin"
	DeleteBackupSharedByMultipleUsersFromAdminLabel                                         TestCaseLabel = "DeleteBackupSharedByMultipleUsersFromAdmin"
	NodeCountForLicensingLabel                                                              TestCaseLabel = "NodeCountForLicensing"
	LicensingCountWithNodeLabelledBeforeClusterAdditionLabel                                TestCaseLabel = "LicensingCountWithNodeLabelledBeforeClusterAddition"
	LicensingCountBeforeAndAfterBackupPodRestartLabel                                       TestCaseLabel = "LicensingCountBeforeAndAfterBackupPodRestart"
	BackupLocationWithEncryptionKeyLabel                                                    TestCaseLabel = "BackupLocationWithEncryptionKey"
	ReplicaChangeWhileRestoreLabel                                                          TestCaseLabel = "ReplicaChangeWhileRestore"
	ResizeOnRestoredVolumeLabel                                                             TestCaseLabel = "ResizeOnRestoredVolume"
	RestoreEncryptedAndNonEncryptedBackupsLabel                                             TestCaseLabel = "RestoreEncryptedAndNonEncryptedBackups"
	ResizeVolumeOnScheduleBackupLabel                                                       TestCaseLabel = "ResizeVolumeOnScheduleBackup"
	BackupClusterVerificationLabel                                                          TestCaseLabel = "BackupClusterVerification"
	UserGroupManagementLabel                                                                TestCaseLabel = "UserGroupManagement"
	BasicBackupCreationLabel                                                                TestCaseLabel = "BasicBackupCreation"
	CreateBackupAndRestoreForAllCombinationsOfSSES3AndDenyPolicyLabel                       TestCaseLabel = "CreateBackupAndRestoreForAllCombinationsOfSSES3AndDenyPolicy"
	BasicSelectiveRestoreLabel                                                              TestCaseLabel = "BasicSelectiveRestore"
	CustomResourceBackupAndRestoreLabel                                                     TestCaseLabel = "CustomResourceBackupAndRestore"
	DeleteAllBackupObjectsLabel                                                             TestCaseLabel = "DeleteAllBackupObjects"
	ScheduleBackupCreationAllNSLabel                                                        TestCaseLabel = "ScheduleBackupCreationAllNS"
	CustomResourceRestoreLabel                                                              TestCaseLabel = "CustomResourceRestore"
	AllNSBackupWithIncludeNewNSOptionLabel                                                  TestCaseLabel = "AllNSBackupWithIncludeNewNSOption"
	BackupSyncBasicTestLabel                                                                TestCaseLabel = "BackupSyncBasicTest"
	BackupMultipleNsWithSameLabelLabel                                                      TestCaseLabel = "BackupMultipleNsWithSameLabel"
	MultipleCustomRestoreSameTimeDiffStorageClassMappingLabel                               TestCaseLabel = "MultipleCustomRestoreSameTimeDiffStorageClassMapping"
	AddMultipleNamespaceLabelsLabel                                                         TestCaseLabel = "AddMultipleNamespaceLabels"
	MultipleInPlaceRestoreSameTimeLabel                                                     TestCaseLabel = "MultipleInPlaceRestoreSameTime"
	CloudSnapsSafeWhenBackupLocationDeleteTestLabel                                         TestCaseLabel = "CloudSnapsSafeWhenBackupLocationDeleteTest"
	SetUnsetNSLabelDuringScheduleBackupLabel                                                TestCaseLabel = "SetUnsetNSLabelDuringScheduleBackup"
	BackupRestoreOnDifferentK8sVersionsLabel                                                TestCaseLabel = "BackupRestoreOnDifferentK8sVersions"
	BackupCRsThenMultipleRestoresOnHigherK8sVersionLabel                                    TestCaseLabel = "BackupCRsThenMultipleRestoresOnHigherK8sVersion"
	ScheduleBackupDeleteAndRecreateNSLabel                                                  TestCaseLabel = "ScheduleBackupDeleteAndRecreateNS"
	DeleteNSDeleteClusterRestoreLabel                                                       TestCaseLabel = "DeleteNSDeleteClusterRestore"
	AlternateBackupBetweenNfsAndS3Label                                                     TestCaseLabel = "AlternateBackupBetweenNfsAndS3"
	BackupNamespaceInNfsRestoredFromS3Label                                                 TestCaseLabel = "BackupNamespaceInNfsRestoredFromS3"
	DeleteS3ScheduleAndCreateNfsScheduleLabel                                               TestCaseLabel = "DeleteS3ScheduleAndCreateNfsSchedule"
	KubeAndPxNamespacesSkipOnAllNSBackupLabel                                               TestCaseLabel = "KubeAndPxNamespacesSkipOnAllNSBackup"
	MultipleBackupLocationWithSameEndpointLabel                                             TestCaseLabel = "MultipleBackupLocationWithSameEndpoint"
	UpgradePxBackupLabel                                                                    TestCaseLabel = "UpgradePxBackupWithHelm"
	StorkUpgradeWithBackupLabel                                                             TestCaseLabel = "StorkUpgradeWithBackup"
	PXBackupEndToEndBackupAndRestoreWithUpgradeLabel                                        TestCaseLabel = "PXBackupEndToEndBackupAndRestoreWithUpgrade"
	IssueDeleteOfIncrementalBackupsAndRestoreLabel                                          TestCaseLabel = "IssueDeleteOfIncrementalBackupsAndRestore"
	DeleteIncrementalBackupsAndRecreateNewLabel                                             TestCaseLabel = "DeleteIncrementalBackupsAndRecreateNew"
	DeleteBucketVerifyCloudBackupMissingLabel                                               TestCaseLabel = "DeleteBucketVerifyCloudBackupMissing"
	DeleteBackupAndCheckIfBucketIsEmptyLabel                                                TestCaseLabel = "DeleteBackupAndCheckIfBucketIsEmpty"
	KubevirtVMBackupRestoreWithDifferentStatesLabel                                         TestCaseLabel = "KubevirtVMBackupRestoreWithDifferentStates"
	BackupAlternatingBetweenLockedAndUnlockedBucketsLabel                                   TestCaseLabel = "BackupAlternatingBetweenLockedAndUnlockedBuckets"
	LockedBucketResizeOnRestoredVolumeLabel                                                 TestCaseLabel = "LockedBucketResizeOnRestoredVolume"
	LockedBucketResizeVolumeOnScheduleBackupLabel                                           TestCaseLabel = "LockedBucketResizeVolumeOnScheduleBackup"
	DeleteLockedBucketUserObjectsFromAdminLabel                                             TestCaseLabel = "DeleteLockedBucketUserObjectsFromAdmin"
	VerifyRBACForInfraAdminLabel                                                            TestCaseLabel = "VerifyRBACForInfraAdmin"
	VerifyRBACForPxAdminLabel                                                               TestCaseLabel = "VerifyRBACForPxAdmin"
	VerifyRBACForAppAdminLabel                                                              TestCaseLabel = "VerifyRBACForAppAdmin"
	VerifyRBACForAppUserLabel                                                               TestCaseLabel = "VerifyRBACForAppUser"
	KubevirtUpgradeTestLabel                                                                TestCaseLabel = "KubevirtUpgradeTest"
	KubevirtVMBackupOrDeletionInProgressLabel                                               TestCaseLabel = "KubevirtVMBackupOrDeletionInProgress"
	KubevirtVMBackupRestoreWithNodeSelectorLabel                                            TestCaseLabel = "KubevirtVMBackupRestoreWithNodeSelector"
	KubevirtVMWithFreezeUnfreezeLabel                                                       TestCaseLabel = "KubevirtVMWithFreezeUnfreeze"
	KubevirtInPlaceRestoreWithReplaceAndRetainLabel                                         TestCaseLabel = "KubevirtInPlaceRestoreWithReplaceAndRetain"
	KubevirtVMRestoreWithAfterChangingVMConfigLabel                                         TestCaseLabel = "KubevirtVMRestoreWithAfterChangingVMConfig"
	DefaultBackupRestoreWithKubevirtAndNonKubevirtNSLabel                                   TestCaseLabel = "DefaultBackupRestoreWithKubevirtAndNonKubevirtNS"
	KubevirtScheduledVMDeleteLabel                                                          TestCaseLabel = "KubevirtScheduledVMDelete"
	CustomBackupRestoreWithKubevirtAndNonKubevirtNSLabel                                    TestCaseLabel = "CustomBackupRestoreWithKubevirtAndNonKubevirtNS"
	BackupAndRestoreSyncDRLabel                                                             TestCaseLabel = "BackupAndRestoreSyncDR"
	ExcludeInvalidDirectoryFileBackupLabel                                                  TestCaseLabel = "ExcludeInvalidDirectoryFileBackup"
	ExcludeDirectoryFileBackupLabel                                                         TestCaseLabel = "ExcludeDirectoryFileBackup"
	MultipleProvisionerCsiSnapshotDeleteBackupAndRestoreLabel                               TestCaseLabel = "MultipleProvisionerCsiSnapshotDeleteBackupAndRestore"
	MultipleMemberProjectBackupAndRestoreForSingleNamespaceLabel                            TestCaseLabel = "MultipleMemberProjectBackupAndRestoreForSingleNamespace"
	BackupNetworkErrorTestLabel                                                             TestCaseLabel = "BackupNetworkErrorTest"
	IssueMultipleBackupsAndRestoreInterleavedCopiesLabel                                    TestCaseLabel = "IssueMultipleBackupsAndRestoreInterleavedCopies"
	ValidateFiftyVolumeBackupsLabel                                                         TestCaseLabel = "ValidateFiftyVolumeBackups"
	BackupAndRestoreWithNonExistingAdminNamespaceAndUpdatedResumeSuspendBackupPoliciesLabel TestCaseLabel = "BackupAndRestoreWithNonExistingAdminNamespaceAndUpdatedResumeSuspendBackupPolicies"
	PXBackupClusterUpgradeTestLabel                                                         TestCaseLabel = "PXBackupClusterUpgradeTest"
	BackupToLockedBucketWithSharedObjectsLabel                                              TestCaseLabel = "BackupToLockedBucketWithSharedObjects"
	RemoveJSONFilesFromNFSBackupLocationLabel                                               TestCaseLabel = "RemoveJSONFilesFromNFSBackupLocation"
	CloudSnapshotMissingValidationForNFSLocationLabel                                       TestCaseLabel = "CloudSnapshotMissingValidationForNFSLocation"
	MultipleProvisionerCsiKdmpBackupAndRestoreLabel                                         TestCaseLabel = "MultipleProvisionerCsiKdmpBackupAndRestore"
	KubevirtVMMigrationTestLabel                                                            TestCaseLabel = "KubevirtVMMigrationTest"
	BackupCSIVolumesWithPartialSuccessLabel                                                 TestCaseLabel = "BackupCSIVolumesWithPartialSuccess"
	BackupStateTransitionForScheduledBackupsLabel                                           TestCaseLabel = "BackupStateTransitionForScheduledBackups"
	EnableNsAndClusterLevelPSAWithBackupAndRestoreLabel                                     TestCaseLabel = "EnableNsAndClusterLevelPSAWithBackupAndRestore"
	RestoreFromHigherPrivilegedNamespaceToLowerLabel                                        TestCaseLabel = "RestoreFromHigherPrivilegedNamespaceToLower"
	PartialBackupSuccessWithPxVolumesLabel                                                  TestCaseLabel = "PartialBackupSuccessWithPxVolumes"
	PartialBackupSuccessWithPxAndKDMPVolumesLabel                                           TestCaseLabel = "PartialBackupSuccessWithPxAndKDMPVolumes"
	PartialBackupWithLowerStorkVersionLabel                                                 TestCaseLabel = "PartialBackupWithLowerStorkVersion"
	PartialBackupSuccessWithAzureEndpointLabel                                              TestCaseLabel = "PartialBackupSuccessWithAzureEndpoint"
	PSALowerPrivilegeToHigherPrivilegeWithProjectMappingLabel                               TestCaseLabel = "PSALowerPrivilegeToHigherPrivilegeWithProjectMapping"
	AzureCloudAccountCreationWithMandatoryAndNonMandatoryFieldsLabel                        TestCaseLabel = "AzureCloudAccountCreationWithMandatoryAndNonMandatoryFields"
	AzureCloudAccountForLockedBucketLabel                                                   TestCaseLabel = "AzureCloudAccountForLockedBucket"
	PXBackupUpgradeWithAzureCredChangeLabel                                                 TestCaseLabel = "PXBackupUpgradeWithAzureCredChange"
	ClusterShareLabel                                                                       TestCaseLabel = "ClusterShare"
	VerifyBackupDeletionWhenRetentionIsMetLabel                                             TestCaseLabel = "VerifyBackupDeletionWhenRetentionIsMet"
	DeleteVerifyBackupDeletionWhenRetentionIsMetLabel                                       TestCaseLabel = "DeleteVerifyBackupDeletionWhenRetentionIsMet"
	VerifyBackupAutoDeletionWhenNewPVCsAreAddedBetweenSchedulesLabel                        TestCaseLabel = "VerifyBackupAutoDeletionWhenNewPVCsAreAddedBetweenSchedules"
	DeleteVerifyBackupAutoDeletionWhenNewPVCsAreAddedBetweenSchedulesLabel                  TestCaseLabel = "DeleteVerifyBackupAutoDeletionWhenNewPVCsAreAddedBetweenSchedules"
	PsaTakeBackupInLowerPrivilegeRestoreInHigherPrivilegeLabel                              TestCaseLabel = "PsaTakeBackupInLowerPrivilegeRestoreInHigherPrivilege"
	SuperAdminLabel                                                                         TestCaseLabel = "SuperAdmin"
	SuperAdminBackupShareLabel                                                              TestCaseLabel = "SuperAdminBackupShare"
	TimeTakenToDeleteBackupWithHeavyLoadLabel                                               TestCaseLabel = "TimeTakenToDeleteBackupWithHeavyLoad"
	TimeTakenToDeleteScheduleBackupWithHeavyLoadAfterSuspendingTheScheduleLabel             TestCaseLabel = "TimeTakenToDeleteScheduleBackupWithHeavyLoadAfterSuspendingTheSchedule"
	BackupSuperAdminRoleForLocalUserLabel                                                   TestCaseLabel = "BackupSuperAdminRoleForLocalUser"
	ClusterShareWithLargeNumberOfUsersAndClustersLabel                                      TestCaseLabel = "ClusterShareWithLargeNumberOfUsersAndClusters"
	ValidateClusterShareWhileBringDownPxBackupPodsLabel                                     TestCaseLabel = "ValidateClusterShareWhileBringDownPxBackupPods"
	ValidateBackupShareUsingBackupsFromSharedClusterLabel                                   TestCaseLabel = "ValidateBackupShareUsingBackupsFromSharedCluster"
	ValidateClusterShareWithConcurrentBackupOperationsLabel                                 TestCaseLabel = "ValidateClusterShareWithConcurrentBackupOperations"
	ValidateUserAccessLevelLabel                                                            TestCaseLabel = "ValidateUserAccessLevel"
	BackupShareLabel                                                                        TestCaseLabel = "BackupShare"
	SoftDeleteAndRecoverBackupOnContainerAndBlobLevelLabel                                  TestCaseLabel = "SoftDeleteAndRecoverBackupOnContainerAndBlobLevel"
	DeleteSoftDeleteAndRecoverBackupOnContainerAndBlobLevelLabel                            TestCaseLabel = "DeleteSoftDeleteAndRecoverBackupOnContainerAndBlobLevel"
	ValidateMetricsLabel                                                                    TestCaseLabel = "ValidateMetrics"
	KDMPBackupLabel                                                                         TestCaseLabel = "ValidateGenericBackupDeletionWithMissingS3Bucket"
	StorkControllerConfigCMLabel                                                            TestCaseLabel = "StorkControllerConfigCM"
	BackupDeletionWithDynamicPVCGenerationLabel                                             TestCaseLabel = "BackupDeletionWithDynamicPVCGeneration"
)

// Common Labels
const (
	PxBackupLabel TestCaseLabel = "px-backup"
)

// Priority labels
const (
	P0 TestCaseLabel = "p0"
	P1 TestCaseLabel = "p1"
	P2 TestCaseLabel = "p2"
)

// Test type labels
const (
	SystemTest      TestCaseLabel = "system-test"
	ScaleTest       TestCaseLabel = "scale-test"
	LongevityTest   TestCaseLabel = "longevity-test"
	PerformanceTest TestCaseLabel = "performance-test"
	FunctionalTest  TestCaseLabel = "functional-test"
)

// PipelineLabels
const (
	ocpPipeline     TestCaseLabel = "ocp-pipeline"
	vanillaPipeline TestCaseLabel = "vanilla-pipeline"
	rkePipeline     TestCaseLabel = "rke-pipeline"
	aksPipeline     TestCaseLabel = "aks-pipeline"
	eksPipeline     TestCaseLabel = "eks-pipeline"
	gkePipeline     TestCaseLabel = "gke-pipeline"
	iksPipeline     TestCaseLabel = "iks-pipeline"
	allPipeline     TestCaseLabel = "all-pipeline"
	roksPipeline    TestCaseLabel = "roks-pipeline"
)

// Parallel/Non-Parallel labels
const (
	ParallelLabel                 TestCaseLabel = "parallel"
	NonParallelLabelTestCaseLabel TestCaseLabel = "non-parallel"
)

// Test duration labels
const (
	Slow TestCaseLabel = "slow"
	Fast TestCaseLabel = "fast"
)

// Backup labels
const (
	PxLabel         TestCaseLabel = "px"
	NonPxLabel      TestCaseLabel = "non-px"
	KDMPLabel       TestCaseLabel = "kdmp"
	CsiLabel        TestCaseLabel = "csi"
	CsiOffloadLabel TestCaseLabel = "csi-offload"
	AnyBackup       TestCaseLabel = "any-backup"
)

// StorkQualificationLabel Stork qualification labels
const (
	StorkQualificationLabel TestCaseLabel = "stork-qualification"
)

// Sanity labels
const (
	SanityLabel TestCaseLabel = "sanity"
)

// Disruptive labels
const (
	DisruptiveLabel TestCaseLabel = "disruptive"
)

// Skip test labels
const (
	SkipTestLabel TestCaseLabel = "skip-test"
)

// Volume labels
const (
	PortworxVolumeLabel TestCaseLabel = "portworx-volume"
	EBSVolumeLabel      TestCaseLabel = "ebs-volume"
	AnyVolumeLabel      TestCaseLabel = "any-volume"
)

// Backup location labels
const (
	NfsBackupLocationLabel   TestCaseLabel = "nfs"
	S3BackupLocationLabel    TestCaseLabel = "s3"
	AzureBackupLocationLabel TestCaseLabel = "Azure"
	S3LockedBucket           TestCaseLabel = "s3-locked-bucket"
	AzureImmutableBucket     TestCaseLabel = "azure-immutable-bucket"
	SSELabel                 TestCaseLabel = "sse"
)

// Labels for locked bucket delete tests
const (
	Day0LockedBucketLabel TestCaseLabel = "day0-locked-bucket"
	Day3LockedBucketLabel TestCaseLabel = "day3-locked-bucket"
)

// Labels for DR pipeline
const (
	DRPipelineLabel TestCaseLabel = "dr-pipeline"
)

// Volume provisioner labels
const (
	FACDLabel TestCaseLabel = "facd"
	FADALabel TestCaseLabel = "fada"
	FBDALabel TestCaseLabel = "fbda"
)

// App labels
const (
	KubevirtAppLabel TestCaseLabel = "kubevirt-app"
)

// Feature labels
const (
	PartialBackupLabel             TestCaseLabel = "partial-backup"
	DiffK8sVersionLabel            TestCaseLabel = "diff-k8s-version"
	ClusterShareAndSuperAdminLabel TestCaseLabel = "cluster-share"
)

/*
Things to remember when assigning labels to a test case:

1. The following categories of labels must be applied:

   - **Priority**: Assign one of P0, P1, or P2. This can be referenced from TestRail.
   - **Test Type**: Examples include System test, Scale test, Longevity test, Performance test, Functional test, etc.
   - **Pipeline**:
     - Assign the "all-pipeline" label only if the test case is critical for the basic functionality of the product.
     - Otherwise, choose the specific pipeline(s) where the test case should run.
     - Example: A backup share test case has no dependency on the underlying platform, so the author can choose vanilla or OCP.
   - **Unique Label**:
     - Assign a unique label to the test case to ensure it is not duplicated in the test suite.
     - The unique label should match the test case name specified in the `Describe` block.
   - **PxBackup Label**: Apply this label to all test cases related to PX-Backup.
   - **Backup Location Labels**:
     - Check if the test case is compatible with both S3 and NFS.
     - If yes, assign both `S3BackupLocationLabel` and `NfsBackupLocationLabel`.
   - **Px and Non-Px Labels**:
     - Check if the test case is compatible with both PX and non-PX environments.
     - If yes, assign both `PxLabel` and `NonPxLabel`.

2. Additional labels may be assigned as required.

   Example: A test case that is part of the Stork qualification suite may have the `StorkQualificationLabel` assigned.
*/

var TestCaseLabelsMap = map[TestCaseName][]TestCaseLabel{
	AddMultipleNamespaceLabels:                                                         {AddMultipleNamespaceLabelsLabel, vanillaPipeline, SystemTest, PxBackupLabel, P2, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FBDALabel, FACDLabel, FADALabel, gkePipeline, iksPipeline, KDMPLabel, CsiOffloadLabel, CsiLabel, roksPipeline},
	AllNSBackupWithIncludeNewNSOption:                                                  {AllNSBackupWithIncludeNewNSOptionLabel, ocpPipeline, SystemTest, PxBackupLabel, P0, S3BackupLocationLabel, NfsBackupLocationLabel, FBDALabel, FACDLabel, PxLabel, FADALabel, aksPipeline, NonPxLabel, iksPipeline, KDMPLabel, CsiOffloadLabel, CsiLabel, roksPipeline},
	AlternateBackupBetweenNfsAndS3:                                                     {AlternateBackupBetweenNfsAndS3Label, vanillaPipeline, SystemTest, PxBackupLabel, P2, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel},
	AzureCloudAccountCreationWithMandatoryAndNonMandatoryFields:                        {AzureCloudAccountCreationWithMandatoryAndNonMandatoryFieldsLabel, SystemTest, PxBackupLabel, P0, aksPipeline},
	AzureCloudAccountForLockedBucket:                                                   {AzureCloudAccountForLockedBucketLabel, SystemTest, PxBackupLabel, P2, aksPipeline},
	BackupAlternatingBetweenLockedAndUnlockedBuckets:                                   {BackupAlternatingBetweenLockedAndUnlockedBucketsLabel, vanillaPipeline, SystemTest, PxBackupLabel, P2, PxLabel, S3LockedBucket, AzureImmutableBucket},
	BackupAndRestoreSyncDR:                                                             {BackupAndRestoreSyncDRLabel, ocpPipeline, SystemTest, PxBackupLabel, P2, vanillaPipeline, S3BackupLocationLabel, NfsBackupLocationLabel, PxLabel, DRPipelineLabel},
	BackupAndRestoreWithNonExistingAdminNamespaceAndUpdatedResumeSuspendBackupPolicies: {allPipeline, SystemTest, PxBackupLabel, P0, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, aksPipeline, NonPxLabel},
	BackupCRsThenMultipleRestoresOnHigherK8sVersion:                                    {BackupCRsThenMultipleRestoresOnHigherK8sVersionLabel, vanillaPipeline, SystemTest, PxBackupLabel, P0, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, DiffK8sVersionLabel},
	BackupCSIVolumesWithPartialSuccess:                                                 {BackupCSIVolumesWithPartialSuccessLabel, PartialBackupLabel, ocpPipeline, iksPipeline, SystemTest, PxBackupLabel, P1},
	BackupClusterVerification:                                                          {BackupClusterVerificationLabel, SystemTest, PxBackupLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel, PxLabel, FADALabel},
	BackupLocationWithEncryptionKey:                                                    {BackupLocationWithEncryptionKeyLabel, vanillaPipeline, SystemTest, PxBackupLabel, P0, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FBDALabel, FACDLabel, FADALabel, aksPipeline, NonPxLabel, gkePipeline, iksPipeline, KDMPLabel, CsiOffloadLabel, CsiLabel, roksPipeline},
	BackupMultipleNsWithSameLabel:                                                      {BackupMultipleNsWithSameLabelLabel, ocpPipeline, SystemTest, PxBackupLabel, P1, S3BackupLocationLabel, NfsBackupLocationLabel, FBDALabel, FACDLabel, PxLabel, FADALabel, aksPipeline, NonPxLabel, gkePipeline, iksPipeline, KDMPLabel, CsiOffloadLabel, CsiLabel, roksPipeline},
	BackupNamespaceInNfsRestoredFromS3:                                                 {BackupNamespaceInNfsRestoredFromS3Label, rkePipeline, SystemTest, PxBackupLabel, P2, FACDLabel, S3BackupLocationLabel, NfsBackupLocationLabel, PxLabel, FADALabel},
	BackupNetworkErrorTest:                                                             {BackupNetworkErrorTestLabel, vanillaPipeline, SystemTest, PxBackupLabel, P1, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel},
	BackupRestartPX:                                                                    {BackupRestartPXLabel, ocpPipeline, SystemTest, PxBackupLabel, P0, PxLabel},
	BackupRestoreOnDifferentK8sVersions:                                                {BackupRestoreOnDifferentK8sVersionsLabel, vanillaPipeline, SystemTest, PxBackupLabel, P1, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, DiffK8sVersionLabel},
	BackupScheduleForOldAndNewNS:                                                       {BackupScheduleForOldAndNewNSLabel, ocpPipeline, rkePipeline, SystemTest, PxBackupLabel, P1, S3BackupLocationLabel, NfsBackupLocationLabel, FBDALabel, FACDLabel, PxLabel, FADALabel, aksPipeline, NonPxLabel, gkePipeline, iksPipeline, KDMPLabel, CsiOffloadLabel, CsiLabel, roksPipeline},
	BackupStateTransitionForScheduledBackups:                                           {BackupStateTransitionForScheduledBackupsLabel, PartialBackupLabel, ocpPipeline, iksPipeline, SystemTest, PxBackupLabel, P0},
	BackupSyncBasicTest:                                                                {BackupSyncBasicTestLabel, allPipeline, SystemTest, PxBackupLabel, P0, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel, FADALabel, ocpPipeline, NonPxLabel, aksPipeline, gkePipeline, iksPipeline, KDMPLabel, CsiOffloadLabel, CsiLabel, roksPipeline},
	BackupToLockedBucketWithSharedObjects:                                              {BackupToLockedBucketWithSharedObjectsLabel, vanillaPipeline, SystemTest, PxBackupLabel, P1, PxLabel, S3LockedBucket, AzureImmutableBucket},
	BasicBackupCreation:                                                                {BasicBackupCreationLabel, allPipeline, SystemTest, PxBackupLabel, P0, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FBDALabel, FACDLabel, FADALabel, aksPipeline, NonPxLabel, gkePipeline, iksPipeline, KDMPLabel, CsiOffloadLabel, CsiLabel, roksPipeline},
	BasicSelectiveRestore:                                                              {BasicSelectiveRestoreLabel, rkePipeline, SystemTest, PxBackupLabel, P0, S3BackupLocationLabel, NfsBackupLocationLabel, FBDALabel, PxLabel, ocpPipeline, NonPxLabel, aksPipeline, gkePipeline, iksPipeline, KDMPLabel, CsiOffloadLabel, CsiLabel, roksPipeline},
	CancelAllRunningBackupJobs:                                                         {CancelAllRunningBackupJobsLabel, ocpPipeline, SystemTest, PxBackupLabel, P2, S3BackupLocationLabel, NfsBackupLocationLabel, FBDALabel, FACDLabel, PxLabel, FADALabel, NonPxLabel, gkePipeline, iksPipeline, KDMPLabel, CsiOffloadLabel, CsiLabel, roksPipeline},
	CancelAllRunningRestoreJobs:                                                        {CancelAllRunningRestoreJobsLabel, ocpPipeline, SystemTest, PxBackupLabel, P2, S3BackupLocationLabel, NfsBackupLocationLabel, FBDALabel, FACDLabel, PxLabel, FADALabel, NonPxLabel, gkePipeline, iksPipeline, KDMPLabel, CsiOffloadLabel, CsiLabel, roksPipeline},
	CancelClusterBackupShare:                                                           {CancelClusterBackupShareLabel, vanillaPipeline, SystemTest, PxBackupLabel, P2, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FBDALabel, FACDLabel, FADALabel},
	CloudSnapsSafeWhenBackupLocationDeleteTest:                                         {CloudSnapsSafeWhenBackupLocationDeleteTestLabel, ocpPipeline, vanillaPipeline, rkePipeline, SystemTest, PxBackupLabel, P0, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel, aksPipeline, NonPxLabel, gkePipeline, iksPipeline, roksPipeline},
	CloudSnapshotMissingValidationForNFSLocation:                                       {CloudSnapshotMissingValidationForNFSLocationLabel, ocpPipeline, SystemTest, PxBackupLabel, P1, FACDLabel, NfsBackupLocationLabel, PxLabel, FADALabel},
	ClusterBackupShareToggle:                                                           {ClusterBackupShareToggleLabel, vanillaPipeline, SystemTest, PxBackupLabel, P0, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel, FADALabel, aksPipeline, NonPxLabel, gkePipeline},
	CreateBackupAndRestoreForAllCombinationsOfSSES3AndDenyPolicy:                       {CreateBackupAndRestoreForAllCombinationsOfSSES3AndDenyPolicyLabel, ocpPipeline, SystemTest, PxBackupLabel, P0, PxLabel, SSELabel},
	CreateMultipleUsersAndGroups:                                                       {CreateMultipleUsersAndGroupsLabel, vanillaPipeline, SystemTest, PxBackupLabel, P2, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FBDALabel, FACDLabel, FADALabel, gkePipeline, iksPipeline, KDMPLabel, CsiOffloadLabel, CsiLabel, roksPipeline},
	CustomBackupRestoreWithKubevirtAndNonKubevirtNS:                                    {CustomBackupRestoreWithKubevirtAndNonKubevirtNSLabel, KubevirtAppLabel, ocpPipeline, vanillaPipeline, SystemTest, PxBackupLabel, P1, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel},
	CustomResourceBackupAndRestore:                                                     {CustomResourceBackupAndRestoreLabel, rkePipeline, SystemTest, PxBackupLabel, P0, S3BackupLocationLabel, NfsBackupLocationLabel, FBDALabel, PxLabel, ocpPipeline, NonPxLabel, aksPipeline, gkePipeline, iksPipeline, KDMPLabel, CsiOffloadLabel, CsiLabel, roksPipeline},
	CustomResourceRestore:                                                              {CustomResourceRestoreLabel, allPipeline, SystemTest, PxBackupLabel, P0, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FBDALabel, FACDLabel, FADALabel, ocpPipeline, NonPxLabel, aksPipeline, gkePipeline, iksPipeline, KDMPLabel, CsiOffloadLabel, CsiLabel, roksPipeline},
	DefaultBackupRestoreWithKubevirtAndNonKubevirtNS:                                   {DefaultBackupRestoreWithKubevirtAndNonKubevirtNSLabel, KubevirtAppLabel, ocpPipeline, vanillaPipeline, SystemTest, PxBackupLabel, P0, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel},
	DeleteAllBackupObjects:                                                             {DeleteAllBackupObjectsLabel, rkePipeline, SystemTest, PxBackupLabel, P0, S3BackupLocationLabel, NfsBackupLocationLabel, FBDALabel, FACDLabel, PxLabel, FADALabel, ocpPipeline, NonPxLabel, aksPipeline, gkePipeline, iksPipeline, KDMPLabel, CsiOffloadLabel, CsiLabel, roksPipeline},
	DeleteBackupAndCheckIfBucketIsEmpty:                                                {DeleteBackupAndCheckIfBucketIsEmptyLabel, allPipeline, SystemTest, PxBackupLabel, P0, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel, FADALabel, ocpPipeline, aksPipeline, iksPipeline, roksPipeline},
	DeleteBackupOfUserNonSharedRBAC:                                                    {DeleteBackupOfUserNonSharedRBACLabel, ocpPipeline, SystemTest, PxBackupLabel, P0, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel, PxLabel, FADALabel, aksPipeline, NonPxLabel, gkePipeline},
	DeleteBackupOfUserSharedRBAC:                                                       {DeleteBackupOfUserSharedRBACLabel, ocpPipeline, SystemTest, PxBackupLabel, P0, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel, PxLabel, FADALabel, gkePipeline, iksPipeline, KDMPLabel, CsiOffloadLabel, CsiLabel, roksPipeline},
	DeleteBackupSharedByMultipleUsersFromAdmin:                                         {DeleteBackupSharedByMultipleUsersFromAdminLabel, vanillaPipeline, SystemTest, PxBackupLabel, P2, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FBDALabel, FACDLabel, FADALabel, gkePipeline, iksPipeline, KDMPLabel, CsiOffloadLabel, CsiLabel, roksPipeline},
	DeleteBucketVerifyCloudBackupMissing:                                               {DeleteBucketVerifyCloudBackupMissingLabel, allPipeline, SystemTest, PxBackupLabel, P0, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel, FADALabel, ocpPipeline, NonPxLabel, aksPipeline, gkePipeline, iksPipeline, KDMPLabel, CsiOffloadLabel, CsiLabel, roksPipeline},
	DeleteFailedInProgressBackupAndRestoreOfUserFromAdmin:                              {DeleteFailedInProgressBackupAndRestoreOfUserFromAdminLabel, vanillaPipeline, SystemTest, PxBackupLabel, P2, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FBDALabel, FACDLabel, FADALabel, gkePipeline, iksPipeline, KDMPLabel, CsiOffloadLabel, CsiLabel, roksPipeline},
	DeleteIncrementalBackupsAndRecreateNew:                                             {DeleteIncrementalBackupsAndRecreateNewLabel, ocpPipeline, rkePipeline, SystemTest, PxBackupLabel, P0, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel, PxLabel, FADALabel, aksPipeline, NonPxLabel, gkePipeline, iksPipeline, roksPipeline},
	DeleteLockedBucketUserObjectsFromAdmin:                                             {DeleteLockedBucketUserObjectsFromAdminLabel, vanillaPipeline, SystemTest, PxBackupLabel, P1, PxLabel, SkipTestLabel},
	DeleteNSDeleteClusterRestore:                                                       {DeleteNSDeleteClusterRestoreLabel, ocpPipeline, SystemTest, PxBackupLabel, P0, S3BackupLocationLabel, NfsBackupLocationLabel, FBDALabel, FACDLabel, PxLabel, FADALabel, NonPxLabel, aksPipeline, iksPipeline, KDMPLabel, CsiOffloadLabel, CsiLabel, roksPipeline},
	DeleteNfsExecutorPodWhileBackupAndRestoreInProgress:                                {DeleteNfsExecutorPodWhileBackupAndRestoreInProgressLabel, ocpPipeline, SystemTest, PxBackupLabel, P2, PxLabel},
	DeleteObjectsByMultipleUsersFromNewAdmin:                                           {DeleteObjectsByMultipleUsersFromNewAdminLabel, vanillaPipeline, SystemTest, PxBackupLabel, P1, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FBDALabel, aksPipeline, NonPxLabel, gkePipeline, iksPipeline, KDMPLabel, CsiOffloadLabel, CsiLabel, roksPipeline},
	DeleteS3ScheduleAndCreateNfsSchedule:                                               {DeleteS3ScheduleAndCreateNfsScheduleLabel, vanillaPipeline, SystemTest, PxBackupLabel, P1, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel, FADALabel},
	DeleteSameNameObjectsByMultipleUsersFromAdmin:                                      {DeleteSameNameObjectsByMultipleUsersFromAdminLabel, ocpPipeline, SystemTest, PxBackupLabel, P1, S3BackupLocationLabel, NfsBackupLocationLabel, FBDALabel, FACDLabel, PxLabel, FADALabel, aksPipeline, NonPxLabel, gkePipeline, iksPipeline, KDMPLabel, CsiOffloadLabel, CsiLabel, roksPipeline},
	DeleteSharedBackup:                                                                 {DeleteSharedBackupLabel, vanillaPipeline, SystemTest, PxBackupLabel, P2, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel, FADALabel, gkePipeline},
	DeleteSharedBackupOfUserFromAdmin:                                                  {DeleteSharedBackupOfUserFromAdminLabel, vanillaPipeline, SystemTest, PxBackupLabel, P1, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FBDALabel, FACDLabel, FADALabel, gkePipeline, iksPipeline, KDMPLabel, CsiOffloadLabel, CsiLabel, roksPipeline},
	DeleteUserBackupsAndRestoresOfDeletedAndInActiveClusterFromAdmin:                   {DeleteUserBackupsAndRestoresOfDeletedAndInActiveClusterFromAdminLabel, vanillaPipeline, SystemTest, PxBackupLabel, P2, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel, FADALabel, gkePipeline},
	DeleteUsersRole:                                                                    {DeleteUsersRoleLabel, vanillaPipeline, SystemTest, PxBackupLabel, P2, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel, FADALabel, gkePipeline},
	DeleteVerifyBackupAutoDeletionWhenNewPVCsAreAddedBetweenSchedules:                  {DeleteVerifyBackupAutoDeletionWhenNewPVCsAreAddedBetweenSchedulesLabel, vanillaPipeline, SystemTest, PxBackupLabel, P1, PxLabel, S3BackupLocationLabel, AzureBackupLocationLabel, Day3LockedBucketLabel, S3LockedBucket, AzureImmutableBucket},
	DeleteVerifyBackupDeletionWhenRetentionIsMet:                                       {DeleteVerifyBackupDeletionWhenRetentionIsMetLabel, vanillaPipeline, SystemTest, PxBackupLabel, P0, PxLabel, S3BackupLocationLabel, AzureBackupLocationLabel, Day3LockedBucketLabel, S3LockedBucket, AzureImmutableBucket},
	DifferentAccessSameUser:                                                            {DifferentAccessSameUserLabel, vanillaPipeline, SystemTest, PxBackupLabel, P1, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FBDALabel, FACDLabel, FADALabel, gkePipeline, iksPipeline, KDMPLabel, CsiOffloadLabel, CsiLabel, roksPipeline},
	DuplicateSharedBackup:                                                              {DuplicateSharedBackupLabel, vanillaPipeline, SystemTest, PxBackupLabel, P2, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FBDALabel, FACDLabel, FADALabel, gkePipeline, iksPipeline, KDMPLabel, CsiOffloadLabel, CsiLabel, roksPipeline},
	EnableNsAndClusterLevelPSAWithBackupAndRestore:                                     {EnableNsAndClusterLevelPSAWithBackupAndRestoreLabel, vanillaPipeline, SystemTest, PxBackupLabel, P0, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel},
	ExcludeDirectoryFileBackup:                                                         {ExcludeDirectoryFileBackupLabel, vanillaPipeline, SystemTest, PxBackupLabel, P1, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FBDALabel, KDMPLabel},
	ExcludeInvalidDirectoryFileBackup:                                                  {ExcludeInvalidDirectoryFileBackupLabel, vanillaPipeline, SystemTest, PxBackupLabel, P1, S3BackupLocationLabel, NfsBackupLocationLabel, FBDALabel, KDMPLabel, PxLabel},
	IssueDeleteOfIncrementalBackupsAndRestore:                                          {IssueDeleteOfIncrementalBackupsAndRestoreLabel, ocpPipeline, rkePipeline, SystemTest, PxBackupLabel, P0, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel, PxLabel, FADALabel, aksPipeline, NonPxLabel, gkePipeline, iksPipeline, roksPipeline},
	IssueMultipleBackupsAndRestoreInterleavedCopies:                                    {IssueMultipleBackupsAndRestoreInterleavedCopiesLabel, ocpPipeline, SystemTest, PxBackupLabel, P1, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel, PxLabel, FADALabel, aksPipeline, NonPxLabel, iksPipeline, KDMPLabel, CsiOffloadLabel, CsiLabel, roksPipeline},
	IssueMultipleDeletesForSharedBackup:                                                {IssueMultipleDeletesForSharedBackupLabel, ocpPipeline, SystemTest, PxBackupLabel, P2, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel, PxLabel, FADALabel, aksPipeline, NonPxLabel, iksPipeline, KDMPLabel, CsiOffloadLabel, CsiLabel, roksPipeline},
	IssueMultipleRestoresWithNamespaceAndStorageClassMapping:                           {IssueMultipleRestoresWithNamespaceAndStorageClassMappingLabel, ocpPipeline, SystemTest, PxBackupLabel, P1, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel, PxLabel, FADALabel, gkePipeline, iksPipeline, roksPipeline},
	KillStorkWithBackupsAndRestoresInProgress:                                          {KillStorkWithBackupsAndRestoresInProgressLabel, ocpPipeline, SystemTest, PxBackupLabel, P0, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel, PxLabel, FADALabel, aksPipeline, NonPxLabel, gkePipeline, iksPipeline, roksPipeline},
	KubeAndPxNamespacesSkipOnAllNSBackup:                                               {KubeAndPxNamespacesSkipOnAllNSBackupLabel, vanillaPipeline, SystemTest, PxBackupLabel, P0, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FBDALabel, FACDLabel, FADALabel},
	KubevirtInPlaceRestoreWithReplaceAndRetain:                                         {KubevirtInPlaceRestoreWithReplaceAndRetainLabel, KubevirtAppLabel, ocpPipeline, vanillaPipeline, SystemTest, PxBackupLabel, P0, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel},
	KubevirtScheduledVMDelete:                                                          {KubevirtScheduledVMDeleteLabel, KubevirtAppLabel, ocpPipeline, vanillaPipeline, SystemTest, PxBackupLabel, P2, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel},
	KubevirtUpgradeTest:                                                                {KubevirtUpgradeTestLabel, KubevirtAppLabel, ocpPipeline, vanillaPipeline, SystemTest, PxBackupLabel, P1, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel},
	KubevirtVMBackupOrDeletionInProgress:                                               {KubevirtVMBackupOrDeletionInProgressLabel, KubevirtAppLabel, ocpPipeline, vanillaPipeline, SystemTest, PxBackupLabel, P1, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel},
	KubevirtVMBackupRestoreWithDifferentStates:                                         {KubevirtVMBackupRestoreWithDifferentStatesLabel, KubevirtAppLabel, ocpPipeline, vanillaPipeline, SystemTest, PxBackupLabel, P0, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel},
	KubevirtVMBackupRestoreWithNodeSelector:                                            {KubevirtVMBackupRestoreWithNodeSelectorLabel, KubevirtAppLabel, ocpPipeline, vanillaPipeline, SystemTest, PxBackupLabel, P2, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel},
	KubevirtVMMigrationTest:                                                            {KubevirtVMMigrationTestLabel, KubevirtAppLabel, ocpPipeline, vanillaPipeline, SystemTest, PxBackupLabel, P0, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel},
	KubevirtVMRestoreWithAfterChangingVMConfig:                                         {KubevirtVMRestoreWithAfterChangingVMConfigLabel, KubevirtAppLabel, ocpPipeline, vanillaPipeline, SystemTest, PxBackupLabel, P1, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel},
	KubevirtVMWithFreezeUnfreeze:                                                       {KubevirtVMWithFreezeUnfreezeLabel, KubevirtAppLabel, ocpPipeline, vanillaPipeline, SystemTest, PxBackupLabel, P2, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel},
	LicensingCountBeforeAndAfterBackupPodRestart:                                       {LicensingCountBeforeAndAfterBackupPodRestartLabel, ocpPipeline, SystemTest, PxBackupLabel, P2, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel, PxLabel, FADALabel, gkePipeline, iksPipeline, KDMPLabel, CsiOffloadLabel, CsiLabel, roksPipeline},
	LicensingCountWithNodeLabelledBeforeClusterAddition:                                {LicensingCountWithNodeLabelledBeforeClusterAdditionLabel, ocpPipeline, SystemTest, PxBackupLabel, P2, S3BackupLocationLabel, NfsBackupLocationLabel, FBDALabel, FACDLabel, PxLabel, FADALabel, gkePipeline},
	LockedBucketResizeOnRestoredVolume:                                                 {LockedBucketResizeOnRestoredVolumeLabel, vanillaPipeline, SystemTest, PxBackupLabel, P2, PxLabel, S3LockedBucket, AzureImmutableBucket},
	LockedBucketResizeVolumeOnScheduleBackup:                                           {LockedBucketResizeVolumeOnScheduleBackupLabel, vanillaPipeline, SystemTest, PxBackupLabel, P2, PxLabel, S3LockedBucket, AzureImmutableBucket},
	ManualAndScheduleBackupUsingNSLabelWithMaxCharLimit:                                {ManualAndScheduleBackupUsingNSLabelWithMaxCharLimitLabel, vanillaPipeline, SystemTest, PxBackupLabel, P2, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel, FADALabel, gkePipeline, iksPipeline, KDMPLabel, CsiOffloadLabel, CsiLabel, roksPipeline},
	ManualAndScheduledBackupUsingNamespaceAndResourceLabel:                             {ManualAndScheduledBackupUsingNamespaceAndResourceLabelLabel, ocpPipeline, rkePipeline, SystemTest, PxBackupLabel, P1, S3BackupLocationLabel, NfsBackupLocationLabel, FBDALabel, FACDLabel, PxLabel, FADALabel, gkePipeline, iksPipeline, KDMPLabel, CsiOffloadLabel, CsiLabel, roksPipeline},
	MultipleBackupLocationWithSameEndpoint:                                             {MultipleBackupLocationWithSameEndpointLabel, vanillaPipeline, ScaleTest, PxBackupLabel, P1, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel},
	MultipleCustomRestoreSameTimeDiffStorageClassMapping:                               {MultipleCustomRestoreSameTimeDiffStorageClassMappingLabel, ocpPipeline, SystemTest, PxBackupLabel, P1, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel, PxLabel, FADALabel, gkePipeline, iksPipeline, roksPipeline},
	MultipleInPlaceRestoreSameTime:                                                     {MultipleInPlaceRestoreSameTimeLabel, allPipeline, SystemTest, PxBackupLabel, P2, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FBDALabel, FACDLabel, FADALabel, ocpPipeline, NonPxLabel, aksPipeline, gkePipeline, iksPipeline, KDMPLabel, CsiOffloadLabel, CsiLabel, roksPipeline},
	MultipleMemberProjectBackupAndRestoreForSingleNamespace:                            {MultipleMemberProjectBackupAndRestoreForSingleNamespaceLabel, rkePipeline, SystemTest, PxBackupLabel, P1, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel},
	MultipleProjectsAndNamespacesBackupAndRestore:                                      {MultipleProjectsAndNamespacesBackupAndRestoreLabel, rkePipeline, SystemTest, PxBackupLabel, P0, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel},
	MultipleProvisionerCsiKdmpBackupAndRestore:                                         {MultipleProvisionerCsiKdmpBackupAndRestoreLabel, iksPipeline, ocpPipeline, SystemTest, PxBackupLabel, P0, NonPxLabel, S3BackupLocationLabel, NfsBackupLocationLabel},
	MultipleProvisionerCsiSnapshotDeleteBackupAndRestore:                               {MultipleProvisionerCsiSnapshotDeleteBackupAndRestoreLabel, iksPipeline, SystemTest, PxBackupLabel, P0, NonPxLabel, S3BackupLocationLabel, NfsBackupLocationLabel},
	NamespaceLabelledBackupOfEmptyNamespace:                                            {NamespaceLabelledBackupOfEmptyNamespaceLabel, ocpPipeline, rkePipeline, SystemTest, PxBackupLabel, P1, S3BackupLocationLabel, NfsBackupLocationLabel, FBDALabel, FACDLabel, PxLabel, FADALabel, aksPipeline, NonPxLabel, gkePipeline},
	NamespaceLabelledBackupSharedWithDifferentAccessMode:                               {NamespaceLabelledBackupSharedWithDifferentAccessModeLabel, allPipeline, SystemTest, PxBackupLabel, P0, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FBDALabel, FACDLabel, FADALabel, gkePipeline, iksPipeline, KDMPLabel, CsiOffloadLabel, CsiLabel, roksPipeline},
	NamespaceMoveFromProjectToProjectToNoProjectWhileRestore:                           {NamespaceMoveFromProjectToProjectToNoProjectWhileRestoreLabel, rkePipeline, SystemTest, PxBackupLabel, P0, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel},
	NodeCountForLicensing:                                                              {NodeCountForLicensingLabel, allPipeline, SystemTest, PxBackupLabel, P0, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FBDALabel, FACDLabel, FADALabel, aksPipeline, NonPxLabel, gkePipeline, iksPipeline, KDMPLabel, CsiOffloadLabel, CsiLabel, roksPipeline},
	PSALowerPrivilegeToHigherPrivilegeWithProjectMapping:                               {PSALowerPrivilegeToHigherPrivilegeWithProjectMappingLabel, rkePipeline, SystemTest, PxBackupLabel, P1, S3BackupLocationLabel, NfsBackupLocationLabel, PxLabel},
	PXBackupClusterUpgradeTest:                                                         {PXBackupClusterUpgradeTestLabel, ocpPipeline, SystemTest, PxBackupLabel, P0},
	PXBackupEndToEndBackupAndRestoreWithUpgrade:                                        {PXBackupEndToEndBackupAndRestoreWithUpgradeLabel, allPipeline, SystemTest, PxBackupLabel, P0, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, iksPipeline},
	PXBackupUpgradeWithAzureCredChange:                                                 {PXBackupUpgradeWithAzureCredChangeLabel, AzureBackupLocationLabel, SystemTest, PxBackupLabel, P1, aksPipeline},
	PartialBackupSuccessWithAzureEndpoint:                                              {PartialBackupSuccessWithAzureEndpointLabel, PartialBackupLabel, AzureBackupLocationLabel, ocpPipeline, SystemTest, PxBackupLabel, P2},
	PartialBackupSuccessWithPxAndKDMPVolumes:                                           {PartialBackupSuccessWithPxAndKDMPVolumesLabel, PartialBackupLabel, ocpPipeline, iksPipeline, SystemTest, PxBackupLabel, P0},
	PartialBackupSuccessWithPxVolumes:                                                  {PartialBackupSuccessWithPxVolumesLabel, PartialBackupLabel, ocpPipeline, iksPipeline, SystemTest, PxBackupLabel, P1},
	PartialBackupWithLowerStorkVersion:                                                 {PartialBackupWithLowerStorkVersionLabel, PartialBackupLabel, ocpPipeline, iksPipeline, SystemTest, PxBackupLabel, P1},
	PsaTakeBackupInLowerPrivilegeRestoreInHigherPrivilege:                              {PsaTakeBackupInLowerPrivilegeRestoreInHigherPrivilegeLabel, vanillaPipeline, rkePipeline, SystemTest, PxBackupLabel, P0, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel},
	RebootNodesWhenBackupsAreInProgress:                                                {RebootNodesWhenBackupsAreInProgressLabel, vanillaPipeline, SystemTest, PxBackupLabel, P0, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel, FADALabel, ocpPipeline, NonPxLabel},
	RemoveJSONFilesFromNFSBackupLocation:                                               {RemoveJSONFilesFromNFSBackupLocationLabel, ocpPipeline, SystemTest, PxBackupLabel, P2, NfsBackupLocationLabel, FACDLabel, PxLabel, FADALabel},
	ReplicaChangeWhileRestore:                                                          {ReplicaChangeWhileRestoreLabel, ocpPipeline, SystemTest, PxBackupLabel, P1, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel, PxLabel, FADALabel, aksPipeline, NonPxLabel, gkePipeline, iksPipeline, roksPipeline},
	ResizeOnRestoredVolume:                                                             {ResizeOnRestoredVolumeLabel, ocpPipeline, SystemTest, PxBackupLabel, P2, S3BackupLocationLabel, PxLabel, gkePipeline, iksPipeline, roksPipeline},
	ResizeVolumeOnScheduleBackup:                                                       {ResizeVolumeOnScheduleBackupLabel, ocpPipeline, SystemTest, PxBackupLabel, P2, S3BackupLocationLabel, PxLabel, iksPipeline, roksPipeline},
	RestartBackupPodDuringBackupSharing:                                                {RestartBackupPodDuringBackupSharingLabel, ocpPipeline, SystemTest, PxBackupLabel, P1, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel, PxLabel, FADALabel, gkePipeline},
	RestoreEncryptedAndNonEncryptedBackups:                                             {RestoreEncryptedAndNonEncryptedBackupsLabel, vanillaPipeline, SystemTest, PxBackupLabel, P2, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FBDALabel, FACDLabel, FADALabel, gkePipeline, iksPipeline, KDMPLabel, CsiOffloadLabel, CsiLabel, roksPipeline},
	RestoreFromHigherPrivilegedNamespaceToLower:                                        {RestoreFromHigherPrivilegedNamespaceToLowerLabel, vanillaPipeline, rkePipeline, SystemTest, PxBackupLabel, P2, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel},
	ScaleDownPxBackupPodWhileBackupAndRestoreIsInProgress:                              {ScaleDownPxBackupPodWhileBackupAndRestoreIsInProgressLabel, ocpPipeline, SystemTest, PxBackupLabel, P0, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel, PxLabel, FADALabel, gkePipeline, iksPipeline, roksPipeline},
	ScaleMongoDBWhileBackupAndRestore:                                                  {ScaleMongoDBWhileBackupAndRestoreLabel, ocpPipeline, SystemTest, PxBackupLabel, P1, S3BackupLocationLabel, FACDLabel, PxLabel, FADALabel, NonPxLabel, NfsBackupLocationLabel, gkePipeline, iksPipeline, roksPipeline},
	ScheduleBackupCreationAllNS:                                                        {ScheduleBackupCreationAllNSLabel, vanillaPipeline, SystemTest, PxBackupLabel, P0, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FBDALabel, FACDLabel, FADALabel, aksPipeline, NonPxLabel, gkePipeline, iksPipeline, KDMPLabel, CsiOffloadLabel, CsiLabel, roksPipeline},
	ScheduleBackupDeleteAndRecreateNS:                                                  {ScheduleBackupDeleteAndRecreateNSLabel, ocpPipeline, SystemTest, PxBackupLabel, P2, S3BackupLocationLabel, NfsBackupLocationLabel, FBDALabel, FACDLabel, PxLabel, FADALabel, gkePipeline, iksPipeline, KDMPLabel, CsiOffloadLabel, CsiLabel, roksPipeline},
	ScheduleBackupWithAdditionAndRemovalOfNS:                                           {ScheduleBackupWithAdditionAndRemovalOfNSLabel, allPipeline, SystemTest, PxBackupLabel, P0, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel, FADALabel, aksPipeline, NonPxLabel, gkePipeline, iksPipeline, KDMPLabel, CsiOffloadLabel, CsiLabel, roksPipeline},
	SetUnsetNSLabelDuringScheduleBackup:                                                {SetUnsetNSLabelDuringScheduleBackupLabel, ocpPipeline, SystemTest, PxBackupLabel, P1, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel, PxLabel, FADALabel, gkePipeline, iksPipeline, KDMPLabel, CsiOffloadLabel, CsiLabel, roksPipeline},
	ShareAndRemoveBackupLocation:                                                       {ShareAndRemoveBackupLocationLabel, vanillaPipeline, SystemTest, PxBackupLabel, P0, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FBDALabel, FACDLabel, FADALabel, aksPipeline, NonPxLabel, gkePipeline},
	ShareBackupAndEdit:                                                                 {ShareBackupAndEditLabel, vanillaPipeline, SystemTest, PxBackupLabel, P2, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FBDALabel, FADALabel, gkePipeline},
	ShareBackupWithDifferentRoleUsers:                                                  {ShareBackupWithDifferentRoleUsersLabel, vanillaPipeline, SystemTest, PxBackupLabel, P1, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel, FADALabel, aksPipeline, NonPxLabel, gkePipeline},
	ShareBackupWithUsersAndGroups:                                                      {ShareBackupWithUsersAndGroupsLabel, vanillaPipeline, SystemTest, PxBackupLabel, P0, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FBDALabel, FACDLabel, FADALabel, aksPipeline, NonPxLabel, gkePipeline},
	ShareBackupsAndClusterWithUser:                                                     {ShareBackupsAndClusterWithUserLabel, vanillaPipeline, SystemTest, PxBackupLabel, P0, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FBDALabel, FACDLabel, FADALabel, aksPipeline, NonPxLabel, gkePipeline},
	ShareLargeNumberOfBackupsWithLargeNumberOfUsers:                                    {ShareLargeNumberOfBackupsWithLargeNumberOfUsersLabel, vanillaPipeline, ScaleTest, PxBackupLabel, P1, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel},
	SharedBackupDelete:                                                                 {SharedBackupDeleteLabel, vanillaPipeline, SystemTest, PxBackupLabel, P1, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FBDALabel, FACDLabel, FADALabel, gkePipeline},
	SingleNamespaceBackupRestoreToNamespaceInSameAndDifferentProject:                   {SingleNamespaceBackupRestoreToNamespaceInSameAndDifferentProjectLabel, rkePipeline, SystemTest, PxBackupLabel, P0, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel},
	StorkUpgradeWithBackup:                                                             {StorkUpgradeWithBackupLabel, SystemTest, PxBackupLabel},
	SuperAdmin:                                                                         {SuperAdminLabel},
	SuperAdminBackupShare:                                                              {SuperAdminBackupShareLabel},
	SwapShareBackup:                                                                    {SwapShareBackupLabel, vanillaPipeline, SystemTest, PxBackupLabel, P1, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FBDALabel, FACDLabel, FADALabel, aksPipeline, NonPxLabel, iksPipeline, roksPipeline},
	TimeTakenToDeleteBackupWithHeavyLoad:                                               {TimeTakenToDeleteBackupWithHeavyLoadLabel, vanillaPipeline, SystemTest, P0, PxLabel, NfsBackupLocationLabel, S3BackupLocationLabel},
	UpdatesBackupOfUserFromAdmin:                                                       {UpdatesBackupOfUserFromAdminLabel, vanillaPipeline, SystemTest, PxBackupLabel, P2, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FBDALabel, iksPipeline, KDMPLabel, CsiOffloadLabel, CsiLabel, roksPipeline},
	UpgradePxBackup:                                                                    {UpgradePxBackupLabel, SystemTest, PxBackupLabel},
	UserGroupManagement:                                                                {UserGroupManagementLabel, SystemTest, PxBackupLabel, S3BackupLocationLabel, FACDLabel, PxLabel, FADALabel},
	ValidateFiftyVolumeBackups:                                                         {ValidateFiftyVolumeBackupsLabel, vanillaPipeline, ScaleTest, PxBackupLabel, P1, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel},
	ValidateUserAccessLevel:                                                            {ValidateUserAccessLevelLabel},
	ValidateMetrics:                                                                    {ValidateMetricsLabel},
	VerifyBackupAutoDeletionWhenNewPVCsAreAddedBetweenSchedules:                        {VerifyBackupAutoDeletionWhenNewPVCsAreAddedBetweenSchedulesLabel, vanillaPipeline, SystemTest, PxBackupLabel, P1, PxLabel, S3BackupLocationLabel, AzureBackupLocationLabel, Day0LockedBucketLabel, S3LockedBucket, AzureImmutableBucket},
	VerifyBackupDeletionWhenRetentionIsMet:                                             {VerifyBackupDeletionWhenRetentionIsMetLabel, vanillaPipeline, SystemTest, PxBackupLabel, P0, PxLabel, S3BackupLocationLabel, AzureBackupLocationLabel, Day0LockedBucketLabel, S3LockedBucket, AzureImmutableBucket},
	VerifyRBACForAppAdmin:                                                              {VerifyRBACForAppAdminLabel, vanillaPipeline, SystemTest, PxBackupLabel, P0, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel, FADALabel, gkePipeline},
	VerifyRBACForAppUser:                                                               {VerifyRBACForAppUserLabel, vanillaPipeline, SystemTest, PxBackupLabel, P0, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel, FADALabel},
	VerifyRBACForInfraAdmin:                                                            {VerifyRBACForInfraAdminLabel, vanillaPipeline, gkePipeline, SystemTest, PxBackupLabel, P0, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel, FADALabel},
	VerifyRBACForPxAdmin:                                                               {VerifyRBACForPxAdminLabel, vanillaPipeline, SystemTest, PxBackupLabel, P2, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel, gkePipeline},
	ViewOnlyFullBackupRestoreIncrementalBackup:                                         {ViewOnlyFullBackupRestoreIncrementalBackupLabel, vanillaPipeline, SystemTest, PxBackupLabel, P1, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel, gkePipeline, iksPipeline, roksPipeline},
	TimeTakenToDeleteScheduleBackupWithHeavyLoadAfterSuspendingTheSchedule:             {TimeTakenToDeleteScheduleBackupWithHeavyLoadAfterSuspendingTheScheduleLabel, PerformanceTest, P0, PxLabel, S3BackupLocationLabel},
	BackupSuperAdminRoleForLocalUser:                                                   {BackupSuperAdminRoleForLocalUserLabel, ClusterShareAndSuperAdminLabel, vanillaPipeline, allPipeline, SystemTest, P0, PxLabel, NfsBackupLocationLabel, S3BackupLocationLabel},
	ClusterShareWithLargeNumberOfUsersAndClusters:                                      {ClusterShareWithLargeNumberOfUsersAndClustersLabel, ClusterShareAndSuperAdminLabel, vanillaPipeline, SystemTest, PxBackupLabel, P0, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel},
	ValidateClusterShareWithConcurrentBackupOperations:                                 {ValidateClusterShareWithConcurrentBackupOperationsLabel, ClusterShareAndSuperAdminLabel, vanillaPipeline, SystemTest, PxBackupLabel, P1, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel},
	ValidateBackupShareUsingBackupsFromSharedCluster:                                   {ValidateBackupShareUsingBackupsFromSharedClusterLabel, ClusterShareAndSuperAdminLabel, vanillaPipeline, SystemTest, PxBackupLabel, P1, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel},
	ValidateClusterShareWhileBringDownPxBackupPods:                                     {ValidateClusterShareWhileBringDownPxBackupPodsLabel, ClusterShareAndSuperAdminLabel, vanillaPipeline, SystemTest, PxBackupLabel, P1, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel},
	StorkControllerConfigCM:                                                            {StorkControllerConfigCMLabel},
	BackupShare:                                                                        {BackupShareLabel},
	SoftDeleteAndRecoverBackupOnContainerAndBlobLevel:                                  {SoftDeleteAndRecoverBackupOnContainerAndBlobLevelLabel, vanillaPipeline, SystemTest, PxBackupLabel, P2, PxLabel, AzureBackupLocationLabel, Day3LockedBucketLabel, AzureImmutableBucket},
	DeleteSoftDeleteAndRecoverBackupOnContainerAndBlobLevel:                            {DeleteSoftDeleteAndRecoverBackupOnContainerAndBlobLevelLabel, vanillaPipeline, SystemTest, PxBackupLabel, P2, PxLabel, AzureBackupLocationLabel, Day3LockedBucketLabel, AzureImmutableBucket},
	BackupDeletionWithDynamicPVCGeneration:                                             {BackupDeletionWithDynamicPVCGenerationLabel, PerformanceTest, P1, PxLabel, S3BackupLocationLabel},
}
