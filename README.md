# Gofi — Router Benchmark Comparison

Comprehensive performance benchmark comparing [**Gofi**](https://github.com/michaelolof/gofi) against industry-standard Go routers.

## Configurations Tested
- **Gofi** — Go 1.22+ `http.ServeMux` wrapper
- **Gofi + Schema** — Gofi with typed schema structs + `ValidateAndBind`
- **Chi** — Standard Chi v5 radix trie router
- **Echo** — Echo v4 high-performance router

---

## Detailed Benchmark Reports

For granular performance data, allocations, and specific use-case results, see the specialized comparison reports:

- [**Gofi vs Chi Benchmarks**](./gofi_vs_chi_benchmarks.md) — Comparison against the standard-bearer for radix trie routing and middleware scalability.
- [**Gofi vs Echo Benchmarks**](./gofi_vs_echo_benchmarks.md) — Comparison against the industry leader in high-parameter routing and concurrency.

---

## High-Level Summary

| Scenario | Gofi Strength | Competitor Strength |
|---|---|---|
| **Static Routes** | High performance | Chi leads on entry cost |
| **Parameterized** | Fast single-params | Echo leads on 5+ params |
| **Middleware** | Constant allocs (3) | Chi leads (constant 2) |
| **Data Handling** | Fastest JSON binding | Tie |
| **Concurrency** | 2x-6x faster than Chi | Echo is fastest |
| **Validation** | [**Automatic**](https://github.com/michaelolof/gofi#type-safety) | Manual / Plug-in |

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

---

## Running Benchmarks

```bash
# Run all benchmarks and generate benchmark-results.md
go run ./cmd/report/

# Run benchmarks manually
go test -bench=. -benchmem -count=3 -timeout=20m

# Run specific category
go test -bench="Github" -benchmem

# Run only schema benchmarks
go test -bench="GofiS|ChiS|EchoS" -benchmem
```

Full raw data available in [benchmark-results.md](./benchmark-results.md).
