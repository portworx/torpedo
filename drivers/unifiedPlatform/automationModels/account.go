package automationModels

type PlatformAccount struct {
	Get     PlatformGetAccount            `copier:"must,nopanic"`
	Onboard PlatformOnboardAccountRequest `copier:"must,nopanic"`
}

type PlatformGetAccount struct {
	AccountId string `copier:"must,nopanic"`
}
