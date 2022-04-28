package pds

import (
	"os/exec"
	"time"
)

type cluster struct {
	kubeconfig string
}

type componentLog struct {
	Name   string
	LogCmd *exec.Cmd
}

const (
	pdsSystemNamespace = "pds-system"
)

func (c *cluster) logComponent(namespace, name string, since time.Time) *exec.Cmd {
	deployment := "deployment/" + name
	timestamp := since.Format(time.RFC3339)
	return c.kubectl(
		"logs",
		deployment,
		"--namespace", namespace,
		"--since-time="+timestamp,
	)
}

func (c *cluster) kubectl(args ...string) *exec.Cmd {
	kubectlArgs := append([]string{"--kubeconfig", c.kubeconfig}, args...)
	return exec.Command("kubectl", kubectlArgs...)
}
