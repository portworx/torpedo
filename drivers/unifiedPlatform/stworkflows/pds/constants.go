package pds

import "time"

// Commands to start, pause or stop data
const (
	DataStop = "Stop"
)

const (
	replicaHealthTimeout  = 60 * time.Minute
	replicaHealthInterval = 10 * time.Second
)
