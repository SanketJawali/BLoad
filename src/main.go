package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"

	"github.com/joho/godotenv"
)

var baseUrl string
var servers []string
var availableServer uint64 = 0

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: Error loading .env file")
	}

	servers = append(servers, strings.Split(os.Getenv("SERVERS"), ",")...)
	fmt.Printf("Server ports: %v\n", servers)
	baseUrl = os.Getenv("BASE_URL")

	router := http.NewServeMux()
	router.HandleFunc("/", requestHandler)

	fmt.Println("Starting server at port 8000.")
	err = http.ListenAndServe(":8000", router)
	if err != nil {
		log.Fatal("Server Crashed: ", err)
	}
}

func requestHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Safeguard against empty server lists
	// if len(servers) == 0 || servers[0] == "" {
	// 	http.Error(w, "No backend servers configured", http.StatusServiceUnavailable)
	// 	return
	// }

	port := servers[atomic.AddUint64(&availableServer, 1)%uint64(len(servers))]

	// 2. Properly construct the URL, including query parameters
	var targetUrl string
	if port == "" {
		targetUrl = fmt.Sprintf("%s%s", baseUrl, r.URL.Path)
	} else {
		targetUrl = fmt.Sprintf("%s:%s%s", baseUrl, port, r.URL.Path)
	}
	if r.URL.RawQuery != "" {
		targetUrl += "?" + r.URL.RawQuery
	}

	if len(targetUrl) > 50 {
		log.Printf("Routing to: %s (truncated)", targetUrl[:100])
	} else {
		log.Println("Routing request to: ", targetUrl)
	}

	// 3. Simplify request creation (no massive switch statement)
	// Using NewRequestWithContext is best practice so the request cancels if the client disconnects early
	proxyReq, err := http.NewRequestWithContext(r.Context(), r.Method, targetUrl, r.Body)
	if err != nil {
		log.Printf("Error creating proxy request: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError) // No more log.Fatal
		return
	}

	// 4. Copy original request headers to the proxy request
	for name, values := range r.Header {
		for _, value := range values {
			proxyReq.Header.Add(name, value)
		}
	}

	// 5. ACTUALLY execute the request using an HTTP client
	client := &http.Client{}
	resp, err := client.Do(proxyReq)
	if err != nil {
		log.Printf("Error reaching backend server at %s: %v", port, err)
		http.Error(w, "Bad Gateway", http.StatusBadGateway) // No more log.Fatal
		return
	}
	defer resp.Body.Close()

	// 6. Copy backend response headers back to the client response
	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	// 7. Write the exact status code returned by the backend
	w.WriteHeader(resp.StatusCode)

	// 8. Stream the body directly to avoid blowing up memory (no io.ReadAll)
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		log.Printf("Error streaming response body: %v", err)
	}
}
