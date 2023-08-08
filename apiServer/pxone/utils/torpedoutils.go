package utils

import (
	"github.com/gin-gonic/gin"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"github.com/portworx/torpedo/drivers/node"
	"github.com/portworx/torpedo/tests"
	"net/http"
	"testing"
)

var (
	IsTorpedoInitDone bool
)

// Method to create a testing instance to call ginkgo tests programmatically
func t() *testing.T {
	return &testing.T{}
}

// InitializeDrivers : This API Call will init all Torpedo Drivers. This needs to be run as ginkgo test
// as multiple ginkgo and gomega dependencies are being called in InitInstance()
func InitializeDrivers(c *gin.Context) {
	var _ = ginkgo.Describe("Initialise Inst", func() {
		ginkgo.It("init instance", func() {
			gomega.Expect(func() {
				tests.ParseFlags()
				tests.InitInstance()
				IsTorpedoInitDone = true
			}).To(gomega.Panic())
		})
	})
	ginkgo.RunSpecs(ginkgo.GinkgoT(), "Initialise Inst")
}

// GetNodes : This API will return list of all nodes in the Cluster
func GetNodes(c *gin.Context) {
	if !IsTorpedoInitDone {
		InitializeDrivers(c)
		if !IsTorpedoInitDone {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Torpedo Init failed due to error",
				"nodes":   nil,
			})
		}
	} else {
		nodes := node.GetWorkerNodes()
		c.JSON(http.StatusOK, gin.H{
			"message": "Nodes are: ",
			"nodes":   nodes,
		})
	}
}
