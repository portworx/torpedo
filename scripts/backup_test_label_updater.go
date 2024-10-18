package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"github.com/pure-px/torpedo/pkg/log"
	"github.com/pure-px/torpedo/tests/backup"
	"io/ioutil"
	"os"
	"regexp"
	"sort"
	"strings"
)

// updateTestCaseLabelsMap updates or deletes the TestCaseLabelsMap
func updateTestCaseLabelsMap(testCaseName tests.TestCaseName, newLabels []tests.TestCaseLabel, delete bool) {
	log.Infof("Updating test case labels map for test case: %s, labels: %v", testCaseName, newLabels)
	if labels, exists := tests.TestCaseLabelsMap[testCaseName]; exists {
		if delete {
			for _, newLabel := range newLabels {
				for i, label := range labels {
					if label == newLabel {
						// Remove the label
						tests.TestCaseLabelsMap[testCaseName] = append(labels[:i], labels[i+1:]...)
						log.Infof("Label %s removed from the test case", newLabel)
						break
					}
				}
			}
			log.Infof("Updated labels: %v", tests.TestCaseLabelsMap[testCaseName])
		} else {
			for _, newLabel := range newLabels {
				exists := false
				for _, label := range labels {
					if label == newLabel {
						exists = true
						break
					}
				}
				if !exists {
					// Append the new label
					tests.TestCaseLabelsMap[testCaseName] = append(tests.TestCaseLabelsMap[testCaseName], newLabel)
					log.Infof("Label %s added to the test case", newLabel)
				}
			}
			log.Infof("Updated labels: %v", tests.TestCaseLabelsMap[testCaseName])
		}
	} else {
		if delete {
			log.Infof("Test case not found, nothing to delete")
		} else {
			log.Infof("Test case not found, adding new test case with the labels")
			// Add the new test case with the labels
			tests.TestCaseLabelsMap[testCaseName] = newLabels
		}
	}
}

// updateMapFromCSV updates the global TestCaseLabelsMap based on the CSV file provided
func updateMapFromCSV(csvFile string, delete bool) error {
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
		if len(record) < 2 {
			return fmt.Errorf("invalid record: %v", record)
		}
		testCaseName := record[0]
		newLabels := record[1:]
		updateTestCaseLabelsMap(testCaseName, newLabels, delete)
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
	inputFile := flag.String("inputfile", "", "Path to the input CSV file")
	outputFile := flag.String("outputfile", "", "Path to the output Go file")
	labelFile := flag.String("labelfile", "../tests/backup/backup_test_labels.go", "Path to the label file")
	deleteFlag := flag.Bool("delete", false, "Delete the label from the test case")
	flag.Parse()

	if *inputFile == "" || *outputFile == "" {
		log.Infof("Usage: go run backup_test_label_updater.go --inputfile=/path/to/input.csv --outputfile=/path/to/output.go [--labelfile=/path/to/label_file.go] [--delete]")
		return
	}

	err := updateMapFromCSV(*inputFile, *deleteFlag)
	if err != nil {
		log.Infof("Error updating map from CSV: %v\n", err)
		return
	}

	err = writeUpdatedMapToFile(*outputFile, *labelFile)
	if err != nil {
		log.Infof("Error writing updated map to file: %v\n", err)
		return
	}

	log.Infof("Test case labels map successfully updated.")
}
