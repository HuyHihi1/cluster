# Welcome to your Go Observability Playground!

I have fully built and deployed the backend API, instrumented with modern tracing, metrics, logging, and Go profiling benchmarks. 

Below is a detailed guide on how to test each learning outcome. You can now use this playground to understand deeply how these tools work.

> [!TIP]
> The entire stack runs via Docker Compose targeting the `docker-compose.yml`. The Go application logic lives in `handlers/sandbox.go` and `handlers/sandbox_test.go`.

---

## 1. Structured Logging (slog)

Standard library `log/slog` is used throughout the app.
When you interact with the API, check the docker logs to see structured JSON logs containing context like tracing info:

```bash
cd ~/Documents/study/go-observability-playground
docker compose logs -f web-app
```

Hit an endpoint like `curl http://localhost:8080/api/v1/compute` and watch the structured logs roll in.

---

## 2. Distributed Tracing (OpenTelemetry & Tempo)

Every request to `/api/v1/*` is automatically traced via OTel and pushed to Grafana Tempo.
The `MockDBHandler` explicitly demonstrates creating sub-spans to track internal logic steps (like DB queries).

**How to verify:**
1. Send a request to the simulated database mock route:
   ```bash
   curl http://localhost:8080/api/v1/mock-db
   ```
2. Open Grafana at [http://localhost:3000/explore](http://localhost:3000/explore) (No login needed initially, or use `admin`/`admin` if prompted).
3. Select the **Tempo** Data Source from the dropdown on the top left.
4. Go to **Search**, select `go-observability-playground` under the **Service Name** filter, and click **Run Query**.
5. Click on the Trace ID to view a beautiful waterfall diagram showing exactly where the database latency comes from!

---

## 3. Custom Metrics (Prometheus)

The app intercepts HTTP requests and tallies them up using custom metrics like `http_requests_total` and specific business metrics like `compute_tasks_total`.

**How to verify:**
1. Hit some endpoints a few times:
   ```bash
   curl http://localhost:8080/api/v1/compute
   curl http://localhost:8080/api/v1/memory
   curl http://localhost:8080/api/v1/mock-db
   ```
2. Navigate to [http://localhost:8080/metrics](http://localhost:8080/metrics) in your browser. You will see the raw Prometheus exposition format. Search the text for `http_requests_total`.
3. Open Grafana [http://localhost:3000/explore](http://localhost:3000/explore), choose the **Prometheus** Data Source.
4. Type `http_requests_total` into the query box and click **Run Query** to see a time-series graph of your API traffic.

---

## 4. Benchmarking

Go comes built-in with powerful benchmarking tools. I've placed two benchmark functions inside `handlers/sandbox_test.go` to measure the unoptimized CPU calculations.

**How to verify:**
Run the benchmark directly on your system (make sure you are in the application directory):
```bash
cd ~/Documents/study/go-observability-playground
go test -bench . -benchmem -v ./handlers/
```
You will get output showing exactly how many ns/op (nanoseconds per operation) and B/op (Bytes allocated per operation) these algorithms consume.

---

## 5. Profiling (pprof)

`pprof` allows you to tap into a live running Go binary and extract CPU and Heap profiles directly. Our server exposes the standard `net/http/pprof` endpoints.

> [!IMPORTANT]
> Since the application runs very fast, we need to generate artificial load to capture a meaningful CPU profile.

**How to verify CPU Profiling:**
1. In one terminal tab, blast the compute endpoint using a quick loop (or `hey`/`ab` if you have them):
   ```bash
   while true; do curl http://localhost:8080/api/v1/compute; done
   ```
2. In another terminal tab, download and view a 15-second CPU profile using the built in `go tool pprof`:
   ```bash
   go tool pprof -http=":8081" http://localhost:8080/debug/pprof/profile?seconds=15
   ```
3. After 15 seconds, a browser window will open on port 8081 showing a Flamegraph and Call Graph of where the CPU spent its time. (You'll clearly see `big.Int.Mul` eating up all the cycles!).

**How to verify Heap (Memory) Profiling:**
1. Hit the memory allocation endpoint once or twice:
   ```bash
   curl http://localhost:8080/api/v1/memory
   ```
2. Run `pprof` against the heap endpoint:
   ```bash
   go tool pprof -http=":8081" http://localhost:8080/debug/pprof/heap
   ```
3. You will visually see the 100MB allocations represented directly in the execution tree.

---

## 6. Multi-Server Asynchronous Distributed Tracing (NATS & Redis)

Distributed tracing becomes especially powerful when tracking requests that move between different systems via message brokers. This playground demonstrates a 2-server architecture:
- **Server A**: Receives an HTTP request and publishes a message to NATS.
- **Server B**: Consumes the message from NATS and saves the data to Redis.

**How to verify:**
1. Trigger the distributed task manually:
   ```bash
   curl http://localhost:8080/api/v1/trigger-async
   ```
   *Note: The `load-generator` also triggers this endpoint every 3 seconds.*

2. Open Grafana [http://localhost:3000/explore](http://localhost:3000/explore) and select the **Tempo** Data Source.
3. Search for traces. You will see a trace that consists of:
   - **`TriggerAsyncHandler` (Server A)**: The HTTP request that published the message.
   - **`ProcessTask` (Server B)**: A child span (kind: consumer) that represents the background worker pulling the message from NATS.
   - **`set` (Server B -> Redis)**: A child span of `ProcessTask` representing the Redis `SET` operation, automatically instrumented via `redisotel`.

4. Observe how the **Trace ID** remains identical across both servers and all three operations (HTTP -> NATS -> Redis), even though they execute in different processes!

---

## 7. Multi-Server Monitoring (Prometheus)

Since we now have two separate servers, Prometheus is configured to scrape metrics from both.

**How to verify:**
1. Open Grafana [http://localhost:3000/explore](http://localhost:3000/explore), choose the **Prometheus** Data Source.
2. Query `http_requests_total`.
3. You can now filter by `job="server-a"` or `job="server-b"` to see the traffic for each specific service.
4. Try `up` to see if both servers are healthy and being scraped.
