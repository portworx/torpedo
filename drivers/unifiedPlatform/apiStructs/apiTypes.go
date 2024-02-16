package apiStructs

import (
	"context"
	"fmt"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	platformv1 "github.com/pure-px/platform-api-go-client/v1alpha1"
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

type ApiResponse struct {
	Meta   Meta
	Config Config
}

type Credentials struct {
	Token string
}

type BackupLocationParams struct {
	BackupLocName     string
	Provider          platformv1.V1Provider
	CloudCredentialId string
	AzureStorage      platformv1.V1AzureBlobStorage
	GoogleStorage     platformv1.V1GoogleCloudStorage
	S3Storage         platformv1.V1S3ObjectStorage
	Validity          platformv1.StatusValidity
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
