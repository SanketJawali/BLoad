# BLoad — Load Balancer in Go
BLoad is a simple HTTP load balancer implemented in Go. It distributes incoming HTTP requests across multiple backend servers, providing basic load balancing and failover capabilities.

## Aim for building this project

- Build a Load Balancer(LB) which distributes the load to multiple servers
- Handle fail-overs of the servers
- Get good at programming in Go

## Benchmarking
The below command uses ApacheBench to send 5000 requests with a concurrency level of 500 to the load balancer running on localhost at port 8000.
```bash
ab -n 5000 -c 500 http://localhost:8000/
```

> TL;DR: The load balancer can handle 5000 requests with a concurrency level of 500 in 27.773 seconds, with an average of 180.03 requests per second and no failed requests.
> **After optimizations** and reducing the concurrency setting the concurrency to 100, it can handle the same number of requests in just 1.415 seconds, with an average of 3532.81 requests per second.

The main reason of reducing the concurrency level from 500 to 100 is to avoid overwhelming the backend servers. With a concurrency level of 500, the load balancer will send too many requests simultaneously, with possible burst increases in concurrency, causing increased latency and slower response times.

When benchmarked on single backend server, the RPS (Resuests Per Second) is around 6200, on local machine. This is much higher than the with LB in between. This is because the LB adds an additional layer of processing, which can introduce some overhead. Additionally, the LB may not be optimized for high concurrency levels, leading to increased latency and reduced performance.

First Benchmarking results:
```
Concurrency Level:      500
Time taken for tests:   27.773 seconds
Complete requests:      5000
Failed requests:        0
Total transferred:      11790000 bytes
HTML transferred:       10855000 bytes
Requests per second:    180.03 [#/sec] (mean)
Time per request:       2777.341 [ms] (mean)
Time per request:       5.555 [ms] (mean, across all concurrent requests)
Transfer rate:          414.56 [Kbytes/sec] received

Connection Times (ms)
              min  mean[+/-sd] median   max
Connect:        0    1   3.2      0      16
Processing:     0  276 1275.6      2   27711
Waiting:        0  276 1275.6      2   27710
Total:          0  278 1276.6      2   27712

Percentage of the requests served within a certain time (ms)
  50%      2
  66%      3
  75%      4
  80%      5
  90%     53
  95%   1454
  98%   3442
  99%   5686
 100%  27712 (longest request)
```

Final Benchmarking results:
```

 ab -n 5000 -c 100 http://localhost:8000/
Concurrency Level:      100
Time taken for tests:   1.415 seconds
Complete requests:      5000
Failed requests:        0
Total transferred:      11790000 bytes
HTML transferred:       10855000 bytes
Requests per second:    3532.81 [#/sec] (mean)
Time per request:       28.306 [ms] (mean)
Time per request:       0.283 [ms] (mean, across all concurrent requests)
Transfer rate:          8135.13 [Kbytes/sec] received

Connection Times (ms)
              min  mean[+/-sd] median   max
Connect:        0    1   0.6      1       3
Processing:     1   15 102.6      5    1057
Waiting:        0   15 102.6      4    1057
Total:          1   16 102.5      6    1057

Percentage of the requests served within a certain time (ms)
  50%      6
  66%      6
  75%      7
  80%      8
  90%      9
  95%     11
  98%     14
  99%   1008
 100%   1057 (longest request)
