package tests

import (
	"encoding/json"
	"fmt"
	"github.com/ktrysmt/go-bitbucket"
	"github.com/libopenstorage/autopilot-api/pkg/apis/autopilot/v1alpha1"
	"github.com/libopenstorage/secrets"
	"github.com/libopenstorage/secrets/k8s"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"

	"net/http"
	"os"
	"strings"
)

const (
	// BitBucketPrStatePending indicates an open PR
	BitBucketPrStatePending = "OPEN"
	// BitBucketPrStateDeclined indicated a closed PR
	BitBucketPrStateDeclined = "DECLINED"
	// BitBucketPrStateMerged indicates a merged PR
	BitBucketPrStateMerged = "MERGED"
	// GitOpsSecretKeyName is the name of the secret, that contains credentials for bitbucket
	GitOpsSecretKeyName = "aut-gitops"
	// BitBucketPasswordEnvKey is the env variable key name for the bitbucket password
	BitBucketPasswordEnvKey = "BITBUCKET_PASSWORD"
)

type bitbucketInst struct {
	baseURL          string
	repo             string
	user             string
	repoFolder       string
	baseBranch       string
	authorName       string
	defaultReviewers []string
	userLogin        string
	userPasswd       string
	projectKey       string
	httpClient       *http.Client // Used for UTs
	bitBucketClient  *bitbucket.Client
}

type bitbucketUser struct {
	Name         string `json:"name"`
	DisplayName  string `json:"displayName"`
	EmailAddress string `json:"emailAddress"`
	Active       bool   `json:"active"`
	Type         string `json:"type"`
	ID           int    `json:"id"`
	Slug         string `json:"slug"`
}

type reviewer struct {
	User               *bitbucketUser `json:"user"`
	LastReviewedCommit string         `json:"lastReviewedCommit"`
	Role               string         `json:"role"`
	Approved           bool           `json:"approved"`
	Status             string         `json:"status"`
}

type bitBucketPR struct {
	Closed       bool       `json:"closed"`
	ID           int        `json:"id"`
	State        string     `json:"state"`
	Open         bool       `json:"open"`
	Title        string     `json:"title"`
	UpdatedTitle string     `json:"updated_title"`
	CreatedDate  float64    `json:"createdDate"`
	Version      int        `json:"version"`
	FromRef      *fromRef   `json:"fromRef"`
	Reviewers    []reviewer `json:"reviewers"`
}

type fromRef struct {
	ID           string `json:"id"`
	LatestCommit string `json:"latestCommit"`
	DisplayID    string `json:"displayId"`
}

type commentAuthor struct {
	Name         string `json:"name"`
	DisplayName  string `json:"displayName"`
	ID           int    `json:"id"`
	EmailAddress string `json:"emailAddress"`
}

type comment struct {
	ID     int            `json:"id"`
	Text   string         `json:"text"`
	Author *commentAuthor `json:"author"`
}

type activity struct {
	ID      int      `json:"id"`
	Comment *comment `json:"comment"`
}

// BitBucketConfig is the configuration for bitbucket for Autopilot
type BitBucketConfig struct {
	User    string `yaml:"user"`
	BaseURL string `yaml:"baseUrl"`
	Repo    string `yaml:"repo"`
	Folder  string `yaml:"folder"`
	Branch  string `yaml:"branch"`
	// ProjectKey is a project name key (for example 'PXAUT')
	ProjectKey       string   `yaml:"projectKey"`
	DefaultReviewers []string `yaml:"defaultReviewers,omitempty"`
	UserLogin        string   `yaml:"login"`
	UserPassword     string   `yaml:"password"`
}

// GitOpsConfig provides configuration data to autopilot used to initialize gitops provider
type GitOpsConfig struct {
	Name   string                 `yaml:"name,omitempty"`
	Type   string                 `yaml:"type"`
	Params map[string]interface{} `yaml:"params"`
}

func initFuncBitBucket(conf GitOpsConfig, namespace string) (*bitbucketInst, error) {
	logrus.Infof("Start initializing BitBucket")

	bitBucketConfig := &BitBucketConfig{}
	data, err := json.Marshal(conf.Params)

	if err != nil {
		return nil, err
	}

	err = yaml.UnmarshalStrict(data, bitBucketConfig)
	if err != nil {
		return nil, err
	}

	// set it silently, as it's required by the library we use, to avoid multiple points of configuration
	err = os.Setenv("BITBUCKET_API_BASE_URL", bitBucketConfig.BaseURL)
	if err != nil {
		return nil, err
	}

	password, err := getPasswordFromSecret(namespace, BitBucketPasswordEnvKey)
	if err != nil {
		return nil, err
	}
	if len(password) == 0 {
		return nil, fmt.Errorf("password is empty")
	}

	inst := &bitbucketInst{
		baseURL:          bitBucketConfig.BaseURL,
		repo:             bitBucketConfig.Repo,
		user:             bitBucketConfig.User,
		baseBranch:       bitBucketConfig.Branch,
		repoFolder:       bitBucketConfig.Folder,
		authorName:       bitBucketConfig.User,
		defaultReviewers: bitBucketConfig.DefaultReviewers,
		userLogin:        bitBucketConfig.User,
		userPasswd:       password,
		projectKey:       bitBucketConfig.ProjectKey,
	}

	// no need to validate credentials in the test mode. Also, the test client is injected later after initialization, so validation will fail
	err = inst.validateCredentials()
	if err != nil {
		return nil, err
	}
	return inst, nil
}

