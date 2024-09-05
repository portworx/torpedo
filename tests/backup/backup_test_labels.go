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
	PxLabel   TestCaseLabel = "px"
	KDMPLabel TestCaseLabel = "kdmp"
	CsiLabel  TestCaseLabel = "csi"
	AnyBackup TestCaseLabel = "any-backup"
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
)

// Labels for locked bucket delete tests
const (
	Day0LockedBucketLabel TestCaseLabel = "day0-locked-bucket"
	Day3LockedBucketLabel TestCaseLabel = "day3-locked-bucket"
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
	PartialBackupLabel  TestCaseLabel = "PartialBackup"
	DiffK8sVersionLabel TestCaseLabel = "DiffK8sVersion"
)

var TestCaseLabelsMap = map[TestCaseName][]TestCaseLabel{
	AddMultipleNamespaceLabels:                                                         {AddMultipleNamespaceLabelsLabel, vanillaPipeline, SystemTest, PxBackupLabel, P2, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel, FBDALabel},
	AllNSBackupWithIncludeNewNSOption:                                                  {AllNSBackupWithIncludeNewNSOptionLabel, ocpPipeline, SystemTest, PxBackupLabel, P0, S3BackupLocationLabel, FACDLabel, NfsBackupLocationLabel, FBDALabel, FADALabel},
	AlternateBackupBetweenNfsAndS3:                                                     {AlternateBackupBetweenNfsAndS3Label, vanillaPipeline, SystemTest, PxBackupLabel, P2, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel},
	AzureCloudAccountCreationWithMandatoryAndNonMandatoryFields:                        {AzureCloudAccountCreationWithMandatoryAndNonMandatoryFieldsLabel, aksPipeline, SystemTest, PxBackupLabel, P0},
	AzureCloudAccountForLockedBucket:                                                   {AzureCloudAccountForLockedBucketLabel, aksPipeline, SystemTest, PxBackupLabel, P2},
	BackupAlternatingBetweenLockedAndUnlockedBuckets:                                   {BackupAlternatingBetweenLockedAndUnlockedBucketsLabel, vanillaPipeline, SystemTest, PxBackupLabel, P2, PxLabel, S3LockedBucket, AzureImmutableBucket},
	BackupAndRestoreSyncDR:                                                             {BackupAndRestoreSyncDRLabel, ocpPipeline, SystemTest, PxBackupLabel, P2},
	BackupAndRestoreWithNonExistingAdminNamespaceAndUpdatedResumeSuspendBackupPolicies: {allPipeline, SystemTest, PxBackupLabel, P0, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel, FADALabel},
	BackupCRsThenMultipleRestoresOnHigherK8sVersion:                                    {BackupCRsThenMultipleRestoresOnHigherK8sVersionLabel, vanillaPipeline, SystemTest, PxBackupLabel, P0, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, DiffK8sVersionLabel},
	BackupCSIVolumesWithPartialSuccess:                                                 {BackupCSIVolumesWithPartialSuccessLabel, PartialBackupLabel, ocpPipeline, iksPipeline, SystemTest, PxBackupLabel, P1},
	BackupClusterVerification:                                                          {BackupClusterVerificationLabel, SystemTest, PxBackupLabel, S3BackupLocationLabel, FACDLabel},
	BackupLocationWithEncryptionKey:                                                    {BackupLocationWithEncryptionKeyLabel, vanillaPipeline, SystemTest, PxBackupLabel, P0, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel, FBDALabel, FADALabel},
	BackupMultipleNsWithSameLabel:                                                      {BackupMultipleNsWithSameLabelLabel, ocpPipeline, SystemTest, PxBackupLabel, P1, S3BackupLocationLabel, FACDLabel, NfsBackupLocationLabel, FBDALabel},
	BackupNamespaceInNfsRestoredFromS3:                                                 {BackupNamespaceInNfsRestoredFromS3Label, rkePipeline, SystemTest, PxBackupLabel, P2},
	BackupNetworkErrorTest:                                                             {BackupNetworkErrorTestLabel, vanillaPipeline, SystemTest, PxBackupLabel, P1, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel},
	BackupRestartPX:                                                                    {BackupRestartPXLabel, ocpPipeline, SystemTest, PxBackupLabel, P0, FADALabel},
	BackupRestoreOnDifferentK8sVersions:                                                {BackupRestoreOnDifferentK8sVersionsLabel, vanillaPipeline, SystemTest, PxBackupLabel, P1, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, DiffK8sVersionLabel},
	BackupScheduleForOldAndNewNS:                                                       {BackupScheduleForOldAndNewNSLabel, ocpPipeline, rkePipeline, SystemTest, PxBackupLabel, P1, S3BackupLocationLabel, FACDLabel, NfsBackupLocationLabel, FBDALabel},
	BackupStateTransitionForScheduledBackups:                                           {BackupStateTransitionForScheduledBackupsLabel, PartialBackupLabel, ocpPipeline, iksPipeline, SystemTest, PxBackupLabel, P0, FADALabel},
	BackupSyncBasicTest:                                                                {BackupSyncBasicTestLabel, allPipeline, SystemTest, PxBackupLabel, P0, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel, FADALabel},
	BackupToLockedBucketWithSharedObjects:                                              {BackupToLockedBucketWithSharedObjectsLabel, vanillaPipeline, SystemTest, PxBackupLabel, P1, PxLabel, S3LockedBucket, AzureImmutableBucket},
	BasicBackupCreation:                                                                {BasicBackupCreationLabel, allPipeline, SystemTest, PxBackupLabel, P0, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel, FBDALabel, FADALabel},
	BasicSelectiveRestore:                                                              {BasicSelectiveRestoreLabel, rkePipeline, SystemTest, PxBackupLabel, P0, S3BackupLocationLabel, FACDLabel, NfsBackupLocationLabel, FBDALabel, FADALabel},
	CancelAllRunningBackupJobs:                                                         {CancelAllRunningBackupJobsLabel, ocpPipeline, SystemTest, PxBackupLabel, P2, S3BackupLocationLabel, FACDLabel, NfsBackupLocationLabel, FBDALabel},
	CancelAllRunningRestoreJobs:                                                        {CancelAllRunningRestoreJobsLabel, ocpPipeline, SystemTest, PxBackupLabel, P2, S3BackupLocationLabel, FACDLabel, NfsBackupLocationLabel, FBDALabel},
	CancelClusterBackupShare:                                                           {CancelClusterBackupShareLabel, vanillaPipeline, SystemTest, PxBackupLabel, P2, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel, FBDALabel},
	CloudSnapsSafeWhenBackupLocationDeleteTest:                                         {CloudSnapsSafeWhenBackupLocationDeleteTestLabel, ocpPipeline, vanillaPipeline, rkePipeline, SystemTest, PxBackupLabel, P0, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel, FADALabel},
	CloudSnapshotMissingValidationForNFSLocation:                                       {CloudSnapshotMissingValidationForNFSLocationLabel, ocpPipeline, SystemTest, PxBackupLabel, P1},
	ClusterBackupShareToggle:                                                           {ClusterBackupShareToggleLabel, vanillaPipeline, SystemTest, PxBackupLabel, P0, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel, FADALabel},
	CreateBackupAndRestoreForAllCombinationsOfSSES3AndDenyPolicy:                       {CreateBackupAndRestoreForAllCombinationsOfSSES3AndDenyPolicyLabel, ocpPipeline, SystemTest, PxBackupLabel, P0, FADALabel},
	CreateMultipleUsersAndGroups:                                                       {CreateMultipleUsersAndGroupsLabel, vanillaPipeline, SystemTest, PxBackupLabel, P2, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel, FBDALabel},
	CustomBackupRestoreWithKubevirtAndNonKubevirtNS:                                    {CustomBackupRestoreWithKubevirtAndNonKubevirtNSLabel, KubevirtAppLabel, ocpPipeline, vanillaPipeline, SystemTest, PxBackupLabel, P1, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel},
	CustomResourceBackupAndRestore:                                                     {CustomResourceBackupAndRestoreLabel, rkePipeline, SystemTest, PxBackupLabel, P0, S3BackupLocationLabel, FACDLabel, NfsBackupLocationLabel, FBDALabel, FADALabel},
	CustomResourceRestore:                                                              {CustomResourceRestoreLabel, allPipeline, SystemTest, PxBackupLabel, P0, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel, FBDALabel, FADALabel},
	DefaultBackupRestoreWithKubevirtAndNonKubevirtNS:                                   {DefaultBackupRestoreWithKubevirtAndNonKubevirtNSLabel, KubevirtAppLabel, ocpPipeline, vanillaPipeline, SystemTest, PxBackupLabel, P0, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FADALabel},
	DeleteAllBackupObjects:                                                             {DeleteAllBackupObjectsLabel, rkePipeline, SystemTest, PxBackupLabel, P0, S3BackupLocationLabel, FACDLabel, NfsBackupLocationLabel, FBDALabel, FADALabel},
	DeleteBackupAndCheckIfBucketIsEmpty:                                                {DeleteBackupAndCheckIfBucketIsEmptyLabel, allPipeline, SystemTest, PxBackupLabel, P0, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel, FADALabel},
	DeleteBackupOfUserNonSharedRBAC:                                                    {DeleteBackupOfUserNonSharedRBACLabel, ocpPipeline, aksPipeline, SystemTest, PxBackupLabel, P0, S3BackupLocationLabel, FACDLabel, FADALabel},
	DeleteBackupOfUserSharedRBAC:                                                       {DeleteBackupOfUserSharedRBACLabel, ocpPipeline, aksPipeline, SystemTest, PxBackupLabel, P0, S3BackupLocationLabel, FACDLabel, NfsBackupLocationLabel, FADALabel},
	DeleteBackupSharedByMultipleUsersFromAdmin:                                         {DeleteBackupSharedByMultipleUsersFromAdminLabel, vanillaPipeline, SystemTest, PxBackupLabel, P2, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel, FBDALabel},
	DeleteBucketVerifyCloudBackupMissing:                                               {DeleteBucketVerifyCloudBackupMissingLabel, allPipeline, SystemTest, PxBackupLabel, P0, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel, FADALabel},
	DeleteFailedInProgressBackupAndRestoreOfUserFromAdmin:                              {DeleteFailedInProgressBackupAndRestoreOfUserFromAdminLabel, vanillaPipeline, SystemTest, PxBackupLabel, P2, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel, FBDALabel},
	DeleteIncrementalBackupsAndRecreateNew:                                             {DeleteIncrementalBackupsAndRecreateNewLabel, ocpPipeline, rkePipeline, SystemTest, PxBackupLabel, P0, S3BackupLocationLabel, FACDLabel, FADALabel},
	DeleteLockedBucketUserObjectsFromAdmin:                                             {DeleteLockedBucketUserObjectsFromAdminLabel, vanillaPipeline, SystemTest, PxBackupLabel, P1, PxLabel, SkipTestLabel},
	DeleteNSDeleteClusterRestore:                                                       {DeleteNSDeleteClusterRestoreLabel, ocpPipeline, SystemTest, PxBackupLabel, P0, S3BackupLocationLabel, FACDLabel, NfsBackupLocationLabel, FBDALabel, FADALabel},
	DeleteNfsExecutorPodWhileBackupAndRestoreInProgress:                                {DeleteNfsExecutorPodWhileBackupAndRestoreInProgressLabel, ocpPipeline, SystemTest, PxBackupLabel, P2},
	DeleteObjectsByMultipleUsersFromNewAdmin:                                           {DeleteObjectsByMultipleUsersFromNewAdminLabel, vanillaPipeline, SystemTest, PxBackupLabel, P1, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel, FBDALabel},
	DeleteS3ScheduleAndCreateNfsSchedule:                                               {DeleteS3ScheduleAndCreateNfsScheduleLabel, vanillaPipeline, SystemTest, PxBackupLabel, P1, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel},
	DeleteSameNameObjectsByMultipleUsersFromAdmin:                                      {DeleteSameNameObjectsByMultipleUsersFromAdminLabel, ocpPipeline, aksPipeline, SystemTest, PxBackupLabel, P1, S3BackupLocationLabel, FACDLabel, NfsBackupLocationLabel, FBDALabel},
	DeleteSharedBackup:                                                                 {DeleteSharedBackupLabel, vanillaPipeline, SystemTest, PxBackupLabel, P2, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel},
	DeleteSharedBackupOfUserFromAdmin:                                                  {DeleteSharedBackupOfUserFromAdminLabel, vanillaPipeline, SystemTest, PxBackupLabel, P1, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel, FBDALabel},
	DeleteUserBackupsAndRestoresOfDeletedAndInActiveClusterFromAdmin:                   {DeleteUserBackupsAndRestoresOfDeletedAndInActiveClusterFromAdminLabel, vanillaPipeline, SystemTest, PxBackupLabel, P2, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel},
	DeleteUsersRole:                                                                    {DeleteUsersRoleLabel, vanillaPipeline, SystemTest, PxBackupLabel, P2, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel},
	DifferentAccessSameUser:                                                            {DifferentAccessSameUserLabel, vanillaPipeline, SystemTest, PxBackupLabel, P1, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel, FBDALabel},
	DuplicateSharedBackup:                                                              {DuplicateSharedBackupLabel, vanillaPipeline, SystemTest, PxBackupLabel, P2, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel, FBDALabel},
	EnableNsAndClusterLevelPSAWithBackupAndRestore:                                     {EnableNsAndClusterLevelPSAWithBackupAndRestoreLabel, vanillaPipeline, SystemTest, PxBackupLabel, P0, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FADALabel},
	ExcludeDirectoryFileBackup:                                                         {ExcludeDirectoryFileBackupLabel, vanillaPipeline, SystemTest, PxBackupLabel, P1, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FBDALabel, KDMPLabel},
	ExcludeInvalidDirectoryFileBackup:                                                  {ExcludeInvalidDirectoryFileBackupLabel, vanillaPipeline, SystemTest, PxBackupLabel, P1, S3BackupLocationLabel, NfsBackupLocationLabel, FBDALabel, KDMPLabel},
	IssueDeleteOfIncrementalBackupsAndRestore:                                          {IssueDeleteOfIncrementalBackupsAndRestoreLabel, ocpPipeline, rkePipeline, SystemTest, PxBackupLabel, P0, S3BackupLocationLabel, FACDLabel, FADALabel},
	IssueMultipleBackupsAndRestoreInterleavedCopies:                                    {IssueMultipleBackupsAndRestoreInterleavedCopiesLabel, ocpPipeline, SystemTest, PxBackupLabel, P1, S3BackupLocationLabel, FACDLabel},
	IssueMultipleDeletesForSharedBackup:                                                {IssueMultipleDeletesForSharedBackupLabel, ocpPipeline, SystemTest, PxBackupLabel, P2, S3BackupLocationLabel, FACDLabel},
	IssueMultipleRestoresWithNamespaceAndStorageClassMapping:                           {IssueMultipleRestoresWithNamespaceAndStorageClassMappingLabel, ocpPipeline, SystemTest, PxBackupLabel, P1, S3BackupLocationLabel, FACDLabel},
	KillStorkWithBackupsAndRestoresInProgress:                                          {KillStorkWithBackupsAndRestoresInProgressLabel, ocpPipeline, SystemTest, PxBackupLabel, P0, S3BackupLocationLabel, FACDLabel, FADALabel},
	KubeAndPxNamespacesSkipOnAllNSBackup:                                               {KubeAndPxNamespacesSkipOnAllNSBackupLabel, vanillaPipeline, SystemTest, PxBackupLabel, P0, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel, FBDALabel, FADALabel},
	KubevirtInPlaceRestoreWithReplaceAndRetain:                                         {KubevirtInPlaceRestoreWithReplaceAndRetainLabel, KubevirtAppLabel, ocpPipeline, vanillaPipeline, SystemTest, PxBackupLabel, P0, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FADALabel},
	KubevirtScheduledVMDelete:                                                          {KubevirtScheduledVMDeleteLabel, KubevirtAppLabel, ocpPipeline, vanillaPipeline, SystemTest, PxBackupLabel, P2, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel},
	KubevirtUpgradeTest:                                                                {KubevirtUpgradeTestLabel, KubevirtAppLabel, ocpPipeline, vanillaPipeline, SystemTest, PxBackupLabel, P1, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel},
	KubevirtVMBackupOrDeletionInProgress:                                               {KubevirtVMBackupOrDeletionInProgressLabel, KubevirtAppLabel, ocpPipeline, vanillaPipeline, SystemTest, PxBackupLabel, P1, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel},
	KubevirtVMBackupRestoreWithDifferentStates:                                         {KubevirtVMBackupRestoreWithDifferentStatesLabel, KubevirtAppLabel, ocpPipeline, vanillaPipeline, SystemTest, PxBackupLabel, P0, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FADALabel},
	KubevirtVMBackupRestoreWithNodeSelector:                                            {KubevirtVMBackupRestoreWithNodeSelectorLabel, KubevirtAppLabel, ocpPipeline, vanillaPipeline, SystemTest, PxBackupLabel, P2, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel},
	KubevirtVMMigrationTest:                                                            {KubevirtVMMigrationTestLabel, KubevirtAppLabel, ocpPipeline, vanillaPipeline, SystemTest, PxBackupLabel, P0, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FADALabel},
	KubevirtVMRestoreWithAfterChangingVMConfig:                                         {KubevirtVMRestoreWithAfterChangingVMConfigLabel, KubevirtAppLabel, ocpPipeline, vanillaPipeline, SystemTest, PxBackupLabel, P1, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel},
	KubevirtVMWithFreezeUnfreeze:                                                       {KubevirtVMWithFreezeUnfreezeLabel, KubevirtAppLabel, ocpPipeline, vanillaPipeline, SystemTest, PxBackupLabel, P2, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel},
	LicensingCountBeforeAndAfterBackupPodRestart:                                       {LicensingCountBeforeAndAfterBackupPodRestartLabel, ocpPipeline, SystemTest, PxBackupLabel, P2, S3BackupLocationLabel, FACDLabel, NfsBackupLocationLabel},
	LicensingCountWithNodeLabelledBeforeClusterAddition:                                {LicensingCountWithNodeLabelledBeforeClusterAdditionLabel, ocpPipeline, SystemTest, PxBackupLabel, P2, S3BackupLocationLabel, FACDLabel, NfsBackupLocationLabel, FBDALabel},
	LockedBucketResizeOnRestoredVolume:                                                 {LockedBucketResizeOnRestoredVolumeLabel, vanillaPipeline, SystemTest, PxBackupLabel, P2, PxLabel, S3LockedBucket, AzureImmutableBucket},
	LockedBucketResizeVolumeOnScheduleBackup:                                           {LockedBucketResizeVolumeOnScheduleBackupLabel, vanillaPipeline, SystemTest, PxBackupLabel, P2, PxLabel, S3LockedBucket, AzureImmutableBucket},
	ManualAndScheduleBackupUsingNSLabelWithMaxCharLimit:                                {ManualAndScheduleBackupUsingNSLabelWithMaxCharLimitLabel, vanillaPipeline, SystemTest, PxBackupLabel, P2, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel},
	ManualAndScheduledBackupUsingNamespaceAndResourceLabel:                             {ManualAndScheduledBackupUsingNamespaceAndResourceLabelLabel, ocpPipeline, rkePipeline, SystemTest, PxBackupLabel, P1, S3BackupLocationLabel, FACDLabel, NfsBackupLocationLabel, FBDALabel},
	MultipleBackupLocationWithSameEndpoint:                                             {MultipleBackupLocationWithSameEndpointLabel, vanillaPipeline, SystemTest, PxBackupLabel, P1, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel},
	MultipleCustomRestoreSameTimeDiffStorageClassMapping:                               {MultipleCustomRestoreSameTimeDiffStorageClassMappingLabel, ocpPipeline, SystemTest, PxBackupLabel, P1, S3BackupLocationLabel, FACDLabel},
	MultipleInPlaceRestoreSameTime:                                                     {MultipleInPlaceRestoreSameTimeLabel, allPipeline, SystemTest, PxBackupLabel, P2, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel, FBDALabel},
	MultipleMemberProjectBackupAndRestoreForSingleNamespace:                            {MultipleMemberProjectBackupAndRestoreForSingleNamespaceLabel, rkePipeline, SystemTest, PxBackupLabel, P1},
	MultipleProjectsAndNamespacesBackupAndRestore:                                      {MultipleProjectsAndNamespacesBackupAndRestoreLabel, rkePipeline, SystemTest, PxBackupLabel, P0, FADALabel},
	MultipleProvisionerCsiKdmpBackupAndRestore:                                         {MultipleProvisionerCsiKdmpBackupAndRestoreLabel, allPipeline, SystemTest, PxBackupLabel, P0, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FADALabel},
	MultipleProvisionerCsiSnapshotDeleteBackupAndRestore:                               {MultipleProvisionerCsiSnapshotDeleteBackupAndRestoreLabel, allPipeline, SystemTest, PxBackupLabel, P0, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel, FADALabel},
	NamespaceLabelledBackupOfEmptyNamespace:                                            {NamespaceLabelledBackupOfEmptyNamespaceLabel, ocpPipeline, rkePipeline, SystemTest, PxBackupLabel, P1, S3BackupLocationLabel, FACDLabel, NfsBackupLocationLabel, FBDALabel},
	NamespaceLabelledBackupSharedWithDifferentAccessMode:                               {NamespaceLabelledBackupSharedWithDifferentAccessModeLabel, allPipeline, SystemTest, PxBackupLabel, P0, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel, FBDALabel, FADALabel},
	NamespaceMoveFromProjectToProjectToNoProjectWhileRestore:                           {NamespaceMoveFromProjectToProjectToNoProjectWhileRestoreLabel, rkePipeline, SystemTest, PxBackupLabel, P0, FADALabel},
	NodeCountForLicensing:                                                              {NodeCountForLicensingLabel, allPipeline, SystemTest, PxBackupLabel, P0, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel, FBDALabel, FADALabel},
	PSALowerPrivilegeToHigherPrivilegeWithProjectMapping:                               {PSALowerPrivilegeToHigherPrivilegeWithProjectMappingLabel, rkePipeline, SystemTest, PxBackupLabel, P1},
	PXBackupClusterUpgradeTest:                                                         {PXBackupClusterUpgradeTestLabel, ocpPipeline, SystemTest, PxBackupLabel, P0, FADALabel},
	PXBackupEndToEndBackupAndRestoreWithUpgrade:                                        {PXBackupEndToEndBackupAndRestoreWithUpgradeLabel, allPipeline, SystemTest, PxBackupLabel, P0, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FADALabel},
	PXBackupUpgradeWithAzureCredChange:                                                 {PXBackupUpgradeWithAzureCredChangeLabel, AzureBackupLocationLabel, aksPipeline, SystemTest, PxBackupLabel, P1},
	PartialBackupSuccessWithAzureEndpoint:                                              {PartialBackupSuccessWithAzureEndpointLabel, PartialBackupLabel, AzureBackupLocationLabel, ocpPipeline, SystemTest, PxBackupLabel, P2},
	PartialBackupSuccessWithPxAndKDMPVolumes:                                           {PartialBackupSuccessWithPxAndKDMPVolumesLabel, PartialBackupLabel, ocpPipeline, iksPipeline, SystemTest, PxBackupLabel, P0, FADALabel},
	PartialBackupSuccessWithPxVolumes:                                                  {PartialBackupSuccessWithPxVolumesLabel, PartialBackupLabel, ocpPipeline, iksPipeline, SystemTest, PxBackupLabel, P1},
	PartialBackupWithLowerStorkVersion:                                                 {PartialBackupWithLowerStorkVersionLabel, PartialBackupLabel, ocpPipeline, iksPipeline, SystemTest, PxBackupLabel, P1},
	PsaTakeBackupInLowerPrivilegeRestoreInHigherPrivilege:                              {PsaTakeBackupInLowerPrivilegeRestoreInHigherPrivilegeLabel, vanillaPipeline, rkePipeline, SystemTest, PxBackupLabel, P0, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel},
	RebootNodesWhenBackupsAreInProgress:                                                {RebootNodesWhenBackupsAreInProgressLabel, vanillaPipeline, SystemTest, PxBackupLabel, P0, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel, FADALabel},
	RemoveJSONFilesFromNFSBackupLocation:                                               {RemoveJSONFilesFromNFSBackupLocationLabel, ocpPipeline, SystemTest, PxBackupLabel, P2},
	ReplicaChangeWhileRestore:                                                          {ReplicaChangeWhileRestoreLabel, ocpPipeline, SystemTest, PxBackupLabel, P1, S3BackupLocationLabel, FACDLabel},
	ResizeOnRestoredVolume:                                                             {ResizeOnRestoredVolumeLabel, ocpPipeline, SystemTest, PxBackupLabel, P2, S3BackupLocationLabel, FACDLabel},
	ResizeVolumeOnScheduleBackup:                                                       {ResizeVolumeOnScheduleBackupLabel, ocpPipeline, SystemTest, PxBackupLabel, P2, S3BackupLocationLabel, FACDLabel},
	RestartBackupPodDuringBackupSharing:                                                {RestartBackupPodDuringBackupSharingLabel, ocpPipeline, SystemTest, PxBackupLabel, P1, S3BackupLocationLabel, FACDLabel, NfsBackupLocationLabel},
	RestoreEncryptedAndNonEncryptedBackups:                                             {RestoreEncryptedAndNonEncryptedBackupsLabel, vanillaPipeline, SystemTest, PxBackupLabel, P2, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel, FBDALabel},
	RestoreFromHigherPrivilegedNamespaceToLower:                                        {RestoreFromHigherPrivilegedNamespaceToLowerLabel, vanillaPipeline, rkePipeline, SystemTest, PxBackupLabel, P2, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel},
	ScaleDownPxBackupPodWhileBackupAndRestoreIsInProgress:                              {ScaleDownPxBackupPodWhileBackupAndRestoreIsInProgressLabel, ocpPipeline, SystemTest, PxBackupLabel, P0, S3BackupLocationLabel, FACDLabel, FADALabel},
	ScaleMongoDBWhileBackupAndRestore:                                                  {ScaleMongoDBWhileBackupAndRestoreLabel, ocpPipeline, SystemTest, PxBackupLabel, P1, S3BackupLocationLabel, FACDLabel},
	ScheduleBackupCreationAllNS:                                                        {ScheduleBackupCreationAllNSLabel, vanillaPipeline, SystemTest, PxBackupLabel, P0, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel, FBDALabel, FADALabel},
	ScheduleBackupDeleteAndRecreateNS:                                                  {ScheduleBackupDeleteAndRecreateNSLabel, ocpPipeline, SystemTest, PxBackupLabel, P2, S3BackupLocationLabel, FACDLabel, NfsBackupLocationLabel, FBDALabel},
	ScheduleBackupWithAdditionAndRemovalOfNS:                                           {ScheduleBackupWithAdditionAndRemovalOfNSLabel, allPipeline, SystemTest, PxBackupLabel, P0, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel, FADALabel},
	SetUnsetNSLabelDuringScheduleBackup:                                                {SetUnsetNSLabelDuringScheduleBackupLabel, ocpPipeline, SystemTest, PxBackupLabel, P1, S3BackupLocationLabel, FACDLabel, NfsBackupLocationLabel},
	ShareAndRemoveBackupLocation:                                                       {ShareAndRemoveBackupLocationLabel, vanillaPipeline, SystemTest, PxBackupLabel, P0, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel, FBDALabel, FADALabel},
	ShareBackupAndEdit:                                                                 {ShareBackupAndEditLabel, vanillaPipeline, SystemTest, PxBackupLabel, P2, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel, FBDALabel},
	ShareBackupWithDifferentRoleUsers:                                                  {ShareBackupWithDifferentRoleUsersLabel, vanillaPipeline, SystemTest, PxBackupLabel, P1, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel},
	ShareBackupWithUsersAndGroups:                                                      {ShareBackupWithUsersAndGroupsLabel, vanillaPipeline, SystemTest, PxBackupLabel, P0, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel, FBDALabel, FADALabel},
	ShareBackupsAndClusterWithUser:                                                     {ShareBackupsAndClusterWithUserLabel, vanillaPipeline, SystemTest, PxBackupLabel, P0, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel, FBDALabel, FADALabel},
	ShareLargeNumberOfBackupsWithLargeNumberOfUsers:                                    {ShareLargeNumberOfBackupsWithLargeNumberOfUsersLabel, vanillaPipeline, SystemTest, PxBackupLabel, P1, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel},
	SharedBackupDelete:                                                                 {SharedBackupDeleteLabel, vanillaPipeline, SystemTest, PxBackupLabel, P1, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel, FBDALabel},
	SingleNamespaceBackupRestoreToNamespaceInSameAndDifferentProject:                   {SingleNamespaceBackupRestoreToNamespaceInSameAndDifferentProjectLabel, rkePipeline, SystemTest, PxBackupLabel, P0, FADALabel},
	StorkUpgradeWithBackup:                                                             {StorkUpgradeWithBackupLabel, SystemTest, PxBackupLabel},
	SwapShareBackup:                                                                    {SwapShareBackupLabel, vanillaPipeline, SystemTest, PxBackupLabel, P1, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel, FBDALabel},
	UpdatesBackupOfUserFromAdmin:                                                       {UpdatesBackupOfUserFromAdminLabel, vanillaPipeline, SystemTest, PxBackupLabel, P2, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel, FBDALabel},
	UpgradePxBackup:                                                                    {UpgradePxBackupLabel, SystemTest, PxBackupLabel},
	UserGroupManagement:                                                                {UserGroupManagementLabel, SystemTest, PxBackupLabel, S3BackupLocationLabel, FACDLabel},
	ValidateFiftyVolumeBackups:                                                         {ValidateFiftyVolumeBackupsLabel, vanillaPipeline, SystemTest, PxBackupLabel, P1, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel},
	VerifyRBACForAppAdmin:                                                              {VerifyRBACForAppAdminLabel, vanillaPipeline, SystemTest, PxBackupLabel, P0, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel, FADALabel},
	VerifyRBACForAppUser:                                                               {VerifyRBACForAppUserLabel, vanillaPipeline, SystemTest, PxBackupLabel, P0, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel, FADALabel},
	VerifyRBACForInfraAdmin:                                                            {VerifyRBACForInfraAdminLabel, vanillaPipeline, SystemTest, PxBackupLabel, P0, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel, FADALabel},
	VerifyRBACForPxAdmin:                                                               {VerifyRBACForPxAdminLabel, vanillaPipeline, SystemTest, PxBackupLabel, P2, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel},
	ViewOnlyFullBackupRestoreIncrementalBackup:                                         {ViewOnlyFullBackupRestoreIncrementalBackupLabel, vanillaPipeline, SystemTest, PxBackupLabel, P1, PxLabel, S3BackupLocationLabel, NfsBackupLocationLabel, FACDLabel},
	VerifyBackupDeletionWhenRetentionIsMet:                                             {VerifyBackupDeletionWhenRetentionIsMetLabel, vanillaPipeline, SystemTest, PxBackupLabel, P0, PxLabel, S3BackupLocationLabel, AzureBackupLocationLabel, Day0LockedBucketLabel, S3LockedBucket, AzureImmutableBucket},
	DeleteVerifyBackupDeletionWhenRetentionIsMet:                                       {DeleteVerifyBackupDeletionWhenRetentionIsMetLabel, vanillaPipeline, SystemTest, PxBackupLabel, P0, PxLabel, S3BackupLocationLabel, AzureBackupLocationLabel, Day3LockedBucketLabel, S3LockedBucket, AzureImmutableBucket},
	VerifyBackupAutoDeletionWhenNewPVCsAreAddedBetweenSchedules:                        {VerifyBackupAutoDeletionWhenNewPVCsAreAddedBetweenSchedulesLabel, vanillaPipeline, SystemTest, PxBackupLabel, P1, PxLabel, S3BackupLocationLabel, AzureBackupLocationLabel, Day0LockedBucketLabel, S3LockedBucket, AzureImmutableBucket},
	DeleteVerifyBackupAutoDeletionWhenNewPVCsAreAddedBetweenSchedules:                  {DeleteVerifyBackupAutoDeletionWhenNewPVCsAreAddedBetweenSchedulesLabel, vanillaPipeline, SystemTest, PxBackupLabel, P1, PxLabel, S3BackupLocationLabel, AzureBackupLocationLabel, Day3LockedBucketLabel, S3LockedBucket, AzureImmutableBucket},
}
