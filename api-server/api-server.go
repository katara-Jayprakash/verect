package main

import (
	"fmt"
	"log"
	"net/http"
)

func handleRequest(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("hello world")
}
func main() {
	Port := ":3000"
	mux := http.NewServeMux()
	mux.HandleFunc("/hello world", handleRequest)
	log.Fatal(http.ListenAndServe(Port, mux))
}
