package aetosutil

import (
	"fmt"
	rest "github.com/portworx/torpedo/pkg/restutil"
	"github.com/sirupsen/logrus"
	"net/http"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var (
	TestSetID  int
	TestCaseID int
)

var testCasesStack = make([]TestCase, 0)
var verifications = make([]result, 0)
var testCaseStartTime time.Time

const (
	dashBoardBaseURL = "http://aetos.pwx.purestorage.com/dashboard" //"http://aetos-dm.pwx.purestorage.com:3939/dashboard"
)

const (
	PASS        = "PASS"
	FAIL        = "FAIL"
	ABORT       = "ABORT"
	TIMEOUT     = "TIMEOUT"
	ERROR       = "ERROR"
	NOT_STARTED = "NOT_STARTED"
	IN_PROGRESS = "IN_PROGRESS"
)

var workflowStatuses = []string{PASS, FAIL, ABORT, ERROR, TIMEOUT, NOT_STARTED, IN_PROGRESS}

type TestSet struct {
	CommitID    string   `json:"commitId"`
	User        string   `json:"user"`
	Product     string   `json:"product"`
	Description string   `json:"description"`
	HostOs      string   `json:"hostOs"`
	Branch      string   `json:"branch"`
	TestType    string   `json:"testType"`
	Tags        []string `json:"tags"`
	Status      string   `json:"status"`
}

type TestCase struct {
	Name       string `json:"name"`
	ShortName  string `json:"shortName"`
	ModuleName string `json:"moduleName"`

	Status      string   `json:"status"`
	Errors      []string `json:"errors"`
	LogFile     string   `json:"logFile"`
	Description string   `json:"description"`
	Command     string   `json:"command"`
	HostOs      string   `json:"hostOs"`
	Tags        []string `json:"tags"`
	TestSetID   int      `json:"testSetID"`
	TestRepoID  string   `json:"testRepoID"`
	Duration    string   `json:"duration"`
}

type result struct {
	TestCaseID   int    `json:"testCaseID"`
	Description  string `json:"description"`
	Actual       string `json:"actual"`
	Expected     string `json:"expected"`
	ResultType   string `json:"type"`
	ResultStatus bool   `json:"result"`
}

// TestSetBegin start testset and push data to dashboard DB
func TestSetBegin(testSet *TestSet) {

	if testSet.Branch == "" {
		logrus.Warn("Branch should not be empty")
	}

	if testSet.Description == "" {
		logrus.Warn("Description should not be empty")
	}

	if testSet.Product == "" {
		testSet.Product = "Portworx Enterprise"
	}

	if testSet.HostOs == "" {
		testSet.HostOs = runtime.GOOS
	}

	createTestSetURL := fmt.Sprintf("%s/testset", dashBoardBaseURL)

	resp, respStatusCode, err := rest.POST(createTestSetURL, testSet, nil, nil)
	if err != nil {
		logrus.Errorf("Error in starting TestSet, Cause: %v", err)
	} else if respStatusCode != http.StatusOK {
		logrus.Errorf("Failed to create TestSet, resp : %s", string(resp))
	} else {
		TestSetID, err = strconv.Atoi(string(resp))
		if err == nil {
			logrus.Infof("TestSetId created : %d", TestSetID)
		} else {
			logrus.Errorf("TestSetId creation failed. Cause : %v", err)
		}
	}

}

// TestSetUpdate update test set  to dashboard DB
func TestSetEnd() {

	if TestSetID == 0 {
		logrus.Fatal("TestSetID is empty")
	}

	updateTestSetURL := fmt.Sprintf("%s/testset/%d/end", dashBoardBaseURL, TestSetID)
	resp, respStatusCode, err := rest.PUT(updateTestSetURL, nil, nil, nil)

	if err != nil {
		logrus.Errorf("Error in updating TestSet, Caose: %v", err)
	} else if respStatusCode != http.StatusOK {
		logrus.Errorf("Failed to end TestSet, Resp : %s", string(resp))
	} else {
		logrus.Infof("TestSetId %d update successfully", TestSetID)
	}
}

func TestCaseEnd() {

	if TestCaseID == 0 {
		logrus.Fatal("TestCaseID is empty")
	}

	url := fmt.Sprintf("%s/testcase/%d/end", dashBoardBaseURL, TestCaseID)
	resp, respStatusCode, err := rest.PUT(url, nil, nil, nil)

	if err != nil {
		logrus.Errorf("Error in updating TestCase, Caose: %v", err)
	} else if respStatusCode != http.StatusOK {
		logrus.Errorf("Failed to end TestCase, Resp : %s", string(resp))
	} else {
		logrus.Infof("TestCase %d ended successfully", TestSetID)
	}
}

// TestSetUpdate update test set  to dashboard DB
func TestSetUpdate(testSet *TestSet) {

	if TestSetID == 0 {
		logrus.Fatal("TestSetID is empty")
	}

	updateTestSetURL := fmt.Sprintf("%s/testset/%d", dashBoardBaseURL, TestSetID)
	resp, respStatusCode, err := rest.PUT(updateTestSetURL, testSet, nil, nil)

	if err != nil {
		logrus.Errorf("Error in updating TestSet, Caose: %v", err)
	} else if respStatusCode != http.StatusOK {
		logrus.Errorf("Failed to update TestSet, Resp : %s", string(resp))
	} else {
		logrus.Infof("TestSetId %d update successfully", TestSetID)
	}
}

// TestCaseBegin start the test case and push data to dashboard DB
func TestCaseBegin(moduleName, description, testRepoId string, tags []string) {

	if TestSetID == 0 {
		logrus.Fatal("TestSetID is empty, cannot update update testcase")
	}

	t := TestCase{}

	pc, file, line, ok := runtime.Caller(1)
	if ok {
		fmt.Printf("Called from %s, line #%d, func: %v\n",
			file, line, runtime.FuncForPC(pc).Name())

		m := regexp.MustCompile(`torpedo`)

		r := m.FindStringIndex(file)
		if r != nil {
			fp := file[r[0]:]
			t.Name = fp
			files := strings.Split(fp, "/")
			t.ShortName = files[len(files)-1]

			logrus.Infof("Running test from file %s, module: %s", fp, moduleName)

		}
		t.ModuleName = moduleName

	}
	//t.StartTime = time.Now().Format(time.RFC3339)
	t.Status = IN_PROGRESS
	t.Description = description
	t.HostOs = runtime.GOOS
	t.TestSetID = TestSetID
	t.TestRepoID = testRepoId
	if tags != nil {
		t.Tags = tags
	}
	testCaseStartTime = time.Now()

	createTestCaseURL := fmt.Sprintf("%s/testcase", dashBoardBaseURL)

	resp, respStatusCode, err := rest.POST(createTestCaseURL, t, nil, nil)
	if err != nil {
		logrus.Infof("Error in starting TesteCase, Cause: %v", err)
	} else if respStatusCode != http.StatusOK {
		logrus.Errorf("Error creating test case, resp :%s", string(resp))
	} else {
		TestCaseID, err = strconv.Atoi(string(resp))
		if err == nil {
			logrus.Infof("TestCaseID created : %d", TestCaseID)
		} else {
			logrus.Errorf("TestCase creation failed. Cause : %v", err)
		}
	}
}

func verify(r result) {

	if r.TestCaseID == 0 {
		logrus.Fatal("TestcaseId should not be empty for updating result")
	}

	commentURL := fmt.Sprintf("%s/result", dashBoardBaseURL)

	resp, respStatusCode, err := rest.POST(commentURL, r, nil, nil)
	if err != nil {
		logrus.Infof("Error in verifying, Cause: %v", err)
	} else if respStatusCode != http.StatusOK {
		logrus.Errorf("Error updating the vrify comment, resp : %s", string(resp))
	} else {
		logrus.Infof("verify response : %s", string(resp))
	}
}

//VerifySafely verify test without aborting the execution
func VerifySafely(actual, expected interface{}, description string) {

	actualVal := fmt.Sprintf("%s", actual)
	expectedVal := fmt.Sprintf("%s", expected)
	res := result{}

	res.Actual = actualVal
	res.Expected = expectedVal
	res.Description = description
	res.TestCaseID = TestCaseID

	logrus.Infof("Verfy Safely: Description : %s", description)
	logrus.Infof("Actual: %v, Expected : %v", actualVal, expectedVal)

	if actualVal == expectedVal {
		res.ResultType = "info"
		res.ResultStatus = true
	} else {
		res.ResultType = "error"
		res.ResultStatus = false
	}
	verifications = append(verifications, res)
	verify(res)

}

//VerifyFatal verify test and abort operation upon failure
func VerifyFatal(actual, expected interface{}, description string) error {

	actualVal := fmt.Sprintf("%s", actual)
	expectedVal := fmt.Sprintf("%s", expected)
	res := result{}

	res.Actual = actualVal
	res.Expected = expectedVal
	res.Description = description
	res.TestCaseID = TestCaseID

	logrus.Infof("Verify Fatal: Description : %s", description)
	logrus.Infof("Actual: %v, Expected : %v", actualVal, expectedVal)

	if actualVal == expectedVal {
		res.ResultType = "info"
		res.ResultStatus = true
	} else {
		res.ResultType = "error"
		res.ResultStatus = false
	}
	verifications = append(verifications, res)
	verify(res)

	if !res.ResultStatus {
		return fmt.Errorf("verification for %s has failed", description)
	}
	return nil

}

// Info logging info message
func Info(message string) {

	res := result{}
	res.TestCaseID = TestCaseID
	res.Description = message
	res.ResultType = "info"
	verify(res)
}

// Warning logging info message
func Warning(message string) {

	res := result{}
	res.TestCaseID = TestCaseID
	res.Description = message
	res.ResultType = "warning"
	verify(res)
}

// Error logging info message
func Error(message string) {

	res := result{}
	res.TestCaseID = TestCaseID
	res.Description = message
	res.ResultType = "error"
	verify(res)
}
