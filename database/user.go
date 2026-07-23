package database

import (
	"crypto/sha256"
	"encoding/hex"
	"log"
	"regexp"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Email          string `gorm:"not null;unique"`
	Username       string `gorm:"not null;unique"`
	Discord        string
	Twitch         string
	Password       string
	WhitelistEntry *WhitelistEntry `gorm:"foreignKey:ID;references:UserID"`
}

func GetUserByCredentials(username, password string) (user *User) {
	if len(username) == 0 {
		return nil
	}

	sha256 := sha256.Sum256([]byte(password))
	password = hex.EncodeToString(sha256[:])

	query := DB.Model(user).Where("password=?", password)
	if regexp.MustCompile(`^\w+$`).MatchString(username) {
		query = query.Where("username=?", username)
	} else {
		query = query.Where("email=?", username)
	}
	query = query.First(&user)

	if err := query.Error; err != nil {
		log.Printf("Failed to get user by credentials: %v", err)
		return nil
	} else if user == nil || (*user == User{}) {
		return nil
	}
	return user
}

func GetMinecraftUUIDByUsername(username string) (UUID uuid.UUID, ID uint) {
	if len(username) == 0 {
		return uuid.UUID{}, 0
	}

	query := DB.Model(User{}).Preload("WhitelistEntry", func(query *gorm.DB) *gorm.DB {
		return query.Select("id", "uuid")
	})
	if regexp.MustCompile(`^\w+$`).MatchString(username) {
		query = query.Where("username=?", username)
	} else {
		query = query.Where("email=?", username)
	}
	var user *User
	query = query.Select("id").First(&user)

	if err := query.Error; err != nil {
		log.Printf("Failed to get minecraft UUID by username: %v", err)
		return uuid.UUID{}, 0
	} else if user == nil || user.WhitelistEntry == nil {
		return uuid.UUID{}, 0
	}
	return user.WhitelistEntry.UUID, user.ID
}
