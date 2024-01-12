package pxbackupworkflows

import (
	"math/rand"
	"time"

	"github.com/portworx/torpedo/pkg/log"
)

// ...

func GetRandomNamespacesForBackup() []string {
	var allNamespacesForBackupMap = make(map[string]bool)
	var allNamepsacesForBackup []string
	rand.Seed(time.Now().Unix()) // initialize global pseudo random generator

	numberOfNamespaces := rand.Intn(len(AllNamespaces))

	for i := 0; i <= numberOfNamespaces; i++ {
		allNamespacesForBackupMap[AllNamespaces[rand.Intn(len(AllNamespaces))]] = true
	}

	for namespaceName, _ := range allNamespacesForBackupMap {
		allNamepsacesForBackup = append(allNamepsacesForBackup, namespaceName)
	}

	log.Infof("Returning This - [%v]", allNamepsacesForBackup)
	return allNamepsacesForBackup
}
