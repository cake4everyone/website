package webserver

import (
	"log"
	"net/http"
	"website/auth"
	"website/database"
)

func handleAccountPage(w http.ResponseWriter, r *http.Request) {
	cookie, _ := r.Cookie("session_user")
	var (
		user database.User
		ok   bool
	)
	if user.ID, ok = auth.IsCookieActive(cookie); !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if err := database.DB.Where(user).Preload("WhitelistEntry").First(&user).Error; err != nil {
		log.Printf("Could not get user from database: %+v", err)
		http.Error(w, "Failed to load your user data", http.StatusInternalServerError)
		return
	}

	ServeTemplate(w, r, "account", user)
}
