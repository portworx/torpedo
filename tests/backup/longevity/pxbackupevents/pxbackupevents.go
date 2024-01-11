package pxbackupevents

import (
	"context"
	"fmt"
	"time"

	"github.com/portworx/torpedo/drivers/backup"
	"github.com/portworx/torpedo/drivers/scheduler"
	. "github.com/portworx/torpedo/tests"
	. "github.com/portworx/torpedo/tests/backup/longevity/pxbackuplongevitytypes"
)

const (
	cloudAccountDeleteTimeout                 = 5 * time.Minute
	cloudAccountDeleteRetryTime               = 30 * time.Second
	storkDeploymentName                       = "stork"
	defaultStorkDeploymentNamespace           = "kube-system"
	upgradeStorkImage                         = "TARGET_STORK_VERSION"
	latestStorkImage                          = "23.9.0"
	restoreNamePrefix                         = "tp-restore"
	destinationClusterName                    = "destination-cluster"
	appReadinessTimeout                       = 10 * time.Minute
	taskNamePrefix                            = "pxb"
	orgID                                     = "default"
	usersToBeCreated                          = "USERS_TO_CREATE"
	groupsToBeCreated                         = "GROUPS_TO_CREATE"
	maxUsersInGroup                           = "MAX_USERS_IN_GROUP"
	maxBackupsToBeCreated                     = "MAX_BACKUPS"
	errorChannelSize                          = 50
	maxWaitPeriodForBackupCompletionInMinutes = 40
	maxWaitPeriodForRestoreCompletionInMinute = 40
	maxWaitPeriodForBackupJobCancellation     = 20
	maxWaitPeriodForRestoreJobCancellation    = 20
	restoreJobCancellationRetryTime           = 30
	restoreJobProgressRetryTime               = 1
	backupJobCancellationRetryTime            = 5
	K8sNodeReadyTimeout                       = 10
	K8sNodeRetryInterval                      = 30
	globalAWSBucketPrefix                     = "global-aws"
	globalAzureBucketPrefix                   = "global-azure"
	globalGCPBucketPrefix                     = "global-gcp"
	globalNFSBucketPrefix                     = "global-nfs"
	globalAWSLockedBucketPrefix               = "global-aws-locked"
	globalAzureLockedBucketPrefix             = "global-azure-locked"
	globalGCPLockedBucketPrefix               = "global-gcp-locked"
	mongodbStatefulset                        = "pxc-backup-mongodb"
	pxBackupDeployment                        = "px-backup"
	backupDeleteTimeout                       = 60 * time.Minute
	backupDeleteRetryTime                     = 30 * time.Second
	backupLocationDeleteTimeout               = 60 * time.Minute
	backupLocationDeleteRetryTime             = 30 * time.Second
	rebootNodeTimeout                         = 1 * time.Minute
	rebootNodeTimeBeforeRetry                 = 5 * time.Second
	latestPxBackupVersion                     = "2.6.0"
	defaultPxBackupHelmBranch                 = "master"
	pxCentralPostInstallHookJobName           = "pxcentral-post-install-hook"
	quickMaintenancePod                       = "quick-maintenance-repo"
	fullMaintenancePod                        = "full-maintenance-repo"
	jobDeleteTimeout                          = 5 * time.Minute
	jobDeleteRetryTime                        = 10 * time.Second
	podStatusTimeOut                          = 20 * time.Minute
	podStatusRetryTime                        = 30 * time.Second
	licenseCountUpdateTimeout                 = 15 * time.Minute
	licenseCountUpdateRetryTime               = 1 * time.Minute
	podReadyTimeout                           = 15 * time.Minute
	storkPodReadyTimeout                      = 20 * time.Minute
	podReadyRetryTime                         = 30 * time.Second
	namespaceDeleteTimeout                    = 10 * time.Minute
	clusterCreationTimeout                    = 5 * time.Minute
	clusterCreationRetryTime                  = 10 * time.Second
	clusterDeleteTimeout                      = 10 * time.Minute
	clusterDeleteRetryTime                    = 5 * time.Second
	vmStartStopTimeout                        = 10 * time.Minute
	vmStartStopRetryTime                      = 30 * time.Second
	cloudCredConfigMap                        = "cloud-config"
	volumeSnapshotClassEnv                    = "VOLUME_SNAPSHOT_CLASS"
	rancherActiveCluster                      = "local"
	rancherProjectDescription                 = "new project"
	multiAppNfsPodDeploymentNamespace         = "kube-system"
	backupScheduleDeleteTimeout               = 60 * time.Minute
	backupScheduleDeleteRetryTime             = 30 * time.Second
)

