package main

import (
	"bench/utils"
	"flag"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	addr := flag.String("addr", "localhost:8080", "server address")
	total := flag.Int64("n", 500_000, "total requests")
	conc := flag.Int("c", 100, "concurrent connections")
	flag.Parse()

	frame, err := utils.PackFrame([]byte("Payload"))
	if err != nil {
		panic(err)
	}

	var remaining = *total
	var success, failed int64
	var wg sync.WaitGroup

	start := time.Now()

	for i := 0; i < *conc; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			conn, err := net.DialTimeout("tcp", *addr, 3*time.Second)
			if err != nil {
				fmt.Println("dial failed:", err)
				return
			}
			defer func() { conn.Close() }()

			for atomic.AddInt64(&remaining, -1) >= 0 {
				conn.SetDeadline(time.Now().Add(5 * time.Second))

				if _, err := conn.Write(frame); err != nil {
					atomic.AddInt64(&failed, 1)
					fmt.Println("write error:", err)
					return
				}
				if _, err := utils.ReadFrame(conn); err != nil {
					atomic.AddInt64(&failed, 1)
					fmt.Println("read error:", err)
					return
				}
				atomic.AddInt64(&success, 1)
			}
		}()
	}

	wg.Wait()
	d := time.Since(start)

	fmt.Printf("%d ok, %d failed in %v\n", success, failed, d)
	fmt.Printf("Throughput: %.0f req/sec\n", float64(success)/d.Seconds())
}
