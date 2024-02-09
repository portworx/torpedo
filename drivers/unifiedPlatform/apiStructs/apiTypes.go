package apiStructs

import (
	"context"
	"fmt"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"time"
)

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

type Credentials struct {
	Token string
}

func (c *Credentials) GetRequestMetadata(_ context.Context, _ ...string) (map[string]string, error) {
	metadata := map[string]string{}

	if c.Token == "" {
		_, token, err := GetBearerToken()
		if err != nil {
			return nil, fmt.Errorf("get berare token, err: %s", err)
		}

		c.Token = token
	}

	metadata["Authorization"] = fmt.Sprintf("Bearer %s", c.Token)

	return metadata, nil
}

func (c *Credentials) RequireTransportSecurity() bool {
	return false
}
