package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"golang.org/x/net/http2"
)

func main() {
	var (
		URL     string
		HTTP2   bool
		VERBOSE bool
	)
	flag.BoolVar(&HTTP2, "http2", false, "是否使用HTTP2")
	flag.BoolVar(&VERBOSE, "verbose", false, "详细信息")
	flag.Parse()
	URL = flag.Arg(0)

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	if HTTP2 {
		http2.ConfigureTransport(transport)
	}

	client := http.Client{Transport: transport}

	var (
		min time.Duration = time.Hour
		max time.Duration
		sum time.Duration

		wg sync.WaitGroup
	)

	const I, N = 2, 1000
	for i := 0; i < I; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for n := 0; n < N; n++ {
				t := time.Now()
				r, err := client.Post(
					fmt.Sprintf("%s?t0=%d", URL, time.Now().UnixNano()/1e3),
					"text/plain", nil)
				if err != nil {
					fmt.Println(err)
					return
				}
				b, err := ioutil.ReadAll(r.Body)
				if err != nil {
					fmt.Println(err)
					return
				}
				fmt.Printf("%s\n", b)
				d := time.Since(t)
				sum += d
				if d < min {
					min = d
				}
				if d > max {
					max = d
				}
			}
		}()
	}
	wg.Wait()
	fmt.Println("min:", min, "max", max, "avg:", sum/(I*N))
}
