# Distributed Unique ID Generator (Go + Fiber + gRPC)

A high-performance, distributed 64-bit unique ID generator using the **Twitter Snowflake** algorithm, implemented in **Go** with **Clean Architecture** principles. It provides both **HTTP (Fiber)** and **gRPC** APIs.

---

![Snowflake ID Bit Layout Diagram]([./doc/0001.jpg](https://github.com/YslamB/go-uuid-generator/blob/main/doc/0001.png))
*Figure: Bit layout of the 64-bit Snowflake ID*

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
