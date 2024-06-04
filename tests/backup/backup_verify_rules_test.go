package tests

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	"github.com/portworx/torpedo/drivers/backup"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
)

var _ = Describe("{VerifyRulesUpdate}", Label(TestCaseLabelsMap[VerifyRulesUpdate]...), func() {

	JustBeforeEach(func() {
		log.InfoD("No pre-configuration needed")
	})

	It("should verify rules update", func() {
		Step(fmt.Sprintf("Creation of pre and post exec rules for applications "), func() {
			log.InfoD("Creation of pre and post exec rules for applications ")
			ctx, err := backup.GetAdminCtxFromSecret()
			preRuleName, postRuleName, err := CreateRuleForBackupWithMultipleApplications(BackupOrgID, Inst().AppList, ctx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of pre and post exec rules for applications from px-admin"))
			if preRuleName != "" {
				preRuleUid, err := Inst().Backup.GetRuleUid(BackupOrgID, ctx, preRuleName)
				log.FailOnError(err, "Fetching pre backup rule [%s] uid", preRuleName)
				log.InfoD("Pre backup rule [%s] uid: [%s]", preRuleName, preRuleUid)
			}
			if postRuleName != "" {
				postRuleUid, err := Inst().Backup.GetRuleUid(BackupOrgID, ctx, postRuleName)
				log.FailOnError(err, "Fetching post backup rule [%s] uid", postRuleName)
				log.InfoD("Post backup rule [%s] uid: [%s]", postRuleName, postRuleUid)
			}
		})
	})

	JustAfterEach(func() {
		log.InfoD("No post-configuration needed")
	})
})
