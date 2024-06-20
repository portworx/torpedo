package users

import (
	"github.com/rancher/shepherd/pkg/namegenerator"
)

const (
	defaultPasswordLength = 12
)

func GenerateUserPassword(password string) string {
	return namegenerator.RandStringLower(defaultPasswordLength)
}
