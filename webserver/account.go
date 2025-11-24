package webserver

import (
	"encoding/json"
	"log"
	"net/http"
	"website/auth"
	"website/database"
)

const (
	QUERY_NICKNAMES = "nicknames"
)

func handleAccountPage(w http.ResponseWriter, r *http.Request) {
	user, abort := auth.GetUser(w, r)
	if abort {
		return
	}
	if r.URL.Query().Has(QUERY_NICKNAMES) {
		ServeTemplate(w, r, user.WhitelistEntry.Nicknames, "account/nicknames")
		return
	}
	switch r.URL.Query().Get("edit") {
	case QUERY_NICKNAMES:
		ServeTemplate(w, r, user, "account/edit/nicknames")
	default:
		if !user.WhitelistEntry.IsAdmin() {
			ServeTemplate(w, r, user, "account", "account/nicknames")
			return
		}
		admin := struct {
			database.User
			WhitelistEntries []database.WhitelistEntry
		}{user, []database.WhitelistEntry{}}
		if err := database.DB.Model(admin.WhitelistEntries).Preload("Reference").Order("ID").Find(&admin.WhitelistEntries).Error; err != nil {
			log.Printf("Failed to load all whitelist entries: %v", err)
			http.Error(w, "Loading whitelist failed!", http.StatusInternalServerError)
			return
		}
		log.Printf("loaded %d entries", len(admin.WhitelistEntries))
		ServeTemplate(w, r, admin, "account", "account/nicknames", "account/admin", "account/whitelistEntry")
	}
}

func handleAccount(w http.ResponseWriter, r *http.Request) {
	editMode := r.URL.Query().Get("edit")
	if editMode == "" {
		http.Error(w, "no edit mode", http.StatusBadRequest)
		return
	}

	user, abort := auth.GetUser(w, r)
	if abort {
		return
	}

	switch editMode {
	case QUERY_NICKNAMES:
		var nicknames database.Nicknames
		err := json.NewDecoder(r.Body).Decode(&nicknames)
		if err != nil {
			log.Printf("Failed to decode nicknames: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		user.WhitelistEntry.Nicknames.Set(nicknames)
		database.DB.Save(&user.WhitelistEntry)
		ServeTemplate(w, r, user.WhitelistEntry.Nicknames, "account/nicknames")
		return
	default:
		http.Error(w, "unknown edit mode: "+editMode, http.StatusBadRequest)
		return
	}
}
