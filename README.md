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

First Benchmarking results:
```
Server Software:        SimpleHTTP/0.6
Server Hostname:        localhost
Server Port:            8000

Document Path:          /
Document Length:        2171 bytes

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
