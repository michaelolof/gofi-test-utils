# Gofi vs Chi — Benchmark Comparison

Detailed performance comparison between [**Gofi**](https://github.com/michaelolof/gofi) and [**Chi**](https://github.com/go-chi/chi) HTTP routers.

## Configurations Tested
- **Gofi** — fasthttp-backed radix tree router
- **Gofi + Schema** — Gofi with typed schema structs + `ValidateAndBind`
- **Chi** — Standard Chi v5 radix trie router
- **Chi + Schema** — Chi with manual struct binding + `go-playground/validator`

Full raw data: [benchmark-results.md](./benchmark-results.md)

---

## Test Environment
- **CPU:** Intel(R) Core(TM) i7-4980HQ CPU @ 2.80GHz
- **RAM:** 16 GB
- **OS:** macOS (darwin/amd64)
## Memory Consumption

| API | Routes | Gofi | Gofi + Schema | Chi |
|---|---|---|---|---|
| Static | 157 | **28 KB** | 304 KB | 0 B |
| GitHub | 203 | **47 KB** | 369 KB | 0 B |
| Google+ | 13 | **3 KB** | 25 KB | 0 B |
| Parse.com | 26 | **6 KB** | 44 KB | 0 B |

> 🥇 **Gofi** — consistently lowest memory for route storage thanks to the fasthttp radix tree. Chi reports 0 B because its memory measurement method differs.

---

## Micro Benchmarks

### Static Route — `GET /`
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Chi** | **318** | **368** | **2** |
| Chi + Schema | 983 | 370 | 3 |
| Gofi | 13,128 | 2,178 | 16 |
| Gofi + Schema | 13,979 | 2,218 | 21 |

> 🥇 **Chi** — 318 ns.

### Single Param — `GET /user/:name`
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Chi** | **542** | **704** | **4** |
| Chi + Schema | 2,437 | 722 | 6 |
| Gofi | 12,935 | 2,202 | 16 |
| Gofi + Schema | 20,411 | 2,482 | 18 |

> 🥇 **Chi** — 542 ns.

### 5 Params — `GET /:a/:b/:c/:d/:e`
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Chi** | **941** | **704** | **4** |
| Chi + Schema | 4,001 | 786 | 6 |
| Gofi | 19,969 | 2,250 | 16 |
| Gofi + Schema | 20,251 | 2,530 | 18 |

> 🥇 **Chi** — 941 ns.

### 20 Params — `GET /:a/:b/.../:t`
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Chi** | **9,130** | **2,504** | **9** |
| Gofi | 16,812 | 2,298 | 16 |

> 🥇 **Chi** — 9,130 ns.

### Param Write — `GET /user/:name`
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Chi** | **1,903** | **704** | **4** |
| Chi + Schema | 2,827 | 720 | 5 |
| Gofi | 13,136 | 2,178 | 16 |
| Gofi + Schema | 16,674 | 2,482 | 18 |

> 🥇 **Chi** — 1,903 ns.

### Multi Param — `GET /users/:id/posts/:id`
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Chi** | **2,169** | **704** | **4** |
| Chi + Schema | 6,872 | 738 | 6 |
| Gofi | 16,750 | 2,226 | 16 |
| Gofi + Schema | 22,185 | 2,506 | 18 |

> 🥇 **Chi** — 2,169 ns.

### Wildcard — `GET /files/*`
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Chi** | **2,190** | **704** | **4** |
| Gofi | 16,724 | 2,226 | 16 |

> 🥇 **Chi** — 2,190 ns.

### Deep Nesting
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Chi** | **1,088** | **368** | **2** |
| Gofi | 23,149 | 2,298 | 16 |

> 🥇 **Chi** — 1,088 ns.

### 404 Handling
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Chi** | **2,385** | **816** | **7** |
| Gofi | 18,353 | 2,186 | 15 |

> 🥇 **Chi** — 2,385 ns.

---

## Middleware & Data Handling

### Middleware Scalability
| Middlewares | Gofi | Chi | Gofi allocs | Chi allocs |
|---|---|---|---|---|
| 5 | 3,075 | **400** | 16 | **2** |
| 10 | 3,289 | **496** | 16 | **2** |
| 20 | 3,619 | **545** | 16 | **2** |

> 🥇 **Chi** — fastest at all counts, constant 2 allocs. Gofi holds steady at 16 allocs regardless of middleware count.

### JSON Binding (Small Payload)
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Gofi** | **19,127** | **3,274** | **24** |
| Gofi + Schema | 22,866 | 2,851 | 30 |
| Chi | 106,855 | 7,101 | 29 |
| Chi + Schema | 117,794 | 7,105 | 29 |

> 🥇 **Gofi** — 19,127 ns.

### JSON Response (100 items)
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| Chi | 587,350 | 9,156 | 21 |
| **Gofi** | **70,199** | **5,280** | **20** |

> 🥇 **Gofi** — 70,199 ns.

---

## Concurrency & Route Groups

### Concurrency (Parallel Requests)
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| Chi | 6,568 | 368 | 2 |
| **Gofi** | **5,852** | **2,177** | **16** |

> 🥇 **Gofi** — 5,852 ns.

### Route Groups
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| Chi | 56,558 | 736 | 6 |
| **Gofi** | **9,447** | **2,226** | **16** |

> 🥇 **Gofi** — 9,447 ns.

---

## Real-World APIs

### GitHub API (203 routes)
| Benchmark | Gofi | Gofi + Schema | Chi |
|---|---|---|---|
| Memory | 47 KB | 369 KB | 0 B |
| Static (ns/op) | 3,219 | 3,180 | **487** |
| Param (ns/op) | 3,588 | 3,742 | **1,013** |
| All (ns/op) | 729,567 | 729,106 | **178,046** |

> 🥇 **Chi** — fastest full traversal (4x faster). **Gofi** — 47 KB route memory.

### Google+ API (13 routes)
| Benchmark | Gofi | Gofi + Schema | Chi |
|---|---|---|---|
| Memory | 3 KB | 25 KB | 0 B |
| Static (ns/op) | 3,094 | 3,109 | **431** |
| 1 Param (ns/op) | 3,629 | 3,512 | **829** |
| 2 Params (ns/op) | 3,652 | 3,508 | **865** |
| All (ns/op) | 48,044 | 47,633 | **9,295** |

> 🥇 **Chi** — fastest per-request (5x faster).

### Parse.com API (26 routes)
| Benchmark | Gofi | Gofi + Schema | Chi |
|---|---|---|---|
| Memory | 6 KB | 44 KB | 0 B |
| Static (ns/op) | 3,346 | 3,354 | **400** |
| 1 Param (ns/op) | 3,191 | 3,390 | **774** |
| 2 Params (ns/op) | 3,494 | 3,495 | **910** |
| All (ns/op) | 86,983 | 92,713 | **19,669** |

> 🥇 **Chi** — fastest per-request (4.4x faster).

---

## Schema Overhead

| Scenario | Gofi | Gofi + Schema | Chi | Chi + Schema |
|---|---|---|---|---|
| Static | 2,981 ns | 3,805 ns (1.3x) | 316 ns | 389 ns (1.2x) |
| 1 param | 3,189 ns | 3,810 ns (1.2x) | 555 ns | 886 ns (1.6x) |
| 5 params | 3,370 ns | 4,468 ns (1.3x) | 831 ns | 1,416 ns (1.7x) |
| JSON bind | 4,221 ns | 6,505 ns (1.5x) | 6,438 ns | 7,255 ns (1.1x) |


---

