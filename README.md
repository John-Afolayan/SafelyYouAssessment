# Fleet Monitoring Service (SafelyYouAssessment)

A Go HTTP service that ingests device telemetry (heartbeats and upload stats) and computes per-device uptime and average upload performance metrics.

## Requirements

- Go 1.22 or later
- `devices.csv` file with device definitions
- `device-simulator` binary to run the simulator against the API

## Quick Start

```
# run using default flags (By default the server listens on localhost:6733 and looks for devices.csv in the current directory.)
go run main/main.go

# Or run with specified flags
go run main/main.go --device_path="path/to/devices.csv" --addr="<ip_address>:<port>"

# Run the device simulator against the service
./device-simulator --port <port>
```

## Running Tests

```bash
make test          # standard
make test-race     # with Go's race detector
go test -bench . -benchmem ./devices/   # benchmarks
```

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/devices/{device_id}/heartbeat` | Register a heartbeat from a device |
| `POST` | `/api/v1/devices/{device_id}/stats` | uptime and avg upload time |
| `GET`  | `/api/v1/devices/{device_id}/stats` | Retrieve uptime and avg upload time |
| `GET`  | `/healthz` | Liveness probe |

## Architecture

```
┌──────────────┐       HTTP        ┌─────────────────────────────────────────┐
│   Device     │  ──────────────▶  │           Middleware Chain              │
│  Simulator   │                   │  Recovery → BodyLimit → RequestLogger  │
└──────────────┘                   └──────────────────┬──────────────────────┘
                                                      │
                                                      ▼
                                              ┌───────────────┐
                                              │   ServeMux     │
                                              │   (routing)    │
                                              └───┬───┬───┬───┘
                                                  │   │   │
                              ┌────────────────┐  │   │   │  ┌────────────────┐
                              │ handleHeartbeat│◀─┘   │   └─▶│ handleGetStats │
                              └───────┬────────┘      │      └───────┬────────┘
                                      │               │              │
                              ┌───────┴────────┐      │      ┌───────┴────────┐
                              │ handlePostStats │◀─────┘      │ CalculateUptime│
                              └───────┬────────┘              │ CalculateAvg…  │
                                      │                       └───────┬────────┘
                                      ▼                               │
                              ┌─────────────────┐                     │
                              │     Store        │◀────────────────────┘
                              │  map[id]*Device  │
                              └─────────────────┘
                                      │
                              ┌───────┴────────┐
                              │    Device       │
                              │  RWMutex        │
                              │  heartbeats []  │
                              │  uploads    []  │
                              └────────────────┘
