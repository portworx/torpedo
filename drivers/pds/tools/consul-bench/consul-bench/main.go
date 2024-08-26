package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"

	consul "github.com/hashicorp/consul/api"
)

func main() {
	consulAddr := flag.String("consul", "127.0.0.1:8500", "Consul address")
	useRPC := flag.Bool("rpc", false, "Use RPC server calls instead of agent HTTP")
	rpcAddr := flag.String("rpc-addr", "127.0.0.1:8300", "When using rpc, the consul rpc addr")
	dc := flag.String("dc", "dc1", "When using rpc, the consul datacenter")
	serviceName := flag.String("service", "srv", "Service to watch")
	serviceTags := flag.String("tags", "", "Comma seperated list of tags to add to registered services")
	registerInstances := flag.Int("register", 0, "Register N -service instances")
	deregister := flag.Bool("deregister", false, "Deregister all instances of -service")
	flapInterval := flag.Duration("flap-interval", 0, "If -register is given, flap each instance between critical and passing state on given interval")
	wait := flag.Duration("query-wait", 10*time.Minute, "Bloquing queries max wait time")
	stale := flag.Bool("query-stale", false, "Run stale blocking queries")
	token := flag.String("token", "", "ACL token")
	watchers := flag.Int("watchers", 1, "Number of concurrnet watchers on service")
	monitor := flag.Int("monitor", 0, "Consul PID")
	runtime := flag.Duration("time", 0, "Time to run the benchmark")
	latepc := flag.Float64("late-ratio", 0, "Ratio of late callers")
	flag.Parse()

	if *token == "" {
		*token = os.Getenv("ACL_TOKEN")
	}

	c, err := consul.NewClient(&consul.Config{
		Address: *consulAddr,
		Token:   *token,
	})
	if err != nil {
		log.Fatal(err)
	}

	stats := make(chan Stat)

	if *registerInstances > 0 {
		err := RegisterServices(c, *serviceName, *registerInstances, *flapInterval, *serviceTags, stats)
		if err != nil {
			log.Fatal(err)
		}
	} else if *deregister {
		err := DeregisterServices(c, *serviceName)
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	done := make(chan struct{})
	var wg sync.WaitGroup

	if *monitor > 0 {
		wg.Add(1)
		go func() {
			Monitor(int32(*monitor), stats, done)
			wg.Done()
		}()
	}

	if *runtime > 0 {
		go func() {
			time.Sleep(*runtime)
			close(done)
		}()
	}

	var qf queryFn
	if !*useRPC {
		qf = QueryAgent(c, *serviceName, *wait, *stale)
	} else {
		qf = QueryServer(*rpcAddr, *dc, *serviceName, *wait, *stale)
	}

	wg.Add(1)
	go func() {
		RunQueries(qf, *watchers, *latepc, stats, done)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		DisplayStats(stats, done)
		wg.Done()
	}()

	go func() {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Interrupt)
		<-signals
		close(done)
	}()

	<-done
	wg.Wait()
	os.Exit(0)
}
