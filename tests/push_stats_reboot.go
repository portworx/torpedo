package main

import (
	"fmt"
	"time"

	"github.com/portworx/torpedo/pkg/stats"
)

func main() {
	rebootStats := &stats.NodeRebootStatsType{
		RebootTime: time.Now().Format("2006-01-02 15:04:05"),
		Node:       "test",
		PxVersion:  "3.0.4",
	}

	err := stats.PushStats(rebootStats)
	if err != nil {
		fmt.Printf("failed to create exportable stats")
	}
}
