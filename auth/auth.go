package auth

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
	"website/database"

	"github.com/golang-jwt/jwt/v4"
	"github.com/spf13/viper"
)

type session struct {
	UserID uint
	Exp    time.Time
}

var (
	activeSessions = make(map[string]session) // map[sessionToken]userID
	sessionsMutex  sync.Mutex
)

func NewToken(userID uint, exp time.Duration) (token string) {
	expiration := time.Now().Add(exp)
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     expiration.Unix(),
		"iat":     time.Now().Unix(),
		"nbf":     time.Now().Unix(),
	}).SignedString([]byte(viper.GetString("auth.jwt_secret")))
	if err != nil {
		log.Fatal(err)
	}
	sessionsMutex.Lock()
	defer sessionsMutex.Unlock()
	activeSessions[token] = session{
		UserID: userID,
		Exp:    expiration,
	}
	return
}

func IsSessionActive(sessionToken string) (userID uint, ok bool) {
	sessionsMutex.Lock()
	defer sessionsMutex.Unlock()
	if s, found := activeSessions[sessionToken]; found {
		if s.Exp.Before(time.Now()) {
			delete(activeSessions, sessionToken)
			return
		}
		return s.UserID, true
	}

	token, err := jwt.Parse(sessionToken, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(viper.GetString("auth.jwt_secret")), nil
	})
	if err != nil {
		log.Printf("failed to parse token: %v", err)
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return
	}
	activeSessions[sessionToken] = session{
		UserID: uint(claims["user_id"].(float64)),
		Exp:    time.Unix(int64(claims["exp"].(float64)), 0),
	}
	return
}

func IsCookieActive(cookie *http.Cookie) (userID uint, ok bool) {
	if cookie == nil {
		return
	}
	return IsSessionActive(cookie.Value)

}

// GetUser returns the user that is logged in. The returned user is preloaded
// with their whitelist entry.
//
// On an error, the a reponse will be written on w and abort is set to true to
// indicate that any further processing should be aborted.
func GetUser(w http.ResponseWriter, r *http.Request) (user database.User, abort bool) {
	cookie, _ := r.Cookie("session_user")
	var ok bool
	if user.ID, ok = IsCookieActive(cookie); !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return user, true
	}
	if err := database.DB.Where(user).Preload("WhitelistEntry").First(&user).Error; err != nil {
		log.Printf("Could not get user from database: %+v", err)
		http.Error(w, "Failed to load your user data", http.StatusInternalServerError)
		return user, true
	}
	return user, false
}
