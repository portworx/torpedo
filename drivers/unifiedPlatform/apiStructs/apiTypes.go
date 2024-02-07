package apiStructs

import "time"

type Meta struct {
	Uid             *string            `copier:"must"`
	Name            *string            `copier:"must"`
	Description     *string            `copier:"must,nopanic"`
	ResourceVersion *string            `copier:"must,nopanic"`
	CreateTime      *time.Time         `copier:"must,nopanic"`
	UpdateTime      *time.Time         `copier:"must,nopanic"`
	Labels          *map[string]string `copier:"must,nopanic"`
	Annotations     *map[string]string `copier:"must,nopanic"`
}

type Config struct {
	UserEmail   *string `copier:"must"`
	DnsName     *string `copier:"must"`
	DisplayName *string `copier:"must"`
}

type Account struct {
	Meta   Meta   `copier:"must"`
	Config Config `copier:"must"`
}
