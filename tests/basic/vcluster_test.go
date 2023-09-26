package tests

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/portworx/torpedo/drivers/vcluster"
	"github.com/portworx/torpedo/pkg/log"
	"github.com/portworx/torpedo/tests"
	. "github.com/portworx/torpedo/tests"
	"os"
	"os/exec"
	"strings"
	"time"
)

var _ = Describe("VCluster", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("VclusterOperations", "Create, Connect and execute a method on Vcluster", nil, 0)
	})
	//var vclusterNames = []string{"my-vcluster1", "my-vcluster2", "my-vcluster3"}
	var vclusterNames = []string{"my-vcluster1"}
	It("Create and connect to vclusters and run a sample method", func() {
		Step("Create vClusters", func() {
			for _, name := range vclusterNames {
				err := vcluster.CreateVCluster(name)
				Expect(err).NotTo(HaveOccurred())
			}
		})
		Step("Wait for all vClusters to come up in Running State", func() {
			for _, name := range vclusterNames {
				err := vcluster.WaitForVClusterRunning(name, 10*time.Minute)
				Expect(err).NotTo(HaveOccurred())
			}
		})
		Step("Connect to each vCluster and execute a method", func() {
			for _, name := range vclusterNames {
				log.Infof("Trying to connect to %v", name)
				err := ConnectVClusterAndExecute(name, SampleMethod, name)
				if err != nil {
					log.FailOnError(err, fmt.Sprintf("Failed to connect and execute method on cluster %v", name))
				}
			}
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		for _, name := range vclusterNames {
			err := vcluster.DeleteVCluster(name)
			if err != nil {
				log.Errorf("Error deleting Vcluster with name %v", name)
			}
		}
	})
})

// This method connects to a vcluster and executes a test function that a testcase wants to run
func ConnectVClusterAndExecute(vclusterName string, testFunc func([]interface{}) []interface{}, args ...interface{}) error {
	killChan := make(chan bool)
	// Running the vcluster connect in the background
	go func() {
		cmd := exec.Command("vcluster", "connect", vclusterName)
		cmd.Start()
		<-killChan
		cmd.Process.Signal(os.Interrupt)
		cmd.Wait()
	}()
	isConnected := false
	for i := 0; i < 60; i++ {
		cmd := exec.Command("kubectl", "config", "current-context")
		out, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("Failed to get current context: %v", err)
		}
		prefix := fmt.Sprintf("vcluster_%s_", vclusterName)
		curContext := strings.TrimSpace(string(out))
		log.Infof("current context is: %v", curContext)
		log.Infof("prefix is: %v", prefix)
		if !strings.Contains(curContext, prefix) {
			log.Infof("Context not yet switched to %v. Retrying.", strings.TrimSpace(string(out)))
			time.Sleep(1 * time.Second)
			continue
		} else {
			log.Infof("Successfully switched to context: %v", strings.TrimSpace(string(out)))
			isConnected = true
			break
		}
	}
	if !isConnected {
		killChan <- true
		return fmt.Errorf("Failed to connect to vCluster: %v", vclusterName)
	}
	results := testFunc(args)
	killChan <- true
	if resError, ok := results[0].(error); ok && resError != nil {
		log.Errorf("Error in Sample test method for vcluster %s: %v", vclusterName, resError)
		return resError
	}
	return nil
}

// This method tries to run nginx app within a vcluster
func SampleMethod(args []interface{}) []interface{} {
	clusterName, ok := args[0].(string)
	if !ok {
		log.Errorf("Expected the first argument type to be string")
		return []interface{}{fmt.Errorf("Expected the first argument type to be string")}
	}
	log.Infof("Within the vCluster: %v", clusterName)
	cmd := exec.Command("vcluster", "list")
	out, err := cmd.Output()
	if err != nil {
		return []interface{}{err}
	} else {
		log.Infof("Output is: %v", string(out))
	}
	cmd = exec.Command("kubectl", "config", "get-contexts")
	out, err = cmd.Output()
	if err != nil {
		return []interface{}{err}
	} else {
		log.Infof("Output is: %v", string(out))
	}

	appToRun := "nginx"
	tests.Inst().AppList = []string{appToRun}
	vcluster.ContextChange = true
	vcluster.CurrentClusterContext = clusterName
	vcluster.UpdatedClusterContext = "host"
	context := tests.ScheduleApplications("vcluster-test-app")
	errChan := make(chan error, 100)
	for _, ctx := range context {
		tests.ValidateContext(ctx, &errChan)
	}
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}
	errStrings := make([]string, 0)
	for _, err := range errors {
		if err != nil {
			errStrings = append(errStrings, err.Error())
		}
	}
	if len(errStrings) > 0 {
		return []interface{}{errStrings}
	}
	return []interface{}{nil}
}
