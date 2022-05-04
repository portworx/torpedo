package tests

type TestCase struct {
	TestName string
	TestArgs map[string]string
}

type TestSuite struct {
	TestCasesList []TestCase
}
