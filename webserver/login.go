package webserver

import (
	"net/http"
	"sync"
	"time"
	"website/auth"
	"website/database"

	"github.com/google/uuid"
)

const (
	sessionCookieName = "session_user"
)

var (
	mcLoginMux sync.RWMutex
	mcLoginMap = make(map[uuid.UUID]chan struct{})
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

func handleLoginMC(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}
	username := r.FormValue("username")

	mcUUID, userID := database.GetMinecraftUUIDByUsername(username)
	if (mcUUID == uuid.UUID{}) {
		http.Error(w, "Unknown username", http.StatusUnauthorized)
		return
	}

	loggedIn := waitForMCLogin(mcUUID, 5*time.Minute)
	if !loggedIn {
		http.Error(w, "Didn't login in time", http.StatusUnauthorized)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:   sessionCookieName,
		Value:  auth.NewToken(userID, time.Hour*24),
		MaxAge: int((time.Hour * 24).Seconds()),
	})
	http.Redirect(w, r, "/account", http.StatusSeeOther)
}

func waitForMCLogin(UUID uuid.UUID, timeout time.Duration) (success bool) {
	loggedInChannel := make(chan struct{}, 1)
	mcLoginMux.Lock()
	mcLoginMap[UUID] = loggedInChannel
	mcLoginMux.Unlock()

	select {
	case <-loggedInChannel:
		success = true
	case <-time.After(timeout):
		success = false
	}

	mcLoginMux.Lock()
	delete(mcLoginMap, UUID)
	mcLoginMux.Unlock()

	return success
}

func MCLoginHasActiveLogin(UUID uuid.UUID) (ch chan struct{}, ok bool) {
	mcLoginMux.RLock()
	ch, ok = mcLoginMap[UUID]
	mcLoginMux.RUnlock()
	return ch, ok
}

func unsetSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:   sessionCookieName,
		Value:  "",
		MaxAge: -1,
	})
}
