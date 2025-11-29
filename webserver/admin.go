package webserver

import (
	"encoding/json"
	"log"
	"net/http"
	"website/auth"
	"website/database"
)

func handleAdminPage(w http.ResponseWriter, r *http.Request) {
	user, abort := auth.GetUser(w, r)
	if abort {
		return
	}
	if !user.WhitelistEntry.IsAdmin() {
		http.Redirect(w, r, "/account", http.StatusSeeOther)
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
	ServeTemplate(w, r, admin, "admin", "account/whitelistEntry")
}

func handleAdmin(w http.ResponseWriter, r *http.Request) {
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
