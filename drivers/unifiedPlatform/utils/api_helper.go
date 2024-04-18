package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/pkg/log"
	accountv1 "github.com/pure-px/platform-api-go-client/platform/v1/account"
	"io/ioutil"
	"net/http"
	"os"
)

const (
	// Control plane environment variables
	envPXCentralUsername = "PX_CENTRAL_USERNAME"
	envPXCentralPassword = "PX_CENTRAL_PASSWORD"
	envPxCentralAPI      = "PX_CENTRAL_API"
)

var RunWithRBAC RunWithRbac

type Credentials struct {
	Token string
}

type RunWithRbac struct {
	RbacFlag  bool
	RbacToken string
}

// BearerToken struct
type BearerToken struct {
	SUCCESS        bool   `json:"SUCCESS"`
	SUCCESSMESSAGE string `json:"SUCCESSMESSAGE"`
	DATA           struct {
		Token string `json:"token"`
	} `json:"DATA"`
}

func (c *Credentials) GetRequestMetadata(_ context.Context, _ ...string) (map[string]string, error) {
	metadata := map[string]string{}

	if c.Token == "" {
		_, token, err := GetBearerToken()
		if err != nil {
			return nil, fmt.Errorf("get berare token, err: %s", err)
		}

		c.Token = token
	}

	metadata["Authorization"] = fmt.Sprintf("Bearer %s", c.Token)

	return metadata, nil
}

func (c *Credentials) RequireTransportSecurity() bool {
	return false
}

// GetBearerToken returns the bearer token
func GetBearerToken() (context.Context, string, error) {
	username := os.Getenv(envPXCentralUsername)
	password := os.Getenv(envPXCentralPassword)
	issuerURL := os.Getenv(envPxCentralAPI)
	if RunWithRBAC.RbacFlag == true {
		return context.Background(), RunWithRBAC.RbacToken, nil
	}
	//log.Infof("user name %s", username)
	//log.Infof("password %s", password)

	url := fmt.Sprintf("%s/login", issuerURL)

	//log.Infof("issuer url %s", issuerURL)

	postBody, err := json.Marshal(map[string]string{
		"email":    username,
		"password": password,
	})
	if err != nil {
		return nil, "", err
	}
	requestBody := bytes.NewBuffer(postBody)
	resp, err := http.Post(url, "application/json", requestBody)
	if err != nil {
		return nil, "", fmt.Errorf("error while fetching bearer token %v", err)
	}

	//Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	defer resp.Body.Close()
	var bearerToken = new(BearerToken)

	err = json.Unmarshal(body, &bearerToken)
	if err != nil {
		return nil, "", err
	}
	return context.Background(), bearerToken.DATA.Token, nil
}

// CustomRegistryConfig :Custom Registry info
type CustomRegistryConfig struct {
	CustomImageRegistryConfig string
	RegistryUrl               string
	RegistryNamespace         string
	RegistryUserName          string
	RegistryPassword          string
	CaCert                    string
}

// ProxyConfig structure
type ProxyConfig struct {
	HttpUrl  string
	HttpsUrl string
	Username string
	Password string
	NoProxy  string
	CaCert   string
}

func GetContext() (context.Context, error) {
	username := os.Getenv(envPXCentralUsername)
	password := os.Getenv(envPXCentralPassword)
	issuerURL := os.Getenv(envPxCentralAPI)
	url := fmt.Sprintf("%s/login", issuerURL)

	postBody, err := json.Marshal(map[string]string{
		"email":    username,
		"password": password,
	})
	if err != nil {
		return nil, err
	}
	requestBody := bytes.NewBuffer(postBody)
	resp, err := http.Post(url, "application/json", requestBody)
	if err != nil {
		return nil, fmt.Errorf("error while fetching bearer token %v", err)
	}
	log.Infof("response %s", resp.Status)

	//Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var bearerToken = new(BearerToken)

	err = json.Unmarshal(body, &bearerToken)
	if err != nil {
		return nil, err
	}

	ctx := context.WithValue(context.Background(), "auth apiKey", map[string]accountv1.APIKey{"ApiKeyAuth": {Key: bearerToken.DATA.Token, Prefix: "Bearer"}})
	log.Infof("ctx value [%s]", ctx.Value("auth apiKey"))

	return ctx, nil

}

func GetWorkflowResponseMap() map[string][]WorkFlowResponse {
	var createdMap = make(map[string][]WorkFlowResponse)
	return createdMap
}

func GetDefaultHeader(token string, accountId string) map[string]string {
	defaultHeader := make(map[string]string)

	defaultHeader["Authorization"] = "Bearer " + token
	if !RunWithRBAC.RbacFlag {
		defaultHeader["px-account-id"] = accountId
	}

	return defaultHeader
}
