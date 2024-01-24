package unifiedControlPlane

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	pdsv2 "github.com/portworx/pds-api-go-client/unifiedcp/v2alpha1"
	"io/ioutil"
	"net/http"
	"os"
)

const (
	// Control plane environment variables
	envControlPlaneURL = "CONTROL_PLANE_URL"
	envUsername        = "PDS_USERNAME"
	envPassword        = "PDS_PASSWORD"
	envPDSISSUERURL    = "PDS_ISSUER_URL"
)

// BearerToken struct
type BearerToken struct {
	SUCCESS        bool   `json:"SUCCESS"`
	SUCCESSMESSAGE string `json:"SUCCESSMESSAGE"`
	DATA           struct {
		Token string `json:"token"`
		USER  struct {
			id string `json:"Id"`
		}
	} `json:"token_type"`
}

func GetContext() (context.Context, error) {
	username := os.Getenv(envUsername)
	password := os.Getenv(envPassword)
	issuerURL := os.Getenv(envPDSISSUERURL)
	url := fmt.Sprintf("%s/api/login", issuerURL)

	postBody, err := json.Marshal(map[string]string{
		"username": username,
		"password": password,
	})
	if err != nil {
		return nil, err
	}
	requestBody := bytes.NewBuffer(postBody)
	resp, err := http.Post(url, "application/json", requestBody)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	//Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var bearerToken = new(BearerToken)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &bearerToken)
	if err != nil {
		return nil, err
	}

	ctx := context.WithValue(context.Background(), pdsv2.ContextAPIKeys, map[string]pdsv2.APIKey{"ApiKeyAuth": {Key: bearerToken.DATA.Token}})

	return ctx, nil

}
