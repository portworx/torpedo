package tests

import (
	"fmt"
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"

	"github.com/portworx/sched-ops/k8s/apps"
	"github.com/portworx/torpedo/drivers/node"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/pkg/testrailuttils"
	. "github.com/portworx/torpedo/tests"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestBasic(t *testing.T) {
	RegisterFailHandler(Fail)

	var specReporters []Reporter
	junitReporter := reporters.NewJUnitReporter("/testresults/junit_secrets.xml")
	specReporters = append(specReporters, junitReporter)
	RunSpecsWithDefaultAndCustomReporters(t, "Torpedo : Secrets", specReporters)
}

var _ = BeforeSuite(func() {
	InitInstance()
})

var _ = Describe("{SecretsVaultFunctional}", func() {
	var testrailID, runID int
	var contexts []*scheduler.Context

	const (
		secretsProvider       = "vault"
		portworxContainerName = "portworx"
	)

	BeforeEach(func() {
		runID = testrailuttils.AddRunsToMilestone(testrailID)
		isOpBased, _ := Inst().V.IsOperatorBasedInstall()
		if !isOpBased {
			k8sApps := apps.Instance()
			deployments, err := k8sApps.ListDaemonSets("kube-system", metav1.ListOptions{
				LabelSelector: "name=portworx",
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(len(deployments)).NotTo(Equal(0))
			Expect(deployments[0].Spec.Template.Spec.Containers).NotTo(BeEmpty())
			usingVault := false
			for _, container := range deployments[0].Spec.Template.Spec.Containers {
				if container.Name == portworxContainerName {
					for _, arg := range container.Args {
						if arg == secretsProvider {
							usingVault = true
						}
					}
				}
			}
			if !usingVault {
				Skip(fmt.Sprintf("Skip test for not using %s", secretsProvider))
			}
		} else {
			spec, err := Inst().V.GetStorageCluster()
			Expect(err).ToNot(HaveOccurred())
			if *spec.Spec.SecretsProvider != secretsProvider {
				Skip(fmt.Sprintf("Skip test for not using %s", secretsProvider))
			}
		}
	})

	// This test performs basic test of starting an application and destroying it (along with storage)
	var _ = Describe("{RunSecretsLogin}", func() {
		var testrailID = 82774
		// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/82774
		var runID int

		It("has to runs secrets login for vault", func() {
			contexts = make([]*scheduler.Context, 0)
			// assumes the rold id and secret id are passed in from parameters
			// so does the
			n := node.GetWorkerNodes()[0]
			err := Inst().V.RunSecretsLogin(n, secretsProvider)
			Expect(err).ToNot(HaveOccurred())

		})
		JustAfterEach(func() {
			AfterEachTest(contexts, testrailID, runID)
		})
	})

	AfterEach(func() {
		AfterEachTest(contexts, testrailID, runID)
	})
})

var _ = AfterSuite(func() {
	PerformSystemCheck()
	ValidateCleanup()
})

func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags
	ParseFlags()
	os.Exit(m.Run())
}
