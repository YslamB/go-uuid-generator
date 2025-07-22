# Distributed Unique ID Generator (Go + Fiber + gRPC)

A high-performance, distributed 64-bit unique ID generator using the **Twitter Snowflake** algorithm, implemented in **Go** with **Clean Architecture** principles. It provides both **HTTP (Fiber)** and **gRPC** APIs.


---

## 🚀 Features

- 64-bit sortable numeric IDs
- Based on Twitter Snowflake:
  - 1-bit sign (always 0)
  - 41-bit timestamp (in ms since epoch)
  - 5-bit datacenter ID (up to 32)
  - 5-bit machine ID (up to 32)
  - 12-bit sequence number (4,096 IDs/ms per node)
- Over **4956 unique IDs/millisecond**
- gRPC & HTTP support
- Cleanly separated layers (Domain, Usecase, Interface, Infra)
- Easily scalable across machines and datacenters

---

## Snowflake UUID Generator Benchmark Test Results

This document contains benchmark results for the custom Snowflake-based UUID generator implementation written in Go.

## 🧪 Benchmark Commands

```bash
go test -bench=. ./internal/infra/snowflake
```
## 🔁 Run #1:
```bash
goos: darwin
goarch: arm64
pkg: id-generator/internal/infra/snowflake
cpu: Apple M2

BenchmarkSnowflakeGenerator-8            4924080               244.0 ns/op
BenchmarkSnowflakeParallel-8             4334239               310.9 ns/op
PASS
ok      id-generator/internal/infra/snowflake   3.251s
```
## 🔁 Run #2
```bash
goos: darwin
goarch: arm64
pkg: id-generator/internal/infra/snowflake
cpu: Apple M2

BenchmarkSnowflakeGenerator-8            4931898               243.9 ns/op
BenchmarkSnowflakeParallel-8             4302860               316.7 ns/op
PASS
ok      id-generator/internal/infra/snowflake   7.425s
```
## 📈 Performance Summary
| Test Case          | ns/op   | Estimated Throughput (IDs/sec) |
| ------------------ | ------- | ------------------------------ |
| SnowflakeGenerator | 243–244 | \~4.1 million                  |
| SnowflakeParallel  | 310–316 | \~3.1 million per goroutine    |

### ✅ Notes
* Each ID is 64-bit, unique, and sortable based on timestamp.

* Easily exceeds the target of 10,000+ IDs/sec, achieving millions/sec.
