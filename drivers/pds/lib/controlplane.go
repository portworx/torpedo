package lib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	api "github.com/portworx/torpedo/drivers/pds/api"
	log "github.com/sirupsen/logrus"
)

// ControlPlane PDS
type ControlPlane struct {
	ControlPlaneURL string
	components      *api.Components
}

// BearerToken struct
type BearerToken struct {
	AccessToken  string `json:"access_token"`
	IDToken      string `json:"id_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    uint64 `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

const (
	defaultAgentWriter = "Default-AgentWriter"
)

// GetBearerToken fetches the token.
func GetBearerToken() string {
	username := GetAndExpectStringEnvVar(envUsername)
	password := GetAndExpectStringEnvVar(envPassword)
	clientID := GetAndExpectStringEnvVar(envPDSClientID)
	clientSecret := GetAndExpectStringEnvVar(envPDSClientSecret)
	issuerURL := GetAndExpectStringEnvVar(envPDSISSUERURL)
	url := fmt.Sprintf("%s/protocol/openid-connect/token", issuerURL)
	grantType := "password"

	postBody, _ := json.Marshal(map[string]string{
		"grant_type":    grantType,
		"client_id":     clientID,
		"client_secret": clientSecret,
		"username":      username,
		"password":      password,
	})

	requestBody := bytes.NewBuffer(postBody)
	resp, err := http.Post(url, "application/json", requestBody)
	if err != nil {
		log.Fatalf("An Error Occured %v", err)
	}
	defer resp.Body.Close()
	//Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("An Error Occured %v", err)
	}
	var bearerToken = new(BearerToken)
	err = json.Unmarshal(body, &bearerToken)
	if err != nil {
		log.Fatalf("An Error Occured %v", err)
	}
	return bearerToken.AccessToken
}

// GetRegistrationToken PDS
func (cp *ControlPlane) GetRegistrationToken(tenantID string) string {
	log.Info("Fetch the registration token.")

	saClient := cp.components.ServiceAccount
	serviceAccounts, _ := saClient.ListServiceAccounts(tenantID)
	var agentWriterID string
	for _, sa := range serviceAccounts {
		if sa.GetName() == defaultAgentWriter {
			agentWriterID = sa.GetId()
		}
	}
	token, _ := saClient.GetServiceAccountToken(agentWriterID)
	return token.GetToken()
}

// GetDNSZone fetches DNS zone for deployment.
func (cp *ControlPlane) GetDNSZone(tenantID string) (string, error) {
	tenantComp := cp.components.Tenant
	tenant, err := tenantComp.GetTenant(tenantID)
	if err != nil {
		log.Errorf("Unable to fetch the tenant info.\n Error - %v", err)
		return "", err
	}
	log.Infof("Get DNS Zone for the tenant. Name -  %s, Id - %s", tenant.GetName(), tenant.GetId())
	dnsModel, err := tenantComp.GetDNS(tenantID)
	if err != nil {
		log.Errorf("Unable to fetch the DNSZone info. \n Error - %v", err)
		return "", err
	}
	return dnsModel.GetDnsZone(), nil
}

// NewControlPlane to create control plane instance.
func NewControlPlane(url string, components *api.Components) *ControlPlane {
	return &ControlPlane{
		ControlPlaneURL: url,
		components:      components,
	}
}
