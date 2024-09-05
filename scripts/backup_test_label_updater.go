package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"github.com/portworx/torpedo/pkg/log"
	"github.com/portworx/torpedo/tests/backup"
	"io/ioutil"
	"os"
	"regexp"
	"sort"
	"strings"
)

// updateTestCaseLabelsMap updates the TestCaseLabelsMap
func updateTestCaseLabelsMap(testCaseName tests.TestCaseName, newLabel tests.TestCaseLabel) {
	log.Infof("Updating test case labels map for test case: %s, label: %s", testCaseName, newLabel)
	if labels, exists := tests.TestCaseLabelsMap[testCaseName]; exists {
		for _, label := range labels {
			if label == newLabel {
				log.Infof("Label already exists, no need to add")
				// Label already exists, no need to add
				return
			}
		}
		// Append the new label
		tests.TestCaseLabelsMap[testCaseName] = append(labels, newLabel)
		log.Infof("Label added to the test case")
		log.Infof("Updated labels: %v", tests.TestCaseLabelsMap[testCaseName])
	} else {
		log.Infof("Test case not found, adding new test case with the label")
		// Add the new test case with the label
		tests.TestCaseLabelsMap[testCaseName] = []tests.TestCaseLabel{newLabel}
	}
}

// updateMapFromCSV updates the global TestCaseLabelsMap based on the CSV file provided
func updateMapFromCSV(csvFile string) error {
	file, err := os.Open(csvFile)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return err
	}

	for _, record := range records {
		log.Infof("Record from csv: %v", record)
		if len(record) != 2 {
			return fmt.Errorf("invalid record: %v", record)
		}
		testCaseName := record[0]
		newLabel := record[1]
		updateTestCaseLabelsMap(testCaseName, newLabel)
	}
	return nil
}

// createLabelNamesMap creates a map of variable names and their corresponding values
func createLabelNamesMap(filename string) (map[tests.TestCaseLabel]string, error) {
	labelNames := make(map[tests.TestCaseLabel]string)

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	constPattern := regexp.MustCompile(`(\w+)\s+TestCaseLabel\s+=\s+"([\w-]+)"`)
	for scanner.Scan() {
		line := scanner.Text()
		if matches := constPattern.FindStringSubmatch(line); matches != nil {
			labelNames[matches[2]] = matches[1]
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return labelNames, nil
}

// Function to write updated map back to the file using variable names
func writeUpdatedMapToFile(filename, labelFilePath string) error {
	var result strings.Builder

	// Map to store variable names and their corresponding values
	labelNames, err := createLabelNamesMap(labelFilePath)
	if err != nil {
		return err
	}

	result.WriteString("var TestCaseLabelsMap = map[TestCaseName][]TestCaseLabel{\n")

	// Retrieve and sort the keys
	var keys []string
	for testCaseName := range tests.TestCaseLabelsMap {
		keys = append(keys, testCaseName)
	}
	sort.Strings(keys)

	// Iterate over the sorted keys
	for _, testCaseName := range keys {
		result.WriteString(fmt.Sprintf("\t%s: {", testCaseName))
		labels := tests.TestCaseLabelsMap[testCaseName]
		for i, label := range labels {
			if i > 0 {
				result.WriteString(", ")
			}
			// Get the variable name from the map
			labelName := labelNames[label]
			result.WriteString(labelName)
		}
		result.WriteString("},\n")
	}
	result.WriteString("}\n")

	return ioutil.WriteFile(filename, []byte(result.String()), 0644)
}

func main() {
	if len(os.Args) < 3 {
		log.Infof("Usage: go run backup_test_label_updater.go <input.csv> <output.go> [label_file.go]")
		return
	}

	inputCSV := os.Args[1]
	outputFile := os.Args[2]
	// Default value
	labelFilePath := "../tests/backup/backup_test_labels.go"

	if len(os.Args) >= 4 {
		labelFilePath = os.Args[3]
	}

	err := updateMapFromCSV(inputCSV)
	if err != nil {
		log.Infof("Error updating map from CSV: %v\n", err)
		return
	}

	err = writeUpdatedMapToFile(outputFile, labelFilePath)
	if err != nil {
		log.Infof("Error writing updated map to file: %v\n", err)
		return
	}

	log.Infof("Test case labels map successfully updated.")
}
