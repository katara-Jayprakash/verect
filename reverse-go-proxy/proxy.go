/**
 * 1.  we are going to given an url = jayprakash.katara.vercel.com
 * 2. extract the jaypraksh which is subdomain of it
 * 3. and now add the keys values in it, and match it with s3 bucket list
 * create proxy and send the request to the proxy
 */

package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

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

	// Create AWS session
	s3Config := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(accessKeyId, secretAccessKey, ""),
		Endpoint:         aws.String(endpoint),
		Region:           aws.String(region),
		S3ForcePathStyle: aws.Bool(true),
	}
	newSession, err := session.NewSession(s3Config)
	S3Client = s3.New(newSession)

	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}

}

// health check point
func health(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "health route is ready")
}

// proxy server handler code
func proxyServer(w http.ResponseWriter, r *http.Request) {
	hostname := r.Host
	subdomain := strings.Split(hostname, ".")[0]
	requestedPath := r.URL.Path
	if requestedPath == "/" {
		requestedPath = "/index.html"
	}

	resolveTo := baseUrl + subdomain + requestedPath
	client := &http.Client{
		Timeout: 15 * time.Second,
	}
	resp, err := client.Get(resolveTo)

	// Handle Fallback
	if err != nil || (resp != nil && resp.StatusCode == http.StatusNotFound) {
		// If the first request actually opened a body, close it now!
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}

		fallbackUrl := baseUrl + subdomain + "/index.html"
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
