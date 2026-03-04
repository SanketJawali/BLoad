package main

import (
	"fmt"
	"io"
	"log"
	"maps"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strings"
	"sync/atomic"

	"github.com/joho/godotenv"
)

type LoadBalancer struct {
	client          *http.Client
	servers         []string
	availableServer uint64
	tokenBucket     map[string](chan int8)
	bucketSize      int
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: Error loading .env file")
	}

	// Create server list from the env
	var serverList []string
	serverList = append(serverList, strings.Split(os.Getenv("SERVERS"), ",")...)
	fmt.Printf("Server added to LB: %v\n", serverList)

	// Creating token buckets for each server
	bucketSize := 6
	tokenBuckets := make(map[string](chan int8))

	for _, server := range serverList {
		// Creating a token bucket for each server with a capacity of n(bucketSize) tokens
		tokenBuckets[server] = make(chan int8, bucketSize)
	}

	// Completely fill the token buckets for each server
	for _, bucket := range tokenBuckets {
		for range bucketSize {
			// Add a token to the bucket
			// we can use any value since we're just counting tokens by the number of items in the channel
			bucket <- 1
		}
	}

	// Initialize the Transport for the HTTP client with connection pooling and keep-alives
	transport := &http.Transport{
		// Adjust based on expected load and server capacity
		// Allow enough idle connections for each server
		MaxIdleConnsPerHost: len(serverList) * bucketSize,
	}

	// Initialize the load balancer with the server list and an HTTP client
	lb := LoadBalancer{
		client:          &http.Client{Transport: transport},
		servers:         serverList,
		availableServer: 0,
		tokenBucket:     tokenBuckets,
	}

	// Setup observability with pprof for profiling and debugging
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	router := http.NewServeMux()
	router.HandleFunc("/", lb.requestHandler)

	fmt.Println("Starting server at port 8000.")
	err = http.ListenAndServe(":8000", router)
	if err != nil {
		log.Fatal("Server Crashed: ", err)
	}
}

func (lb *LoadBalancer) requestHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Get the server port using round-robin load balancing algorithm
	allocatedServer := lb.servers[atomic.AddUint64(&lb.availableServer, 1)%uint64(len(lb.servers))]

	// 2. Properly construct the URL, including query parameters
	// The old url creation logic was removed. It was very inefficient
	// creating a new URL for each request is unnecessary
	// we can just concatenate the base server URL with the request path and query parameters.
	//
	// The old method used to use a base URL, with port number, and then add the URL path and query params to it.
	// Get the query parameters and append them to the target URL if they exist
	var targetUrl string

	if r.URL.RawQuery != "" {
		targetUrl = fmt.Sprintf("%s%s?%s", allocatedServer, r.URL.Path, r.URL.RawQuery)
	} else {
		targetUrl = fmt.Sprintf("%s%s", allocatedServer, r.URL.Path)
	}

	// Trauncate very long URLs, log the server we're routing to
	if len(targetUrl) > 80 {
		log.Printf("Routing request to: %s (truncated)\n", targetUrl[:80])
	} else {
		log.Println("Routing request to: ", targetUrl)
	}

	// 3. Simplify request creation
	// Using NewRequestWithContext is best practice so the request cancels if the client disconnects early
	proxyReq, err := http.NewRequestWithContext(r.Context(), r.Method, targetUrl, r.Body)
	if err != nil {
		log.Printf("Error creating proxy request: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// 4. Copy original request headers to the proxy request
	// maps.Copy is a convenient way to copy all headers without needing to loop through them manually
	// manually copying headers with .Header.Add does more things behind the scenes, like checking for correct header formatting
	// which is unnecessary here since we're just copying them as-is,
	// so using maps.Copy is more efficient and less error-prone.
	maps.Copy(proxyReq.Header, r.Header)

	// 5. Actually execute the request using an HTTP client
	// NOTE: Each request consumes a token from the bucket of the allocated server.
	// If the bucket is empty, it will block until a token is available.
	// Stays blocked until a token is available.

	<-lb.tokenBucket[allocatedServer]
	defer func() {
		// Return the token to the bucket after the request is done
		lb.tokenBucket[allocatedServer] <- 1
	}()

	resp, err := lb.client.Do(proxyReq)
	if err != nil {
		log.Printf("Error reaching backend server at '%s': %v", allocatedServer, err)
		http.Error(w, "Bad Gateway", http.StatusBadGateway) // No more log.Fatal
		return
	}
	defer resp.Body.Close()

	// 6. Copy backend response headers back to the client response
	// Refer to point 4 for why maps.Copy is used here as well
	maps.Copy(w.Header(), resp.Header)

	// 7. Write the exact status code returned by the backend
	w.WriteHeader(resp.StatusCode)

	// 8. Stream the body directly to avoid blowing up memory
	// Copying the response body directly prevents loading the entire response into memory
	// which is crucial for large responses
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		log.Printf("Error streaming response body: %v", err)
	}
}
