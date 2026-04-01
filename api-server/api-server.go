package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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

func getProjectSlug() string {
	// generating human readable slug from namegenerator
	rawSlug := namesgenerator.GetRandomName(0)
	generatorSlug := strings.ReplaceAll(rawSlug, "_", "-")

	// Get first 4 characters of UUID (without dashes)
	uuidSlug := strings.ReplaceAll(uuid.New().String(), "-", "")[:4]

	// Combine to create slug like "names generator + uuid"
	slug := generatorSlug + "-" + uuidSlug
	return slug
}

func int32Ptr(i int32) *int32 {
	return &i
}

func getEnvOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func K8sJobDefination(ProjectId string, githubUrl string) *batchv1.Job {
	imagePullSecret := getEnvOrDefault("IMAGE_PULL_SECRET", "ghcr-secret")
	backendImage := getEnvOrDefault("BACKEND_IMAGE", "ghcr.io/sharma-jayprakash/verect-backend:latest")
	appSecretName := getEnvOrDefault("APP_SECRET_NAME", "verect-secrets")

	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: ProjectId,
		},
		Spec: batchv1.JobSpec{
			BackoffLimit:            int32Ptr(3),   // retry upto 3 times on failure
			TTLSecondsAfterFinished: int32Ptr(300), // clean up after 5 minutes
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					ImagePullSecrets: []corev1.LocalObjectReference{
						{Name: imagePullSecret},
					},
					Containers: []corev1.Container{
						{
							Name:  "verect-backend",
							Image: backendImage,
							Env: []corev1.EnvVar{
								{Name: "PROJECT_ID", Value: ProjectId},
								{Name: "GIT_REPOSITORY_URL", Value: githubUrl},
							},
							EnvFrom: []corev1.EnvFromSource{
								{
									SecretRef: &corev1.SecretEnvSource{
										LocalObjectReference: corev1.LocalObjectReference{Name: appSecretName},
									},
								},
							},
						},
					},
					RestartPolicy: corev1.RestartPolicyNever,
				},
			},
		},
	}
}

func DeployProject(w http.ResponseWriter, r *http.Request) {
	// check its post request or not
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	// converting Json into struct
	var data RequestData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	// checking github url is empty or not
	githubUrl := data.GithubUrl
	if githubUrl == "" {
		http.Error(w, "githubUrl is required", http.StatusBadRequest)
		return
	}
	// creating slug
	ProjectId := getProjectSlug()
	job := K8sJobDefination(ProjectId, githubUrl)
	namespace := getEnvOrDefault("JOB_NAMESPACE", "default")

	createdJob, err := kubeClient.BatchV1().Jobs(namespace).Create(context.Background(), job, metav1.CreateOptions{})
	if err != nil {
		log.Printf("failed to create job: %v", err)
		http.Error(w, "failed to create build job", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message":   "build job created",
		"projectId": ProjectId,
		"jobName":   createdJob.Name,
		"namespace": namespace,
	})

}
func main() {
	Port, exists := os.LookupEnv("PORT")
	if !exists {
		Port = "9000"
	}
	log.Println("k8s client connected successfully!")

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/project", DeployProject)

	if err := http.ListenAndServe(":"+Port, mux); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