var (
	// User should keep updating preRuleApp, postRuleApp, appsWithCRDsAndWebhooks
	preRuleApp                  = []string{"cassandra", "postgres"}
	postRuleApp                 = []string{"cassandra"}
	appsWithCRDsAndWebhooks     = []string{"elasticsearch-crd-webhook"} // The apps which have CRDs and webhooks
	globalAWSBucketName         string
	globalAzureBucketName       string
	globalGCPBucketName         string
	globalNFSBucketName         string
	globalAWSLockedBucketName   string
	globalAzureLockedBucketName string
	globalGCPLockedBucketName   string
	cloudProviders              = []string{"aws"}
	commonPassword              string
	backupPodLabels             = []map[string]string{
		{"app": "px-backup"}, {"app.kubernetes.io/component": "pxcentral-apiserver"},
		{"app.kubernetes.io/component": "pxcentral-backend"},
		{"app.kubernetes.io/component": "pxcentral-frontend"},
		{"app.kubernetes.io/component": "keycloak"},
		{"app.kubernetes.io/component": "pxcentral-lh-middleware"},
		{"app.kubernetes.io/component": "pxcentral-mysql"}}
	cloudPlatformList          = []string{"rke", "aws", "azure", "gke"}
	nfsBackupExecutorPodLabel  = map[string]string{"kdmp.portworx.com/driver-name": "nfsbackup"}
	nfsRestoreExecutorPodLabel = map[string]string{"kdmp.portworx.com/driver-name": "nfsrestore"}
)

type userRoleAccess struct {
	user     string
	roles    backup.PxBackupRole
	accesses BackupAccess
	context  context.Context
}

type userAccessContext struct {
	user     string
	accesses BackupAccess
	context  context.Context
}

var backupAccessKeyValue = map[BackupAccess]string{
	1: "ViewOnlyAccess",
	2: "RestoreAccess",
	3: "FullAccess",
}

var storkLabel = map[string]string{
	"name": "stork",
}

type BackupAccess int32

type ReplacePolicyType int32

const (
	ReplacePolicyInvalid ReplacePolicyType = 0
	ReplacePolicyRetain  ReplacePolicyType = 1
	ReplacePolicyDelete  ReplacePolicyType = 2
)

const (
	ViewOnlyAccess BackupAccess = 1
	RestoreAccess               = 2
	FullAccess                  = 3
)

type ExecutionMode int32

const (
	Sequential ExecutionMode = iota
	Parallel
)

var (
	// AppRuleMaster is a map of struct for all the value for rules
	// This map needs to be updated for new applications as and whe required
	AppRuleMaster = map[string]backup.AppRule{
		"cassandra": {
			PreRule: backup.PreRule{
				Rule: backup.RuleSpec{
					ActionList: []string{"nodetool flush -- keyspace1;", "echo 'test"}, PodSelectorList: []string{"app=cassandra", "app=cassandra1"}, Background: []string{"false", "false"}, RunInSinglePod: []string{"false", "false"}, Container: []string{},
				},
			},
			PostRule: backup.PostRule{
				Rule: backup.RuleSpec{
					ActionList: []string{"nodetool verify -- keyspace1;", "nodetool verify -- keyspace1;"}, PodSelectorList: []string{"app=cassandra", "app=cassandra1"}, Background: []string{"false", "false"}, RunInSinglePod: []string{"false", "false"}, Container: []string{},
				},
			},
		},
		"mysql": {
			PreRule: backup.PreRule{
				Rule: backup.RuleSpec{
					ActionList: []string{"mysql --user=root --password=$MYSQL_ROOT_PASSWORD -Bse 'FLUSH TABLES WITH READ LOCK;system ${WAIT_CMD};'"}, PodSelectorList: []string{"app=mysql"}, Background: []string{"true"}, RunInSinglePod: []string{"false"}, Container: []string{},
				},
			},
			PostRule: backup.PostRule{
				Rule: backup.RuleSpec{
					ActionList: []string{"mysql --user=root --password=$MYSQL_ROOT_PASSWORD -Bse 'FLUSH LOGS; UNLOCK TABLES;'"}, PodSelectorList: []string{"app=mysql"}, Background: []string{"false"}, RunInSinglePod: []string{"false"}, Container: []string{},
				},
			},
		},
		"mysql-backup": {
			PreRule: backup.PreRule{
				Rule: backup.RuleSpec{
					ActionList: []string{"mysql --user=root --password=$MYSQL_ROOT_PASSWORD -Bse 'FLUSH TABLES WITH READ LOCK;system ${WAIT_CMD};'"}, PodSelectorList: []string{"app=mysql"}, Background: []string{"true"}, RunInSinglePod: []string{"false"}, Container: []string{},
				},
			},
			PostRule: backup.PostRule{
				Rule: backup.RuleSpec{
					ActionList: []string{"mysql --user=root --password=$MYSQL_ROOT_PASSWORD -Bse 'FLUSH LOGS; UNLOCK TABLES;'"}, PodSelectorList: []string{"app=mysql"}, Background: []string{"false"}, RunInSinglePod: []string{"false"}, Container: []string{},
				},
			},
		},
		"postgres": {
			PreRule: backup.PreRule{
				Rule: backup.RuleSpec{
					ActionList: []string{"PGPASSWORD=$POSTGRES_PASSWORD; psql -U \"$POSTGRES_USER\" -c \"CHECKPOINT\""}, PodSelectorList: []string{"app=postgres"}, Background: []string{"false"}, RunInSinglePod: []string{"false"}, Container: []string{},
				},
			},
		},
		"postgres-backup": {
			PreRule: backup.PreRule{
				Rule: backup.RuleSpec{
					ActionList: []string{"PGPASSWORD=$POSTGRES_PASSWORD; psql -U \"$POSTGRES_USER\" -c \"CHECKPOINT\""}, PodSelectorList: []string{"app=postgres"}, Background: []string{"false"}, RunInSinglePod: []string{"false"}, Container: []string{},
				},
			},
		},
		"postgres-csi": {
			PreRule: backup.PreRule{
				Rule: backup.RuleSpec{
					ActionList: []string{"PGPASSWORD=$POSTGRES_PASSWORD; psql -U \"$POSTGRES_USER\" -c \"CHECKPOINT\""}, PodSelectorList: []string{"app=postgres"}, Background: []string{"false"}, RunInSinglePod: []string{"false"}, Container: []string{},
				},
			},
		},
	}
)

