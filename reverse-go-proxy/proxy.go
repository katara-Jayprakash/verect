/**
 * 1.  we are going to given an url = jayprakash.katara.vercel.com
 * 2. extract the jaypraksh which is subdomain of it
 * 3. and now add the keys values in it, and match it with s3 bucket list
 * create proxy and send the request to the proxy
 */

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type Response struct {
	Message string `json:"message"`
	Method  string `json:"method"`
	Host    string `json:"host"`
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.Method)
	fmt.Println(r.URL.String())
	fmt.Fprintf(w, "server got your message")

}
func health(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "health route is ready")
}
func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		response := Response{
			Message: "hi got your message thanx client",
			Method:  r.Method,
			Host:    r.Host,
		}
		json.NewEncoder(w).Encode(response)

	}
}
func main() {
	Port := ":80"
	mux := http.NewServeMux()

	fmt.Println("server is running on Port:", Port)
	mux.HandleFunc("/", handleRequest)
	mux.HandleFunc("/health", health)
	mux.HandleFunc("/api", handler)

	log.Fatal(http.ListenAndServe(Port, mux))

}
