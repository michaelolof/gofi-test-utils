# Gofi vs Fiber — Benchmark Comparison

Detailed performance comparison between [**Gofi**](https://github.com/michaelolof/gofi) and [**Fiber**](https://github.com/gofiber/fiber) HTTP routers.

## Configurations Tested
- **Gofi** — fasthttp-backed radix tree router
- **Gofi + Schema** — Gofi with typed schema structs + `ValidateAndBind`
- **Fiber** — Fiber v2 fasthttp-based router
- **Fiber + Schema** — Fiber with `ParamsParser` + `go-playground/validator`

Full raw data: [benchmark-results.md](./benchmark-results.md)

---

## Memory Consumption

| API | Routes | Gofi | Gofi + Schema | Fiber |
|---|---|---|---|---|
| Static | 157 | **28 KB** | 304 KB | 0 B |
| GitHub | 203 | **47 KB** | 369 KB | 0 B |
| Google+ | 13 | **3 KB** | 25 KB | 0 B |
| Parse.com | 26 | **6 KB** | 44 KB | 0 B |

> 🥇 **Gofi** — consistently lowest memory for route storage thanks to the fasthttp radix tree. Fiber reports 0 B because its memory measurement method differs.

---

## Micro Benchmarks

### Static Route — `GET /`
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Gofi** | **13,128** | **2,178** | **16** |
| Gofi + Schema | 13,979 | 2,218 | 21 |
| Fiber | 38,154 | 6,486 | 25 |
| Fiber + Schema | 41,120 | 6,484 | 25 |

> 🥇 **Gofi** — 13,128 ns.

### Single Param — `GET /user/:name`
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Gofi** | **12,935** | **2,202** | **16** |
| Gofi + Schema | 20,411 | 2,482 | 18 |
| Fiber | 42,091 | 6,479 | 25 |
| Fiber + Schema | 78,406 | 6,809 | 37 |

> 🥇 **Gofi** — 12,935 ns.

### 5 Params — `GET /:a/:b/:c/:d/:e`
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Gofi** | **19,969** | **2,250** | **16** |
| Gofi + Schema | 20,251 | 2,530 | 18 |
| Fiber | 42,264 | 6,478 | 25 |
| Fiber + Schema | 80,503 | 7,523 | 69 |

> 🥇 **Gofi** — 19,969 ns.

### 20 Params — `GET /:a/:b/.../:t`
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Gofi** | **16,812** | **2,298** | **16** |
| Fiber | 46,241 | 6,622 | 26 |

> 🥇 **Gofi** — 16,812 ns.

### Param Write — `GET /user/:name`
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Gofi** | **13,136** | **2,178** | **16** |
| Gofi + Schema | 16,674 | 2,482 | 18 |
| Fiber | 39,932 | 6,485 | 25 |
| Fiber + Schema | 64,845 | 6,810 | 37 |

> 🥇 **Gofi** — 13,136 ns.

### Multi Param — `GET /users/:id/posts/:id`
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Gofi** | **16,750** | **2,226** | **16** |
| Gofi + Schema | 22,185 | 2,506 | 18 |
| Fiber | 40,499 | 6,482 | 25 |
| Fiber + Schema | 75,507 | 6,990 | 45 |

> 🥇 **Gofi** — 16,750 ns.

### Wildcard — `GET /files/*`
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Gofi** | **16,724** | **2,226** | **16** |
| Fiber | 39,318 | 6,482 | 25 |

> 🥇 **Gofi** — 16,724 ns.

### Deep Nesting
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Gofi** | **23,149** | **2,298** | **16** |
| Fiber | 37,862 | 6,626 | 26 |

> 🥇 **Gofi** — 23,149 ns.

### 404 Handling
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Gofi** | **18,353** | **2,186** | **15** |
| Fiber | 44,383 | 6,576 | 29 |

> 🥇 **Gofi** — 18,353 ns.

---

## Middleware & Data Handling

### Middleware Scalability
| Middlewares | Gofi | Fiber | Gofi allocs | Fiber allocs |
|---|---|---|---|---|
| 5 | **3,075** | 19,322 | **16** | 25 |
| 10 | **3,289** | 20,400 | **16** | 25 |
| 20 | **3,619** | 29,843 | **16** | 25 |

> 🥇 **Gofi** — 6-8x faster at all middleware depths. Both maintain constant allocations (16 vs 25), but Fiber degrades noticeably at 20 middlewares (29,843 ns).

### JSON Binding (Small Payload)
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Gofi** | **19,127** | **3,274** | **24** |
| Gofi + Schema | 22,866 | 2,851 | 30 |
| Fiber + Schema | 46,430 | 6,860 | 31 |
| Fiber | 46,001 | 6,866 | 31 |

> 🥇 **Gofi** — 19,127 ns.

### JSON Response (100 items)
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Gofi** | **70,199** | **5,280** | **20** |
| Fiber | 84,645 | 12,238 | 28 |

> 🥇 **Gofi** — 70,199 ns.

---

## Concurrency & Route Groups

### Concurrency (Parallel Requests)
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Gofi** | **5,852** | **2,177** | **16** |
| Fiber | 14,783 | 6,421 | 25 |

> 🥇 **Gofi** — 5,852 ns.

### Route Groups
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Gofi** | **9,447** | **2,226** | **16** |
| Fiber | 53,562 | 6,482 | 25 |

> 🥇 **Gofi** — 9,447 ns.

---

## Real-World APIs

### GitHub API (203 routes)
| Benchmark | Gofi | Gofi + Schema | Fiber |
|---|---|---|---|
| Memory | 47 KB | 369 KB | 0 B |
| Static (ns/op) | **3,219** | 3,180 | 21,321 |
| Param (ns/op) | **3,588** | 3,742 | 21,765 |
| All (ns/op) | **729,567** | 729,106 | 5,207,706 |

> 🥇 **Gofi** — 7.1x faster full traversal (729K vs 5.2M ns). Dominant across every sub-benchmark.

### Google+ API (13 routes)
| Benchmark | Gofi | Gofi + Schema | Fiber |
|---|---|---|---|
| Memory | 3 KB | 25 KB | 0 B |
| Static (ns/op) | **3,094** | 3,109 | 20,006 |
| 1 Param (ns/op) | 3,629 | **3,512** | 20,573 |
| 2 Params (ns/op) | 3,652 | **3,508** | 22,180 |
| All (ns/op) | 48,044 | **47,633** | 287,720 |

> 🥇 **Gofi** — 6x faster per-request.

### Parse.com API (26 routes)
| Benchmark | Gofi | Gofi + Schema | Fiber |
|---|---|---|---|
| Memory | 6 KB | 44 KB | 0 B |
| Static (ns/op) | **3,346** | 3,354 | 20,640 |
| 1 Param (ns/op) | **3,191** | 3,390 | 21,099 |
| 2 Params (ns/op) | **3,494** | 3,495 | 21,031 |
| All (ns/op) | **86,983** | 92,713 | 553,807 |

> 🥇 **Gofi** — 6.4x faster full traversal.

---

## Schema Overhead

| Scenario | Gofi | Gofi + Schema | Fiber | Fiber + Schema |
|---|---|---|---|---|
| Static | 2,981 ns | 3,805 ns (1.3x) | 18,461 ns | 18,360 ns (1.0x) |
| 1 param | 3,189 ns | 3,810 ns (1.2x) | 19,170 ns | 31,968 ns (1.7x) |
| 5 params | 3,370 ns | 4,468 ns (1.3x) | 19,402 ns | 38,013 ns (2.0x) |
| JSON bind | 4,221 ns | 6,505 ns (1.5x) | 28,137 ns | 24,752 ns (0.9x) |


---

