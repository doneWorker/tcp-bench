package main

import (
	"encoding/binary"
	"flag"
	"io"
	"log"
	"net"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	packetsProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tcp_packets_processed_total",
		Help: "Total number of TCP connections handled",
	})
	bytesProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tcp_bytes_processed_total",
		Help: "Total raw bytes echoed over TCP",
	})
	activeConnections = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "tcp_active_connections",
		Help: "Current number of active TCP client connections",
	})
)

func main() {
	addr := flag.String("addr", ":8081", "listen address")
	metrics := flag.String("metrics", ":2113", "metrics address")
	flag.Parse()

	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		log.Println("Metric Server Started on", *metrics)
		if err := http.ListenAndServe(*metrics, mux); err != nil {
			log.Fatalf("unable to start server: %v", err)
		}
	}()

	listener, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	log.Println("TCP Server Started on", *addr)
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}

		activeConnections.Inc()

		go func(c net.Conn) {
			defer activeConnections.Dec()
			defer c.Close()

			buf := make([]byte, 4+4096)

			for {
				if _, err := io.ReadFull(c, buf[:4]); err != nil {
					if err != io.EOF {
						log.Println("Error in read frame:", err)
					}
					return
				}
				n := int(binary.BigEndian.Uint32(buf[:4]))
				if n > len(buf)-4 {
					log.Println("frame too large:", n)
					return
				}
				if _, err := io.ReadFull(c, buf[4:4+n]); err != nil {
					log.Println("Error in read frame:", err)
					return
				}

				if _, err := c.Write(buf[:4+n]); err != nil {
					log.Println("Error writing echo response:", err)
					return
				}
				packetsProcessed.Inc()
			}
		}(conn)
	}
}
