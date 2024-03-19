package ginkgo_dsl

import (
	"github.com/onsi/ginkgo/v2"
	"github.com/portworx/torpedo/pkg/log"
)

var NewDescribe = func(text string, args ...interface{}) bool {
	log.Infof("NewDescribe: %s", text)
	return ginkgo.Describe(text, args...)
}

var NewIt = func(text string, args ...interface{}) bool {
	log.Infof("NewIt: %s", text)
	newArgs := make([]interface{}, 0)
	if text == "It 1B" {
		//newArgs = []interface{}{ginkgo.Label("test-It1")}
	}
	newArgs = append(newArgs, args...)
	return ginkgo.It(text, newArgs)
}

var NewStep = func(text string, args ...func()) {
	log.Infof("NewStep: %s", text)
	ginkgo.By(text, args...)
}
