package main

import (
	"encoding/binary"
	"flag"
	"log"
	"net/http"
	"runtime"

	"github.com/panjf2000/gnet/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	packetsProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tcp_packets_processed_total",
		Help: "Total number of TCP packets handled",
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

type server struct {
	gnet.BuiltinEventEngine
}

func (s *server) OnOpen(c gnet.Conn) ([]byte, gnet.Action) {
	activeConnections.Inc()
	return nil, gnet.None
}

func (s *server) OnClose(c gnet.Conn, err error) gnet.Action {
	activeConnections.Dec()
	return gnet.None
}

func (s *server) OnTraffic(c gnet.Conn) gnet.Action {
	for {
		data, _ := c.Peek(-1)
		if len(data) < 4 {
			break
		}
		n := int(binary.BigEndian.Uint32(data))
		if n > 4096 {
			return gnet.Close
		}
		total := 4 + n
		if len(data) < total {
			break
		}
		c.Write(data[:total])
		c.Discard(total)
		packetsProcessed.Inc()
	}
	return gnet.None
}

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

	log.Println("gnet TCP Server Started on", *addr)
	log.Fatal(gnet.Run(&server{}, "tcp://"+*addr, gnet.WithNumEventLoop(runtime.GOMAXPROCS(0))))
}
