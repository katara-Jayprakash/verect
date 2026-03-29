package main

import (
	"fmt"
	"net/http"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "health route is ready")
}
func main() {
	Port := ":9000"

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	http.ListenAndServe(Port, mux)
}
