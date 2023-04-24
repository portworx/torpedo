package controlplane

import (
	"github.com/portworx/torpedo/drivers/pds/api"
	"github.com/portworx/torpedo/pkg/log"
)

// ControlPlane PDS
type ControlPlane struct {
	ControlPlaneURL string
	components      *api.Components
}

// PDS const
const (
	storageTemplateName   = "QaDefault"
	resourceTemplateName  = "Small"
	appConfigTemplateName = "QaDefault"
)

// PDS vars
var (
	isavailable                bool
	isTemplateavailable        bool
	resourceTemplateID         string
	appConfigTemplateID        string
	storageTemplateID          string
	isStorageTemplateAvailable bool
)

// GetRegistrationToken return token to register a target cluster.
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
func (cp *ControlPlane) GetDNSZone(tenantID string) (string, error) {
	tenantComp := cp.components.Tenant
	tenant, err := tenantComp.GetTenant(tenantID)
	if err != nil {
		log.Panicf("Unable to fetch the tenant info.\n Error - %v", err)
	}
	log.Infof("Get DNS Zone for the tenant. Name -  %s, Id - %s", tenant.GetName(), tenant.GetId())
	dnsModel, err := tenantComp.GetDNS(tenantID)
	if err != nil {
		log.Infof("Unable to fetch the DNSZone info. \n Error - %v", err)
	}
	return dnsModel.GetDnsZone(), err
}

//// GetResourceTemplate get the resource template id
//func (cp *ControlPlane) GetResourceTemplate(tenantID string, supportedDataService string) (string, error) {
//	log.Infof("Get the resource template for each data services")
//	resourceTemplates, err := cp.components.ResourceSettingsTemplate.ListTemplates(tenantID)
//	if err != nil {
//		return "", err
//	}
//	isavailable = false
//	isTemplateavailable = false
//	for i := 0; i < len(resourceTemplates); i++ {
//		if resourceTemplates[i].GetName() == resourceTemplateName {
//			isTemplateavailable = true
//			dataService, err := cp.components.DataService.GetDataService(resourceTemplates[i].GetDataServiceId())
//			if err != nil {
//				return "", err
//			}
//			if dataService.GetName() == supportedDataService {
//				log.Infof("Data service name: %v", dataService.GetName())
//				log.Infof("Resource template details ---> Name %v, Id : %v ,DataServiceId %v , StorageReq %v , Memoryrequest %v",
//					resourceTemplates[i].GetName(),
//					resourceTemplates[i].GetId(),
//					resourceTemplates[i].GetDataServiceId(),
//					resourceTemplates[i].GetStorageRequest(),
//					resourceTemplates[i].GetMemoryRequest())
//
//				isavailable = true
//				resourceTemplateID = resourceTemplates[i].GetId()
//			}
//		}
//	}
//	if !(isavailable && isTemplateavailable) {
//		log.Errorf("Template with Name %v does not exis", resourceTemplateName)
//	}
//	return resourceTemplateID, nil
//}
//
//// GetStorageTemplate return the storage template id
//func (cp *ControlPlane) GetStorageTemplate(tenantID string) (string, error) {
//	log.InfoD("Get the storage template")
//	storageTemplates, err := cp.components.StorageSettingsTemplate.ListTemplates(tenantID)
//	if err != nil {
//		return "", err
//	}
//	isStorageTemplateAvailable = false
//	for i := 0; i < len(storageTemplates); i++ {
//		if storageTemplates[i].GetName() == storageTemplateName {
//			isStorageTemplateAvailable = true
//			log.InfoD("Storage template details -----> Name %v,Repl %v , Fg %v , Fs %v",
//				storageTemplates[i].GetName(),
//				storageTemplates[i].GetRepl(),
//				storageTemplates[i].GetFg(),
//				storageTemplates[i].GetFs())
//			storageTemplateID = storageTemplates[i].GetId()
//		}
//	}
//	if !isStorageTemplateAvailable {
//		log.Fatalf("storage template %v is not available ", storageTemplateName)
//	}
//	return storageTemplateID, nil
//}
//
//// GetAppConfTemplate returns the app config template id
//func (cp *ControlPlane) GetAppConfTemplate(tenantID string, supportedDataService string) (string, error) {
//	appConfigs, err := cp.components.AppConfigTemplate.ListTemplates(tenantID)
//	var d dataservice.DataserviceType
//	if err != nil {
//		return "", err
//	}
//	isavailable = false
//	isTemplateavailable = false
//	dataServiceId := d.GetDataServiceID(supportedDataService)
//	for i := 0; i < len(appConfigs); i++ {
//		if appConfigs[i].GetName() == appConfigTemplateName {
//			isTemplateavailable = true
//			if dataServiceId == appConfigs[i].GetDataServiceId() {
//				appConfigTemplateID = appConfigs[i].GetId()
//				isavailable = true
//			}
//		}
//	}
//	if !(isavailable && isTemplateavailable) {
//		log.Errorf("App Config Template with name %v does not exist", appConfigTemplateName)
//	}
//	return appConfigTemplateID, nil
//}

// NewControlPlane to create control plane instance.
func NewControlPlane(url string, components *api.Components) *ControlPlane {
	return &ControlPlane{
		ControlPlaneURL: url,
		components:      components,
	}
}
