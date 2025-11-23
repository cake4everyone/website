package webserver

import (
	"net/http"
	"time"
	"website/auth"
	"website/database"
)

const (
	sessionCookieName = "session_user"
)

func handleLoginPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Has("logout") {
		query := r.URL.Query()
		query.Del("logout")
		r.URL.RawQuery = query.Encode()

		unsetSessionCookie(w)
	} else {
		cookie, _ := r.Cookie(sessionCookieName)
		if _, ok := auth.IsCookieActive(cookie); ok {
			http.Redirect(w, r, "/account", http.StatusSeeOther)
			return
		}
	}
	ServeFile(w, r, "login")
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}
	username := r.FormValue("username")
	password := r.FormValue("password")
	user := database.GetUserByCredentials(username, password)
	if user != nil {
		http.SetCookie(w, &http.Cookie{
			Name:   sessionCookieName,
			Value:  auth.NewToken(user.ID, time.Hour*24),
			MaxAge: int((time.Hour * 24).Seconds()),
		})
		http.Redirect(w, r, "/account", http.StatusSeeOther)
		return
	}
	http.Error(w, "Invalid username or password", http.StatusUnauthorized)
}

func unsetSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:   sessionCookieName,
		Value:  "",
		MaxAge: -1,
	})
}
