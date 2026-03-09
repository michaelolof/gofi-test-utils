# Gofi — Router Benchmark Comparison

Comprehensive performance benchmark comparing [**Gofi**](https://github.com/michaelolof/gofi) against industry-standard Go routers (Chi, Echo, Gin, and Fiber).

This repository contains two distinct benchmark suites that measure different aspects of routing performance:
1. **Routing Microbenchmarks (In-Process)** — Measures pure routing efficiency, trie traversal speed, and CPU/Memory overhead.
2. **HTTP Load Tests (End-to-End)** — Measures real-world throughput, connection handling, and total HTTP stack performance.

---

## Test Environment
All benchmarks in this repository were executed on the following hardware/software:
- **CPU:** Intel(R) Core(TM) i7-4980HQ CPU @ 2.80GHz
- **RAM:** 16 GB
- **OS:** macOS (darwin/amd64)
## 1. HTTP Load Tests (End-to-End)

These tests spin up real HTTP servers on localhost and hit them with `bombardier` to measure true network-level throughput (Requests per Second). 

Because **Gofi** and **Fiber** are built on [fasthttp](https://github.com/valyala/fasthttp) (which hyper-optimizes TCP connection pooling and memory recycling), they significantly outperform standard `net/http` routers in heavy-load scenarios.

### How to Run

```bash
# Generate the httpbench binaries and run network load tests (takes ~2 mins)
python3 ./cmd/httpbench/runner.py
```
*Note: Requires `go` and will install `bombardier` automatically.*

### Results Overview (Reqs/sec)

| Case | Endpoint | Gofi (fasthttp) | Fiber (fasthttp) | Chi (net/http) | Gin (net/http) | Echo (net/http) | Winner |
|---|---|---|---|---|---|---|---|
| Static Route | `GET /` | **163,880** | 83,525 | 60,663 | 43,530 | 14,799 | 🏆 **Gofi** |
| Single Param | `GET /user/gordon` | **139,461** | 72,217 | 43,594 | 40,774 | 14,151 | 🏆 **Gofi** |
| Multi Param | `GET /users/123/posts/456` | **134,118** | 63,252 | 30,245 | 34,869 | 16,030 | 🏆 **Gofi** |
| Middleware Chain | `GET /middlewares` | **135,437** | 66,113 | 27,912 | 34,058 | 15,901 | 🏆 **Gofi** |
| Query Processing | `GET /query?q=searchterm&limit=10` | **129,563** | 59,226 | 25,752 | 33,093 | 15,661 | 🏆 **Gofi** |
| JSON Bind (Small) | `POST /json` | **75,308** | 45,575 | 20,473 | 29,130 | 15,214 | 🏆 **Gofi** |
| JSON Response (Small) | `GET /json-response` | **119,048** | 27,631 | 15,263 | 22,990 | 13,483 | 🏆 **Gofi** |
| JSON Bind (Large) | `POST /json-large` | **2,377** | 2,128 | 1,133 | 2,080 | 1,355 | 🏆 **Gofi** |
| JSON Response (Large) | `GET /json-response-large` | **117,543** | 4,658 | 2,831 | 5,154 | 3,758 | 🏆 **Gofi** |
| JSON Validate (Small) | `POST /json-validate-small` | **80,958** | 39,710 | 13,278 | 27,656 | 20,150 | 🏆 **Gofi** |
| JSON Validate Response (Small) | `GET /json-response-validate-small` | **98,963** | 51,098 | 12,405 | 34,316 | 26,632 | 🏆 **Gofi** |
| JSON Validate (Large) | `POST /json-validate-large` | **6,124** | 2,092 | 831 | 1,393 | 1,625 | 🏆 **Gofi** |
| JSON Validate Response (Large) | `GET /json-response-validate-large` | **76,345** | 5,023 | 2,357 | 4,765 | 4,322 | 🏆 **Gofi** |
| Multipart Bind | `POST /multipart` | **49,574** | 23,896 | 7,547 | 17,307 | 16,894 | 🏆 **Gofi** |
| FormData Bind | `POST /formdata` | **88,269** | 36,532 | 8,766 | 24,873 | 22,766 | 🏆 **Gofi** |

**What this means:** Under heavy concurrent network load, Gofi processes about **2x more requests per second** than Chi, Echo, or Gin. Even against Fiber (which shares the fasthttp backend), Gofi's optimized radix tree and `fastjson` engine deliver 20% to 60% higher throughput.

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
