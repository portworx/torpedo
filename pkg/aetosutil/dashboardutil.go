package aetosutil

import (
	"fmt"
	"github.com/portworx/torpedo/pkg/log"
	rest "github.com/portworx/torpedo/pkg/restutil"
	"github.com/sirupsen/logrus"
	"net/http"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var testCasesStack = make([]TestCase, 0)
var verifications = make([]result, 0)
var testCaseStartTime time.Time

const (
	dashBoardBaseURL = "http://aetos.pwx.purestorage.com/dashboard" //"http://aetos-dm.pwx.purestorage.com:3939/dashboard"
)

const (
	//PASS status for testset/testcase
	PASS = "PASS"
	//FAIL status for testset/testcase
	FAIL = "FAIL"
	//ABORT status for testset/testcase
	ABORT = "ABORT"
	//TIMEOUT status for testset/testcase
	TIMEOUT = "TIMEOUT"
	//ERROR status for testset/testcase
	ERROR = "ERROR"
	// NOTSTARTED  status for testset/testcase
	NOTSTARTED = "NOT_STARTED"
	// INPROGRESS  status for testset/testcase
	INPROGRESS = "IN_PROGRESS"
)

var workflowStatuses = []string{PASS, FAIL, ABORT, ERROR, TIMEOUT, NOTSTARTED, INPROGRESS}

//Dashboard aetos dashboard structure
type Dashboard struct {
	//IsEnabled enable/disable dashboard logging
	IsEnabled bool
	//TestSetID test set ID to post the test logs and results
	TestsetID         int
	testcaseID        int
	verifications     []result
	testsetStartTime  time.Time
	testcaseStartTime time.Time
}

//TestSet struct
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

//TestCase struct
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

type comment struct {
	TestCaseID  int    `json:"testCaseID"`
	Description string `json:"description"`
	ResultType  string `json:"type"`
}

var tpLog *logrus.Logger

// TestSetBegin start testset and push data to dashboard DB
func (d *Dashboard) TestSetBegin(testSet *TestSet) {

	tpLog = log.GetLogInstance()

	tpLog = log.GetLogInstance()

	if testSet.Branch == "" {
		tpLog.Warn("Branch should not be empty")
	}

	if testSet.Description == "" {
		tpLog.Warn("Description should not be empty")
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
		tpLog.Errorf("Error in starting TestSet, Cause: %v", err)
	} else if respStatusCode != http.StatusOK {
		tpLog.Errorf("Failed to create TestSet, resp : %s", string(resp))
	} else {
		d.TestsetID, err = strconv.Atoi(string(resp))
		if err == nil {

			tpLog.Infof("TestSetId created : %d", d.TestsetID)
		} else {
			tpLog.Errorf("TestSetId creation failed. Cause : %v", err)
		}
		tpLog.Infof("Dashbaord URL : %s", fmt.Sprintf("http://aetos.pwx.purestorage.com/resultSet/testSetID/%d", d.TestsetID))

	}

}

// TestSetEnd  end testset and update  to dashboard DB
func (d *Dashboard) TestSetEnd() {

	if d.TestsetID == 0 {

		tpLog.Errorf("TestSetID is empty")
		return
	}

	updateTestSetURL := fmt.Sprintf("%s/testset/%d/end", dashBoardBaseURL, d.TestsetID)
	resp, respStatusCode, err := rest.PUT(updateTestSetURL, nil, nil, nil)

	if err != nil {
		tpLog.Errorf("Error in updating TestSet, Caose: %v", err)
	} else if respStatusCode != http.StatusOK {
		tpLog.Errorf("Failed to end TestSet, Resp : %s", string(resp))
	} else {

		tpLog.Infof("TestSetId %d update successfully", d.TestsetID)

	}
}

// TestCaseEnd update testcase  to dashboard DB

func (d *Dashboard) TestCaseEnd() {

	if d.testcaseID == 0 {

		tpLog.Error("TestCaseID is empty")
		return
	}

	url := fmt.Sprintf("%s/testcase/%d/end", dashBoardBaseURL, d.testcaseID)
	resp, respStatusCode, err := rest.PUT(url, nil, nil, nil)

	if err != nil {
		tpLog.Errorf("Error in updating TestCase, Caose: %v", err)
	} else if respStatusCode != http.StatusOK {
		tpLog.Errorf("Failed to end TestCase, Resp : %s", string(resp))
	} else {

		tpLog.Infof("TestCase %d ended successfully", d.testcaseID)

	}
}

// TestSetUpdate update test set  to dashboard DB
func (d *Dashboard) TestSetUpdate(testSet *TestSet) {

	if d.TestsetID == 0 {

		tpLog.Error("TestSetID is empty")
	}

	updateTestSetURL := fmt.Sprintf("%s/testset/%d", dashBoardBaseURL, d.TestsetID)
	resp, respStatusCode, err := rest.PUT(updateTestSetURL, testSet, nil, nil)

	if err != nil {
		tpLog.Errorf("Error in updating TestSet, Caose: %v", err)
	} else if respStatusCode != http.StatusOK {
		tpLog.Errorf("Failed to update TestSet, Resp : %s", string(resp))
	} else {
		tpLog.Infof("TestSetId %d update successfully", d.TestsetID)

	}
}

// TestCaseBegin start the test case and push data to dashboard DB

func (d *Dashboard) TestCaseBegin(moduleName, description, testRepoID string, tags []string) {

	if d.TestsetID == 0 {

		tpLog.Errorf("TestSetID is empty, cannot update update testcase")
		return
	}

	t := TestCase{}

	_, file, _, ok := runtime.Caller(1)
	if ok {

		m := regexp.MustCompile(`torpedo`)

		r := m.FindStringIndex(file)
		if r != nil {
			fp := file[r[0]:]
			t.Name = fp
			files := strings.Split(fp, "/")
			t.ShortName = files[len(files)-1]

			tpLog.Infof("Running test from file %s, module: %s", fp, moduleName)

		}
		t.ModuleName = moduleName

	}
	//t.StartTime = time.Now().Format(time.RFC3339)
	t.Status = INPROGRESS
	t.Description = description
	t.HostOs = runtime.GOOS

	t.TestSetID = d.TestsetID

	t.TestRepoID = testRepoID
	if tags != nil {
		t.Tags = tags
	}
	testCaseStartTime = time.Now()

	createTestCaseURL := fmt.Sprintf("%s/testcase", dashBoardBaseURL)

	resp, respStatusCode, err := rest.POST(createTestCaseURL, t, nil, nil)
	if err != nil {
		tpLog.Infof("Error in starting TesteCase, Cause: %v", err)
	} else if respStatusCode != http.StatusOK {
		tpLog.Errorf("Error creating test case, resp :%s", string(resp))
	} else {
		d.testcaseID, err = strconv.Atoi(string(resp))
		if err == nil {
			tpLog.Infof("TestCaseID created : %d", d.testcaseID)
		} else {
			tpLog.Errorf("TestCase creation failed. Cause : %v", err)
		}
	}
}

func (d *Dashboard) verify(r result) {

	if r.TestCaseID == 0 {
		tpLog.Errorf("TestcaseId should not be empty for updating result")
	}

	commentURL := fmt.Sprintf("%s/result", dashBoardBaseURL)

	resp, respStatusCode, err := rest.POST(commentURL, r, nil, nil)
	if err != nil {
		tpLog.Errorf("Error in verifying, Cause: %v", err)
	} else if respStatusCode != http.StatusOK {
		tpLog.Errorf("Error updating the vrify comment, resp : %s", string(resp))
	} else {
		tpLog.Tracef("verify response : %s", string(resp))

	}
}

//VerifySafely verify test without aborting the execution
func (d *Dashboard) VerifySafely(actual, expected interface{}, description string) {

	actualVal := fmt.Sprintf("%s", actual)
	expectedVal := fmt.Sprintf("%s", expected)
	res := result{}

	res.Actual = actualVal
	res.Expected = expectedVal
	res.Description = description
	res.TestCaseID = d.testcaseID

	tpLog.Infof("Verfy Safely: Description : %s", description)
	tpLog.Infof("Actual: %v, Expected : %v", actualVal, expectedVal)

	if actualVal == expectedVal {
		res.ResultType = "info"
		res.ResultStatus = true
	} else {
		res.ResultType = "error"
		res.ResultStatus = false
	}
	verifications = append(verifications, res)
	d.verify(res)

}

//VerifyFatal verify test and abort operation upon failure
func (d *Dashboard) VerifyFatal(actual, expected interface{}, description string) error {

	actualVal := fmt.Sprintf("%s", actual)
	expectedVal := fmt.Sprintf("%s", expected)
	res := result{}

	res.Actual = actualVal
	res.Expected = expectedVal
	res.Description = description
	res.TestCaseID = d.testcaseID

	tpLog.Infof("Verify Fatal: Description : %s", description)
	tpLog.Infof("Actual: %v, Expected : %v", actualVal, expectedVal)

	if actualVal == expectedVal {
		res.ResultType = "info"
		res.ResultStatus = true
	} else {
		res.ResultType = "error"
		res.ResultStatus = false
	}
	verifications = append(verifications, res)
	d.verify(res)

	if !res.ResultStatus {
		return fmt.Errorf("verification for %s has failed", description)
	}
	return nil

}

// Info logging info message
func (d *Dashboard) Info(message string) {

	res := comment{}
	res.TestCaseID = d.testcaseID
	res.Description = message
	res.ResultType = "info"
	d.addComment(res)
}

// Infof logging info with formated message
func (d *Dashboard) Infof(message string, args ...interface{}) {
	fmtMsg := fmt.Sprintf(message, args...)

	res := comment{}
	res.TestCaseID = d.testcaseID
	res.Description = fmtMsg
	res.ResultType = "info"
	d.addComment(res)
}

// Warnf logging formatted warn message
func (d *Dashboard) Warnf(message string, args ...interface{}) {
	fmtMsg := fmt.Sprintf(message, args...)
	res := comment{}
	res.TestCaseID = d.testcaseID
	res.Description = fmtMsg
	res.ResultType = "warning"
	d.addComment(res)
}

// Warn logging warn message
func (d *Dashboard) Warn(message string) {

	res := comment{}
	res.TestCaseID = d.testcaseID
	res.Description = message
	res.ResultType = "warning"
	d.addComment(res)
}

// Error logging error message
func (d *Dashboard) Error(message string) {

	res := comment{}
	res.TestCaseID = d.testcaseID
	res.Description = message
	res.ResultType = "error"
	d.addComment(res)
}

// Errorf logging formatted error message
func (d *Dashboard) Errorf(message string, args ...interface{}) {
	fmtMsg := fmt.Sprintf(message, args...)
	res := comment{}
	res.TestCaseID = d.testcaseID
	res.Description = fmtMsg
	res.ResultType = "error"
	d.addComment(res)
}

func (d *Dashboard) addComment(c comment) {

	if c.TestCaseID == 0 {
		tpLog.Errorf("TestcaseId should not be empty for updating result")
	}

	commentURL := fmt.Sprintf("%s/result", dashBoardBaseURL)

	resp, respStatusCode, err := rest.POST(commentURL, c, nil, nil)
	if err != nil {
		tpLog.Errorf("Error in verifying, Cause: %v", err)
	} else if respStatusCode != http.StatusOK {
		tpLog.Errorf("Error updating the vrify comment, resp : %s", string(resp))
	} else {
		tpLog.Tracef("verify response : %s", string(resp))
	}
}

//Get returns the dashboard struct instance
func Get() *Dashboard {
	return &Dashboard{}
}
