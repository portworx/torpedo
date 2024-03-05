package stworkflows

import (
	_ "github.com/gobwas/glob/syntax/ast"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/platformLibs"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
)

func RunTestWithRbac(accName, roleName, ResourceId string) error {
	jwtToken, err := platformLibs.RunWithRbac(accName, roleName, ResourceId)
	if err != nil {
		return err
	}
	utils.RunWithRBAC = utils.RunWithRbac{
		RbacFlag:  true,
		RbacToken: jwtToken,
	}
	return nil
}
