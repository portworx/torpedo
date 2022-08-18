package aetosutil

import (
	"encoding/json"
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

// GetTestSet returns testset from dashboard DB
func GetTestSet(testSetId int) *TestSet {

	if testSetId == 0 {
		testSetId = TestSetID
	}

	if testSetId == 0 {
		logrus.Fatal("TestSetID is empty")
	}

	getTestSetURL := fmt.Sprintf("%s/testset/%d", dashBoardBaseURL, testSetId)
	resp, respStatusCode, err := rest.Get(getTestSetURL, nil, nil)

	if err != nil {
		logrus.Errorf("Error in gettting TestSet: %d, Cause: %v", testSetId, err)
		return nil
	}

	if respStatusCode != http.StatusOK {
		logrus.Errorf("Error in gettting TestSet: %d, Resp: %v", testSetId, string(resp))
		return nil
	}

	var testSet TestSet
	err = json.Unmarshal(resp, &testSet)
	if err != nil {
		logrus.Errorf("Error un marshalling testset get, Resp :%s", string(resp))
		return nil
	}

	return &testSet
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

func GetTestCase(testCaseID int) *TestCase {
	if testCaseID == 0 {
		testCaseID = TestCaseID
	}

	if testCaseID == 0 {
		logrus.Fatal("TestCaseID is empty")

	}

	getTestCaseURL := fmt.Sprintf("%s/testcase/%d", dashBoardBaseURL, testCaseID)
	resp, respStatusCode, err := rest.Get(getTestCaseURL, nil, nil)

	if err != nil {
		logrus.Infof("Error in gettting TestCase: %d, Cause: %v", testCaseID, err)
		return nil
	}

	if respStatusCode != http.StatusOK {
		logrus.Errorf("Error in gettting TestCase: %d, Resp: %v", testCaseID, string(resp))
		return nil
	}

	var testCase TestCase
	err = json.Unmarshal(resp, &testCase)
	if err != nil {
		logrus.Errorf("Error un marshalling testcase get, Resp :%s,Cause : %v ", string(resp), err)
		return nil
	}

	return &testCase
}

func TestCaseUpdate(testCase *TestCase) {
	if TestCaseID == 0 {
		logrus.Fatal("TestCaseID is empty")
	}

	updateTestCaseURL := fmt.Sprintf("%s/testcase/%d", dashBoardBaseURL, TestCaseID)
	resp, respStatusCode, err := rest.PUT(updateTestCaseURL, testCase, nil, nil)

	if err != nil {
		logrus.Infof("Error in updating TestCase, Cause: %v", err)
	} else if respStatusCode != http.StatusOK {
		logrus.Errorf("Error updating test case, Resp : %s", string(resp))
	} else {
		logrus.Infof("TestCaseId %d update successfully", TestCaseID)
	}
}

func TestCaseEnd(status string) {
	time.Sleep(10 * time.Second)

	testCaseStatus := PASS

	if TestCaseID == 0 {
		logrus.Fatal("TestCaseID is empty")
	}

	testcase := GetTestCase(TestCaseID)

	endTime := time.Now()
	duration := endTime.Sub(testCaseStartTime)
	testcase.Duration = fmt.Sprint(duration.Minutes())
	fmt.Printf("Duration : %v\n", duration.Minutes())
	//testcase.EndTime = endTime
	//errors, verifications
	if status != "" {
		testcase.Status = status
	} else {
		for _, r := range verifications {
			if !r.ResultStatus {
				testCaseStatus = FAIL
				break
			}
		}
		logrus.Infof("Updating testcase status as %s", testCaseStatus)
		testcase.Status = testCaseStatus
	}

	logrus.Infof("test case before update : %v", testcase)

	TestCaseUpdate(testcase)

}

func getAllTestCases() []TestCase {

	if TestSetID == 0 {
		logrus.Fatal("TestCaseID is empty")

	}

	getTestCaseURL := fmt.Sprintf("%s/testcases/%d", dashBoardBaseURL, TestSetID)
	resp, respStatusCode, err := rest.Get(getTestCaseURL, nil, nil)

	if err != nil {
		logrus.Infof("Error in getting TestCases for TestSet: %d, Cause: %v", TestSetID, err)
		return nil
	}

	if respStatusCode != http.StatusOK {
		logrus.Errorf("Error in getting TestCases: %d, Resp: %v", TestSetID, string(resp))
		return nil
	}

	var testCases []TestCase
	err = json.Unmarshal(resp, &testCases)
	if err != nil {
		logrus.Errorf("Error un marshalling testcases get, Resp :%s", string(resp))
		return nil
	}

	return testCases
}

func TestSetEnd() {

	testcases := getAllTestCases()
	overallStatus := PASS

	for _, t := range testcases {
		if t.Status == IN_PROGRESS || t.Status == NOT_STARTED {
			overallStatus = t.Status
			break
		}
	}

	for _, t := range testcases {
		logrus.Infof("Verification : %v", t)
		if isFailure(t.Status) {
			overallStatus = t.Status
			break
		}
	}

	testSet := GetTestSet(0)
	logrus.Infof("Updating testset status as %s", overallStatus)
	testSet.Status = overallStatus
	TestSetUpdate(testSet)

}

func getFailureArray() []string {
	failureArray := make([]string, 0)

	for _, v := range workflowStatuses {
		if v != PASS && v != NOT_STARTED && v != IN_PROGRESS {
			failureArray = append(failureArray, v)
		}
	}

	return failureArray
}

func isFailure(status string) bool {

	failureArray := getFailureArray()

	for _, v := range failureArray {
		if status == v {
			return true
		}
	}
	return false
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
