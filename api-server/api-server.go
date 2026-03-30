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
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var kubeClient *kubernetes.Clientset

func healthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "health route is ready")
}

func getconfig() (*rest.Config, error) {
	// if kubeConfigBase exist then decode its and use it;
	if kubeconfigDecode := os.Getenv("KUBECONFIG_BASE64"); kubeconfigDecode != "" {
		kubeconfigBytes, err := base64.StdEncoding.DecodeString(kubeconfigDecode)
		if err != nil {
			return nil, fmt.Errorf("failed to decode kubeconfig: %w", err)
		}
		// This is used when your kubeconfig is in memory (byte array), not a file.
		kubeConfig, err := clientcmd.RESTConfigFromKubeConfig(kubeconfigBytes)
		if err != nil {
			return nil, fmt.Errorf("cannot create kube config from Base64: %w", err)
		}

		log.Println("Using kubeconfig from Base64")
		return kubeConfig, nil
	}
	//in-cluster config (AKS, EKS, GKE)
	kubeConfig, err := rest.InClusterConfig()
	if err == nil {
		log.Println("Using in-cluster kubeconfig")
		return kubeConfig, nil
	}
	// in local developement;
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("cannot find home dir: %w", err)
	}
	kubeConfigPath := filepath.Join(homeDir, ".kube", "config")
	kubeConfig, err = clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		return nil, fmt.Errorf("cannot find kubeConfigdir: %w", err)
	}
	log.Println("Using local kubeconfig file")
	return kubeConfig, nil
}
func init() {
	// load variable from the .env into the system
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file founded shift to local .kube/config")
	}
	config, err := getconfig()
	if err != nil {
		log.Fatal("could not get config", err)
	}
	kubeClient, err = kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal("could not create k8s client", err)
	}

}
func main() {
	Port, exists := os.LookupEnv("PORT")
	if !exists {
		Port = ":9000"
	}
	log.Println("k8s client connected successfully!")

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	// mux.HandleFunc("/")

	http.ListenAndServe(Port, mux)
}
