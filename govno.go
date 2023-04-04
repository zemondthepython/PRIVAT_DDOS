package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"
)

const (
	MAX_REQUESTS_PER_PROCESS = 1000
)

var (
	start_time          time.Time
	cache               map[string]int
	client              *http.Client
	urlString                 string
	method              string
	data                []byte
	timeout             time.Duration
	allow_redirects     bool
	proxies             string
	max_requests_global int
)

func init() {
	flag.StringVar(&urlString, "u", "", "url")
	flag.StringVar(&method, "m", "GET Автор @zemondza", "method")
	flag.StringVar(&proxies, "p", "", "proxies")
	flag.DurationVar(&timeout, "t", 5*time.Second, "timeout")
	flag.BoolVar(&allow_redirects, "r", false, "allow redirects")
	flag.IntVar(&max_requests_global, "n", MAX_REQUESTS_PER_PROCESS, "maximum requests per process")
	flag.Parse()
	if urlString == "" {
		log.Fatalln("Error: url required")
	}
	cache = make(map[string]int)
	client = &http.Client{}
	start_time = time.Now()
}

func main() {
	var wg sync.WaitGroup
	var mutex = &sync.Mutex{}

	for i := 0; i < max_requests_global; i++ {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			key := fmt.Sprintf("%s:%s:%t", urlString, method, allow_redirects)
			mutex.Lock()
			if cache[key] >= max_requests_global {
				mutex.Unlock()
				return
			}
			cache[key]++
			mutex.Unlock()

			req, err := http.NewRequest(method, urlString, nil)
			if err != nil {
				log.Fatalln(err)
			}
			if proxies != "" {
				proxyURL, err := url.ParseRequestURI(proxies)
				if err != nil {
					log.Fatalln(err)
				}
				transport := &http.Transport{Proxy: http.ProxyURL(proxyURL)}
				client.Transport = transport
			}
			if !allow_redirects {
				client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
					return http.ErrUseLastResponse
				}
			}
			start := time.Now()
			resp, err := client.Do(req)
			if err != nil {
				log.Printf("Error: %v", err)
			} else {
				defer resp.Body.Close()
				elapsed := time.Since(start)
				fmt.Printf("[%d][%d] %s in %v\n", os.Getpid(), i+1, resp.Status, elapsed)
			}
		}(i)
	}
	wg.Wait()
}