```

## Docker

```bash
docker build -t fleet-monitor .
docker run -p 6733:6733 fleet-monitor
```

---

## Write-up

### How long did you spend? What was the hardest part?

I spent around 3 hours working on the MVP. The hardest part was getting the data model right before writing any handler logic. It was tempting to jump straight into HTTP code, but I found that taking time upfront to think through what raw data needs to be stored per device versus what gets computed on demand, made the rest of the implementation straightforward. Specifically, deciding to store raw slices of heartbeat timestamps and upload durations, rather than pre-computing running totals, kept the model simple and correct for an MVP.

This next part was not tricky but it was something important to keep in mind that can be missed if not familiar with using Golang for API development, this being concurrent access. The simulator sends requests from multiple goroutines simultaneously, so I needed to protect each device's data with a sync.Mutex to avoid race conditions. I placed the mutex directly on the Device struct and exposed AddHeartbeat and AddStat methods that handle locking internally, which keeps the handlers clean and prevents the locking logic from leaking into the wrong layer.

### How would you modify the data model to account for more kinds of metrics?

The current Device struct has explicit fields for each metric type (UploadTime, SentAt). This works fine for a few metrics but doesn't scale well as I add more. A cleaner approach would be to make metrics generic. For example, I would introduce a Metric interface and store a map of metric name to a slice of values:

```go
type Device struct {
    mu      sync.Mutex
    DeviceID string
    Metrics  map[string][]float64
}
```

Each new metric type (battery level, CPU usage, signal strength) would just be a new key in the map. The service layer would look up the relevant slice by name and apply the appropriate calculation. This means adding a new metric requires no changes to the `Device` struct, just a new handler and a new service function. You'd also want a corresponding change to the CSV format or a separate configuration file to define which metrics each device type supports.

### Runtime Complexity

- **POST /heartbeat** and **POST /stats**: O(1), appending to a slice is constant time.
- **GET /stats**:
  - `CalculateUptime`: O(1), reads slice length and two elements (first and last heartbeat).
  - `CalculateAvgUploadTime`: O(n), iterates over all recorded upload times to compute the sum, where n is the number of stats recorded for that device.
- **Startup (LoadFromCSV)**: O(n) where n is the number of devices in the CSV.
- **Memory**: O(n × m) where n is the number of devices and m is the number of messages received per device over the lifetime of the server.

### Security, Testing, and Deployment

**Security:**
- Panic recovery middleware catches unexpected panics, logs the stack trace, and returns a clean 500 instead of crashing the process.
- Request body size limit (1 MB via `http.MaxBytesReader`) prevents memory exhaustion from oversized payloads.
- HTTP server timeouts (read 5s, write 10s, idle 60s) prevent slowloris-style resource exhaustion.
- Input validation rejects malformed or semantically invalid payloads with structured 400 JSON responses.
- The Docker image runs as a non-root user (`appuser`, UID 1000).
- For production: add TLS termination (via reverse proxy or `http.ListenAndServeTLS`), rate limiting middleware, and authentication (API keys or mTLS for device-to-service trust).

**Testing:**
- Unit tests cover calculation logic, edge cases (zero heartbeats, single entry, division by zero), and concurrent access with the race detector.
- Integration tests exercise the full HTTP handler stack with `httptest`, verifying status codes and response bodies for happy paths, 404s, invalid JSON, and boundary conditions.
- Middleware tests verify body size rejection and panic recovery.
- `make test-race` runs all tests with Go's race detector to catch data races.
- Benchmarks (`go test -bench . -benchmem ./devices/`) profile the hot paths: heartbeat and upload ingestion, uptime calculation, and average upload time calculation.

**Deployment:**
- Multi-stage Docker build produces a minimal Alpine image with a stripped binary (`-ldflags="-s -w"`).
- `/healthz` endpoint enables Kubernetes liveness and readiness probes.
- Graceful shutdown on SIGINT/SIGTERM lets in-flight requests complete (5s drain) before exit.
- Structured logging (method, path, status, latency) on every request supports observability.

-----------

### AI Tool Usage

Claude (Anthropic) was used to review and refactor the codebase for production readiness, specifically for identifying improvements in encapsulation, error handling patterns, test coverage, middleware hardening, and deployment configuration. I provided my code to Claude and prompted it to identify code smells, areas of enhancement, optimization, and security improvements. It provided the test files, helped optimize my code and also expanded on this README with the Architecture, Security, Testing, and Deployment sections.

Claude provided the below as its reasoning for its changes:

---

```
The core architectural change was improving encapsulation and thread safety on the `Device` struct. In the original code, fields like `SentAt` and `UploadTime` were exported, meaning any handler could read or mutate them without holding the mutex. I made all fields unexported and added accessor methods that operate under a `sync.RWMutex` (upgraded from a plain `Mutex` so concurrent reads no longer block each other). The `Store` map was similarly hidden behind a `Lookup` method instead of being a public field. On the write side, `AddHeartbeat` and `AddUpload` now maintain running aggregates (heartbeat count, first/last timestamps, upload sum and count) in addition to appending to the raw slices. This is the O(1) optimization you called out in your write-up: `CalculateUptime` and `CalculateAvgUploadTime` now read a handful of pre-computed scalars under a read lock instead of copying and iterating entire slices, making the stats endpoint constant time regardless of data volume.

The second layer of changes was production hardening. The server now performs graceful shutdown on SIGINT/SIGTERM with a 5-second drain window, enforces HTTP read/write/idle timeouts to prevent slowloris attacks, and runs through a three-stage middleware chain: panic recovery (catches crashes and returns a clean 500 with a logged stack trace), a 1 MB request body size limit via `http.MaxBytesReader`, and request logging that records method, path, status code, and latency on every request. The handlers themselves were tightened up: 404 responses now include the `{"msg": "..."}` JSON body the OpenAPI spec requires, malformed input returns 400 instead of 500, and a shared `writeJSON` helper plus `lookupDevice` method eliminated duplicated code across all three handlers. I also added a `/healthz` liveness endpoint for Kubernetes probes.

The third area was testing, tooling, and deployment. I added unit tests covering calculation logic and edge cases (zero heartbeats, single entry, division by zero), a concurrent stress test that hammers a device from 200 goroutines, HTTP integration tests using `httptest` for every handler path (success, 404, invalid JSON, negative upload time, empty stats), middleware tests for body size rejection and panic recovery, and benchmarks profiling the hot ingestion and calculation paths. On the deployment side, I added a multi-stage Dockerfile that produces a minimal Alpine image with a stripped binary running as a non-root user, a Makefile for common workflows (`make test`, `make test-race`, `make run`), and a `.gitignore`. The `go.mod` and Dockerfile Go version were also updated to match your local Go 1.26 environment, which fixed the Docker build failure. Finally, `LoadFromCSV` was corrected to accept a header-only CSV as valid (returning an empty store) rather than treating it as an error, which fixed the failing `TestLoadFromCSV_EmptyFile` test.
```