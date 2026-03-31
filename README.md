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

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/devices/{device_id}/heartbeat` | Register a heartbeat from a device |
| `POST` | `/api/v1/devices/{device_id}/stats` | uptime and avg upload time |
| `GET`  | `/api/v1/devices/{device_id}/stats` | Retrieve uptime and avg upload time |


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

Overall the server is lightweight and capaable of high-frequency telemetry ingestion since writes are O(1) and the only linear scan happens at read time across a bounded set of per-device records. However, improvements can be made and I will explore those shortly integrating it with a LLM.

-----------

Overall, the MVP satisfies the requirements.
I indentified that the most unoptimized part of the project is `GET stats` which gets the uptime and average upload time, it is O(n) per call due to performing all calculation each time. In production, this could be reduced to O(1) by maintaining a running sum and count (incremental average), trading write-time simplicity for read-time performance. However, I kept it like this to verify the MVP. Shortly, I will integrate the project with AI to implement this as well as identify areas of improvement and enhancements.
