package kubevirt

import (
	"fmt"
	. "github.com/pure-px/torpedo/drivers/utilities"
	"github.com/pure-px/torpedo/pkg/log"
	"strconv"
)

var (
	defaultFilePath = "/home/cirros/"
)

// GenerateRandomCommandToCreateFiles creates random textfiles with random data
func GenerateRandomCommandToCreateFiles(count int) map[string][]string {
	var randomFileCommands = make(map[string][]string)
	var filePath = defaultFilePath + RandomString(10)
	var insertCommands []string
	var selectCommands []string
	var deleteCommands []string
	var updateCommands []string

	// Generating command to create the dir to hold files if not exists
	createDir := fmt.Sprintf("mkdir -p -m777 %s", filePath)
	log.Infof("Command to create Dir - [%s]", createDir)
	insertCommands = append(insertCommands, createDir)

	for counter := 0; counter < count; counter++ {
		currentCounter := strconv.Itoa(counter)
		fileName := fmt.Sprintf("%s/%s_%s.txt", filePath, currentCounter, RandomString(4))
		// fileContent := fmt.Sprintf("%s", RandomString(10))
		insertCommands = append(insertCommands, fmt.Sprintf("touch %s", fileName))
		selectCommands = append(selectCommands, fmt.Sprintf("ls %s", fileName))
		updateCommands = append(updateCommands, fmt.Sprintf("echo '%s' >> %s", RandomString(5), fileName))
		deleteCommands = append(deleteCommands, fmt.Sprintf("rm %s", fileName))
	}

	randomFileCommands["insert"] = insertCommands
	randomFileCommands["select"] = selectCommands
	randomFileCommands["update"] = updateCommands
	randomFileCommands["delete"] = deleteCommands

	return randomFileCommands
}
