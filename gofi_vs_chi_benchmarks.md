# Gofi vs Chi — Benchmark Comparison

Detailed performance comparison between [**Gofi**](https://github.com/michaelolof/gofi) and [**Chi**](https://github.com/go-chi/chi) HTTP routers.

## Configurations Tested
- **Gofi** — Go 1.22+ `http.ServeMux` wrapper
- **Gofi + Schema** — Gofi with typed schema structs + `ValidateAndBind`
- **Chi** — Standard Chi v5 radix trie router
- **Chi + Schema** — Chi with manual struct binding + `go-playground/validator`

Full raw data: [benchmark-results.md](./benchmark-results.md)

---

## Memory Consumption

| API | Routes | Gofi | Gofi + Schema | Chi |
|---|---|---|---|---|
| Static | 157 | 91 KB | 314 KB | **78 KB** |
| GitHub | 203 | 135 KB | 382 KB | **91 KB** |
| Google+ | 13 | 10 KB | 27 KB | **6 KB** |
| Parse.com | 26 | 17 KB | 46 KB | **8 KB** |

> 🥇 **Chi** — consistently lowest memory for route storage (radix trie is ~40% more compact than stdlib wrapper).

---

## Micro Benchmarks

### Static Route — `GET /`
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Chi** | **305** | **368** | **2** |
| Gofi | 414 | 416 | 3 |
| Chi + Schema | 447 | 370 | 3 |
| Gofi + Schema | 921 | 488 | 10 |

> 🥇 **Chi** — 305 ns (26% faster than Gofi).

### Single Param — `GET /user/:name`
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Gofi** | **534** | **432** | **4** |
| Chi | 601 | 704 | 4 |
| Chi + Schema | 1,080 | 722 | 6 |
| Gofi + Schema | 1,374 | 536 | 13 |

> 🥇 **Gofi** — 534 ns (11% faster than Chi).

### 5 Params — `GET /:a/:b/:c/:d/:e`
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Chi** | **911** | 704 | **4** |
| Gofi | 958 | **656** | 7 |
| Chi + Schema | 1,820 | 786 | 6 |
| Gofi + Schema | 2,531 | 888 | 20 |

> 🥇 **Chi** — 911 ns.

### 20 Params — `GET /:a/:b/.../:t`
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Gofi** | **1,961** | **1,424** | **9** |
| Chi | 3,416 | 2,505 | 9 |

> 🥇 **Gofi** — 1,961 ns (74% faster than Chi for extremely deep param counts).

### Param Write — `GET /user/:name`
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Gofi** | **245** | **16** | **1** |
| Chi | 634 | 704 | 4 |
| Chi + Schema | 1,067 | 720 | 5 |
| Gofi + Schema | 1,079 | 120 | 10 |

> 🥇 **Gofi** — 245 ns (2.5x faster than Chi for raw param resolution).

### Multi Param — `GET /users/:id/posts/:id`
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Chi** | **692** | 704 | **4** |
| Gofi | 737 | **464** | 5 |
| Chi + Schema | 1,251 | 738 | 6 |
| Gofi + Schema | 1,688 | 600 | 15 |

> 🥇 **Chi** — 692 ns.

### Wildcard — `GET /files/*`
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Chi** | **655** | 704 | **4** |
| Gofi | 981 | **504** | 8 |

> 🥇 **Chi** — 655 ns.

### Deep Nesting
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Chi** | **361** | **368** | **2** |
| Gofi | 707 | 416 | 3 |

> 🥇 **Chi** — 361 ns (1.9x faster than Gofi via single trie traversal).

### 404 Handling
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Gofi** | **700** | **464** | **6** |
| Chi | 890 | 816 | 7 |

> 🥇 **Gofi** — 700 ns (21% faster than Chi).

---

## Middleware & Data Handling

### Middleware Scalability
| Middlewares | Gofi | Chi | Gofi allocs | Chi allocs |
|---|---|---|---|---|
| 5 | 557 | **393** | 3 | **2** |
| 10 | 658 | **475** | 3 | **2** |
| 20 | 849 | **554** | 3 | **2** |

> 🥇 **Chi** — fastest at all counts, constant 2 allocs (best in class).

### JSON Binding (Small Payload)
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Gofi** | **6,541** | 7,469 | 30 |
| Chi | 6,567 | **7,101** | **29** |
| Chi + Schema | 7,316 | 7,106 | 29 |
| Gofi + Schema | 8,704 | 7,398 | 43 |

> 🥇 **Gofi** — 6,541 ns (effectively tied with Chi).

### JSON Response (100 items)
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Gofi** | **20,323** | **8,778** | **19** |
| Chi | 23,894 | 9,144 | 21 |

> 🥇 **Gofi** — 20,323 ns (15% faster than Chi, less memory).

---

## Concurrency & Route Groups

### Concurrency (Parallel Requests)
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Gofi** | **186** | **21** | **1** |
| Chi | 499 | 368 | 2 |

> 🥇 **Gofi** — 186 ns (2.6x faster than Chi).

### Route Groups
| Router | ns/op | B/op | allocs/op |
|---|---|---|---|
| **Gofi** | **1,553** | **432** | **4** |
| Chi | 2,568 | 736 | 6 |

> 🥇 **Gofi** — 1,553 ns (39% faster than Chi).

---

## Real-World APIs

### GitHub API (203 routes)
| Benchmark | Gofi | Chi |
|---|---|---|
| Memory | 135 KB | **91 KB** |
| Static (ns/op) | 842 | **565** |
| Param (ns/op) | 1,222 | **1,112** |
| **All (ns/op)** | 256,278 | **221,660** |

> 🥇 **Chi** — fastest full traversal and lowest memory.

---

## Schema Overhead

| Scenario | Gofi | Gofi + Schema | Chi | Chi + Schema |
|---|---|---|---|---|
| Static | 414 ns | 921 ns (2.2x) | 305 ns | 447 ns (1.5x) |
| 1 param | 534 ns | 1,374 ns (2.6x) | 601 ns | 1,080 ns (1.8x) |
| JSON bind | 6,541 ns | 8,704 ns (1.3x) | 6,567 ns | 7,316 ns (1.1x) |

---

## Key Takeaways

### Gofi excels at:
- **Single Param Routing** — High performance for standard REST patterns.
- **Param Resolution** — 2.5x faster raw parameter access.
- **Concurrency** — 2.6x faster parallel throughput.
- **JSON Response Memory** — Lowest per-request footprint.
- **Automatic Validation** — Seamless validation via `ValidateAndBind`.

### Chi excels at:
- **Static Routes** — Radix trie optimization.
- **Memory Storage** — 40% more compact route storage.
- **Deep Nesting** — Single trie traversal for deep paths.
- **Middleware** — Zero-cost middleware scalability.
- **Low Registry Cost** — Schema overhead is significantly lower than Gofi's compiler.
