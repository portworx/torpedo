package main

import (
	"sync"
	"time"

	. "github.com/portworx/torpedo/tests/backup/longevity/pxbackuplongevitytriggers"
)

func main() {
	var wg sync.WaitGroup
	var startTime = time.Now()

	for {
		startTime = TriggerLongevityWorkflows(startTime, &wg)
		time.Sleep(2 * time.Second)
	}
}
