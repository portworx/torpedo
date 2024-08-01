package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/onsi/ginkgo/v2"
	"github.com/portworx/torpedo/drivers/pds/parameters"
	"github.com/portworx/torpedo/pkg/log"
	"io/ioutil"
	"os"
	"strings"

	"net/http"

	"net/url"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
)

const (
	// Control plane environment variables
	envControlPlaneURL = "CONTROL_PLANE_URL"
	envUsername        = "PDS_USERNAME"
	envPassword        = "PDS_PASSWORD"
	envPDSClientSecret = "PDS_CLIENT_SECRET"
	envPDSClientID     = "PDS_CLIENT_ID"
	envPDSISSUERURL    = "PDS_ISSUER_URL"
)

// BearerToken struct
type BearerToken struct {
	SUCCESS        bool   `json:"SUCCESS"`
	SUCCESSMESSAGE string `json:"SUCCESSMESSAGE"`
	DATA           struct {
		Token string `json:"token"`
	} `json:"DATA"`
}

var (
	customParams *parameters.Customparams
	siID         *ServiceIdentity
)

// GetContext return context for api call.
func GetContext() (context.Context, error) {
	var token string
	currentSpecReport := ginkgo.CurrentSpecReport()
	testName := strings.Split(currentSpecReport.FullText(), " ")[0]
	serviceIdFlag := customParams.ReturnServiceIdentityFlag()
	PDSControlPlaneURL := os.Getenv("CONTROL_PLANE_URL")
	endpointURL, err := url.Parse(PDSControlPlaneURL)
	if err != nil {
		return nil, fmt.Errorf("Unable to connect to ControlPlane URL: %v\n", err)
	}
	apiConf := pds.NewConfiguration()
	apiConf.Host = endpointURL.Host
	apiConf.Scheme = endpointURL.Scheme
	if serviceIdFlag == true && strings.Contains(testName, "ServiceIdentity") {
		serviceIdToken := siID.ReturnServiceIdToken()
		if serviceIdToken == "" {
			token, err = getBearerToken()
		} else {
			token = serviceIdToken
			log.InfoD("ServiceIdentity Token being used")
		}

	} else {
		token, err = getBearerToken()
		if err != nil {
			return nil, err
		}
	}
	ctx := context.WithValue(context.Background(), pds.ContextAPIKeys, map[string]pds.APIKey{"ApiKeyAuth": {Key: token, Prefix: "Bearer"}})
	return ctx, nil
}

func getBearerToken() (string, error) {
	username := os.Getenv(envUsername)
	password := os.Getenv(envPassword)
	issuerURL := os.Getenv(envPDSISSUERURL)
	url := fmt.Sprintf("%s/login", issuerURL)

	log.Debugf("username [%s]", username)
	log.Debugf("password [%s]", password)
	log.Debugf("url [%s]", url)

	postBody, err := json.Marshal(map[string]string{
		"email":    username,
		"password": password,
	})
	if err != nil {
		return "", err
	}
	requestBody := bytes.NewBuffer(postBody)
	resp, err := http.Post(url, "application/json", requestBody)
	if err != nil {
		return "", fmt.Errorf("error while fetching bearer token %v", err)
	}
	log.Infof("response %s", resp.Status)

	//Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	var bearerToken = new(BearerToken)

	err = json.Unmarshal(body, &bearerToken)
	if err != nil {
		return "", err
	}

	log.Debugf("token [%s]", bearerToken.DATA.Token)

	return bearerToken.DATA.Token, nil

}
