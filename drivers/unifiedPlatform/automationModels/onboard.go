package automationModels

// PlatformOnboardAccountRequest to create account.
type PlatformOnboardAccountRequest struct {
	// account to be created and onboarded
	Register *PlatformRegisterAccount `copier:"must,nopanic"`
}

// PlatformOnboardAccountRequest to create account.
type PlatformOnboardAccountResponse struct {
	// account to be created and onboarded
	Register *AccountRegistration `copier:"must,nopanic"`
}

type PlatformRegisterAccount struct {
	AccountRegistration *AccountRegistration `copier:"must,nopanic"`
}

type AccountRegistration struct {
	// Metadata of the account.
	Meta *Meta `copier:"must,nopanic"`
	// Configuration info used for registering the account
	Config *AccountConfig `copier:"must,nopanic"`
}

type AccountConfig struct {
	// Desired configuration of the Account
	AccountConfig *PlatformAccountConfig `copier:"must,nopanic"`
}

type PlatformAccountConfig struct {
	// email of the first account user
	UserEmail string `copier:"must,nopanic"`
	// Account DNS name
	DnsName string `copier:"must,nopanic"`
	// Display name of Account
	DisplayName string `copier:"must,nopanic"`
	// Describes whether it is a Freemium or an Enterprise account
	AccountType AccountType `copier:"must,nopanic"`
}

type AccountType int32
