package main

import (
	"fmt"
	"net/http"
)

func handleRequest(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("hello world")
}
func main() {
	Port := ":3000"
	http.ListenAndServe(Port, handleRequest)

}