func validateConfig(bucketConfig *BitBucketConfig) error {
	if len(bucketConfig.Repo) == 0 {
		return fmt.Errorf("BitBucket repo name is required")
	}
	if len(bucketConfig.User) == 0 {
		return fmt.Errorf("BitBucket user name is required")
	}
	if len(bucketConfig.BaseURL) == 0 {
		return fmt.Errorf("BitBucket base url is required")
	}
	if len(bucketConfig.ProjectKey) == 0 {
		return fmt.Errorf("BitBucket project key is required")
	}
	return nil
}

func generateCommitBranchName(actionApproval *v1alpha1.ActionApproval) string {
	separator := "-"

	items := []string{
		"aut", actionApproval.GetNamespace(), actionApproval.GetName(),
	}
	return strings.ToLower(strings.Join(items, separator))
}

func (s *bitbucketInst) getLastPRForApproval(actionApproval *v1alpha1.ActionApproval) (*bitBucketPR, error) {
	commitBranch := generateCommitBranchName(actionApproval)

	client := s.getClient()
	data, err := client.Repositories.PullRequests.APIv1Get(&bitbucket.PullRequestsOptions{
		ProjectKey:   s.projectKey,
		RepoSlug:     s.repo,
		States:       []string{BitBucketPrStateDeclined, BitBucketPrStateMerged, BitBucketPrStatePending},
		SourceBranch: commitBranch,
	})

	if err != nil {
		return nil, err
	}
	prs, err := unmarshalPR(data)
	if err != nil {
		return nil, err
	}

	for _, pr := range prs {
		if pr.FromRef != nil && pr.FromRef.DisplayID == commitBranch {
			return &pr, nil
		}
	}

	return nil, nil
}

func (s *bitbucketInst) getCommentsForPR(prID int) ([]activity, error) {
	client := s.getClient()
	data, err := client.Repositories.PullRequests.API1v1GetActivities(&bitbucket.PullRequestsOptions{
		ID:         fmt.Sprintf("%d", prID),
		ProjectKey: s.projectKey,
		RepoSlug:   s.repo,
	})

	if err != nil {
		return nil, err
	}
	comments, err := unmarshalComments(data)
	if err != nil {
		return nil, err
	}

	return comments, nil
}

func (s *bitbucketInst) mergePR(pr *bitBucketPR) error {
	client := s.getClient()
	_, err := client.Repositories.PullRequests.APIv1Merge(&bitbucket.PullRequestsOptions{
		ID:         fmt.Sprintf("%d", pr.ID),
		ProjectKey: s.projectKey,
		RepoSlug:   s.repo,
		Version:    pr.Version,
	})

	if err != nil {
		return err
	}
	return nil
}

func unmarshalPR(pr interface{}) ([]bitBucketPR, error) {
	byteDataValues, err := getResponseValue(pr)
	if err != nil {
		return nil, err
	}

	vals := []bitBucketPR{}
	err = json.Unmarshal(byteDataValues, &vals)
	if err != nil {
		return nil, err
	}

	return vals, nil
}

func unmarshalComments(comments interface{}) ([]activity, error) {
	byteDataValues, err := getResponseValue(comments)
	if err != nil {
		return nil, err
	}
	vals := []activity{}
	err = json.Unmarshal(byteDataValues, &vals)
	if err != nil {
		return nil, err
	}
	return vals, nil
}

func getResponseValue(resp interface{}) ([]byte, error) {
	byteData, err := json.Marshal(resp)
	if err != nil {
		return nil, err
	}
	rawPR := map[string]interface{}{}
	err = json.Unmarshal(byteData, &rawPR)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshall PR data: %s", err)
	}

	byteDataValues, err := json.Marshal(rawPR["values"])
	if err != nil {
		return nil, err
	}
	return byteDataValues, err
}

func (s *bitbucketInst) validateCredentials() error {
	_, err := s.getClient().Repositories.PullRequests.APIv1Get(&bitbucket.PullRequestsOptions{RepoSlug: s.repo, ProjectKey: s.projectKey})
	return err
}

func (s *bitbucketInst) getClient() *bitbucket.Client {

	var client *bitbucket.Client

	if s.bitBucketClient != nil {
		return s.bitBucketClient
	}

	if s.httpClient != nil {
		client = bitbucket.NewBasicAuth(s.userLogin, s.userPasswd)
		client.HttpClient = s.httpClient
	} else {
		client = bitbucket.NewBasicAuth(s.userLogin, s.userPasswd)
	}
	return client
}

func (s *bitbucketInst) buildRepoURL() string {
	if strings.HasSuffix(s.baseURL, "/") {
		s.baseURL = strings.TrimRight(s.baseURL, "/")
	}
	url := fmt.Sprintf("%s/scm/%s/%s.git", s.baseURL, s.projectKey, s.repo)
	logrus.Infof("Repo URL: %s", url)
	return url
}

func (s *bitbucketInst) approvedBy(pr *bitBucketPR) (string, bool) {
	for _, reviewer := range pr.Reviewers {
		if reviewer.Approved {
			return reviewer.User.DisplayName, true
		}
	}
	return "", false
}

func (s *bitbucketInst) setHTTPClient(client *http.Client) {
	s.httpClient = client
}

func getPasswordFromSecret(namespace string, key string) (string, error) {

	var val string
	k8sSecrets, err := secrets.New(k8s.Name, nil)
	if err != nil {
		return "", fmt.Errorf("failed to instantiate k8s secrets manager due to: %v", err)
	}

	gitopsSecret, err := k8sSecrets.GetSecret(GitOpsSecretKeyName, map[string]string{
		k8s.SecretNamespace: namespace,
	})
	if err != nil {
		if !k8s_errors.IsNotFound(err) {
			logrus.Warnf("failed to fetch secret: %s/%s due to: %v", namespace, GitOpsSecretKeyName, err)
		}
	} else {
		passwdVal, present := gitopsSecret[key]
		if present {
			val = passwdVal.(string)
		}
	}
	return val, err
}
