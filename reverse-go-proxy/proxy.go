/**
1. user → jayprakash.myvercel.in/assets/main.js
2. extract subdomain → jayprakash
3. build object path → __outputs/jayprakash/assets/main.js
4. fetch from storage
5. stream to client
*/

package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
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
	baseUrl    string
	bucketName string
	S3Client   *s3.Client
)

func init() {
	// load variable from the .env into the system
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
	baseUrl = os.Getenv("B2_BASE_URL")
	bucketName = os.Getenv("BUCKET_NAME")

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
		o.BaseEndpoint = aws.String(endpoint) // Backblaze endpoint
		o.UsePathStyle = true
	})
}

// health check point
func health(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "reverse proxy server health route is ready")
}

// proxy server handler code
func proxyServer(w http.ResponseWriter, r *http.Request) {
	// Extract subdomain from hostname
	hostname := r.Host
	parts := strings.Split(hostname, ".")
	if len(parts) < 3 {
		http.Error(w, "Invalid domain", http.StatusBadRequest)
		return
	}
	subdomain := parts[0]
	requestedPath := path.Clean(r.URL.Path)
	if requestedPath == "/" {
		requestedPath = "/index.html"
	}

	resolveToKey := "__outputs/" + subdomain + requestedPath
	//  Create the request for the S3 object
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(resolveToKey),
	}
	// creating the context for timeout
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Try to get object
	resp, err := S3Client.GetObject(ctx, input)
	// handle handing error
	if err != nil {
		fallbackKey := "__outputs/" + subdomain + "/index.html"

		// If the first one failed, we try the index.html
		resp, err = S3Client.GetObject(ctx, &s3.GetObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(fallbackKey),
		})

		if err != nil {
			http.Error(w, "Site Not Found", http.StatusNotFound)
			return
		}
	}

	defer resp.Body.Close()
	// Setting Headers (S3 SDK specific way)
	if resp.ContentType != nil {
		w.Header().Set("Content-Type", *resp.ContentType)
	}
	// If the code reached here, it's a 200 OK.
	w.WriteHeader(http.StatusOK)
	// Stream the data in chunks
	io.Copy(w, resp.Body)
}

func main() {
	Port, exits := os.LookupEnv("PORT")
	if exits == false {
		Port = ":80"
	}

	mux := http.NewServeMux()

	fmt.Println("reverse proxy server is running on Port:", Port)
	mux.HandleFunc("/health", health)
	mux.HandleFunc("/", proxyServer)

	log.Fatal(http.ListenAndServe(Port, mux))

}
