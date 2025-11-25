package webserver

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

var mutex sync.Mutex

var cache = make(map[string]string)

func handleAPINameLookup(w http.ResponseWriter, r *http.Request) {
	uuid := mux.Vars(r)["uuid"]
	if r.URL.Query().Has("force") {
		delete(cache, uuid)
	}
	if name, ok := cache[uuid]; ok {
		w.Write([]byte(name))
		return
	}
	mutex.Lock()
	defer mutex.Unlock()
	if name, ok := cache[uuid]; ok {
		w.Write([]byte(name))
		return
	}
	cache[uuid] = GetMCName(uuid)
	w.Write([]byte(cache[uuid]))
	time.Sleep(time.Second)
}

func GetMCName(uuid string) string {
	var tries int

retry:
	resp, err := http.Get(fmt.Sprintf("https://api.minecraftservices.com/minecraft/profile/lookup/%s", uuid))
	if err != nil {
		log.Printf("Failed to resolve MC name: %v", err)
		return "ERROR:?"
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		if tries > 5 {
			return "ERROR:" + resp.Status
		}
		tries++
		log.Printf("Retry %d: UUID '%s'", tries, uuid)
		time.Sleep(10 * time.Second)
		goto retry
	}
	if resp.StatusCode != http.StatusOK {
		return "ERROR:" + resp.Status
	}

	var data struct {
		Name string `json:"name"`
	}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		log.Printf("Failed to decode MC name response: %v", err)
		return "ERROR:?"
	}
	return data.Name
}
