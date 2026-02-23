package main

import (
	"fmt"
	"log"
	"net/http"
)

// Setup for the server queue
// This array contains the list of ports the servers are running at
var servers = []int{5000, 6969, 7070}

// A number which will keep incrementing
// finding the server involves getting the mod of this int
// and retreving the server address from the servers array
var availableServer = 0

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
	port := servers[availableServer%len(servers)]
	log.Println("Request routed to server at port: ", port)
	availableServer++

	fmt.Fprintln(w, "Inside Request handler. Server port: ", port)
}
