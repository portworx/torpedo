package tests

import (
	"fmt"
	"os"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	"github.com/portworx/sched-ops/k8s/core"
	"github.com/portworx/torpedo/drivers/node"
	"github.com/portworx/torpedo/drivers/scheduler"
	torpedovolume "github.com/portworx/torpedo/drivers/volume"
	"github.com/portworx/torpedo/pkg/jirautils"
	. "github.com/portworx/torpedo/tests"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
)

func TestTelemetryBasic(t *testing.T) {
	RegisterFailHandler(Fail)

	var specReporters []Reporter
	junitReporter := reporters.NewJUnitReporter("/testresults/junit_telemetry.xml")
	specReporters = append(specReporters, junitReporter)
	RunSpecsWithDefaultAndCustomReporters(t, "Torpedo : Telemetry", specReporters)
}

var _ = BeforeSuite(func() {
	InitInstance()
})

// // This test performs basic test of starting an application and destroying it (along with storage)
// var _ = Describe("{DiagsBasic}", func() {
// 	var contexts []*scheduler.Context
// 	It("has to setup, validate, try to get diags on nodes and teardown apps", func() {
// 		contexts = make([]*scheduler.Context, 0)
// 		for i := 0; i < Inst().GlobalScaleFactor; i++ {
// 			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("diagsbasic-%d", i))...)
// 		}

// 		ValidateApplications(contexts)

// 		// One node at a time, collect diags and verify in S3
// 		for _, currNode := range node.GetWorkerNodes() {
// 			Step(fmt.Sprintf("collect diags on node: %s | %s", currNode.Name, currNode.Type), func() {

// 				config := &torpedovolume.DiagRequestConfig{
// 					DockerHost:    "unix:///var/run/docker.sock",
// 					OutputFile:    fmt.Sprintf("/var/cores/torpedo-diagsbasic-%s-%d.tar.gz", currNode.Name, time.Now().Unix()),
// 					ContainerName: "",
// 					OnHost:        true,
// 				}
// 				err := Inst().V.CollectDiags(currNode, config, torpedovolume.DiagOps{Validate: true})
// 				Expect(err).NotTo(HaveOccurred())
// 			})
// 		}

// 		for _, ctx := range contexts {
// 			TearDownContext(ctx, nil)
// 		}
// 	})
// 	JustAfterEach(func() {
// 		AfterEachTest(contexts)
// 	})
// })

// // This test performs basic test of starting an application and destroying it (along with storage)
// var _ = Describe("{DiagsAsyncBasic}", func() {
// 	var contexts []*scheduler.Context
// 	It("has to setup, validate, try to get a-sync diags on nodes and teardown apps", func() {
// 		contexts = make([]*scheduler.Context, 0)
// 		for i := 0; i < Inst().GlobalScaleFactor; i++ {
// 			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("diagsasyncbasic-%d", i))...)
// 		}

// 		ValidateApplications(contexts)

// 		// One node at a time, collect diags and verify in S3
// 		for _, currNode := range node.GetWorkerNodes() {
// 			Step(fmt.Sprintf("collect diags on node: %s", currNode.Name), func() {

// 				config := &torpedovolume.DiagRequestConfig{
// 					DockerHost:    "unix:///var/run/docker.sock",
// 					OutputFile:    fmt.Sprintf("/var/cores/torpedo-diagsasync-%s-%d.tar.gz", currNode.Name, time.Now().Unix()),
// 					ContainerName: "",
// 					OnHost:        true,
// 				}
// 				err := Inst().V.CollectDiags(currNode, config, torpedovolume.DiagOps{Validate: true, Async: true})
// 				Expect(err).NotTo(HaveOccurred())
// 			})
// 		}

// 		for _, ctx := range contexts {
// 			TearDownContext(ctx, nil)
// 		}
// 	})

// 	JustAfterEach(func() {
// 		AfterEachTest(contexts)
// 	})
// })

