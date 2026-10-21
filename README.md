# TCP Echo Benchmark: Go `net` vs gnet

Compares two TCP echo servers under identical load, running in Kubernetes (Docker Desktop) with Prometheus and Grafana.

- **go-std**: standard library `net`, one goroutine per connection
- **go-gnet**: [gnet v2](https://github.com/panjf2000/gnet), event-loop based

Both use the same framing (4-byte big-endian length + payload), the same 4 KB frame limit, and the same resource limits (2 CPU, 2Gi).

![Alt text for the image](/images/example.png)

## Structure

```
go-std/        std server + Dockerfile
go-gnet/       gnet server + Dockerfile
client/        load generator + Dockerfile
k8s/           servers, Prometheus, Grafana
jobs/          client Jobs (one per server)
dashboards/    Grafana dashboard JSON (auto-loaded)
Makefile
```

## Requirements

- Docker Desktop with Kubernetes enabled (kubeadm), about 8 CPUs / 16 GB for the VM
- `make`, `kubectl`

## Run

```
make build          # build server images
make build-client   # build client image
make deploy         # deploy servers, Prometheus, Grafana
make forward        # Grafana on http://localhost:3000 (admin / admin)
make bench          # start both client Jobs at the same time
make results        # print req/sec of each client after they finish
```

Other targets: `make down`, `make restart`, `make restart-grafana`.

Change the load in `jobs/client.yml` (`-c` connections, `-n` total requests).

## Metrics

Grafana shows both servers on the same charts: CPU, resident memory, goroutines, threads, active connections, total processed bytes.