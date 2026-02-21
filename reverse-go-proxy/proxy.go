package main

import (
	"fmt"
	"net/http"
)

func handleRequest(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hi there, i love being here, user: %v!", r.URL.User)
	fmt.Println("Debug:", r.URL.User)
	fmt.Fprintf(w, "something went up!", r.URL.Path)
}

func main() {
	Port := ":80"
	http.HandleFunc("/", handleRequest)
	http.ListenAndServe(Port, nil)
}
