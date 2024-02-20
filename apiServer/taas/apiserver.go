package main

import (
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/portworx/torpedo/apiServer/taas/utils"
	"log"
)

// We will define all API calls here.
// Once Gin Server starts, it will initialise all APIs it contains.
// Future work : To have segregated APIs based on need -> We will have to create multiple main calls for initialising.
func main() {

	val1 := flag.String("key1", "value1", "test1")
	val2 := flag.String("key2", "value2", "test2")
	val3 := flag.String("key3", "value3", "test3")
	fmt.Println("Got flag - " + *val1)
	log.Println("Got flag - " + *val1)
	fmt.Println("Got flag - " + *val2)
	log.Println("Got flag - " + *val2)
	fmt.Println("Got flag - " + *val3)
	log.Println("Got flag - " + *val3)
	router := gin.Default()
	router.DELETE("taas/deletens/:namespace", utils.DeleteNS)
	router.POST("taas/createns", utils.CreateNS)
	router.POST("taas/inittorpedo", utils.InitializeDrivers)
	router.GET("taas/getnodes", utils.GetNodes)
	router.POST("taas/rebootnode/:nodename", utils.RebootNode)
	router.GET("taas/storagenodes", utils.GetStorageNodes)
	router.GET("taas/storagelessnodes", utils.GetStorageLessNodes)
	router.POST("taas/collectsupport", utils.CollectSupport)
	router.POST("taas/scheduleapps/:appName", utils.ScheduleAppsAndValidate)
	router.POST("taas/deploypxagent", utils.ExecuteHelmCmd)
	router.GET("taas/getclusterid/:namespace", utils.GetNamespaceID)
	router.GET("taas/getclusternodestatus", utils.GetNodeStatus)
	router.POST("taas/runhelmcmd", utils.ExecuteHelmCmd)
	router.GET("taas/pxversion", utils.GetPxVersion)
	router.GET("taas/ispxinstalled", utils.IsPxInstalled)
	router.GET("taas/getpxctloutput", utils.GetPxctlStatusOutput)
	//tests.ParseFlags()
	//flag.Parse()
	log.Fatal(router.Run(":8080"))
}
