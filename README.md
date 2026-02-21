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
