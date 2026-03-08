# Gofi vs Gin — Benchmark Comparison

Detailed performance comparison between [**Gofi**](https://github.com/michaelolof/gofi) and [**Gin**](https://github.com/gin-gonic/gin) HTTP routers.

## Configurations Tested
- **Gofi** — fasthttp-backed radix tree router
- **Gofi + Schema** — Gofi with typed schema structs + `ValidateAndBind`
- **Gin** — Gin high-performance HTTP framework
- **Gin + Schema** — Gin with `ShouldBindUri` / `ShouldBindJSON` struct validation

Full raw data: [benchmark-results.md](./benchmark-results.md)

---

## Memory Consumption

| API | Routes | Gofi | Gofi + Schema | Gin |
|---|---|---|---|---|
| Static | 157 | **28 KB** | 304 KB | 0 B |
| GitHub | 203 | **47 KB** | 369 KB | 0 B |
| Google+ | 13 | **3 KB** | 25 KB | 0 B |
| Parse.com | 26 | **6 KB** | 44 KB | 0 B |

> 🥇 **Gofi** — consistently lowest memory for route storage thanks to the fasthttp radix tree. Gin reports 0 B because its memory measurement method differs.

---

## Micro Benchmarks

### Static Route — `GET /`
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Gin** | **1,140** | **448** | **3** |
| Gin + Schema | 1,159 | 448 | 3 |
| Gofi | 13,128 | 2,178 | 16 |
| Gofi + Schema | 13,979 | 2,218 | 21 |

> 🥇 **Gin** — 1,140 ns.

### Single Param — `GET /user/:name`
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Gin** | **1,305** | **448** | **3** |
| Gin + Schema | 4,613 | 896 | 8 |
| Gofi | 12,935 | 2,202 | 16 |
| Gofi + Schema | 20,411 | 2,482 | 18 |

> 🥇 **Gin** — 1,305 ns.

### 5 Params — `GET /:a/:b/:c/:d/:e`
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Gin** | **1,382** | **448** | **3** |
| Gofi | 19,969 | 2,250 | 16 |
| Gofi + Schema | 20,251 | 2,530 | 18 |
| Gin + Schema | 8,268 | 1,089 | 12 |

> 🥇 **Gin** — 1,382 ns.

### 20 Params — `GET /:a/:b/.../:t`
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Gin** | **1,856** | **448** | **3** |
| Gofi | 16,812 | 2,298 | 16 |

> 🥇 **Gin** — 1,856 ns.

### Param Write — `GET /user/:name`
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Gin** | **206** | **0** | **0** |
| Gin + Schema | 2,719 | 448 | 5 |
| Gofi | 13,136 | 2,178 | 16 |
| Gofi + Schema | 16,674 | 2,482 | 18 |

> 🥇 **Gin** — 206 ns.

### Multi Param — `GET /users/:id/posts/:id`
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Gin** | **1,051** | **448** | **3** |
| Gin + Schema | 4,641 | 944 | 9 |
| Gofi | 16,750 | 2,226 | 16 |
| Gofi + Schema | 22,185 | 2,506 | 18 |

> 🥇 **Gin** — 1,051 ns.

### Wildcard — `GET /files/*`
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Gin** | **1,125** | **448** | **3** |
| Gofi | 16,724 | 2,226 | 16 |

> 🥇 **Gin** — 1,125 ns.

### Deep Nesting
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Gin** | **1,089** | **448** | **3** |
| Gofi | 23,149 | 2,298 | 16 |

> 🥇 **Gin** — 1,089 ns.

### 404 Handling
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Gin** | **959** | **400** | **2** |
| Gofi | 18,353 | 2,186 | 15 |

> 🥇 **Gin** — 959 ns.

---

## Middleware & Data Handling

### Middleware Scalability
| Middlewares | Gofi | Gin | Gofi allocs | Gin allocs |
|---|---|---|---|---|
| 5 | 3,075 | **908** | 16 | **3** |
| 10 | 3,289 | **935** | 16 | **3** |
| 20 | 3,619 | **1,103** | 16 | **3** |

> 🥇 **Gin** — 3-4x faster with constant 3 allocs at any depth. Gofi holds steady at 16 allocs.

### JSON Binding (Small Payload)
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Gofi** | **19,127** | **3,274** | **24** |
| Gofi + Schema | 22,866 | 2,851 | 30 |
| Gin | 21,355 | 7,501 | 30 |
| Gin + Schema | 21,115 | 7,506 | 30 |

> 🥇 **Gofi** — 19,127 ns.

### JSON Response (100 items)
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| Gofi | 70,199 | 5,280 | 20 |
| **Gin** | **61,704** | **11,482** | **20** |

> 🥇 **Gin** — 61,704 ns.

---

## Concurrency & Route Groups

### Concurrency (Parallel Requests)
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Gin** | **269** | **54** | **1** |
| Gofi | 5,852 | 2,177 | 16 |

> 🥇 **Gin** — 269 ns.

### Route Groups
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Gin** | **1,916** | **448** | **3** |
| Gofi | 9,447 | 2,226 | 16 |

> 🥇 **Gin** — 1,916 ns.

---

## Real-World APIs

### GitHub API (203 routes)
| Benchmark | Gofi | Gofi + Schema | Gin |
|---|---|---|---|
| Memory | 47 KB | 369 KB | 0 B |
| Static (ns/op) | 3,219 | 3,180 | **670** |
| Param (ns/op) | 3,588 | 3,742 | **801** |
| All (ns/op) | 729,567 | 729,106 | **173,287** |

> 🥇 **Gin** — fastest full traversal (4.2x faster).

### Google+ API (13 routes)
| Benchmark | Gofi | Gofi + Schema | Gin |
|---|---|---|---|
| Memory | 3 KB | 25 KB | 0 B |
| Static (ns/op) | 3,094 | 3,109 | **690** |
| 1 Param (ns/op) | 3,629 | 3,512 | **779** |
| 2 Params (ns/op) | 3,652 | 3,508 | **791** |
| All (ns/op) | 48,044 | 47,633 | **9,202** |

> 🥇 **Gin** — fastest per-request (5.2x faster).

### Parse.com API (26 routes)
| Benchmark | Gofi | Gofi + Schema | Gin |
|---|---|---|---|
| Memory | 6 KB | 44 KB | 0 B |
| Static (ns/op) | 3,346 | 3,354 | **679** |
| 1 Param (ns/op) | 3,191 | 3,390 | **704** |
| 2 Params (ns/op) | 3,494 | 3,495 | **763** |
| All (ns/op) | 86,983 | 92,713 | **17,930** |

> 🥇 **Gin** — fastest per-request (4.9x faster).

---

## Schema Overhead

| Scenario | Gofi | Gofi + Schema | Gin | Gin + Schema |
|---|---|---|---|---|
| Static | 2,981 ns | 3,805 ns (1.3x) | 609 ns | 658 ns (1.1x) |
| 1 param | 3,189 ns | 3,810 ns (1.2x) | 797 ns | 2,457 ns (3.1x) |
| 5 params | 3,370 ns | 4,468 ns (1.3x) | 804 ns | 4,624 ns (5.8x) |
| JSON bind | 4,221 ns | 6,505 ns (1.5x) | 8,057 ns | 9,723 ns (1.2x) |


---

