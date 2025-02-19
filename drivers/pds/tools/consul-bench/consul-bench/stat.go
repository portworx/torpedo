package main

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"time"
)

type Stat struct {
	Label string
	Value float64
}

func DisplayStats(ch <-chan Stat, done chan struct{}) {
	type statC struct {
		Stat
		Count int
	}

	var l sync.Mutex
	avg := make(map[string]*statC)
	entries := make(map[string]Stat)

	go func() {
		for {
			e := <-ch
			l.Lock()
			entries[e.Label] = e
			if _, ok := avg[e.Label]; !ok {
				avg[e.Label] = &statC{
					Count: 0,
					Stat:  e,
				}
			}
			avg[e.Label].Value = (avg[e.Label].Value*float64(avg[e.Label].Count) + e.Value) / float64(avg[e.Label].Count+1)
			avg[e.Label].Count++
			l.Unlock()
		}
	}()

	printLine := func(entries []Stat) {
		sort.Slice(entries, func(i, j int) bool {
			return strings.Compare(entries[i].Label, entries[j].Label) < 0
		})
		var s []string
		for _, e := range entries {
			s = append(s, fmt.Sprintf("%s: %.2f", e.Label, e.Value))
		}

		log.Println(strings.Join(s, ", "))
	}

	start := time.Now()
	tick := time.Tick(time.Second)

	for {
		select {
		case <-done:
			l.Lock()
			entriesSlice := []Stat{{
				Label: "Runtime (s)",
				Value: time.Since(start).Seconds(),
			}}
			for _, e := range avg {
				entriesSlice = append(entriesSlice, e.Stat)
			}

			log.Println("====== Summary ======")
			printLine(entriesSlice)
			l.Unlock()
			return
		case <-tick:
			l.Lock()
			var entriesSlice []Stat
			for _, e := range entries {
				entriesSlice = append(entriesSlice, e)
			}

			printLine(entriesSlice)

			entriesSlice = entriesSlice[:0]
			l.Unlock()
		}
	}
}
