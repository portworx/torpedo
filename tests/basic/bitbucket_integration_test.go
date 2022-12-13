package tests

import (
	"fmt"
	"github.com/portworx/torpedo/pkg/osutils"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	http2 "github.com/go-git/go-git/v5/plumbing/transport/http"
	apapi "github.com/libopenstorage/autopilot-api/pkg/apis/autopilot/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/pkg/aututils"
	. "github.com/portworx/torpedo/tests"
)

const (
	directory = "/tmp/autopilot/"
)

// This testsuite is used for performing basic scenarios with Autopilot rules where it
// schedules apps and wait until workload is completed on the volumes and then validates
// PVC sizes of the volumes
var _ = Describe(fmt.Sprintf("{%sGitopsBasic}", testSuiteName), func() {

	var namespace string
	bitbucket, err := initFuncBitBucket(GitOpsConfig{
		Name: "gitops",
		Type: "bitbucket-scm",
		Params: map[string]interface{}{
			"defaultReviewers": []string{"vc-cnbu-devops"},
			"user":             "svc-cnbu-devops",
			"repo":             "autopilot-torpedo-bb",
			"folder":           "workloads",
			"baseUrl":          "https://bitbucket-staging.dev.purestorage.com/",
			"projectKey":       "PXAUT",
			"branch":           "master",
		},
	}, "kube-system")

	Expect(err).NotTo(HaveOccurred())

	It("has to create a volume, approve PR and check if volume has been resized", func() {
		var contexts []*scheduler.Context
		testName := strings.ToLower(fmt.Sprintf("%sGitopsBasic", testSuiteName))

		apRules := []apapi.AutopilotRule{
			aututils.PVCRuleByTotalSizeApprovalRequired(25, 150, "20Gi"),
		}
		apRules[0].Spec.ActionsCoolDownPeriod = int64(10)

		Step("schedule applications", func() {
			taskName := fmt.Sprintf("%s-aprule%s", testName, GenerateUUID())
			namespace = taskName
			apRules[0].Name = fmt.Sprintf("%s", apRules[0].Name)
			labels := map[string]string{
				"autopilot": apRules[0].Name,
				"app":       "postgres",
			}
			apRules[0].Spec.ActionsCoolDownPeriod = int64(60)
			context, err := Inst().S.Schedule(taskName, scheduler.ScheduleOptions{
				AppKeys:            Inst().AppList,
				StorageProvisioner: Inst().Provisioner,
				AutopilotRule:      apRules[0],
				Labels:             labels,
				Namespace:          namespace,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(context).NotTo(BeEmpty())
			contexts = append(contexts, context...)
		})

		Step("check if approval has been created", func() {
			// wait for event Normal or Initializing => Triggered (sometimes autopilot triggers from Initializing state)
			err := aututils.WaitForAutopilotEvent(apRules[0], "", []string{aututils.AnyToTriggeredEvent})
			Expect(err).NotTo(HaveOccurred())

			// wait for event
			err = aututils.WaitForAutopilotEvent(apRules[0], "", []string{aututils.TriggeredToActionAwaitingApprovalEvent})
			Expect(err).NotTo(HaveOccurred())

			err = aututils.WaitForActionApprovalsObjects(namespace, "")
			Expect(err).NotTo(HaveOccurred())

			actionApprovalList, err := Inst().S.ListActionApprovals(namespace)
			Expect(err).NotTo(HaveOccurred())
			Expect(actionApprovalList.Items).NotTo(BeEmpty())
		})

		Step("check if PR is out", func() {
			// vendor bitbucket library and get list of prs
			// check if pr is out for this particular rule
			actionApprovalList, err := Inst().S.ListActionApprovals(namespace)
			Expect(err).NotTo(HaveOccurred())
			Expect(actionApprovalList.Items).NotTo(BeEmpty())
			for i := 0; i < 10; i++ {
				for _, approval := range actionApprovalList.Items {
					pr, err := bitbucket.getLastPRForApproval(&approval)
					if err != nil {
						logrus.Infof("error getting last PR for approval: %v", err)
						continue
					}
					if pr != nil {
						logrus.Infof("PR found: %v", pr.Title)
						return
					}
				}
				time.Sleep(time.Second * 4)
			}
		})

		Step("approve and merge the PR", func() {
			// vendor bitbucket library and get list of prs
			// check if pr is out for this particular rule
			actionApprovalList, err := Inst().S.ListActionApprovals(namespace)
			Expect(err).NotTo(HaveOccurred())
			Expect(actionApprovalList.Items).NotTo(BeEmpty())

			for _, approval := range actionApprovalList.Items {
				pr, err := bitbucket.getLastPRForApproval(&approval)
				if err != nil {
					Expect(err).NotTo(HaveOccurred())
				}
				if pr != nil {
					logrus.Infof("attempting to merge RP: %v", pr.Title)
					err = bitbucket.mergePR(pr)
					Expect(err).NotTo(HaveOccurred())
				}
			}
		})
		password, err := getPasswordFromSecret(namespace, BitBucketPasswordEnvKey)
		Expect(err).NotTo(HaveOccurred())

		Step("applying merged specs", func() {
			_, err := git.PlainClone(directory, false, &git.CloneOptions{
				RemoteName: "origin",
				Auth: &http2.BasicAuth{
					Username: "svc-cnbu-devops",
					Password: password,
				},
				URL:        bitbucket.buildRepoURL(),
				Progress:   os.Stdout,
				NoCheckout: false,
			})
			Expect(err).NotTo(HaveOccurred())
			workloadsDir := filepath.Join(directory, "workloads")
			_, err = os.Stat(workloadsDir)
			Expect(err).NotTo(HaveOccurred())
			pathToApply := fmt.Sprintf("%s/.", workloadsDir)
			cmdArgs := []string{"apply", "-f", pathToApply}
			err = osutils.Kubectl(cmdArgs)
			Expect(err).NotTo(HaveOccurred())
		})

		Step("wait action awaiting approval to action pending", func() {
			err := aututils.WaitForAutopilotEvent(apRules[0], "", []string{aututils.ActionAwaitingApprovalToActiveActionsPending})
			Expect(err).NotTo(HaveOccurred())
		})
		Step("wait action pending to action in progress", func() {
			err := aututils.WaitForAutopilotEvent(apRules[0], "", []string{aututils.ActiveActionsPendingToActiveActionsInProgress})
			Expect(err).NotTo(HaveOccurred())
		})

		Step("wait action in progress to action taken", func() {
			err := aututils.WaitForAutopilotEvent(apRules[0], "", []string{aututils.ActiveActionsInProgressToActiveActionsTaken})
			Expect(err).NotTo(HaveOccurred())
		})

		Step("validating volumes and verifying size of volumes", func() {
			for _, ctx := range contexts {
				ValidateVolumes(ctx)
			}
		})

		Step("make sure approval is gone", func() {
			time.Sleep(2 * time.Minute)
		})

		Step("destroy apps", func() {
			opts := make(map[string]bool)
			opts[scheduler.OptionsWaitForResourceLeakCleanup] = true
			for _, ctx := range contexts {
				TearDownContext(ctx, opts)
			}
		})
	})
})
