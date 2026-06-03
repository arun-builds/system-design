package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	var (
		url      = flag.String("url", "http://localhost:8080/api/products", "endpoint URL")
		workers  = flag.Int("c", 50, "concurrent workers")
		duration = flag.Duration("d", 30*time.Second, "test duration")
	)
	flag.Parse()

	log.Printf("load test: url=%s workers=%d duration=%s", *url, *workers, *duration)

	var success, failure, bytes atomic.Int64
	deadline := time.Now().Add(*duration)
	var wg sync.WaitGroup

	for i := 0; i < *workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			client := &http.Client{Timeout: 10 * time.Second}
			for time.Now().Before(deadline) {
				resp, err := client.Get(*url)
				if err != nil {
					failure.Add(1)
					continue
				}
				n, _ := io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
				bytes.Add(n)
				if resp.StatusCode == http.StatusOK {
					success.Add(1)
				} else {
					failure.Add(1)
				}
			}
		}()
	}

	wg.Wait()

	total := success.Load() + failure.Load()
	fmt.Printf("\n=== Results ===\n")
	fmt.Printf("URL:      %s\n", *url)
	fmt.Printf("Workers:  %d\n", *workers)
	fmt.Printf("Duration: %s\n", *duration)
	fmt.Printf("Requests: %d\n", total)
	fmt.Printf("Success:  %d\n", success.Load())
	fmt.Printf("Failure:  %d\n", failure.Load())
	fmt.Printf("RPS:      %.2f\n", float64(total)/duration.Seconds())
	fmt.Printf("Bytes:    %d (%.2f MB)\n", bytes.Load(), float64(bytes.Load())/1024/1024)
}
