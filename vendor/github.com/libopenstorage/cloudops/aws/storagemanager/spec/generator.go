package main

import (
	"fmt"

	"github.com/libopenstorage/cloudops"
	"github.com/libopenstorage/cloudops/aws/storagemanager"
	"github.com/libopenstorage/cloudops/pkg/parser"
)

const (
	awsYamlPath = "aws.yaml"
)

func main() {
	matrixRows := append(
		getGp2StorageDecisionMatrixRows(),
		getIo1StorageDecisionMatrixRows()...,
	)
	matrix := cloudops.StorageDecisionMatrix{Rows: matrixRows}
	if err := parser.NewStorageDecisionMatrixParser().MarshalToYaml(&matrix, awsYamlPath); err != nil {
		fmt.Println("Failed to generate aws storage decision matrix yaml: ", err)
		return
	}
	fmt.Println("Generated aws storage decision matrix yaml at ", awsYamlPath)
}

// getGp2StorageDecisionMatrixRows will programmatically generate rows for gp2 drive type
// for the storage decision matrix
func getGp2StorageDecisionMatrixRows() []cloudops.StorageDecisionMatrixRow {
	rows := []cloudops.StorageDecisionMatrixRow{}
	// First row has min and max 100 IOPS for 0 - 33Gi
	row := getCommonRow(0)
	row.DriveType = storagemanager.DriveTypeGp2
	row.MinIOPS = 100
	row.MaxIOPS = 100
	row.MinSize = 0
	row.MaxSize = 33
	rows = append(rows, row)
	for iops := 100; iops <= 16000; iops = iops + 50 {
		row := getCommonRow(0)
		row.DriveType = storagemanager.DriveTypeGp2
		row.MinIOPS = uint64(iops)
		row.MaxIOPS = uint64(iops + 50)
		row.MinSize = row.MinIOPS / uint64(storagemanager.Gp2IopsMultiplier)
		row.MaxSize = row.MaxIOPS / uint64(storagemanager.Gp2IopsMultiplier)
		rows = append(rows, row)
	}
	// Last row has min and max 16000 IOPS from 5333Gi - 16TiB
	row = getCommonRow(0)
	row.DriveType = storagemanager.DriveTypeGp2
	row.MinIOPS = 16000
	row.MaxIOPS = 16000
	row.MinSize = 5334
	row.MaxSize = 16000
	rows = append(rows, row)
	return rows
}

// getIo1StorageDecisionMatrixRows will programmatically generate rows for io1 drive type
// for the storage decision matrix.
// For io1 volumes,
// the MinIOPS is chosen from the MaxSize
// According to AWS docs, IOPS-GiB ratio should be greater than 2:1
// We are keeping the ratio to 4:1. (gp2 already provides 3:1 ratio)
// the MaxIOPS is chosen from MinSize
// According to AWS docs, the provisioned IOPS can go upto 50 times the
// size of the volume in GiB.
func getIo1StorageDecisionMatrixRows() []cloudops.StorageDecisionMatrixRow {
	rows := []cloudops.StorageDecisionMatrixRow{}
	for startSize := 50; startSize < 6400; startSize = startSize * 2 {
		row := getCommonRow(1)
		row.DriveType = storagemanager.DriveTypeIo1
		row.MinSize = uint64(startSize)
		row.MaxSize = uint64(startSize * 2)
		// Keeping the ratio of IOPS to GiB as 4:1
		minIops := row.MaxSize * 4
		if minIops >= 32000 {
			// AWS caps IOPS at 32000
			minIops = 32000
		}
		row.MinIOPS = minIops
		// AWS provides the max ratio of IOPS to GiB as 50:1
		maxIops := row.MinSize * 50
		if maxIops >= 32000 {
			maxIops = 32000
		}
		row.MaxIOPS = maxIops
		rows = append(rows, row)
	}
	// Last row has minSize as 6400 and maxSize as 16TiB
	row := getCommonRow(1)
	row.MinIOPS = 32000
	row.MaxIOPS = 32000
	row.MinSize = 6400
	row.MaxSize = 16000
	row.DriveType = storagemanager.DriveTypeIo1
	rows = append(rows, row)
	return rows
}

func getCommonRow(priority int) cloudops.StorageDecisionMatrixRow {
	return cloudops.StorageDecisionMatrixRow{
		InstanceType:      "*",
		InstanceMaxDrives: 8,
		InstanceMinDrives: 1,
		Region:            "*",
		Priority:          priority,
	}
}
