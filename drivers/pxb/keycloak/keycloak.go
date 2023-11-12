package keycloak

import (
	"context"
	"fmt"
	. "github.com/portworx/torpedo/drivers/pxb/pxbutils"
	"github.com/portworx/torpedo/pkg/log"
	"io"
	"net/http"
	"net/url"
)

type Keycloak struct {
	SignIn struct {
		Username string
		Password string
	}
	Client *http.Client
}

func (k *Keycloak) GetBasePath(admin bool) string {
	if admin {
		return "/auth/admin/realms/master"
	}
	return "/auth/realms/master"
}

func (k *Keycloak) Process(ctx context.Context, method string, admin bool, route string, body io.Reader, headerMap map[string]string) ([]byte, error) {
	reqURL, err := url.JoinPath(k.URI, k.GetBasePath(admin), route)
	if err != nil {
		return nil, ProcessError(err)
	}
	req, err := http.NewRequestWithContext(ctx, method, reqURL, body)
	if err != nil {
		return nil, ProcessError(err)
	}
	for key, val := range headerMap {
		req.Header.Set(key, val)
	}
	resp, err := k.Client.Do(req)
	if err != nil {
		return nil, ProcessError(err)
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Errorf("failed to close response body. Err: [%v]", ProcessError(err))
		}
	}()
	statusCode := resp.StatusCode
	statusText := http.StatusText(statusCode)
	switch {
	case statusCode >= 200 && statusCode < 300:
		return io.ReadAll(resp.Body)
	default:
		err = fmt.Errorf("[%s] [%s] returned status [%d]: [%s]", method, url, statusCode, statusText)
		return nil, ProcessError(err)
	}
}
