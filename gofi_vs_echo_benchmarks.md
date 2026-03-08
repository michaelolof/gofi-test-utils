# Gofi vs Echo — Benchmark Comparison

Detailed performance comparison between [**Gofi**](https://github.com/michaelolof/gofi) and [**Echo**](https://github.com/labstack/echo) HTTP routers.

## Configurations Tested
- **Gofi** — fasthttp-backed radix tree router
- **Gofi + Schema** — Gofi with typed schema structs + `ValidateAndBind`
- **Echo** — Echo v4 high-performance router
- **Echo + Schema** — Echo with `c.Bind()` + `c.Validate()` + validator

Full raw data: [benchmark-results.md](./benchmark-results.md)

---

## Test Environment
- **CPU:** Intel(R) Core(TM) i7-4980HQ CPU @ 2.80GHz
- **RAM:** 16 GB
- **OS:** macOS (darwin/amd64)
## Memory Consumption

| API | Routes | Gofi | Gofi + Schema | Echo |
|---|---|---|---|---|
| Static | 157 | **28 KB** | 304 KB | 0 B |
| GitHub | 203 | **47 KB** | 369 KB | 0 B |
| Google+ | 13 | **3 KB** | 25 KB | 0 B |
| Parse.com | 26 | **6 KB** | 44 KB | 0 B |

> 🥇 **Gofi** — consistently lowest memory for route storage thanks to the fasthttp radix tree. Echo reports 0 B because its memory measurement method differs.

---

## Micro Benchmarks

### Static Route — `GET /`
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Echo** | **5,968** | **424** | **4** |
| Echo + Schema | 13,908 | 424 | 4 |
| Gofi | 13,128 | 2,178 | 16 |
| Gofi + Schema | 13,979 | 2,218 | 21 |

> 🥇 **Echo** — 5,968 ns.

### Single Param — `GET /user/:name`
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| Echo | 14,301 | 424 | 4 |
| Echo + Schema | 31,630 | 504 | 7 |
| **Gofi** | **12,935** | **2,202** | **16** |
| Gofi + Schema | 20,411 | 2,482 | 18 |

> 🥇 **Gofi** — 12,935 ns.

### 5 Params — `GET /:a/:b/:c/:d/:e`
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Echo** | **14,553** | **424** | **4** |
| Gofi | 19,969 | 2,250 | 16 |
| Echo + Schema | 103,675 | 632 | 11 |
| Gofi + Schema | 20,251 | 2,530 | 18 |

> 🥇 **Echo** — 14,553 ns.

### 20 Params — `GET /:a/:b/.../:t`
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| Echo | 46,093 | 424 | 4 |
| **Gofi** | **16,812** | **2,298** | **16** |

> 🥇 **Gofi** — 16,812 ns.

### Param Write — `GET /user/:name`
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Echo** | **3,986** | **8** | **1** |
| Echo + Schema | 34,446 | 88 | 4 |
| Gofi | 13,136 | 2,178 | 16 |
| Gofi + Schema | 16,674 | 2,482 | 18 |

> 🥇 **Echo** — 3,986 ns.

### Multi Param — `GET /users/:id/posts/:id`
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Echo** | **13,452** | **424** | **4** |
| Echo + Schema | 44,872 | 536 | 8 |
| Gofi | 16,750 | 2,226 | 16 |
| Gofi + Schema | 22,185 | 2,506 | 18 |

> 🥇 **Echo** — 13,452 ns.

### Wildcard — `GET /files/*`
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Echo** | **5,156** | **424** | **4** |
| Gofi | 16,724 | 2,226 | 16 |

> 🥇 **Echo** — 5,156 ns.

### Deep Nesting
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Echo** | **4,732** | **424** | **4** |
| Gofi | 23,149 | 2,298 | 16 |

> 🥇 **Echo** — 4,732 ns.

### 404 Handling
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Echo** | **10,080** | **896** | **10** |
| Gofi | 18,353 | 2,186 | 15 |

> 🥇 **Echo** — 10,080 ns.

---

## Middleware & Data Handling

### Middleware Scalability
| Middlewares | Gofi | Echo | Gofi allocs | Echo allocs |
|---|---|---|---|---|
| 5 | 3,075 | **1,230** | **16** | 9 |
| 10 | 3,289 | **1,497** | **16** | 14 |
| 20 | 3,619 | **2,057** | **16** | 24 |

> 🥇 **Echo** — faster at all counts. However, **Gofi** maintains constant 16 allocs while Echo grows linearly (9 → 14 → 24). At scale, Gofi's allocation pattern is more predictable.

### JSON Binding (Small Payload)
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| Gofi | 19,127 | 3,274 | 24 |
| Gofi + Schema | 22,866 | 2,851 | 30 |
| Echo | 14,255 | 7,477 | 31 |
| **Echo + Schema** | **13,657** | **7,482** | **31** |

> 🥇 **Echo + Schema** — 13,657 ns.

### JSON Response (100 items)
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| Gofi | 70,199 | 5,280 | 20 |
| **Echo** | **28,277** | **8,827** | **20** |

> 🥇 **Echo** — 28,277 ns.

---

## Concurrency & Route Groups

### Concurrency (Parallel Requests)
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Echo** | **111** | **14** | **1** |
| Gofi | 5,852 | 2,177 | 16 |

> 🥇 **Echo** — 111 ns.

### Route Groups
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Echo** | **1,543** | **472** | **7** |
| Gofi | 9,447 | 2,226 | 16 |

> 🥇 **Echo** — 1,543 ns.

---

## Real-World APIs

### GitHub API (203 routes)
| Benchmark | Gofi | Gofi + Schema | Echo | Echo + Schema |
|---|---|---|---|---|
| Memory | 47 KB | 369 KB | 0 B | 0 B |
| Static (ns/op) | 3,219 | 3,180 | **799** | 749 |
| Param (ns/op) | 3,588 | 3,742 | 1,148 | **850** |
| All (ns/op) | 729,567 | 729,106 | 207,693 | **182,704** |

> 🥇 **Echo** — fastest full traversal (4x faster). **Gofi** — 47 KB route memory.

### Google+ API (13 routes)
| Benchmark | Gofi | Gofi + Schema | Echo | Echo + Schema |
|---|---|---|---|---|
| Memory | 3 KB | 25 KB | 0 B | 0 B |
| Static (ns/op) | 3,094 | 3,109 | 911 | **690** |
| 1 Param (ns/op) | 3,629 | 3,512 | 937 | **738** |
| 2 Params (ns/op) | 3,652 | 3,508 | 922 | **833** |
| All (ns/op) | 48,044 | 47,633 | 12,033 | **10,353** |

> 🥇 **Echo** — fastest per-request (4.6x faster).

### Parse.com API (26 routes)
| Benchmark | Gofi | Gofi + Schema | Echo | Echo + Schema |
|---|---|---|---|---|
| Memory | 6 KB | 44 KB | 0 B | 0 B |
| Static (ns/op) | 3,346 | 3,354 | 785 | **715** |
| 1 Param (ns/op) | 3,191 | 3,390 | 834 | **734** |
| 2 Params (ns/op) | 3,494 | 3,495 | 801 | **762** |
| All (ns/op) | 86,983 | 92,713 | 22,056 | **21,010** |

> 🥇 **Echo** — fastest per-request (4.1x faster).

---

## Schema Overhead

| Scenario | Gofi | Gofi + Schema | Echo | Echo + Schema |
|---|---|---|---|---|
| Static | 2,981 ns | 3,805 ns (1.3x) | 629 ns | 652 ns (1.0x) |
| 1 param | 3,189 ns | 3,810 ns (1.2x) | 722 ns | 1,925 ns (2.7x) |
| 5 params | 3,370 ns | 4,468 ns (1.3x) | 915 ns | 4,079 ns (4.5x) |
| JSON bind | 4,221 ns | 6,505 ns (1.5x) | 8,498 ns | 9,736 ns (1.1x) |


---

