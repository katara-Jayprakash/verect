package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "health route is ready")
}
func init() {
	// load variable from the .env into the system
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
	if kubeConfigBase64 := os.Getenv("KUBECONFIG_BASE64"); kubeConfigBase64 != "" {
		// if kubeConfigBase exist then decode its and use it;
		// and if its does not exist then try to find out locally config with /.kube/config
		data := make([]byte, base64.StdEncoding.DecodedLen(len(kubeConfigBase64)))
		n, err := base64.StdEncoding.Decode(data, []byte(kubeConfigBase64))

		// fmt.Println(kubeConfigBase64)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(string(data[:n]))

	}

}
func main() {
	Port := ":9000"

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	http.ListenAndServe(Port, mux)
}
