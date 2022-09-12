package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"os"
)

func main() {

	//t := aetosutil.TestSet{
	//	CommitID:    "2.12.0-serfdf",
	//	User:        "lsrinivas",
	//	Product:     "PxEnp",
	//	Description: "torpedo desc",
	//	HostOs:      "linux",
	//	Branch:      "master",
	//	TestType:    "SystemTest",
	//	Tags:        []string{"sample1", "samp2"},
	//	Status:      aetosutil.NOTSTARTED,
	//}
	//
	//aetosutil.TestSetBegin(&t)
	//
	//aetosutil.TestCaseBegin("Aetos", "Aetos dashboard integration", "", nil)
	//
	//aetosutil.Info("This is info message")
	//aetosutil.Warn("This is warning message")
	//aetosutil.VerifySafely("true", "true", "Equating true")
	//aetosutil.VerifySafely("true", "false", "Equating false")
	//aetosutil.VerifyFatal(2, 2, "harding equating 2")
	//aetosutil.VerifyFatal(2, 3, "harding equating 3")
	//aetosutil.Error("This is error string")
	//
	//time.Sleep(10 * time.Second)
	//aetosutil.Error("came out of sleep")
	//aetosutil.TestCaseEnd()
	//
	//aetosutil.TestSetEnd()

	f, err := os.OpenFile("/Users/leela/workspace/testbeds/test-clus/log.txt", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("Failed to create logfile" + "log.txt")
		panic(err)
	}
	f2, err := os.OpenFile("/Users/leela/workspace/testbeds/test-clus/temp.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("Failed to create logfile" + "log.txt")
		panic(err)
	}
	//defer f.Close()
	//defer f2.Close()

	log := &logrus.Logger{
		// Log into f file handler and on os.Stdout
		Out:       io.MultiWriter(f, f2, os.Stdout),
		Level:     logrus.DebugLevel,
		Formatter: &logrus.TextFormatter{},
	}

	log.Trace("Trace message")
	log.Info("Info message")
	log.Warn("Warn message")
	log.Error("Error message")
	log.Fatal("Fatal message")

}
