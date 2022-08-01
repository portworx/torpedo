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
	controlPlaneUrl string
	components      *api.Components
}

type BearerToken struct {
	Access_token  string `json:"access_token"`
	Token_type    string `json:"token_type"`
	Expires_in    uint64 `json:"expires_in"`
	Refresh_token string `json:"refresh_token"`
}

func GetBearerToken() string {
	username := os.Getenv("PDS_USERNAME")
	password := os.Getenv("PDS_PASSWORD")
	clientId := os.Getenv("PDS_CLIENT_ID")
	clientSecret := os.Getenv("PDS_CLIENT_SECRET")
	issuer_url := os.Getenv("PDS_ISSUER_URL")
	url := fmt.Sprintf("%s/protocol/openid-connect/token", issuer_url)
	grantType := "password"

	postBody, _ := json.Marshal(map[string]string{
		"grant_type":    grantType,
		"client_id":     clientId,
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

	return bearerToken.Access_token

}

func (cp *ControlPlane) GetRegistrationToken(tenantId string) string {
	log.Info("Fetch the registration token.")

	saClient := cp.components.ServiceAccount
	serviceAccounts, _ := saClient.ListServiceAccounts(tenantId)
	var agentWriterId string
	for _, sa := range serviceAccounts {
		if sa.GetName() == "Default-AgentWriter" {
			agentWriterId = sa.GetId()
		}
	}
	token, _ := saClient.GetServiceAccountToken(agentWriterId)
	return token.GetToken()
}

func (cp *ControlPlane) GetDnsZone(tenantId string) string {
	tenantComp := cp.components.Tenant
	tenant, err := tenantComp.GetTenant(tenantId)
	if err != nil {
		log.Panicf("Unable to fetch the tenant info.\n Error - %v", err)
	}
	log.Infof("Get DNS Zone for the tenant. Name -  %s, Id - %s", tenant.GetName(), tenant.GetId())
	dnsModel, err := tenantComp.GetDns(tenantId)
	if err != nil {
		log.Panicf("Unable to fetch the DNSZone info. \n Error - %v", err)
	}
	return dnsModel.GetDnsZone()
}

func NewControlPlane(url string, components *api.Components) *ControlPlane {
	return &ControlPlane{
		controlPlaneUrl: url,
		components:      components,
	}
}
