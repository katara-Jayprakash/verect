package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var kubeClient *kubernetes.Clientset

func healthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "health route is ready")
}
func init() {
	/**
	 *  load the .env file and if not present then switch to ~/.kube/config
	 */
	// load variable from the .env into the system
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file founded shift to local .kube/config")
	}

	// if kubeConfigBase exist then decode its and use it;
	if kubeConfigBase64 := os.Getenv("KUBECONFIG_BASE64"); kubeConfigBase64 != "" {
		container := make([]byte, base64.StdEncoding.DecodedLen(len(kubeConfigBase64)))
		n, err := base64.StdEncoding.Decode(container, []byte(kubeConfigBase64))
		if err != nil {
			log.Fatal("failed to decode kubeconfig:", err)
		}
		// This is used when your kubeconfig is in memory (byte array), not a file.
		kubeConfig, err := clientcmd.RESTConfigFromKubeConfig(container[:n])
		if err != nil {
			log.Fatal("cannot create k8s client :", err)
		}
		// creating a client for interecting with k8s
		kubeClient, err = kubernetes.NewForConfig(kubeConfig)
		if err != nil {
			log.Fatal("cannot create k8s client :", err)
		}

	} else {
		currDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatal("cannot find home dir:", err)
		}
		kubeConfigPath := filepath.Join(currDir, ".kube", "config")

		// Create the client configuration from the kubeconfig file (currently we are taking configuration from <file>).
		kubeConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
		if err != nil {
			fmt.Printf("Error encountered: %v\n", err)
			return
		}
		// creating a client for interecting with k8s
		kubeClient, err = kubernetes.NewForConfig(kubeConfig)
		if err != nil {
			log.Fatal("cannot create k8s client :", err)
		}
	}

}
func main() {
	Port := ":9000"

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	http.ListenAndServe(Port, mux)
}
