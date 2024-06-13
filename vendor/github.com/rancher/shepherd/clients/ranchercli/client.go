package ranchercli

import (
	"fmt"
	"os/exec"

	management "github.com/rancher/shepherd/clients/rancher/generated/management/v3"
	namegen "github.com/rancher/shepherd/pkg/namegenerator"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/sirupsen/logrus"
)

const (
	rancher   = "rancher"
	createCmd = "create"
	deleteCmd = "delete"
	moveCmd   = "move"
)

type Client struct {
	Session *session.Session
}

func (c *Client) Create(name string, args ...string) error {
	args = append([]string{name, createCmd}, args...)
	err := c.ExecuteCommand(rancher, args...)
	if err != nil {
		return err
	}

	c.Session.RegisterCleanupFunc(func() error {
		return c.Delete(name, args...)
	})

	return nil
}

func (c *Client) Delete(name string, args ...string) error {
	args = append([]string{name, deleteCmd}, args...)

	return c.ExecuteCommand(rancher, args...)
}

func (c *Client) Exists(rancher, resourceType, name string) error {
	cmdStr := fmt.Sprintf("%s %s ls | grep %s", rancher, resourceType, name)
	cmd := exec.Command("sh", "-c", cmdStr)
	err := cmd.Run()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			if exitError.ExitCode() != 0 {
				return fmt.Errorf("error validating %s exists: %v", name, err)
			}
		}

		return err
	}

	return nil
}

// ExecuteCommand will execute a command and log any errors.
func (c *Client) ExecuteCommand(name string, args ...string) error {
	command := exec.Command(name, args...)
	_, err := command.Output()
	if err != nil {
		return fmt.Errorf("error executing command '%s': %v", name, err)
	}

	return nil
}

// NewClient will download the CLI and login to the Rancher server.
func NewClient(session *session.Session, token, host string, client *management.Client) (*Client, error) {
	c := &Client{
		Session: session,
	}

	projectConfig := &management.Project{
		ClusterID: "local",
		Name:      namegen.AppendRandomString("project"),
	}

	testProject, err := client.Project.Create(projectConfig)
	if err != nil {
		return nil, err
	}

	logrus.Infof("Logging into Rancher server...")
	err = c.ExecuteCommand(rancher, "login", "--token", token, "https://"+host, "--skip-verify", "--context", testProject.ID)
	if err != nil {
		return nil, err
	}

	return c, nil
}
