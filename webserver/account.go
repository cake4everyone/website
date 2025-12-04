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
		if user.WhitelistEntry != nil {
			if err := database.DB.Model(user.WhitelistEntry).Preload("Markers").Preload("Markers.Users").Find(&user.WhitelistEntry).Error; err != nil {
				log.Printf("Failed to load markers: %v", err)
			}
		}
		ServeTemplate(w, r, user, "account", "account/nicknames", "account/marker")
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
