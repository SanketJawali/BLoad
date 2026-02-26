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
var servers = []int{5000, 6969, 7070}

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
	url := fmt.Sprintf("%v:%v", baseUrl, port)
	log.Println("Request Host", r.Host)
	log.Println(url)

	// Make the request to the backend server
	res, err := http.Get(url)
	if err != nil {
		log.Fatalln("Error occured while making GET request to server, at port: ", port)
	}
	defer res.Body.Close()

	// Make sure to get all the response
	body, err := io.ReadAll(res.Body)

	log.Printf("Made a GET request to port: %v\nResponse: %v", port, string(body))

	// Return the response back to the client
	fmt.Fprintln(w, string(body))
}
