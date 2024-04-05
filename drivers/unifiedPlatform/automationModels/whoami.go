package automationModels

type WhoamiResponse struct {
	Id       string     `copier:"must,nopanic"`
	Email    string     `copier:"must,nopanic"`
	Accounts []*Account `copier:"must,nopanic"`
}

type Account struct {
	Id      string `copier:"must,nopanic"`
	DnsName string `copier:"must,nopanic"`
	Name    string `copier:"must,nopanic"`
}
