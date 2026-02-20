# Gofi vs Chi vs Echo — Benchmark Comparison

Comprehensive performance benchmark comparing [**Gofi**](https://github.com/michaelolof/gofi), [**Chi**](https://github.com/go-chi/chi), and [**Echo**](https://github.com/labstack/echo) HTTP routers for Go — each tested plain and with schema validation/binding.

**Six configurations tested:**
- **Gofi** — Go 1.22+ `http.ServeMux` wrapper
- **Gofi + Schema** — Gofi with typed schema structs + `ValidateAndBind`
- **Chi** — Standard Chi v5 radix trie router
- **Chi + Schema** — Chi with manual struct binding + `go-playground/validator`
- **Echo** — Echo v4 high-performance router
- **Echo + Schema** — Echo with `c.Bind()` + `c.Validate()` + validator

Full raw data: [benchmark-results.md](./benchmark-results.md)

---

## Memory Consumption

| API | Routes | Gofi | Gofi + Schema | Chi | Echo |
|---|---|---|---|---|---|
| Static | 157 | 97 KB | 341 KB | **83 KB** | 88 KB |
| GitHub | 203 | 135 KB | 423 KB | **91 KB** | 114 KB |
| Google+ | 13 | 10 KB | 29 KB | **6 KB** | 10 KB |
| Parse.com | 26 | 17 KB | 51 KB | **8 KB** | 13 KB |

> 🥇 **Chi** — consistently lowest memory for route storage
> 🥈 **Echo** — moderate footprint
> 🥉 **Gofi** — slightly higher than Echo. Gofi + Schema uses ~3.5x more due to schema compilation at registration

---

## Micro Benchmarks

### Static Route — `GET /`

| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Chi** | **315** | **368** | **2** |
| Chi + Schema | 399 | 370 | 3 |
| Gofi | 424 | 416 | 3 |
| Echo | 502 | 424 | 4 |
| Echo + Schema | 503 | 424 | 4 |
| Gofi + Schema | 590 | 432 | 5 |

> 🥇 **Chi** — 315 ns (26% faster than Gofi)
> 🥈 **Gofi** — 424 ns
> 🥉 **Echo** — 502 ns

### Single Param — `GET /user/:name`

| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Echo** | **526** | **424** | **4** |
| Gofi | 543 | 432 | 4 |
| Chi | 623 | 704 | 4 |
| Chi + Schema | 946 | 722 | 6 |
| Echo + Schema | 1,365 | 504 | 7 |
| Gofi + Schema | 2,007 | 736 | 15 |

> 🥇 **Echo** — 526 ns
> 🥈 **Gofi** — 543 ns (within 3% — effectively tied)
> 🥉 **Chi** — 623 ns

### 5 Params — `GET /:a/:b/:c/:d/:e`

| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Echo** | **640** | **424** | **4** |
| Gofi | 923 | 656 | 7 |
| Chi | 939 | 704 | 4 |
| Chi + Schema | 1,538 | 786 | 6 |
| Echo + Schema | 2,949 | 632 | 11 |
| Gofi + Schema | 4,020 | 1,088 | 22 |

> 🥇 **Echo** — 640 ns (constant 424 B / 4 allocs regardless of param count)
> 🥈 **Gofi** — 923 ns
> 🥉 **Chi** — 939 ns (effectively tied with Gofi)

### 20 Params — `GET /:a/:b/.../:t`

| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Echo** | **921** | **424** | **4** |
| Gofi | 1,897 | 1,424 | 9 |
| Chi | 3,478 | 2,504 | 9 |

> 🥇 **Echo** — 921 ns (constant allocation design dominates at high param counts)
> 🥈 **Gofi** — 1,897 ns (46% faster than Chi)
> 🥉 **Chi** — 3,478 ns

### Param Write — `GET /user/:name` (reads param + writes to response)

| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Echo** | **126** | **8** | **1** |
| Gofi | 243 | 16 | 1 |
| Chi | 691 | 704 | 4 |
| Echo + Schema | 805 | 88 | 4 |
| Chi + Schema | 883 | 720 | 5 |
| Gofi + Schema | 1,783 | 320 | 12 |

> 🥇 **Echo** — 126 ns / 8 B / 1 alloc
> 🥈 **Gofi** — 243 ns / 16 B / 1 alloc (1.9x slower, still **2.8x faster** than Chi)
> 🥉 **Chi** — 691 ns / 704 B / 4 allocs

### Multi Param — `GET /users/:userID/posts/:postID`

| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Echo** | **534** | **424** | **4** |
| Gofi | 707 | 464 | 5 |
| Chi | 713 | 704 | 4 |
| Chi + Schema | 1,153 | 738 | 6 |
| Echo + Schema | 1,747 | 536 | 8 |
| Gofi + Schema | 2,575 | 800 | 17 |

> 🥇 **Echo** — 534 ns
> 🥈 **Gofi** — 707 ns (tied with Chi, 34% less memory)
> 🥉 **Chi** — 713 ns

### Wildcard — `GET /files/*`

| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Echo** | **501** | **424** | **4** |
| Chi | 635 | 704 | 4 |
| Gofi | 940 | 504 | 8 |

> 🥇 **Echo** — 501 ns
> 🥈 **Chi** — 635 ns
> 🥉 **Gofi** — 940 ns (lowest memory at 504 B)

### Deep Nesting — `GET /v1/api/deep/nested/resource/action`

| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Chi** | **371** | **368** | **2** |
| Echo | 537 | 424 | 4 |
| Gofi | 721 | 416 | 3 |

> 🥇 **Chi** — 371 ns (1.9x faster than Gofi via single trie traversal)
> 🥈 **Echo** — 537 ns
> 🥉 **Gofi** — 721 ns

### 404 Handling

| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Gofi** | **707** | **464** | **6** |
| Chi | 971 | 816 | 7 |
| Echo | 1,320 | 896 | 10 |

> 🥇 **Gofi** — 707 ns (27% faster than Chi, 43% less memory)
> 🥈 **Chi** — 971 ns
> 🥉 **Echo** — 1,320 ns

---

## Middleware Scalability

| Middlewares | Gofi | Chi | Echo | Gofi allocs | Chi allocs | Echo allocs |
|---|---|---|---|---|---|---|
| 5 | 513 | **371** | 705 | 3 | **2** | 9 |
| 10 | 566 | **402** | 919 | 3 | **2** | 14 |
| 20 | 669 | **476** | 1,350 | 3 | **2** | 24 |

> 🥇 **Chi** — fastest at all counts, constant 2 allocs (best in class)
> 🥈 **Gofi** — constant 3 allocs, reasonably close to Chi
> 🥉 **Echo** — allocations grow linearly (9 → 14 → 24)

---

## Data Handling

### JSON Binding (Small Payload)

| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Echo** | **5,744** | 7,477 | 31 |
| Chi | 5,954 | **7,101** | **29** |
| Gofi | 6,052 | 7,469 | 30 |
| Echo + Schema | 7,021 | 7,482 | 31 |
| Chi + Schema | 7,414 | 7,106 | 29 |
| Gofi + Schema | 7,751 | 7,598 | 45 |

> 🥇 **Echo** — 5,744 ns
> 🥈 **Chi** — 5,954 ns (lowest memory, fewest allocs)
> 🥉 **Gofi** — 6,052 ns
>
> All three within ~5% — bottleneck is `encoding/json`.

### JSON Response (100 items)

| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Gofi** | **14,474** | **8,778** | **19** |
| Chi | 15,227 | 9,146 | 21 |
| Echo | 18,031 | 8,824 | 20 |

> 🥇 **Gofi** — 14,474 ns (5% faster than Chi, least memory and allocs)
> 🥈 **Chi** — 15,227 ns
> 🥉 **Echo** — 18,031 ns

---

## Concurrency

| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Gofi** | **106** | 21 | **1** |
| Echo | 146 | **14** | 1 |
| Chi | 264 | 368 | 2 |

> 🥇 **Gofi** — 106 ns (27% faster than Echo, 2.5x faster than Chi)
> 🥈 **Echo** — 146 ns (least memory at 14 B)
> 🥉 **Chi** — 264 ns

---

## Route Groups (nested middleware)

| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Gofi** | **640** | **432** | **4** |
| Chi | 1,252 | 736 | 6 |
| Echo | 1,747 | 472 | 7 |

> 🥇 **Gofi** — 640 ns (49% faster than Chi, least memory and allocs)
> 🥈 **Chi** — 1,252 ns
> 🥉 **Echo** — 1,747 ns

---

## Real-World APIs

### GitHub API (203 routes)

| Benchmark | Gofi | Gofi + Schema | Chi | Echo | Echo + Schema |
|---|---|---|---|---|---|
| Memory | 135 KB | 423 KB | **91 KB** | 114 KB | 115 KB |
| Static (ns/op) | 575 | 579 | **415** | 1,266 | 1,687 |
| Param (ns/op) | **811** | 809 | 767 | 1,520 | **755** |
| **All (ns/op)** | 167,636 | 175,840 | **162,471** | 189,340 | 553,661 |
| All (B/op) | 94,233 | 94,234 | 130,866 | **86,105** | **86,105** |
| All (allocs) | 946 | 946 | **740** | 812 | 812 |

> 🥇 **Chi** — 162,471 ns full traversal (fastest, fewest allocs at 740)
> 🥈 **Gofi** — 167,636 ns (within 3%, lowest memory per iteration alongside Gofi + Schema)
> 🥉 **Echo** — 189,340 ns (86 KB per iteration — least memory)

### Google+ API (13 routes)

| Benchmark | Gofi | Gofi + Schema | Chi | Echo | Echo + Schema |
|---|---|---|---|---|---|
| Memory | 10 KB | 29 KB | **6 KB** | 10 KB | 10 KB |
| Static (ns/op) | 502 | 488 | **363** | 1,255 | 550 |
| 1 Param (ns/op) | **635** | 644 | 671 | 2,292 | 581 |
| 2 Params (ns/op) | 854 | 822 | 737 | 2,132 | **623** |
| **All (ns/op)** | 9,692 | 9,459 | **8,742** | 16,447 | 26,420 |
| All (B/op) | 5,746 | 5,746 | 8,483 | **5,514** | **5,514** |

> 🥇 **Chi** — 8,742 ns full traversal (fewest allocs)
> 🥈 **Gofi + Schema** — 9,459 ns
> 🥉 **Gofi** — 9,692 ns (Echo/Gofi tied on per-iteration memory)

### Parse.com API (26 routes)

| Benchmark | Gofi | Gofi + Schema | Chi | Echo | Echo + Schema |
|---|---|---|---|---|---|
| Memory | 17 KB | 51 KB | **8 KB** | 13 KB | 13 KB |
| Static (ns/op) | 560 | 565 | **384** | 1,903 | 535 |
| 1 Param (ns/op) | 673 | 687 | 695 | 1,815 | **552** |
| 2 Params (ns/op) | 829 | 791 | 735 | 1,932 | **576** |
| **All (ns/op)** | 19,098 | 19,074 | **16,338** | 35,177 | 55,257 |
| All (B/op) | 11,173 | 11,172 | 14,949 | **11,028** | **11,028** |

> 🥇 **Chi** — 16,338 ns full traversal (fastest, fewest allocs)
> 🥈 **Gofi + Schema** — 19,074 ns
> 🥉 **Gofi** — 19,098 ns (Echo + Schema fastest for individual param lookups)

---

## Schema Overhead

The cost of adding schema validation/binding to each router:

| Scenario | Gofi | Gofi + Schema | Chi | Chi + Schema | Echo | Echo + Schema |
|---|---|---|---|---|---|---|
| Static | 424 ns | 590 ns (1.4x) | 315 ns | 399 ns (1.3x) | 502 ns | 503 ns (1.0x) |
| 1 param | 543 ns | 2,007 ns (**3.7x**) | 623 ns | 946 ns (1.5x) | 526 ns | 1,365 ns (2.6x) |
| 5 params | 923 ns | 4,020 ns (**4.4x**) | 939 ns | 1,538 ns (1.6x) | 640 ns | 2,949 ns (**4.6x**) |
| JSON bind | 6,052 ns | 7,751 ns (1.3x) | 5,954 ns | 7,414 ns (1.2x) | 5,744 ns | 7,021 ns (1.2x) |

> 🥇 **Echo + Schema** — lowest overhead for static (1.0x) and JSON (1.2x)
> 🥈 **Chi + Schema** — consistently low (~1.2-1.6x) via direct struct assignment
> 🥉 **Gofi + Schema** — highest per-request overhead (~3.7-4.4x for params) — the cost of **automatic, type-safe `ValidateAndBind`**

---

## Key Takeaways

### Gofi excels at:
- **Concurrency** — 🥇 106 ns (27% faster than Echo, 2.5x faster than Chi)
- **Route groups** — 🥇 640 ns (49% faster than Chi)
- **JSON response** — 🥇 14,474 ns (fastest, least memory)
- **404 handling** — 🥇 fastest unmatched route resolution
- **Raw param access** — 🥈 243 ns / 16 B / 1 alloc
- **Gofi + Schema** — automatic `ValidateAndBind` with type safety (unique feature)

### Chi excels at:
- **Static routes** — 🥇 fastest trie lookup (26% faster than Gofi)
- **Deep nesting** — 🥇 1.9x faster than Gofi
- **Middleware scalability** — 🥇 constant 2 allocs (best in class)
- **Route storage memory** — 🥇 40-60% less than Gofi at registration
- **Full API traversal** — 🥇 fastest for GitHub, Google+, Parse.com
- **Schema overhead** — 🥈 Chi + Schema adds only ~30-60%

### Echo excels at:
- **Parameterized routing** — 🥇 constant 424 B / 4 allocs regardless of param count
- **Param write** — 🥇 126 ns / 8 B (fastest raw param access)
- **Wildcard** — 🥇 501 ns
- **JSON binding** — 🥇 5,744 ns
- **Schema overhead** — 🥇 lowest for static routes (1.0x)

### The trade-off:

| | Gofi | Chi | Echo |
|---|---|---|---|
| **Fastest for** | Concurrency, route groups, JSON response, 404s | Static, deep nesting, middleware, API traversal | Params, wildcard, param write, JSON bind |
| **Memory model** | Low per-request, moderate storage | Low storage, higher per-request | Constant per-request, moderate storage |
| **Schema cost** | ~3.7-4.4x (full auto ValidateAndBind) | ~1.3-1.6x (manual struct binding) | ~1.0-4.6x (built-in Bind + Validate) |
| **Best for** | Type-safe APIs with validation | Many static routes, heavy middleware | High-param APIs, raw performance |

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
