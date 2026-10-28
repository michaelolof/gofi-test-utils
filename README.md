# Gofi — Router Benchmark Comparison

Comprehensive performance benchmark comparing [**Gofi**](https://github.com/michaelolof/gofi) against industry-standard Go routers (Chi, Echo, Gin, and Fiber).

This repository contains two distinct benchmark suites that measure different aspects of routing performance:
1. **Routing Microbenchmarks (In-Process)** — Measures pure routing efficiency, trie traversal speed, and CPU/Memory overhead.
2. **HTTP Load Tests (End-to-End)** — Measures real-world throughput, connection handling, and total HTTP stack performance.

---

## Test Environment
All benchmarks in this repository were executed on the following hardware/software:
- **CPU:** Apple M5 (10 cores)
- **RAM:** 32 GB
- **OS:** macOS 27 (darwin/arm64), Go 1.27.1
## 1. HTTP Load Tests (End-to-End)

These tests spin up real HTTP servers on the IPv4 loopback (`127.0.0.1`) and hit them with `bombardier` to measure true network-level throughput (Requests per Second). 

Because **Gofi** and **Fiber** are built on [fasthttp](https://github.com/valyala/fasthttp) (which hyper-optimizes TCP connection pooling and memory recycling), they significantly outperform standard `net/http` routers in heavy-load scenarios.

### How to Run

```bash
# Build the httpbench binaries and run the network load tests (takes ~8 mins)
python3 ./scripts/httpbench_runner.py
```
*Note: Requires `go` and will install `bombardier` automatically.*

### Results Overview (Reqs/sec)

| Case | Endpoint | Gofi (fasthttp) | Fiber (fasthttp) | Chi (net/http) | Gin (net/http) | Echo (net/http) | Winner |
|---|---|---|---|---|---|---|---|
| Static Route | `GET /` | **182,132** | **182,936** | 177,707 | 175,759 | 175,235 | Tie (Gofi, Fiber) |
| Single Param | `GET /user/gordon` | **183,237** | **183,499** | 175,923 | 179,416 | 176,035 | Tie (Gofi, Fiber) |
| Multi Param | `GET /users/123/posts/456` | **183,126** | **183,517** | 176,021 | 176,327 | 176,024 | Tie (Gofi, Fiber) |
| Middleware Chain | `GET /middlewares` | **183,677** | **182,817** | 177,522 | 176,519 | 172,736 | Tie (Gofi, Fiber) |
| Query Processing | `GET /query?q=searchterm&limit=10` | **184,330** | **183,221** | 173,712 | 173,003 | 173,112 | Tie (Gofi, Fiber) |
| JSON Bind (Small) | `POST /json` | **180,901** | **182,530** | 174,090 | 172,370 | 170,566 | Tie (Gofi, Fiber) |
| JSON Response (Small) | `GET /json-response` | **180,966** | 171,507 | 160,047 | 146,454 | 163,378 | 🏆 **Gofi** |
| JSON Bind Only (Large) | `POST /json-large` | **42,619** | 38,403 | 26,243 | 25,364 | 26,203 | 🏆 **Gofi** |
| JSON Response (Large) | `GET /json-response-large` | **84,072** | 57,795 | 61,253 | 45,612 | 63,976 | 🏆 **Gofi** |
| JSON Bind + Validate (Small) | `POST /json-validate-small` | **182,737** | **182,729** | 172,762 | 172,405 | 171,504 | Tie (Gofi, Fiber) |
| JSON Validate Response (Small) | `GET /json-response-validate-small` | **183,160** | **183,251** | 172,393 | 174,757 | 174,860 | Tie (Gofi, Fiber) |
| JSON Validate Only (Large) | `POST /json-validate-large` | **37,630** | 20,143 | 15,714 | 15,545 | 15,510 | 🏆 **Gofi** |
| JSON Bind + Validate (Large) | `POST /json-bind-validate-large` | **26,045** | 19,825 | 15,529 | 15,499 | 15,340 | 🏆 **Gofi** |
| JSON Validation Failure (Early) | `POST /json-bind-validate-large` | **84,954** | 34,758 | 24,456 | 15,438 | 24,206 | 🏆 **Gofi** |
| JSON Validation Failure (Late/Deep) | `POST /json-bind-validate-large` | **25,890** | 19,811 | 15,404 | 15,448 | 15,242 | 🏆 **Gofi** |
| JSON Validate Response (Large) | `GET /json-response-validate-large` | **47,902** | 24,241 | 21,903 | 19,729 | 21,638 | 🏆 **Gofi** |
| Multipart Bind | `POST /multipart` | **178,818** | 173,716 | 135,448 | 138,348 | 136,260 | 🏆 **Gofi** |
| FormData Bind | `POST /formdata` | **182,676** | **182,940** | 166,609 | 166,478 | 164,727 | Tie (Gofi, Fiber) |

Results within 2% of the best are a tie. Requests with small bodies saturate the `bombardier` client at about 180k req/s on this machine, so those rows show the client limit rather than a difference between routers. The validated-response endpoints validate the payload before encoding it in every router.

For full details, see the generated `httpbench-results.md`.

---

## 2. Routing Microbenchmarks (In-Process)

These tests use Go's standard `testing.B` benchmark harness tool to evaluate the raw speed and memory footprint of the routing logic itself, completely stripping out the TCP stack.

In these tests, the pure `net/http` routers (Chi, Echo, Gin) appear faster in `ns/op` because they don't incur `fasthttp.RequestCtx` allocation overhead during mock tests. However, Gofi dominates in routing storage efficiency and JSON binding logic.

### How to Run

```bash
# Run all Go microbenchmarks and generate benchmark-results.md and individual markdown reports
go run ./cmd/report/
```

### Detailed Markdown Reports

The `go run ./cmd/report/` script generates specific head-to-head comparison markdown files that detail exactly where each framework excels:

- [**Gofi vs Chi**](./gofi_vs_chi_benchmarks.md) — Chi wins in pure radix traversal speed; Gofi wins in memory footprint and JSON processing.
- [**Gofi vs Echo**](./gofi_vs_echo_benchmarks.md) — Echo wins in high-concurrency (in-process) routing; Gofi wins in schema predictability.
- [**Gofi vs Gin**](./gofi_vs_gin_benchmarks.md) — Gin wins raw parameter writes; Gofi wins data handling and automatic validation.
- [**Gofi vs Fiber**](./gofi_vs_fibre_benchmarks.md) — Gofi outperforms Fiber across the board in routing, memory, and JSON parsing.

### What these tests mean:
- **`ns/op` (Nanoseconds per operation):** Lower is better. Measures the CPU time required to resolve a route. Chi usually wins this metric.
- **`B/op` (Bytes per operation):** Lower is better. Measures heap allocations required per request.
- **`Memory Consumption`:** Gofi consistently wins this. Gofi's fasthttp radix tree requires ~50% less memory to store routes than Chi or Echo.
- **`Schema Overhead`:** By using `<Router> + Schema` benchmarks, we test the extra CPU cost of strongly-typed struct validation. Gofi's compiler-based schema binding is much more predictable (~20-30% overhead) compared to Echo or Gin which can spike up to 500% overhead on complex routes.

---

## Core Test Suites

The [`internal/suites`](./internal/suites) package contains a comprehensive suite of tests that verify Gofi's core functionality and ensure cross-router consistency:

- **Binding & Validation** — JSON, Form, and Multipart body parsing with schema validation.
- **Routing & Middleware** — Parameterized routes, group-scoped handlers, and middleware chains.
- **Data Handling** — Dependency injection, context storage, and custom response types.
- **Error Handling** — Custom error responders and centralized validation failure handling.

---

## Usage Examples

Explore practical implementations of Gofi in the [`cmd/examples`](./cmd/examples) directory:

| Example | Description | Link |
|---|---|---|
| **JSON API** | High-performance JSON request/response handling. | [View](./cmd/examples/json) |
| **Form Data** | Handling `application/x-www-form-urlencoded` payloads. | [View](./cmd/examples/formdata) |
| **Multipart** | Managing single and multiple file uploads. | [View](./cmd/examples/multipart) |
