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
	"net/http"
	"sync"
)

var mu sync.RWMutex
var Port string = ":3000"
var database map[int]task

type task struct {
	Id        int    `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tasklist := []task{}
		mu.RLock() //locking for safety purpose
		for _, t := range database {
			tasklist = append(tasklist, t)
		}
		mu.RUnlock() //unlocking thread
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tasklist)
	}

	if r.Method == http.MethodPost {
		nextId := len(database) + 1
		var newTask task
		json.NewDecoder(r.Body).Decode(&newTask)
		mu.Lock()
		newTask.Id = nextId
		database[nextId] = newTask

		mu.Unlock()
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("server sent you message!"))
	}

}
func main() {
	database = make(map[int]task)

	// http server
	http.HandleFunc("/", handleRequest)
	fmt.Println("server is running on port" + Port)

	http.ListenAndServe(Port, nil)

}
