package automationModels

type PlatformAccount struct {
	Get     PlatformGetAccount            `copier:"must,nopanic"`
	Onboard PlatformOnboardAccountRequest `copier:"must,nopanic"`
}

type PlatformAccountResponse struct {
	Get V1Account
}

type V1Account struct {
	Meta   *V1Meta          `json:"meta,omitempty"`
	Config *V1Config        `json:"config,omitempty"`
	Status *Accountv1Status `json:"status,omitempty"`
}

type Accountv1Status struct {
	Reason *string      `json:"reason,omitempty"`
	Phase  *V1PhaseType `json:"phase,omitempty"`
}

type PlatformGetAccount struct {
	AccountId string `copier:"must,nopanic"`
}
