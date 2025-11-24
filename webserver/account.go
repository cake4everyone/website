package webserver

import (
	"encoding/json"
	"net/http"
	"path"
	"website/auth"
	"website/database"
)

func handleAccountPage(w http.ResponseWriter, r *http.Request) {
	user, abort := auth.GetUser(w, r)
	if abort {
		return
	}
	editMode := r.URL.Query().Get("edit")
	name := path.Join("account", editMode)
	ServeTemplate(w, r, name, user)
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
	case "nicknames":
		var nicknames database.Nicknames
		err := json.NewDecoder(r.Body).Decode(&nicknames)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		user.WhitelistEntry.Nicknames.Set(nicknames)
		database.DB.Save(&user.WhitelistEntry)
		w.WriteHeader(http.StatusOK)
		return
	default:
		http.Error(w, "unknown edit mode: "+editMode, http.StatusBadRequest)
		return
	}
}
