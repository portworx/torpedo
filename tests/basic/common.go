package tests

import (
	"time"
)

const (
	replicationUpdateTimeout         = 4 * time.Hour
	retryTimeout                     = time.Minute * 2
	addDriveUpTimeOut                = time.Minute * 15
	poolResizeTimeout                = time.Minute * 120
	poolExpansionStatusCheckInterval = time.Minute * 3
	JournalDeviceSizeInGB            = 3
)

var isJournalEnabled bool
var bufferSizeInGB uint64
var targetSizeInBytes uint64
var originalSizeInBytes uint64
var testDescription string
var testName string
