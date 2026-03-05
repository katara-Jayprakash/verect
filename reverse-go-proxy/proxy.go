/**
 * 1.  we are going to given an url = jayprakash.katara.vercel.com
 * 2. extract the jaypraksh which is subdomain of it
 * 3. and now add the keys values in it, and match it with s3 bucket list
 * create proxy and send the request to the proxy
 */

package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

func health(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "health route is ready")
}
func proxyServer(w http.ResponseWriter, r *http.Request) {
	hostname := r.Host
	subdomain := strings.Split(hostname, ".")[0]
	fmt.Fprint(w, subdomain)

}

func main() {
	Port := ":80"
	mux := http.NewServeMux()

	fmt.Println("server is running on Port:", Port)
	mux.HandleFunc("/health", health)
	mux.HandleFunc("/", proxyServer)

	log.Fatal(http.ListenAndServe(Port, mux))

}
