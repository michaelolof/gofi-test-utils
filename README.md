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
| Static Route | `GET /` | **183,482** | **182,663** | 176,702 | 175,828 | 175,154 | Tie (Gofi, Fiber) |
| Single Param | `GET /user/gordon` | **183,883** | **182,662** | 176,055 | 179,195 | 176,124 | Tie (Gofi, Fiber) |
| Multi Param | `GET /users/123/posts/456` | **183,032** | **183,660** | 175,498 | 175,774 | 175,492 | Tie (Gofi, Fiber) |
| Middleware Chain | `GET /middlewares` | **183,838** | **183,015** | 177,806 | 176,164 | 175,751 | Tie (Gofi, Fiber) |
| Query Processing | `GET /query?q=searchterm&limit=10` | **183,403** | **182,200** | 173,712 | 169,408 | 173,788 | Tie (Gofi, Fiber) |
| JSON Bind (Small) | `POST /json` | **181,703** | **182,680** | 173,939 | 171,914 | 171,072 | Tie (Gofi, Fiber) |
| JSON Response (Small) | `GET /json-response` | 86,989 | **172,713** | 160,834 | 148,352 | 164,404 | 🏆 **Fiber** |
| JSON Bind Only (Large) | `POST /json-large` | **43,026** | 38,903 | 26,438 | 25,634 | 26,457 | 🏆 **Gofi** |
| JSON Response (Large) | `GET /json-response-large` | 18,828 | **62,885** | 62,061 | 46,334 | **63,716** | Tie (Fiber, Echo) |
| JSON Bind + Validate (Small) | `POST /json-validate-small` | **182,675** | **183,442** | 173,819 | 172,390 | 170,556 | Tie (Gofi, Fiber) |
| JSON Validate Response (Small) | `GET /json-response-validate-small` | **182,905** | **183,540** | 173,250 | 175,518 | 175,253 | Tie (Gofi, Fiber) |
| JSON Validate Only (Large) | `POST /json-validate-large` | **37,426** | 20,068 | 15,861 | 15,865 | 15,591 | 🏆 **Gofi** |
| JSON Bind + Validate (Large) | `POST /json-bind-validate-large` | **25,556** | 20,063 | 15,775 | 15,868 | 15,641 | 🏆 **Gofi** |
| JSON Validation Failure (Early) | `POST /json-bind-validate-large` | **85,707** | 34,793 | 24,668 | 15,818 | 24,517 | 🏆 **Gofi** |
| JSON Validation Failure (Late/Deep) | `POST /json-bind-validate-large` | **26,211** | 19,809 | 15,467 | 15,609 | 15,442 | 🏆 **Gofi** |
| JSON Validate Response (Large) | `GET /json-response-validate-large` | 15,932 | 58,796 | **61,485** | 45,186 | **62,157** | Tie (Chi, Echo) |
| Multipart Bind | `POST /multipart` | **178,506** | 172,044 | 135,955 | 138,073 | 137,357 | 🏆 **Gofi** |
| FormData Bind | `POST /formdata` | **183,491** | **182,910** | 165,997 | 166,050 | 163,199 | Tie (Gofi, Fiber) |

Results within 2% of the best are a tie. Requests with small bodies saturate the `bombardier` client at about 180k req/s on this machine, so those rows show the client limit rather than a difference between routers.

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
