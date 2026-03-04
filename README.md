# BLoad — Load Balancer in Go
This project implements a simple HTTP load balancer written in Go.
It was originally built to learn Go's `net/http` package, concurrency patterns, and performance behavior under load.

During benchmarking, the project evolved into a deeper exploration of **queueing behavior, backend saturation, and load shaping**.

---

# Benchmarking

Benchmarks were performed using **ApacheBench (`ab`)**.

Example command:

```bash
ab -n 5000 -c 500 http://localhost:8000/
```

Where:

* `-n` = total number of requests
* `-c` = concurrency level (simultaneous requests)

All tests were run on **localhost**.

Three systems were benchmarked:

1. **Single Python HTTP server**
2. **Load Balancer without concurrency limiting**
3. **Load Balancer with backend concurrency limiting**

---

# Test Architecture

```
Client (ApacheBench)
        │
        ▼
   Go Load Balancer
        │
        ▼
  Python Backend Servers
```

10 backend servers were used in most LB tests.

---

# Benchmark Results Summary

## Python Server

| Metric           | Value             |
| ---------------- | ----------------- |
| Runs             | 9 / 18 successful |
| Max load handled | 6,500 requests    |
| Max concurrency  | 1,020             |
| Avg throughput   | **6,052 req/s**   |
| Peak throughput  | **6,421 req/s**   |
| Avg mean latency | 99 ms             |
| p50 latency      | 1 ms              |
| p99 latency      | 1 ms              |
| p100 latency     | 55 ms             |
| Transfer rate    | 13,937 KB/s       |

### Observations

* Extremely low latency under moderate load.
* **Fails quickly when saturated**.
* Has a **hard capacity cliff**.

---

## Load Balancer (with concurrency limiting)

| Metric           | Value                  |
| ---------------- | ---------------------- |
| Runs             | 19 / 25 successful     |
| Max load handled | **1,000,000 requests** |
| Max concurrency  | **1,020**              |
| Avg throughput   | **10,364 req/s**       |
| Peak throughput  | **11,603 req/s**       |
| Avg mean latency | 64 ms                  |
| p50 latency      | 62 ms                  |
| p99 latency      | 84 ms                  |
| p100 latency     | 113 ms                 |
| Transfer rate    | **23,867 KB/s**        |

### Observations

* Highest throughput of all configurations.
* Stable latency distribution.
* Handles **very large workloads without collapse**.
* Backend overload is prevented using concurrency limiting.

---

## Load Balancer (no rate limit)

| Metric           | Value            |
| ---------------- | ---------------- |
| Runs             | 7 / 7 successful |
| Max load handled | 10,000 requests  |
| Max concurrency  | 1,000            |
| Avg throughput   | **1,546 req/s**  |
| Peak throughput  | 4,458 req/s      |
| Avg mean latency | **2538 ms**      |
| p50 latency      | 7.6 ms           |
| p99 latency      | **4153 ms**      |
| p100 latency     | **14,763 ms**    |
| Transfer rate    | 3,560 KB/s       |

### Observations

* Extremely poor tail latency.
* Queue collapse occurs under load.
* Backend servers become saturated.
* Throughput drops drastically.

---

# Key Findings

## 1. More Concurrency Does Not Mean Higher Throughput

A common misconception is that increasing concurrency increases throughput.

In practice, excessive concurrency causes:

* backend saturation
* longer service times
* queue buildup
* exploding latency

This ultimately **reduces throughput**.

This phenomenon is often called **queue collapse**.

---

## 2. Backend Servers Have an Optimal Concurrency Window

Testing revealed that each backend server performed best with roughly:

```
5–6 concurrent requests per backend
```

Beyond this point:

* response time increases
* requests pile up
* throughput drops

The load balancer prevents this by limiting concurrency.

---

## 3. Load Balancing Is More Than Request Distribution

A naive load balancer simply forwards requests.

A good load balancer:

* limits backend concurrency
* shapes traffic
* prevents overload
* stabilizes latency

This project implements **load shaping via concurrency control**.

---

## 4. Tail Latency Matters More Than Median Latency

Without concurrency control:

```
p50 latency: ~7 ms
p99 latency: ~4000 ms
p100 latency: ~15,000 ms
```

This means a small percentage of requests experience **extreme delays**.

Concurrency limiting dramatically reduces tail latency.

---

## 5. Controlled Systems Can Achieve Higher Throughput

Surprisingly, the load balancer achieved **higher throughput than the raw Python server**.

This happens because the load balancer prevents backend saturation.

By keeping the system within its optimal operating region, overall throughput increases.

---

# Why the Load Balancer Handles 1M Requests

The Python server fails after roughly **6500 requests under high concurrency**.

However, the load balancer successfully handled:

```
1,000,000 requests
```

This does **not** mean the backend servers could handle 1M requests instantly.

Instead:

* the load balancer **spreads the requests over time**
* concurrency limits prevent overload
* requests are processed gradually

Example:

```
Without limiter
1000 requests arrive instantly
→ backend overload
→ collapse
```

```
With limiter
1000 requests arrive
→ only 60 allowed at a time
→ rest wait briefly
→ system remains stable
```

This is known as **load shaping**.

---

# Key Concepts Demonstrated

This project demonstrates several important distributed systems concepts:

### Queueing behavior

Too many concurrent requests cause queue buildup and latency explosion.

---

### Backend saturation

Servers degrade before they fail.

Throughput drops as service time increases.

---

### Tail latency

High percentiles (p99 / p100) dominate user experience.

---

### Load shaping

Limiting concurrency stabilizes the system.

---

### Backpressure

The load balancer slows incoming traffic when backends are busy.

---

# Future Improvements

Possible extensions:

* adaptive concurrency limits
* least-connections load balancing
* health checks
* circuit breakers
* dynamic backend scaling

---

# Final Takeaway

A load balancer is not just a traffic router.

It is a **system stabilizer** that:

* prevents overload
* smooths traffic bursts
* protects backend services
* improves overall throughput.

---

# Author Notes

This project started as a small Go learning exercise but became a deep exploration of:

* concurrency control
* benchmarking
* system behavior under load
* latency distribution

The most important takeaway:

> **Throughput is maximized not by sending more requests, but by sending the right number of requests.**
