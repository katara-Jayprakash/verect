/**
 * 1.  we are going to given an url = jayprakash.katara.vercel.com
 * 2. extract the jaypraksh which is subdomain of it
 * 3. and now add the keys values in it, and match it with s3 bucket list
 * create proxy and send the request to the proxy
 */

package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	// "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/joho/godotenv"
)

var (
	baseUrl  string
	S3Client *s3.Client
)

func init() {
	// load variable from the .env into the system
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
	baseUrl = os.Getenv("B2_BASE_URL")

	endpoint := os.Getenv("END_POINT")
	region := os.Getenv("REGION")
	accessKeyId := os.Getenv("accessKeyId")
	secretAccessKey := os.Getenv("secretAccessKey")

	// Load configuration
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyId,
			secretAccessKey, "",
		)), config.WithRegion(region))
	if err != nil {
		log.Fatalf("Unable to load SDK config, %v", err)
	}
	// create s3 client
	S3Client = s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true
	})
}

// health check point
func health(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "health route is ready")
}

// proxy server handler code
func proxyServer(w http.ResponseWriter, r *http.Request) {
	// creating the context for timeout
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	hostname := r.Host
	subdomain := strings.Split(hostname, ".")[0]
	requestedPath := r.URL.Path
	if requestedPath == "/" {
		requestedPath = "/index.html"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	resolveTo := baseUrl + subdomain + requestedPath
	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	//
	resp, err := client.Get(resolveTo)

	// Handle Fallback
	if err != nil || (resp != nil && resp.StatusCode == http.StatusNotFound) {
		// If the first request actually opened a body, close it now!
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}

		fallbackUrl := baseUrl + subdomain + "/index.html"
		// change from this side
		//Get Object

		resp, err = http.Get(fallbackUrl)
		if err != nil {
			http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
			return
		}
	}
	defer resp.Body.Close()
	// Copy headers from bucket to client
	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	//Set the status code (e.g., 200, 404, etc.)
	w.WriteHeader(resp.StatusCode)
	// Stream the data in chunks
	io.Copy(w, resp.Body)
}

func main() {
	Port, exits := os.LookupEnv("PORT")
	if exits == false {
		Port = ":80"
	}

	mux := http.NewServeMux()

	fmt.Println("server is running on Port:", Port)
	mux.HandleFunc("/health", health)
	mux.HandleFunc("/", proxyServer)

	log.Fatal(http.ListenAndServe(Port, mux))

}
