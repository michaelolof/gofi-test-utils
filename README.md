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
| Static Route | `GET /` | **164,049** | 109,097 | 51,392 | 47,795 | 28,080 | 🏆 **Gofi** |
| Single Param | `GET /user/gordon` | **103,889** | 85,414 | 40,731 | 44,584 | 21,611 | 🏆 **Gofi** |
| Multi Param | `GET /users/123/posts/456` | **126,116** | 77,671 | 40,336 | 39,286 | 18,398 | 🏆 **Gofi** |
| Middleware Chain | `GET /middlewares` | **128,225** | 77,501 | 33,169 | 37,314 | 16,538 | 🏆 **Gofi** |
| Query Processing | `GET /query?q=searchterm&limit=10` | **126,361** | 68,762 | 30,487 | 35,612 | 20,287 | 🏆 **Gofi** |
| JSON Bind (Small) | `POST /json` | **75,164** | 58,023 | 32,040 | 30,205 | 13,758 | 🏆 **Gofi** |
| JSON Response (Small) | `GET /json-response` | **120,821** | 36,524 | 26,297 | 24,360 | 12,489 | 🏆 **Gofi** |
| JSON Bind (Large) | `POST /json-large` | 2,545 | **3,019** | 1,972 | 2,137 | 1,240 | 🏆 **Fiber** |
| JSON Response (Large) | `GET /json-response-large` | **98,193** | 6,837 | 4,825 | 5,670 | 4,071 | 🏆 **Gofi** |
| JSON Validate (Small) | `POST /json-validate-small` | **77,564** | 54,143 | 23,761 | 32,319 | 19,306 | 🏆 **Gofi** |
| JSON Validate Response (Small) | `GET /json-response-validate-small` | **106,625** | 64,104 | 24,806 | 41,042 | 27,612 | 🏆 **Gofi** |
| JSON Validate (Large) | `POST /json-validate-large` | **6,648** | 2,841 | 1,708 | 1,682 | 1,702 | 🏆 **Gofi** |
| JSON Validate Response (Large) | `GET /json-response-validate-large` | **95,677** | 6,902 | 4,033 | 5,894 | 4,878 | 🏆 **Gofi** |
| Multipart Bind | `POST /multipart` | **49,474** | 29,958 | 14,597 | 21,083 | 16,749 | 🏆 **Gofi** |
| FormData Bind | `POST /formdata` | **78,102** | 48,143 | 20,093 | 28,930 | 23,473 | 🏆 **Gofi** |

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
