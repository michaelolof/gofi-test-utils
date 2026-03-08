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

| Endpoint | Gofi (fasthttp) | Fiber (fasthttp) | Chi (net/http) | Gin (net/http) | Echo (net/http) | Winner |
|---|---|---|---|---|---|---|
| `GET /` | **157,530** | 124,215 | 79,836 | 63,284 | 62,819 | 🏆 **Gofi** |
| `GET /user/gordon` | **145,365** | 100,184 | 63,788 | 60,210 | 49,817 | 🏆 **Gofi** |
| `GET /users/123/posts/456` | **134,095** | 95,929 | 60,242 | 52,943 | 45,659 | 🏆 **Gofi** |
| `GET /middlewares` | **138,207** | 97,649 | 59,914 | 44,947 | 44,942 | 🏆 **Gofi** |
| `GET /query?q=searchterm&limit=10` | **134,956** | 92,097 | 55,574 | 41,640 | 42,214 | 🏆 **Gofi** |
| `POST /json` | **82,164** | 74,025 | 48,473 | 42,401 | 36,825 | 🏆 **Gofi** |
| `GET /json-response` | **125,944** | 46,499 | 39,302 | 32,628 | 29,620 | 🏆 **Gofi** |
| `POST /json-large` | 2,217 | **3,560** | 3,118 | 2,683 | 2,536 | 🏆 **Fiber** |
| `GET /json-response-large` | **119,141** | 8,561 | 8,430 | 6,963 | 7,119 | 🏆 **Gofi** |
| `POST /json-validate-small` | **86,312** | 69,329 | 36,687 | 42,351 | 35,551 | 🏆 **Gofi** |
| `GET /json-response-validate-small` | **110,221** | 79,549 | 40,144 | 50,780 | 43,985 | 🏆 **Gofi** |
| `POST /json-validate-large` | **7,250** | 3,444 | 2,741 | 2,095 | 2,833 | 🏆 **Gofi** |
| `GET /json-response-validate-large` | **110,811** | 8,828 | 7,655 | 7,374 | 8,083 | 🏆 **Gofi** |
| `POST /multipart` | **63,514** | 43,477 | 24,795 | 27,011 | 27,109 | 🏆 **Gofi** |
| `POST /formdata` | **111,096** | 66,186 | 36,580 | 37,934 | 37,893 | 🏆 **Gofi** |

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