const (
	EventBuilder1                    = "EventBuilder1"
	EventBuilder1Fail                = "EventBuilder1Fail"
	EventScheduleApps                = "eventScheduleApps"
	EventValidateScheduleApplication = "EventValidateScheduleApplication"
)

type PxBackupEventBuilder func(*PxBackupLongevity) (error, string, EventData)

var AllBuilders = map[string]PxBackupEventBuilder{
	EventBuilder1:                    eventBuilder1,
	EventBuilder1Fail:                eventBuilder1Fail,
	EventScheduleApps:                eventScheduleApps,
	EventValidateScheduleApplication: eventValidateScheduleApplication,
}

func eventScheduleApps(inputsForEventBuilder *PxBackupLongevity) (error, string, EventData) {
	eventData := &EventData{}

	scheduledAppContexts := make([]*scheduler.Context, 0)
	var bkpNamespaces = make([]string, 0)

	for i := 0; i < Inst().GlobalScaleFactor; i++ {
		taskName := fmt.Sprintf("%s-%d-%s", taskNamePrefix, i, RandomString(5))
		appContexts := ScheduleApplications(taskName)
		for _, ctx := range appContexts {
			ctx.ReadinessTimeout = appReadinessTimeout
			namespace := GetAppNamespace(ctx, taskName)
			bkpNamespaces = append(bkpNamespaces, namespace)
			scheduledAppContexts = append(scheduledAppContexts, ctx)
		}
	}

	eventData.BackupNamespaces = bkpNamespaces
	eventData.SchedulerContext = scheduledAppContexts

	return nil, "", *eventData
}

func eventValidateScheduleApplication(inputsForEventBuilder *PxBackupLongevity) (error, string, EventData) {
	eventData := &EventData{}
	ValidateApplications(inputsForEventBuilder.ApplicationData.SchedulerContext)
	return nil, "", *eventData
}

func eventBuilder1(inputsForEventBuilder *PxBackupLongevity) (error, string, EventData) {
	eventData := &EventData{}
	time.Sleep(time.Second * time.Duration(inputsForEventBuilder.CustomData.Integers["timeToBlock"]))
	return nil, "", *eventData
}

func eventBuilder1Fail(inputsForEventBuilder *PxBackupLongevity) (error, string, EventData) {
	eventData := &EventData{}
	time.Sleep(time.Second * time.Duration(inputsForEventBuilder.CustomData.Integers["timeToBlock"]))
	return fmt.Errorf("This is the returned error"), "This is the highlight event from - EventBuilder1Fail", *eventData
}

func RunBuilder(eventBuilderName string, inputsForEventBuilder *PxBackupLongevity, eventResponse *EventResponse) {
	eventBuilder := AllBuilders[eventBuilderName]
	eventBuilderIdentifier := eventBuilderName + "-" + time.Now().Format("15:04:05.000")
	eventResponse.EventBuilders[eventBuilderIdentifier] = &EventBuilderResponse{}

	startTime := time.Now()

	err, highlightEvent, eventData := eventBuilder(inputsForEventBuilder)
	if err != nil {
		eventResponse.EventBuilders[eventBuilderIdentifier].Error = err
	}
	if highlightEvent != "" {
		eventResponse.EventBuilders[eventBuilderIdentifier].HighlightEvent = highlightEvent
	}
	eventResponse.EventBuilders[eventBuilderIdentifier].EventData = eventData
	eventResponse.EventBuilders[eventBuilderIdentifier].TimeTakenInMinutes = float32(time.Since(startTime).Minutes())

}
