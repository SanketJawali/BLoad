# BLoad — Load Balancer in Go
BLoad is a simple HTTP load balancer implemented in Go. It distributes incoming HTTP requests across multiple backend servers, providing basic load balancing and failover capabilities.

## Aim for building this project

- Build a Load Balancer(LB) which distributes the load to multiple servers
- Handle fail-overs of the servers
- Get good at programming in Go

## Implementation concerns

### 1. Core Architecture & Network Flow

- **The Reverse Proxy Identity:** An HTTP load balancer is a reverse proxy distributing traffic across multiple backends.
- **The Client Illusion:** Clients only know the LB's IP/Port. Backend servers remain completely hidden and isolated.
- **TCP Connection Pooling:** The LB maintains long-lived, warm TCP connections to backend servers. It multiplexes incoming client HTTP requests over these existing connections to avoid the heavy latency of constant 3-way handshakes.

### 2. Concurrency Model (Go-Specific)

- **Thread-Per-Request is Dead:** Spawning an OS thread per connection exhausts RAM immediately. Modern systems use async I/O (`epoll`/`kqueue`) to handle thousands of sockets on a few OS threads.
- **Go's Native Advantage:** Go handles network concurrency for you. The `net/http` package automatically spawns a lightweight goroutine for every incoming request.
- **No Manual Spawning:** You do not manually spawn goroutines inside your HTTP handler. The handler *is* the concurrent execution context.

### 3. Routing Strategy

- **The Dumb Pipe:** Do not duplicate backend API routes on the LB.
- **Catch-All Listener:** Bind the LB to a root/wildcard route (`/`). Extract the incoming path and query parameters, select a backend target, stitch the URL together, and forward the raw HTTP payload.

### 4. State Management & Race Conditions

- **The Mutex Bottleneck:** Putting a standard Mutex lock over your routing algorithm (like a Round Robin index) forces 10,000 concurrent handlers into a single-file line. This tanks throughput.
- **Lock-Free Math:** To increment an index safely across thousands of threads, use hardware-level atomic operations (like Go's `sync/atomic` package) instead of blocking locks.
- **Writer Starvation:** If thousands of readers hold locks on a queue, a single writer (your health manager) will be starved and fail to update the server list.
- **Atomic Swaps:** Never mutate an array that thousands of threads are currently reading. Build a new array in the background, then perform an atomic pointer swap to seamlessly redirect new traffic.

### 5. Health Checking & Fault Tolerance

- **Passive vs. Active:** Production requires both. Active checks (background pings) catch quiet failures. Passive checks (TCP connection drops in the handler) act as instant circuit breakers.
- **The Thundering Herd:** If a backend dies, 1,000 handlers will fail simultaneously and trigger a flood of failure signals.
- **The Manager Actor:** Use a dedicated background goroutine (Manager) listening to a channel to handle health state changes, but you must implement signal de-duplication so it doesn't process 1,000 redundant health checks for the same dead server.
