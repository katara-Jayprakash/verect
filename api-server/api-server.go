package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/moby/moby/pkg/namesgenerator"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	kubeClient *kubernetes.Clientset
)

type RequestData struct {
	GithubUrl string `json:"githubUrl"`
	// ProjectId string `json:"projectId"`
}

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

func getProjectSlug() (string, error) {
	// generating human readable slug from namegenerator
	rawSlug := namesgenerator.GetRandomName(0)
	generatorSlug := strings.ReplaceAll(rawSlug, "_", "-")

	// Get first 4 characters of UUID (without dashes)
	uuidSlug := strings.ReplaceAll(uuid.New().String(), "-", "")[:4]

	// Combine to create slug like "names generator + uuid"
	slug := generatorSlug + "-" + uuidSlug
	return slug, nil
}
func DeployProject(w http.ResponseWriter, r *http.Request) {
	// check its post request or not
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
	// converting Json into struct
	var data RequestData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	// projectId := data.ProjectId
	githubId := data.GithubUrl
	fmt.Println(githubId)

	ProjectId, err := getProjectSlug()
	if err != nil {
		http.Error(w, "Something wrong with creating slug", http.StatusInternalServerError)
		return
	}
	fmt.Println(ProjectId)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Data received successfully"))

}
func main() {
	Port, exists := os.LookupEnv("PORT")
	if !exists {
		Port = ":9000"
	}
	log.Println("k8s client connected successfully!")

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/project", DeployProject)

	if err := http.ListenAndServe(":"+Port, mux); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
