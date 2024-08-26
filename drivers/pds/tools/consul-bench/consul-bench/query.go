package main

import (
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/hashicorp/consul/agent/pool"
	"github.com/hashicorp/consul/agent/structs"
	consul "github.com/hashicorp/consul/api"
)

type queryFn func(uint64) (uint64, error)

func RunQueries(fn queryFn, count int, lateRatio float64, stats chan Stat, done chan struct{}) error {
	log.Println("Starting", count, "watchers...")

	var qps int32
	var lateQps int32

	errs := make(chan error, 1)
	for i := 0; i < count; i++ {
		go func() {
			index := uint64(1)
			var err error
			var i int
			for {
				if rand.Float64() <= lateRatio {
					atomic.AddInt32(&lateQps, 1)
					index -= 50
					if index <= 0 {
						index = 1
					}
				}
				index, err = fn(index)
				if err != nil {
					select {
					case errs <- err:
					default:
					}
					return
				}
				atomic.AddInt32(&qps, 1)
				i++
			}
		}()
	}
	go func() {
		for range time.Tick(time.Second) {
			c := atomic.SwapInt32(&qps, 0)
			lc := atomic.SwapInt32(&lateQps, 0)
			stats <- Stat{"QPS", float64(c)}
			stats <- Stat{"Late QPS", float64(lc)}
		}
	}()
	log.Println("Watchers started.")

	<-done
	select {
	case err := <-errs:
		return err
	default:
	}
	return nil
}

func QueryAgent(client *consul.Client, serviceName string, wait time.Duration, allowStale bool) queryFn {
	return func(index uint64) (uint64, error) {
		_, meta, err := client.Health().Service(serviceName, "", false, &consul.QueryOptions{
			WaitTime:   wait,
			WaitIndex:  index,
			AllowStale: allowStale,
		})
		if err != nil {
			return 0, err
		}

		return meta.LastIndex, nil
	}
}

func QueryServer(addr string, dc string, serviceName string, wait time.Duration, allowStale bool) queryFn {
	connPool := &pool.ConnPool{
		SrcAddr:    nil,
		LogOutput:  os.Stderr,
		MaxTime:    time.Hour,
		MaxStreams: 1000000,
		ForceTLS:   false,
	}

	args := structs.ServiceSpecificRequest{
		Datacenter:  dc,
		ServiceName: serviceName,
		Source: structs.QuerySource{
			Datacenter: dc,
			Node:       "test-1",
		},
		QueryOptions: structs.QueryOptions{
			MaxQueryTime: 10 * time.Minute,
		},
	}

	ip, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		log.Fatal(err)
	}
	port, _ := strconv.Atoi(portStr)
	if port == 0 {
		port = 8300
	}
	srvAddr := &net.TCPAddr{net.ParseIP(ip), port, ""}

	return func(index uint64) (uint64, error) {
		args.QueryOptions.MinQueryIndex = index
		var resp *structs.IndexedCheckServiceNodes
		err := connPool.RPC(dc, srvAddr, 3, "Health.ServiceNodes", false, &args, &resp)
		if err != nil {
			return 0, err
		}
		return resp.QueryMeta.Index, nil
	}
}
