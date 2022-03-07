package lib

import (
	"bytes"
	"os"
	"os/exec"

	"github.com/sirupsen/logrus"
)

type cluster struct {
	kubeconfig string
}

const (
	pdsSystemNamespace = "pds-system"
)

func (c *cluster) logComponent(namespace, name string) {
	logrus.Infof("Component %v", name)
	deployment := "deployment/" + name
	output, err := c.mustKubectl(
		"logs",
		deployment,
		"--namespace", namespace,
	)
	if err != nil {
		logrus.Errorf("Error: %v", err)
	}
	logrus.Infof("Output : %v", output)
}

func (c *cluster) mustKubectl(args ...string) (string, error) {
	kubectlArgs := append([]string{"--kubeconfig", c.kubeconfig}, args...)
	cmd := kubectl(kubectlArgs...)
	return mustRun(cmd)
}

func (c *cluster) describePods(namespace string) {
	logrus.Infof("Pods in %s:", namespace)
	output, err := c.mustKubectl("describe", "pods", "--namespace", namespace)
	if err != nil {
		logrus.Info(err)
	}
	logrus.Info(output)
}

func kubectl(args ...string) *exec.Cmd {
	return exec.Command("kubectl", args...)
}

func mustRun(cmd *exec.Cmd) (string, error) {
	out, err := runWithOutput(cmd)
	if err != nil {
		return err.Error(), err
	}
	return out, nil
}

func runWithOutput(cmd *exec.Cmd) (string, error) {
	b := new(bytes.Buffer)
	cmd.Stdout = b
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return b.String(), nil
}
