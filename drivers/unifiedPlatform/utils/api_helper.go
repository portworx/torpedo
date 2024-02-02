package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/portworx/torpedo/pkg/log"
	platformv2 "github.com/pure-px/platform-api-go-client/v1alpha1"
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

// BearerToken struct
type BearerToken struct {
	SUCCESS        bool   `json:"SUCCESS"`
	SUCCESSMESSAGE string `json:"SUCCESSMESSAGE"`
	DATA           struct {
		Token string `json:"token"`
	} `json:"DATA"`
}

// GetBearerToken returns the bearer token
func GetBearerToken() (context.Context, string, error) {
	username := os.Getenv(envPXCentralUsername)
	password := os.Getenv(envPXCentralPassword)
	issuerURL := os.Getenv(envPxCentralAPI)
	url := fmt.Sprintf("%s/login", issuerURL)

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

func GetContext() (context.Context, error) {
	username := os.Getenv(envPXCentralUsername)
	password := os.Getenv(envPXCentralPassword)
	issuerURL := os.Getenv(envPxCentralAPI)
	url := fmt.Sprintf("%s/login", issuerURL)

	log.Infof("issuerURL [%s]", issuerURL)
	log.Infof("email [%s]", username)
	log.Infof("password [%s]", password)

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

	log.Infof("Bearer Token %s", bearerToken.DATA.Token)
	ctx := context.WithValue(context.Background(), "auth apiKey", map[string]platformv2.APIKey{"ApiKeyAuth": {Key: bearerToken.DATA.Token, Prefix: "Bearer"}})
	log.Infof("ctx value [%s]", ctx.Value("auth apiKey"))

	return ctx, nil

}
