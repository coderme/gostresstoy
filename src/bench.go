package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

type result struct {
	success bool
	timeIt  time.Duration
	size    int64
}

type stats struct {
	Requests         int64         `json:"requests"`
	Errors           int64         `json:"errors,omitempty"`
	Bytes            int64         `json:"received_bytes,omitempty"`
	RequestPerSecond float64       `json:"requests/sec"`
	BytesPerSecond   float64       `json:"Bytes/sec"`
	AvgLatency       time.Duration `json:"avg_latency"`
	Latencies        time.Duration `json:"latencies"`
	Slowest          time.Duration `json:"slowest"`
	Fastest          time.Duration `json:"fastest"`
}

const (
	maxPool            = 1 << 15
	estimatedRPS       = 1 << 12
	defaultCPU         = 1
	defaultConcurrency = 10
)

var (
	numVCPU = runtime.NumCPU()
	stop    = make(chan struct{}, 1)

	timeout     <-chan time.Time
	netStats    = &stats{}
	run         chan struct{}
	resultsPipe chan *result

	// we like to reuse client
	client = &http.Client{}
	pool   chan *http.Request
	url    string
)

func init() {
	if os.Getuid() == 0 {
		fmt.Println("Bad idea, running as 'root'")
		os.Exit(2)
	}

	flag.CommandLine.Usage = usage
	flag.Parse()

	url = flag.Arg(0)

	if *showVersion {
		printVersion()
	}

	if *showLicense {
		printLicense()
	}

	if *help || *showhelp || url == "" {
		usage()
	}

	if *vcpu < 1 || *vcpu > numVCPU {
		*vcpu = defaultCPU
	}
	if *concurency < 1 {
		*concurency = defaultConcurrency
	}
	if *numRequests < 1 {
		*numRequests = 0
	}
	resultsPipe = make(chan *result, 12000)

}

func listenSignal() {
	go func() {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals)
		for {
			switch s := <-signals; s {
			case syscall.SIGKILL, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM:
				//showStat()
				stop <- struct{}{}
			}
		}

	}()
}

func main() {
	runtime.GOMAXPROCS(*vcpu)
	listenSignal()
	defer showStat()

	run = make(chan struct{}, *concurency)

	// ready-made requests
	err := initRequests()
	if err != nil {
		fmt.Println("Error making requests,", err)
		os.Exit(1)
	}

	if *numRequests > 0 {
		stressWithNumber()

	} else {
		stressWithTime()
	}

	for {
		select {
		case <-stop:
			return
		case r := <-resultsPipe:
			aggr(r)
		//nothing to do; then refill the pool
		case pool <- newRequest():
		}
	}
}

func stressWithNumber() {
	finished := make(chan struct{}, *numRequests)
	go func() {
		for c := 1; c <= *numRequests; c++ {
			stressIt(url, resultsPipe, finished)
		}
	}()
	// wait for all requests to finish
	go func() {
		ticker := time.Tick(700 * time.Millisecond)
		for {
			select {
			case <-ticker:
				if len(finished) == *numRequests {
					stop <- struct{}{}
					return
				}
			}
		}
	}()
}

func stressWithTime() {
	timeout = time.After(*duration)
	go func() {
		for {
			select {
			case <-timeout:
				stop <- struct{}{}
				return
			default:
				stressIt(url, resultsPipe, nil)
			}

		}
	}()
}

func showStat() {
	defer os.Exit(0)

	drainResultPipe()

	var rps float64
	req := float64(netStats.Requests)
	latency := float64(netStats.Latencies) / req
	if latency > 0 {
		rps = 1.0 / (latency / 1e9)
		rps *= float64(*concurency)
	} else {
		rps = req
	}
	size := float64(netStats.Bytes) / req
	netStats.BytesPerSecond = size * rps

	netStats.AvgLatency = time.Duration(latency)

	netStats.RequestPerSecond = rps
	if *dumpJSON {
		encoded, err := json.MarshalIndent(netStats, "", "\t")
		if err != nil {
			fmt.Println("Error encoding JSON", err)
		} else {
			fmt.Printf("%s", encoded)
			return
		}
	}

	stat := fmt.Sprintf("\nRequests: %d\n", netStats.Requests)
	if netStats.Errors > 0 {
		stat += fmt.Sprintf("Errors: %d\n", netStats.Errors)
	}

	fmt.Printf(stat+"BytesReceived: %s\nRequests/sec: %.2f\nTransfer/sec: %s\nAvgLatency: %s\nLatencies: %s\nFastest: %s\nSlowest: %s\n\n",
		formatBytes(netStats.Bytes),
		rps,
		formatBytes(int64(netStats.BytesPerSecond)),
		netStats.AvgLatency,
		netStats.Latencies,
		netStats.Fastest,
		netStats.Slowest,
	)

}

func aggr(r *result) {
	if netStats.Slowest == 0 || r.timeIt > netStats.Slowest {
		netStats.Slowest = r.timeIt
	}

	if netStats.Fastest == 0 || r.timeIt < netStats.Fastest {
		netStats.Fastest = r.timeIt
	}
	netStats.Bytes += r.size
	if !r.success {
		netStats.Errors++
	}
	netStats.Requests++
	netStats.Latencies += r.timeIt

}

func drainResultPipe() {
	close(resultsPipe)
	for {
		r, ok := <-resultsPipe
		if !ok {
			return
		}
		aggr(r)
	}
}

func formatBytes(b int64) string {
	const (
		kb = 1024
		mb = kb * kb
		gb = mb * kb
		tb = gb * kb
		pb = tb * kb
		eb = pb * kb
	)

	type unit struct {
		length int64
		suffix string
	}

	units := []unit{
		unit{eb, "EB"},
		unit{pb, "PB"},
		unit{tb, "TB"},
		unit{gb, "GB"},
		unit{mb, "MB"},
		unit{kb, "KB"},
	}
	for _, u := range units {
		if b >= u.length {
			x := float64(b) / float64(u.length)
			return fmt.Sprintf("%.2f %s", x, u.suffix)
		}
	}
	return fmt.Sprintf("%d Bytes", b)
}

func initRequests() error {
	poolSize := 0

	if *numRequests > 0 {
		poolSize = *numRequests

	} else {
		estimated := (*duration).Seconds() * float64(estimatedRPS)
		poolSize = int(estimated)

	}
	if poolSize > *maxPoolSize {
		poolSize = *maxPoolSize
	}

	pool = make(chan *http.Request, poolSize)
	for c := 0; c < poolSize; c++ {
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return err
		}
		pool <- req
	}

	return nil
}

func newRequest() *http.Request {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		fmt.Println("Error newRequest()", err)
		return nil
	}
	return req
}

func stressIt(url string, r chan *result, d chan struct{}) {
	run <- struct{}{}

	go func() {
		t := &result{}
		defer func() {
			<-run
			r <- t
			if d != nil {
				d <- struct{}{}
			}
		}()

		timeit := time.Now()
		resp, err := client.Do(<-pool)
		t.timeIt = time.Since(timeit)

		if err == nil {
			defer resp.Body.Close()
		}

		if err != nil || resp == nil {
			t.success = false
			return
		}

		t.success = true

		if !*countBytes {
			return
		}

		nbytes, _ := io.Copy(ioutil.Discard, resp.Body)
		t.size = nbytes
	}()
}
