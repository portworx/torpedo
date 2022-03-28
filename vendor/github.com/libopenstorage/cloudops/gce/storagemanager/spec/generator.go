package main

import (
	"fmt"
	"github.com/libopenstorage/cloudops"
	"github.com/libopenstorage/cloudops/gce/storagemanager"
	"github.com/libopenstorage/cloudops/pkg/parser"
	"math"
)

const (
	gceYamlPath = "gce.yaml"
)

func main() {
	matrixRows := append(
		getStandardDecisionMatrixRows(),
		getSSDDecisionMatrixRows()...,
	)
	matrix := cloudops.StorageDecisionMatrix{Rows: matrixRows}
	if err := parser.NewStorageDecisionMatrixParser().MarshalToYaml(&matrix, gceYamlPath); err != nil {
		fmt.Println("Failed to generate aws storage decision matrix yaml: ", err)
		return
	}
	fmt.Println("Generated gce storage decision matrix yaml at ", gceYamlPath)

}

func getSSDDecisionMatrixRows() []cloudops.StorageDecisionMatrixRow {
	rows := []cloudops.StorageDecisionMatrixRow{}
	// First row has min and max 100 IOPS for 30Gi
	row := getCommonRow(1)
	row.DriveType = storagemanager.DriveTypeSSD
	// 15000 IOPS is max read IOPS for SSD persistent disks
	for iops := 100; iops <= 15000; iops = iops + 50 {
		row := getCommonRow(1)
		row.DriveType = storagemanager.DriveTypeSSD
		row.MinIOPS = uint64(iops)
		row.MaxIOPS = uint64(iops + 50)
		row.MinSize = uint64(math.Ceil(float64(iops) / storagemanager.SSDIopsMultiplier))
		row.MaxSize = uint64(math.Ceil(float64(iops+50) / storagemanager.SSDIopsMultiplier))
		rows = append(rows, row)
	}
	// Last row has min and max 7500 IOPS and size 500Gi
	row = getCommonRow(1)
	row.DriveType = storagemanager.DriveTypeSSD
	row.MinIOPS = 15000
	row.MaxIOPS = 15000
	row.MinSize = 500
	row.MaxSize = 500
	rows = append(rows, row)
	return rows
}

func getStandardDecisionMatrixRows() []cloudops.StorageDecisionMatrixRow {
	rows := []cloudops.StorageDecisionMatrixRow{}
	// First row has min and max 100 IOPS for 0 - 134Gi
	row := getCommonRow(0)
	row.DriveType = storagemanager.DriveTypeStandard
	row.MinIOPS = 100
	row.MaxIOPS = 100
	row.MinSize = 0
	row.MaxSize = 134
	rows = append(rows, row)
	// 7500 IOPS is max read IOPS for Zonal standard persistent disks
	for iops := 100; iops <= 7500; iops = iops + 50 {
		row := getCommonRow(0)
		row.DriveType = storagemanager.DriveTypeStandard
		row.MinIOPS = uint64(iops)
		row.MaxIOPS = uint64(iops + 50)
		row.MinSize = uint64(math.Ceil(float64(iops) / storagemanager.StandardIopsMultiplier))
		row.MaxSize = uint64(math.Ceil(float64(iops+50) / storagemanager.StandardIopsMultiplier))
		rows = append(rows, row)
	}
	// Last row has min and max 7500 IOPS and size 10TiB
	row = getCommonRow(0)
	row.DriveType = storagemanager.DriveTypeStandard
	row.MinIOPS = 7500
	row.MaxIOPS = 7500
	row.MinSize = 10000
	row.MaxSize = 10000
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
