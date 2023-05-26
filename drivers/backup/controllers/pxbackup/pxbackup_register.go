package pxbackup

import (
	"fmt"
	"github.com/portworx/torpedo/drivers/backup"
	"github.com/portworx/torpedo/drivers/backup/utils"
)

type RegisterNewUserConfig struct {
	username  string
	firstName string
	lastName  string
	email     string
	password  string
}

func NewUser(username string, password string) *RegisterNewUserConfig {
	return &RegisterNewUserConfig{
		username:  username,
		password:  password,
		firstName: "first-" + username,
		lastName:  "last-" + username,
		email:     username + "@cnbu.com",
	}
}

func (c *RegisterNewUserConfig) Register() error {
	err := backup.AddUser(c.username, c.firstName, c.lastName, c.email, c.password)
	if err != nil {
		debugMessage := fmt.Sprintf("username [%s]; first-name [%s]; last-name [%s]; email [%s]", c.username, c.firstName, c.lastName, c.email)
		return utils.ProcessError(err, debugMessage)
	}
	return nil
}
