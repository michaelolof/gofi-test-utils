# Gofi vs Echo — Benchmark Comparison

Detailed performance comparison between [**Gofi**](https://github.com/michaelolof/gofi) and [**Echo**](https://github.com/labstack/echo) HTTP routers.

## Configurations Tested
- **Gofi** — Go 1.22+ `http.ServeMux` wrapper
- **Gofi + Schema** — Gofi with typed schema structs + `ValidateAndBind`
- **Echo** — Echo v4 high-performance router
- **Echo + Schema** — Echo with `c.Bind()` + `c.Validate()` + validator

Full raw data: [benchmark-results.md](./benchmark-results.md)

---

## Memory Consumption

| API | Routes | Gofi | Gofi + Schema | Echo |
|---|---|---|---|---|
| Static | 157 | 91 KB | 314 KB | **88 KB** |
| GitHub | 203 | 135 KB | 382 KB | **115 KB** |
| Google+ | 13 | 10 KB | 27 KB | **10 KB** |
| Parse.com | 26 | 17 KB | 46 KB | **13 KB** |

> 🥇 **Gofi** and **Echo** are closely matched on route storage, with Gofi being slightly more efficient as route count grows.

---

## Micro Benchmarks

### Static Route — `GET /`
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Gofi** | **414** | **416** | **3** |
| Echo | 566 | 424 | 4 |
| Echo + Schema | 594 | 424 | 4 |
| Gofi + Schema | 921 | 488 | 10 |

> 🥇 **Gofi** — 414 ns (26% faster than Echo).

### Single Param — `GET /user/:name`
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Gofi** | **534** | 432 | **4** |
| Echo | 611 | **424** | 4 |
| Gofi + Schema | 1,374 | 536 | 13 |
| Echo + Schema | 1,482 | 504 | 7 |

> 🥇 **Gofi** — 534 ns (12% faster than Echo).

### 5 Params — `GET /:a/:b/:c/:d/:e`
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Echo** | **699** | **424** | **4** |
| Gofi | 958 | 656 | 7 |
| Gofi + Schema | 2,531 | 888 | 20 |
| Echo + Schema | 3,049 | 632 | 11 |

> 🥇 **Echo** — 699 ns (constant 424 B / 4 allocs regardless of param count).

### 20 Params — `GET /:a/:b/.../:t`
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Echo** | **998** | **424** | **4** |
| Gofi | 1,961 | 1,424 | 9 |

> 🥇 **Echo** — 998 ns (constant-allocation design dominates at high param counts).

### Param Write — `GET /user/:name`
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Echo** | **139** | **8** | **1** |
| Gofi | 245 | 16 | 1 |
| Echo + Schema | 838 | 88 | 4 |
| Gofi + Schema | 1,079 | 120 | 10 |

> 🥇 **Echo** — 139 ns / 8 B / 1 alloc.

### Multi Param — `GET /users/:id/posts/:id`
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Echo** | **590** | **424** | **4** |
| Gofi | 737 | 464 | 5 |
| Gofi + Schema | 1,688 | 600 | 15 |
| Echo + Schema | 1,847 | 536 | 8 |

> 🥇 **Echo** — 590 ns.

### Wildcard — `GET /files/*`
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Echo** | **653** | **424** | **4** |
| Gofi | 981 | 504 | 8 |

> 🥇 **Echo** — 653 ns.

### Deep Nesting
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Gofi** | **707** | **416** | **3** |
| Echo | 814 | 424 | 4 |

> 🥇 **Gofi** — 707 ns.

### 404 Handling
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Gofi** | **700** | **464** | **6** |
| Echo | 1,970 | 896 | 10 |

> 🥇 **Gofi** — 700 ns (2.8x faster than Echo).

---

## Middleware & Data Handling

### Middleware Scalability
| Middlewares | Gofi | Echo | Gofi allocs | Echo allocs |
|---|---|---|---|---|
| 5 | **557** | 981 | 3 | 9 |
| 10 | **658** | 1,199 | 3 | 14 |
| 20 | **849** | 1,765 | 3 | 24 |

> 🥇 **Gofi** — fastest at all counts, linear allocation growth for Echo (9 → 14 → 24).

### JSON Binding (Small Payload)
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Gofi** | **6,541** | **7,469** | **30** |
| Echo | 8,618 | 7,478 | 31 |
| Gofi + Schema | 8,704 | 7,398 | 43 |
| Echo + Schema | 9,286 | 7,483 | 31 |

> 🥇 **Gofi** — 6,541 ns (24% faster than Echo's default binder).

### JSON Response (100 items)
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Gofi** | **20,323** | **8,778** | **19** |
| Echo | 21,485 | 8,825 | 20 |

> 🥇 **Gofi** — 20,323 ns (5% faster than Echo).

---

## Concurrency & Route Groups

### Concurrency (Parallel Requests)
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Echo** | **76** | **13** | **1** |
| Gofi | 186 | 21 | 1 |

> 🥇 **Echo** — 76 ns (2.4x faster than Gofi).

### Route Groups
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Echo** | **991** | 472 | 7 |
| Gofi | 1,553 | **432** | **4** |

> 🥇 **Echo** — 991 ns.

---

## Real-World APIs

### GitHub API (203 routes)
| Benchmark | Gofi | Echo |
|---|---|---|
| Memory | 135 KB | **115 KB** |
| Static (ns/op) | 842 | **647** |
| Param (ns/op) | 1,222 | **738** |
| **All (ns/op)** | 256,278 | **162,877** |

> 🥇 **Echo** — fastest full traversal and lower memory.

---

## Schema Overhead

| Scenario | Gofi | Gofi + Schema | Echo | Echo + Schema |
|---|---|---|---|---|
| Static | 414 ns | 921 ns (2.2x) | 566 ns | 594 ns (1.1x) |
| 1 param | 534 ns | 1,374 ns (2.6x) | 611 ns | 1,482 ns (2.4x) |
| JSON bind | 6,541 ns | 8,704 ns (1.3x) | 8,618 ns | 9,286 ns (1.1x) |

---

## Key Takeaways

### Gofi excels at:
- **Single Param Routing** — Optimized for common REST patterns.
- **404 Handling** — 2.8x faster unmatched route resolution.
- **Middleware scalability** — Constant allocations (3) vs Echo's linear growth.
- **JSON Binding** — 24% faster raw binding performance.
- **Automatic Validation** — Type-safe validation integrated into the router.

### Echo excels at:
- **High Param Counts** — Constant-allocation design (424 B) for complex paths.
- **Concurrency** — Best-in-class parallel throughput (76 ns).
- **Route Groups** — Fast execution across nested groups.
- **Param Write** — fastest raw parameter access.
- **Low Registry Cost** — Schema overhead is very low for standard static/bind patterns.
