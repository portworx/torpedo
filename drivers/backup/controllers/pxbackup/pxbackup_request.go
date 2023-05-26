package pxbackup

import (
	"context"
	"fmt"
	api "github.com/portworx/px-backup-api/pkg/apis/v1"
	"github.com/portworx/torpedo/drivers/backup"
	"github.com/portworx/torpedo/drivers/backup/utils"
	"github.com/portworx/torpedo/tests"
)

func (p *PxbController) processPxBackupRequest(request interface{}) (response interface{}, err error) {
	var ctx context.Context
	if p.profile.isAdmin {
		ctx, err = backup.GetAdminCtxFromSecret()
		if err != nil {
			return nil, utils.ProcessError(err)
		}
	} else {
		ctx, err = backup.GetNonAdminCtx(p.profile.username, p.profile.password)
		if err != nil {
			debugMessage := fmt.Sprintf("profile: username [%s]", p.profile.username)
			return nil, utils.ProcessError(err, debugMessage)
		}
	}
	switch request.(type) {
	// CloudCredential
	case *api.CloudCredentialCreateRequest:
		response, err = tests.Inst().Backup.CreateCloudCredential(ctx, request.(*api.CloudCredentialCreateRequest))
	case *api.CloudCredentialUpdateRequest:
		response, err = tests.Inst().Backup.UpdateCloudCredential(ctx, request.(*api.CloudCredentialUpdateRequest))
	case *api.CloudCredentialInspectRequest:
		response, err = tests.Inst().Backup.InspectCloudCredential(ctx, request.(*api.CloudCredentialInspectRequest))
	case *api.CloudCredentialEnumerateRequest:
		response, err = tests.Inst().Backup.EnumerateCloudCredential(ctx, request.(*api.CloudCredentialEnumerateRequest))
	case *api.CloudCredentialDeleteRequest:
		response, err = tests.Inst().Backup.DeleteCloudCredential(ctx, request.(*api.CloudCredentialDeleteRequest))
	case *api.CloudCredentialOwnershipUpdateRequest:
		response, err = tests.Inst().Backup.UpdateOwnershipCloudCredential(ctx, request.(*api.CloudCredentialOwnershipUpdateRequest))
	// Cluster
	case *api.ClusterCreateRequest:
		response, err = tests.Inst().Backup.CreateCluster(ctx, request.(*api.ClusterCreateRequest))
	case *api.ClusterUpdateRequest:
		response, err = tests.Inst().Backup.UpdateCluster(ctx, request.(*api.ClusterUpdateRequest))
	case *api.ClusterEnumerateRequest:
		response, err = tests.Inst().Backup.EnumerateCluster(ctx, request.(*api.ClusterEnumerateRequest))
	case *api.ClusterInspectRequest:
		response, err = tests.Inst().Backup.InspectCluster(ctx, request.(*api.ClusterInspectRequest))
	case *api.ClusterDeleteRequest:
		response, err = tests.Inst().Backup.DeleteCluster(ctx, request.(*api.ClusterDeleteRequest))
	case *api.ClusterBackupShareUpdateRequest:
		response, err = tests.Inst().Backup.ClusterUpdateBackupShare(ctx, request.(*api.ClusterBackupShareUpdateRequest))
	// BLocation
	case *api.BackupLocationCreateRequest:
		response, err = tests.Inst().Backup.CreateBackupLocation(ctx, request.(*api.BackupLocationCreateRequest))
	case *api.BackupLocationUpdateRequest:
		response, err = tests.Inst().Backup.UpdateBackupLocation(ctx, request.(*api.BackupLocationUpdateRequest))
	case *api.BackupLocationEnumerateRequest:
		response, err = tests.Inst().Backup.EnumerateBackupLocation(ctx, request.(*api.BackupLocationEnumerateRequest))
	case *api.BackupLocationInspectRequest:
		response, err = tests.Inst().Backup.InspectBackupLocation(ctx, request.(*api.BackupLocationInspectRequest))
	case *api.BackupLocationDeleteRequest:
		response, err = tests.Inst().Backup.DeleteBackupLocation(ctx, request.(*api.BackupLocationDeleteRequest))
	case *api.BackupLocationValidateRequest:
		response, err = tests.Inst().Backup.ValidateBackupLocation(ctx, request.(*api.BackupLocationValidateRequest))
	case *api.BackupLocationOwnershipUpdateRequest:
		response, err = tests.Inst().Backup.UpdateOwnershipBackupLocation(ctx, request.(*api.BackupLocationOwnershipUpdateRequest))
	// Backup
	case *api.BackupCreateRequest:
		response, err = tests.Inst().Backup.CreateBackup(ctx, request.(*api.BackupCreateRequest))
	case *api.BackupUpdateRequest:
		response, err = tests.Inst().Backup.UpdateBackup(ctx, request.(*api.BackupUpdateRequest))
	case *api.BackupEnumerateRequest:
		response, err = tests.Inst().Backup.EnumerateBackup(ctx, request.(*api.BackupEnumerateRequest))
	case *api.BackupInspectRequest:
		response, err = tests.Inst().Backup.InspectBackup(ctx, request.(*api.BackupInspectRequest))
	case *api.BackupDeleteRequest:
		response, err = tests.Inst().Backup.DeleteBackup(ctx, request.(*api.BackupDeleteRequest))
	// Restore
	case *api.RestoreCreateRequest:
		response, err = tests.Inst().Backup.CreateRestore(ctx, request.(*api.RestoreCreateRequest))
	case *api.RestoreUpdateRequest:
		response, err = tests.Inst().Backup.UpdateRestore(ctx, request.(*api.RestoreUpdateRequest))
	case *api.RestoreEnumerateRequest:
		response, err = tests.Inst().Backup.EnumerateRestore(ctx, request.(*api.RestoreEnumerateRequest))
	case *api.RestoreInspectRequest:
		response, err = tests.Inst().Backup.InspectRestore(ctx, request.(*api.RestoreInspectRequest))
	case *api.RestoreDeleteRequest:
		response, err = tests.Inst().Backup.DeleteRestore(ctx, request.(*api.RestoreDeleteRequest))
	// SchedulePolicy
	case *api.SchedulePolicyCreateRequest:
		response, err = tests.Inst().Backup.CreateSchedulePolicy(ctx, request.(*api.SchedulePolicyCreateRequest))
	case *api.SchedulePolicyUpdateRequest:
		response, err = tests.Inst().Backup.UpdateSchedulePolicy(ctx, request.(*api.SchedulePolicyUpdateRequest))
	case *api.SchedulePolicyEnumerateRequest:
		response, err = tests.Inst().Backup.EnumerateSchedulePolicy(ctx, request.(*api.SchedulePolicyEnumerateRequest))
	case *api.SchedulePolicyInspectRequest:
		response, err = tests.Inst().Backup.InspectSchedulePolicy(ctx, request.(*api.SchedulePolicyInspectRequest))
	case *api.SchedulePolicyDeleteRequest:
		response, err = tests.Inst().Backup.DeleteSchedulePolicy(ctx, request.(*api.SchedulePolicyDeleteRequest))
	case *api.SchedulePolicyOwnershipUpdateRequest:
		response, err = tests.Inst().Backup.UpdateOwnershiSchedulePolicy(ctx, request.(*api.SchedulePolicyOwnershipUpdateRequest))
	// BackupSchedule
	case *api.BackupScheduleCreateRequest:
		response, err = tests.Inst().Backup.CreateBackupSchedule(ctx, request.(*api.BackupScheduleCreateRequest))
	case *api.BackupScheduleUpdateRequest:
		response, err = tests.Inst().Backup.UpdateBackupSchedule(ctx, request.(*api.BackupScheduleUpdateRequest))
	case *api.BackupScheduleEnumerateRequest:
		response, err = tests.Inst().Backup.EnumerateBackupSchedule(ctx, request.(*api.BackupScheduleEnumerateRequest))
	case *api.BackupScheduleInspectRequest:
		response, err = tests.Inst().Backup.InspectBackupSchedule(ctx, request.(*api.BackupScheduleInspectRequest))
	case *api.BackupScheduleDeleteRequest:
		response, err = tests.Inst().Backup.DeleteBackupSchedule(ctx, request.(*api.BackupScheduleDeleteRequest))
	// License
	case *api.LicenseActivateRequest:
		response, err = tests.Inst().Backup.ActivateLicense(ctx, request.(*api.LicenseActivateRequest))
	case *api.LicenseInspectRequest:
		response, err = tests.Inst().Backup.InspectLicense(ctx, request.(*api.LicenseInspectRequest))
	// Rule
	case *api.RuleCreateRequest:
		response, err = tests.Inst().Backup.CreateRule(ctx, request.(*api.RuleCreateRequest))
	case *api.RuleUpdateRequest:
		response, err = tests.Inst().Backup.UpdateRule(ctx, request.(*api.RuleUpdateRequest))
	case *api.RuleEnumerateRequest:
		response, err = tests.Inst().Backup.EnumerateRule(ctx, request.(*api.RuleEnumerateRequest))
	case *api.RuleInspectRequest:
		response, err = tests.Inst().Backup.InspectRule(ctx, request.(*api.RuleInspectRequest))
	case *api.RuleDeleteRequest:
		response, err = tests.Inst().Backup.DeleteRule(ctx, request.(*api.RuleDeleteRequest))
	case *api.RuleOwnershipUpdateRequest:
		response, err = tests.Inst().Backup.UpdateOwnershipRule(ctx, request.(*api.RuleOwnershipUpdateRequest))
	default:
		err = fmt.Errorf("unsupported request [%v] of type [%T] for px-backup", request, request)
		return nil, utils.ProcessError(err)
	}
	if err != nil {
		debugMessage := fmt.Sprintf("request: [%v]; profile: username [%s], is-admin [%t]; context-context: [%v]", request, p.profile.username, p.profile.isAdmin, ctx)
		return nil, utils.ProcessError(err, debugMessage)
	}
	return response, nil
}