var _ = Describe("{GenerateLogs}", func() {
	var contexts []*scheduler.Context
	It("has to setup, validate, try to get a-sync diags on nodes and teardown apps", func() {
		contexts = make([]*scheduler.Context, 0)
		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("diagsasyncbasic-%d", i))...)
		}

		//ValidateApplications(contexts)

		// One node at a time, collect diags and verify in S3
		for _, currNode := range node.GetWorkerNodes() {
			_, err := runCmd("pwd", currNode)
			if err != nil {
				logrus.Warnf("SSH to node %v not working", currNode.Name)
				logrus.Error(err)
			} else {
				logrus.Infof("Creating directors logs in the node %v", currNode.Name)
				cmdOut, _ := runCmd("mkdir /root/logs", currNode)
				logrus.Infof("mkdir o/p: %v", cmdOut)
				logrus.Info("Mounting nfs diags directory")
				cmdOut, _ = runCmd("mount -t nfs diags.pwx.dev.purestorage.com:/var/lib/osd/pxns/688230076034934618 /root/logs", currNode)
				logrus.Infof("Mount o/p: %v", cmdOut)

			}

			Step(fmt.Sprintf("collect diags on node: %s", currNode.Name), func() {

				config := &torpedovolume.DiagRequestConfig{
					DockerHost:    "unix:///var/run/docker.sock",
					OutputFile:    fmt.Sprintf("/var/cores/diags-%s-%d.tar.gz", currNode.Name, time.Now().Unix()),
					ContainerName: "",
					Profile:       false,
					Live:          true,
					Upload:        false,
					All:           true,
					Force:         true,
					OnHost:        true,
					Extra:         false,
				}
				err := Inst().V.CollectDiags(currNode, config, torpedovolume.DiagOps{Validate: true, Async: true})
				Expect(err).NotTo(HaveOccurred())
			})
		}

		Step("collect stork logs", func() {
			storkLabel := make(map[string]string)
			storkLabel["name"] = "stork"
			podList, err := core.Instance().GetPods("", storkLabel)
			if err == nil {
				logsByPodName := map[string]string{}
				for _, p := range podList.Items {
					logOptions := corev1.PodLogOptions{
						// Getting 250 lines from the pod logs to get the io_bytes
						TailLines: getInt64Address(250),
					}
					output, err := core.Instance().GetPodLog(p.Name, p.Namespace, &logOptions)
					if err != nil {
						logrus.Error(fmt.Errorf("failed to get logs for the pod %s/%s: %w", p.Namespace, p.Name, err))
					}
					logsByPodName[p.Name] = output
				}
			} else {
				logrus.Errorf("Error in getting stork pods, Err: %v", err.Error())
			}

		})

		Step("collect autopilot logs", func() {
			podLabel := make(map[string]string)
			podLabel["name"] = "autopilot"
			podList, err := core.Instance().GetPods("", podLabel)
			if err == nil {
				logsByPodName := map[string]string{}
				for _, p := range podList.Items {
					logOptions := corev1.PodLogOptions{
						// Getting 250 lines from the pod logs to get the io_bytes
						TailLines: getInt64Address(250),
					}
					output, err := core.Instance().GetPodLog(p.Name, p.Namespace, &logOptions)
					if err != nil {
						logrus.Error(fmt.Errorf("failed to get logs for the pod %s/%s: %w", p.Namespace, p.Name, err))
					}
					logsByPodName[p.Name] = output
				}
			} else {
				logrus.Errorf("Error in getting autopilot pods, Err: %v", err.Error())
			}

		})

		Step("collect PX operator logs", func() {
			podLabel := make(map[string]string)
			podLabel["name"] = "portworx-operator"
			podList, err := core.Instance().GetPods("", podLabel)
			if err == nil {
				logsByPodName := map[string]string{}
				for _, p := range podList.Items {
					logOptions := corev1.PodLogOptions{
						// Getting 250 lines from the pod logs to get the io_bytes
						TailLines: getInt64Address(250),
					}
					output, err := core.Instance().GetPodLog(p.Name, p.Namespace, &logOptions)
					if err != nil {
						logrus.Error(fmt.Errorf("failed to get logs for the pod %s/%s: %w", p.Namespace, p.Name, err))
					}
					logsByPodName[p.Name] = output
				}
			} else {
				logrus.Errorf("Error in getting portworx-operator pods, Err: %v", err.Error())
			}

		})

		for _, ctx := range contexts {
			TearDownContext(ctx, nil)
		}
	})

	JustAfterEach(func() {
		AfterEachTest(contexts)
	})
})

func getInt64Address(x int64) *int64 {
	return &x
}

var _ = Describe("{JiraTest}", func() {

	It("Jira test method", func() {
		jirautils.CreateIssue("lsrinivas@purestorage.com", "jCSvaoTxEtPFHZeD3jB0B6FF")

	})

})

var _ = AfterSuite(func() {
	//PerformSystemCheck()
	//ValidateCleanup()
})

func runCmd(cmd string, n node.Node) (string, error) {
	output, err := Inst().N.RunCommand(n, cmd, node.ConnectionOpts{
		Timeout:         1 * time.Minute,
		TimeBeforeRetry: 5 * time.Second,
		Sudo:            true,
	})
	if err != nil {
		logrus.Warnf("failed to run cmd: %s. err: %v", cmd, err)
	}
	return output, err
}

func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags
	ParseFlags()
	os.Exit(m.Run())
}
