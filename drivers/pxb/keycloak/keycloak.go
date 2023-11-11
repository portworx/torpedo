package keycloak

import (
	"net/http"
)

type Keycloak struct {
	URI    string
	SignIn struct {
		Username string
		Password string
	}
	Client *http.Client
}
