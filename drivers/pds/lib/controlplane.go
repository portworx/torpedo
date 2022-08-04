package lib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

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
	AccessToken  string `json:"AccessToken"`
	TokenType    string `json:"TokenType"`
	ExpiresIn    uint64 `json:"ExpiresIn"`
	RefreshToken string `json:"RefreshToken"`
}

// GetBearerToken fetches the token.
func GetBearerToken() (string, error) {
	username := os.Getenv("PDSUsername")
	password := os.Getenv("PDSPassword")
	clientID := os.Getenv("PDSClientID")
	clientSecret := os.Getenv("PDSClientSecret")
	issuerURL := os.Getenv("PDSIssuerURL")
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
		return "", err
	}
	defer resp.Body.Close()
	//Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var bearerToken = new(BearerToken)
	err = json.Unmarshal(body, &bearerToken)
	if err != nil {
		return "", err
	}

	return bearerToken.AccessToken, nil

}

// GetRegistrationToken PDS
func (cp *ControlPlane) GetRegistrationToken(tenantID string) (string, error) {
	log.Info("Fetch the registration token.")

	saClient := cp.components.ServiceAccount
	serviceAccounts, _ := saClient.ListServiceAccounts(tenantID)
	var agentWriterID string
	for _, sa := range serviceAccounts {
		if sa.GetName() == "Default-AgentWriter" {
			agentWriterID = sa.GetId()
		}
	}
	token, err := saClient.GetServiceAccountToken(agentWriterID)
	if err != nil {
		return "", err
	}
	return token.GetToken(), nil
}

// GetDNSZone fetches DNS zone for deployment.
func (cp *ControlPlane) GetDNSZone(tenantID string) string {
	tenantComp := cp.components.Tenant
	tenant, err := tenantComp.GetTenant(tenantID)
	if err != nil {
		log.Panicf("Unable to fetch the tenant info.\n Error - %v", err)
	}
	log.Infof("Get DNS Zone for the tenant. Name -  %s, Id - %s", tenant.GetName(), tenant.GetId())
	dnsModel, err := tenantComp.GetDNS(tenantID)
	if err != nil {
		log.Panicf("Unable to fetch the DNSZone info. \n Error - %v", err)
	}
	return dnsModel.GetDnsZone()
}

// NewControlPlane to create control plane instance.
func NewControlPlane(url string, components *api.Components) *ControlPlane {
	return &ControlPlane{
		ControlPlaneURL: url,
		components:      components,
	}
}
