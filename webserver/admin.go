package webserver

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"website/auth"
	"website/database"
	"website/types"
)

const (
	QUERY_NEW_ENTRY = "new_entry"
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
	if r.URL.Query().Has(QUERY_NEW_ENTRY) {
		ServeTemplate(w, r, user, "admin/newEntry")
		return
	}
	switch r.URL.Query().Get("edit") {
	case QUERY_NEW_ENTRY:
		ServeTemplate(w, r, user, "admin/edit/newEntry")
	default:
		admin := struct {
			database.User
			WhitelistEntries []database.WhitelistEntry
		}{user, []database.WhitelistEntry{}}
		if err := database.DB.Model(admin.WhitelistEntries).Preload("Reference").Order("ID").Find(&admin.WhitelistEntries).Error; err != nil {
			log.Printf("Failed to load all whitelist entries: %v", err)
			http.Error(w, "Loading whitelist failed!", http.StatusInternalServerError)
			return
		}
		ServeTemplate(w, r, admin, "admin", "account/whitelistEntry", "account/whitelistEntryCheckbox", "admin/newEntry")
	}
}

func handleAdmin(w http.ResponseWriter, r *http.Request) {
	user, abort := auth.GetUser(w, r)
	if abort {
		return
	}
	if !user.WhitelistEntry.IsAdmin() {
		http.Redirect(w, r, "/account", http.StatusSeeOther)
		return
	}

	editMode := r.URL.Query().Get("edit")
	if editMode == "" {
		http.Error(w, "no edit mode", http.StatusBadRequest)
		return
	}

	switch editMode {
	case QUERY_NEW_ENTRY:
		var newUserInput struct {
			Name        string
			Nicknames   string
			Minecraft   string
			ReferenceID string `json:"reference_id"`
			Active      string
		}
		err := json.NewDecoder(r.Body).Decode(&newUserInput)
		if err != nil {
			log.Printf("Failed to decode new user: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		log.Printf("new user: %+v", newUserInput)

		referenceID, err := strconv.ParseUint(newUserInput.ReferenceID, 10, 64)
		if err != nil {
			http.Error(w, "Invalid reference ID", http.StatusBadRequest)
			return
		}
		var refuser database.User
		refuser.ID = uint(referenceID)
		if err = database.DB.Find(&refuser).Error; err != nil {
			log.Printf("Failed to get reference user %d: %+v", referenceID, err)
			http.Error(w, "reference user does not exist", http.StatusBadRequest)
			return
		}

		minecraft, err := types.GetMinecraftProfileByName(newUserInput.Minecraft)
		if err != nil {
			log.Printf("New whitelist entry failed: get mc uuid: %v", err)
			http.Error(w, "Failed to resolve Minecraft Name", http.StatusInternalServerError)
			return
		}

		_, err = database.NewUser(newUserInput.Name, minecraft.UUID, uint(referenceID), newUserInput.Active == "true")
		if err != nil {
			log.Printf("Creating new user failed: %v", err)
			http.Error(w, "Creating new user failed", http.StatusInternalServerError)
			return
		}
		ServeTemplate(w, r, nil, "admin/newEntry")
		return
	default:
		http.Error(w, "unknown edit mode: "+editMode, http.StatusBadRequest)
		return
	}
}
