package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/portworx/torpedo/drivers/node"
	"github.com/portworx/torpedo/drivers/pds/lib"
	"github.com/portworx/torpedo/tests"
	"log"
	"net/http"
)

func DeleteNamespace(ctx *gin.Context) {
	ns := ctx.Param("namespace")
	err := lib.DeleteK8sNamespace(ns)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	} else {
		ctx.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Namespace %s Deleted", ns)})
	}
}

func CreateNS(c *gin.Context) {
	ns, err := lib.CreateTempNS(6)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to Create Namespace",
			"error":   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message":   "Namespace created successfully",
		"namespace": ns,
	})
}

func InitializeDrivers(c *gin.Context) {
	tests.InitInstance()
}

func GetNodes(c *gin.Context) {
	nodes := node.GetWorkerNodes()
	var list []string
	for _, workerNode := range nodes {
		list = append(list, fmt.Sprintf("%v", workerNode.Name))
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Nodes are: ",
		"nodes":   list,
	})
}

func main() {
	router := gin.Default()
	router.GET("pxone/init", InitializeDrivers)
	router.GET("pxone/nodes", GetNodes)
	router.DELETE("pxone/deletens/:namespace", DeleteNamespace)
	router.POST("pxone/createns", CreateNS)
	log.Fatal(router.Run(":8080"))
}
