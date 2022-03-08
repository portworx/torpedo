package jirautils

import (
	"bytes"

	jira "github.com/andygrunwald/go-jira"
	"github.com/sirupsen/logrus"
)

var (
	client *jira.Client
)

const (
	jiraAPIToken = "jCSvaoTxEtPFHZeD3jB0B6FF"
	jiraURL      = "https://portworx.atlassian.net/"
)

// CreateIssue creates issue in jira
func CreateIssue(username, token string) {
	httpClient := jira.BasicAuthTransport{
		Username: username,
		Password: token,
	}
	var err error

	client, err = jira.NewClient(httpClient.Client(), jiraURL)
	if err != nil {
		logrus.Error(err)
	} else {
		getProjects(client)

	}
	createPTX(client, "6098fbce2614ec006818c402", "Test Description", "Test Summary")
	getPTX(client)

}

func getPTX(client *jira.Client) {

	issue, _, err := client.Issue.Get("PTX-5278", nil)
	logrus.Infof("Error: %v", err)

	logrus.Infof("%s: %+v\n", issue.Key, issue.Fields.Summary)
	logrus.Infof("%+v\n", issue)

	logrus.Infof("%s: %s\n", issue.ID, issue.Fields.Summary)
	logrus.Info(issue.Fields.FixVersions[0].Name)

}

func createPTX(client *jira.Client, accountId, description, summary string) {

	i := jira.Issue{
		Fields: &jira.IssueFields{
			Assignee: &jira.User{
				AccountID: accountId,
			},
			Description: description,
			Type: jira.IssueType{
				Name: "Bug",
			},
			Project: jira.Project{
				Key: "PTX",
			},
			FixVersions: []*jira.FixVersion{
				{
					Name: "master",
				},
			},
			AffectsVersions: []*jira.AffectsVersion{
				{
					Name: "master",
				},
			},
			Summary: summary,
		},
	}
	issue, resp, err := client.Issue.Create(&i)

	logrus.Infof("Resp: %v", resp.StatusCode)
	if resp.StatusCode == 201 {
		logrus.Info("Successfully created new jira issue.")
		logrus.Infof("Jira Issue: %+v\n", issue.Key)

	} else {
		logrus.Infof("Error while creating jira issue: %v", err)
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		newStr := buf.String()
		logrus.Infof(newStr)
	}

}

func getProjects(client *jira.Client) {
	req, _ := client.NewRequest("GET", "rest/api/3/project/recent", nil)

	projects := new([]jira.Project)
	_, err := client.Do(req, projects)
	if err != nil {
		logrus.Info("Error while getting project")
		logrus.Error(err)
		return
	}

	for _, project := range *projects {

		logrus.Infof("%s: %s\n", project.Key, project.Name)
	}
}
