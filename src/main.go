package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"sync/atomic"
)

// Defining baseUrl
var baseUrl = "http://localhost"

// Setup for the server queue
// This array contains the list of ports the servers are running at
// var servers = []int{5000, 6969, 7070}
var servers = []int{4321, 4321, 4321}

// A number which will keep incrementing
// finding the server involves getting the mod of this int
// and retreving the server address from the servers array
var availableServer uint64 = 0

func main() {

	router := http.NewServeMux()

	router.HandleFunc("/", requestHandler)

	log.Println("Starting server at port 8000.")
	err := http.ListenAndServe(":8000", router)

	if err != nil {
		log.Fatal("Error occured: ", err)
		panic("Server Crashed")
	}
}

func requestHandler(w http.ResponseWriter, r *http.Request) {
	// Get the server to forward the request to
	port := servers[atomic.AddUint64(&availableServer, 1)%uint64(len(servers))]
	log.Println("Request routed to server at port: ", port)

	// Construct the url for the request to backend server
	url := fmt.Sprintf("%v:%v%v", baseUrl, port, r.URL.Path)

	// Make the request to the backend server
	serverRes, err := http.Get(url)
	if err != nil {
		log.Fatalln("Error occured while making GET request to server, at port: ", port)
	}
	defer serverRes.Body.Close()

	// Make sure to get all the response
	body, err := io.ReadAll(serverRes.Body)

	// Copy headers from the backend response to the client response
	for key, values := range serverRes.Header {
		// Skip content length header as it will be set automatically
		// by the http package when writing the response
		// Adding content length header manually can cause issues with chunked transfer encoding
		if key == "Content-Length" {
			continue
		}
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(serverRes.StatusCode)

	// Return the response back to the client
	fmt.Fprintln(w, string(body))
}
