package webserver

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"website/auth"
	"website/database"
)

const (
	QUERY_NICKNAMES       = "nicknames"
	QUERY_MARKERS         = "markers"
	QUERY_MARKER_ADD      = "marker_add"
	QUERY_MARKER_USER_ADD = "add_user_to_marker"
)

func handleAccountPage(w http.ResponseWriter, r *http.Request) {
	user, abort := auth.GetUser(w, r)
	if abort {
		return
	}
	showAccountPage(w, r, user)
}

func showAccountPage(w http.ResponseWriter, r *http.Request, user database.User) {
	if r.URL.Query().Has(QUERY_NICKNAMES) {
		ServeTemplate(w, r, user.WhitelistEntry.Nicknames, "account/nicknames")
		return
	} else if r.URL.Query().Has(QUERY_MARKERS) {
		ServeTemplate(w, r, user.WhitelistEntry.Markers, "account/markers", "account/marker")
		return
	}

	switch editMode := r.URL.Query().Get("edit"); editMode {
	case QUERY_NICKNAMES:
		ServeTemplate(w, r, user, "account/edit/nicknames")
	case QUERY_MARKER_ADD:
		ServeTemplate(w, r, user, "account/edit/marker_add")
	case "":
		ServeTemplate(w, r, user, "account", "account/nicknames", "account/markers", "account/marker")
	default:
		log.Printf("unknown account edit: %s", editMode)
		http.Error(w, "Unknown Mode "+editMode, http.StatusBadRequest)
	}
}

func handleAccount(w http.ResponseWriter, r *http.Request) {
	user, abort := auth.GetUser(w, r)
	if abort {
		return
	}

	serveAccountEdit(w, r, user)
}

func serveAccountEdit(w http.ResponseWriter, r *http.Request, user database.User) {
	editMode := r.URL.Query().Get("edit")
	if editMode == "" {
		http.Error(w, "no edit mode", http.StatusBadRequest)
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
	case QUERY_MARKER_ADD:
		var newMarkerInput struct {
			X1    string
			Z1    string
			X2    string
			Z2    string
			Color string
		}
		err := json.NewDecoder(r.Body).Decode(&newMarkerInput)
		if err != nil {
			log.Printf("Failed to decode marker: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		newMarkerInput.Color, _ = strings.CutPrefix(newMarkerInput.Color, "#")
		marker := &database.Marker{}
		err = marker.Color.Scan(newMarkerInput.Color)
		if err != nil {
			log.Printf("Cannot scan color '%s' in new marker pos: %+v", newMarkerInput.Color, err)
			http.Error(w, "Invalid color "+newMarkerInput.Color, http.StatusBadRequest)
			return
		}
		marker.Point1.X, err = strconv.Atoi(newMarkerInput.X1)
		if err != nil {
			log.Printf("Failed to parse integer X1 '%s' in new marker pos: %+v", newMarkerInput.X1, err)
			http.Error(w, "Invalid From X Coordinate "+newMarkerInput.X1, http.StatusBadRequest)
			return
		}
		marker.Point1.Z, err = strconv.Atoi(newMarkerInput.Z1)
		if err != nil {
			log.Printf("Failed to parse integer Z1 '%s' in new marker pos: %+v", newMarkerInput.Z1, err)
			http.Error(w, "Invalid From Z Coordinate "+newMarkerInput.X2, http.StatusBadRequest)
			return
		}
		marker.Point2.X, err = strconv.Atoi(newMarkerInput.X2)
		if err != nil {
			log.Printf("Failed to parse integer X2 '%s' in new marker pos: %+v", newMarkerInput.X2, err)
			http.Error(w, "Invalid To X Coordinate "+newMarkerInput.X2, http.StatusBadRequest)
			return
		}
		marker.Point2.Z, err = strconv.Atoi(newMarkerInput.Z2)
		if err != nil {
			log.Printf("Failed to parse integer Z2 '%s' in new marker pos: %+v", newMarkerInput.Z2, err)
			http.Error(w, "Invalid To X Coordinate "+newMarkerInput.Z2, http.StatusBadRequest)
			return
		}
		marker.Users = append(marker.Users, user.WhitelistEntry)

		user.WhitelistEntry.Markers = append(user.WhitelistEntry.Markers, marker)
		database.DB.Save(&user.WhitelistEntry)
		ServeTemplate(w, r, user.WhitelistEntry.Markers, "account/markers", "account/marker")
		return
	case QUERY_MARKER_USER_ADD:
		handleAddUserToMarker(w, r, user)
		return
	default:
		http.Error(w, "unknown edit mode: "+editMode, http.StatusBadRequest)
		return
	}
}

func handleAddUserToMarker(w http.ResponseWriter, r *http.Request, user database.User) {
	var userIdInput struct {
		ID string
	}
	err := json.NewDecoder(r.Body).Decode(&userIdInput)
	if err != nil {
		log.Printf("Failed to decode user: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID, err := strconv.ParseUint(userIdInput.ID, 10, 64)
	if err != nil {
		log.Printf("Failed to parse user ID '%s': %+v", userIdInput.ID, err)
		http.Error(w, "Not a valid user ID "+userIdInput.ID, http.StatusBadRequest)
		return
	}
	userAdd := &database.User{}
	userAdd.ID = uint(userID)
	if user.ID == userAdd.ID {
		http.Error(w, "Cant add self to a marker", http.StatusBadRequest)
		return
	}

	if err = database.DB.Preload("WhitelistEntry").First(userAdd).Error; err != nil {
		log.Printf("failed to find user '%d' to add to marker: %+v", userAdd.ID, err)
		http.Error(w, "User does not exist "+userIdInput.ID, http.StatusBadRequest)
		return
	}
	markerIdInput := r.URL.Query().Get("id")
	if markerIdInput == "" {
		http.Error(w, "No marker ID", http.StatusBadRequest)
		return
	}
	markerID, err := strconv.ParseUint(markerIdInput, 10, 64)
	if err != nil {
		log.Printf("Failed to parse marker ID '%s': %+v", userIdInput.ID, err)
		http.Error(w, "Not a valid user ID", http.StatusBadRequest)
		return
	}
	marker := &database.Marker{}
	marker.ID = uint(markerID)
	if err = database.DB.Preload("Users").First(marker).Error; err != nil {
		log.Printf("failed to find user '%d' to add to marker: %+v", marker.ID, err)
		http.Error(w, "Marker does not exist"+userIdInput.ID, http.StatusBadRequest)
		return
	}

	var userInMarker, userAddInMarker bool
	for _, markerUser := range marker.Users {
		switch markerUser.UserID {
		case user.ID:
			userInMarker = true
		case userAdd.ID:
			userAddInMarker = true
		}
		if userInMarker && userAddInMarker {
			break
		}
	}
	if !userInMarker {
		http.Error(w, "This is not your marker", http.StatusForbidden)
		return
	}
	if userAddInMarker {
		http.Error(w, "User "+userIdInput.ID+" already in this marker", http.StatusPreconditionFailed)
		return
	}

	marker.Users = append(marker.Users, userAdd.WhitelistEntry)
	if err = database.DB.Save(marker).Error; err != nil {
		log.Printf("failed to save marker '%d' after user add '%d': %+v", marker.ID, userAdd.ID, err)
		http.Error(w, "Failed to save marker", http.StatusInternalServerError)
		return
	}

	ServeTemplate(w, r, marker, "account/marker")
}
