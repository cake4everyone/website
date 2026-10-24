package webserver

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"sync"
	"time"
	"website/auth"
	"website/database"

	"github.com/gorilla/mux"
)

var mutex sync.Mutex

var cache = make(map[string]string)

const (
	CACHE_FILE = "cache/mc_names.json"
)

func init() {
	openCache()
}

func openCache() {
	f, err := os.Open(CACHE_FILE)
	if os.IsNotExist(err) {
		return
	} else if err != nil {
		log.Printf("Failed to open cache: %v", err)
	}
	defer f.Close()

	mutex.Lock()
	defer mutex.Unlock()
	err = json.NewDecoder(f).Decode(&cache)
	if err != nil {
		log.Printf("Failed to decode cache: %v", err)
	}
}

func WriteCache() {
	var err error
	if err = os.MkdirAll(path.Dir(CACHE_FILE), 0770); err != nil {
		log.Printf("Failed to create cache directory: %v", err)
		return
	}
	f, err := os.Create(CACHE_FILE)
	if err != nil {
		log.Printf("Failed to create cache: %v", err)
		return
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "	")
	err = enc.Encode(cache)
	if err != nil {
		log.Printf("Failed to encode cache: %v", err)
	}
}

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

func handleAPISetActive(w http.ResponseWriter, r *http.Request) {
	user, abort := auth.GetUser(w, r)
	if abort {
		return
	}
	if !user.WhitelistEntry.IsAdmin() {
		http.Redirect(w, r, "/account", http.StatusSeeOther)
		return
	}

	bodyData, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("failed to read body: %+v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	bodyVal, err := url.ParseQuery(string(bodyData))
	if err != nil {
		log.Printf("failed to parse body: %+v", err)
		http.Error(w, "Invalid body!", http.StatusBadRequest)
		return
	}
	var body struct {
		ID     uint
		Active bool
	}
	if !bodyVal.Has("id") || !bodyVal.Has("active") {
		log.Printf("incomplete body: %s", bodyData)
		http.Error(w, "missing key(s)", http.StatusBadRequest)
		return
	}
	bodyID, err := strconv.ParseUint(bodyVal.Get("id"), 10, 64)
	if err != nil {
		log.Printf("invalid number: %s", bodyVal.Get("id"))
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	body.ID = uint(bodyID)
	body.Active, err = strconv.ParseBool(bodyVal.Get("active"))
	if err != nil {
		log.Printf("invalid bool: %s", bodyVal.Get("active"))
		http.Error(w, "invalid active", http.StatusBadRequest)
		return
	}
	body.ID = uint(bodyID)

	entry := database.WhitelistEntry{UserID: body.ID}
	if err = database.DB.Where(entry).Find(&entry).Error; err != nil {
		log.Printf("Failed to get whitelist entry by ID: %+v", err)
		http.Error(w, "Couldn't load whitelist!", http.StatusInternalServerError)
		return
	}
	if entry.IsActive() == body.Active {
		http.Error(w, "Unchanged", http.StatusPreconditionFailed)
		return
	}
	entry.SetActive(body.Active)
	if err = database.DB.Save(entry).Error; err != nil {
		log.Printf("Failed to save whitelist entry: %+v", err)
		http.Error(w, "Failed to update whitelist!", http.StatusInternalServerError)
		return
	}

	ServeTemplate(w, r, entry, "account/whitelistEntryCheckbox")
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
