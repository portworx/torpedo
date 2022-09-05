package main

import (
	"github.com/portworx/torpedo/pkg/aetosutil"
	"time"
)

func main() {

	t := aetosutil.TestSet{
		CommitID:    "2.12.0-serfdf",
		User:        "lsrinivas",
		Product:     "PxEnp",
		Description: "torpedo desc",
		HostOs:      "linux",
		Branch:      "master",
		TestType:    "SystemTest",
		Tags:        []string{"sample1", "samp2"},
		Status:      aetosutil.NOT_STARTED,
	}

	aetosutil.TestSetBegin(&t)

	aetosutil.TestCaseBegin("Aetos", "Aetos dashboard integration", "", nil)

	aetosutil.Info("This is info message")
	aetosutil.Warning("This is warning message")
	aetosutil.VerifySafely("true", "true", "Equating true")
	aetosutil.VerifySafely("true", "false", "Equating false")
	aetosutil.VerifyFatal(2, 2, "harding equating 2")
	aetosutil.VerifyFatal(2, 3, "harding equating 3")
	aetosutil.Error("This is error string")
	time.Sleep(10 * time.Second)
	aetosutil.Error("came out of sleep")
	aetosutil.TestCaseEnd()
	aetosutil.TestSetEnd()

}
