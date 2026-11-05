package database

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"regexp"
	"slices"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Email          string `gorm:"not null;unique"`
	Username       string `gorm:"not null;unique"`
	Name           string
	OnlineName     string
	Nicknames      Nicknames
	ReferenceID    uint
	Reference      *User
	Flags          UserFlags `gorm:"NOT NULL;default:0"`
	Discord        string
	Twitch         string
	Password       string
	WhitelistEntry *WhitelistEntry `gorm:"foreignKey:ID;references:UserID"`
}

func NewUser(username string, mcUUID uuid.UUID, referenceID uint, isActive bool) (newUser *User, err error) {
	newUser = &User{
		Email:       fmt.Sprintf("%s@local", username),
		Username:    username,
		Password:    "",
		ReferenceID: referenceID,
	}

	if err = DB.Save(newUser).Error; err != nil {
		return nil, err
	}
	log.Printf("Creating new user with ID %d", newUser.ID)

	if mcUUID == uuid.Nil {
		return newUser, nil
	}

	mcAccount := &WhitelistEntry{
		UserID: newUser.ID,
		User:   *newUser,
		UUID:   mcUUID,
	}
	mcAccount.SetActive(isActive)

	if err = DB.Save(mcAccount).Error; err != nil {
		return nil, err
	}

	newUser.WhitelistEntry = mcAccount
	return newUser, err
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
	} else if user == nil || user.Equal(User{}) {
		return nil
	}
	return user
}

func GetUserByID(id uint) (user *User, err error) {
	user = new(User)
	user.ID = id
	if err := DB.Preload("WhitelistEntry").First(&user).Error; err != nil {
		return nil, fmt.Errorf("get user by ID '%d' from database: %w", id, err)
	}
	return user, nil
}

func GetUserByUsername(username string) (user *User) {
	if len(username) == 0 {
		return nil
	}

	query := DB.Model(user)
	if regexp.MustCompile(`^\w+$`).MatchString(username) {
		query = query.Where("username=?", username)
	} else {
		query = query.Where("email=?", username)
	}
	query = query.Preload("WhitelistEntry").First(&user)

	if err := query.Error; err != nil {
		log.Printf("Failed to get user by username: %v", err)
		return nil
	} else if user == nil || user.Equal(User{}) {
		return nil
	}
	return user
}

func GetMinecraftUUIDByUsername(username string) (UUID uuid.UUID, ID uint) {
	user := GetUserByUsername(username)
	if user == nil {
		return uuid.UUID{}, 0
	}
	return user.WhitelistEntry.UUID, user.ID
}

func (u *User) Equal(other User) bool {
	return u.ID == other.ID &&
		u.Email == other.Email &&
		u.Username == other.Username &&
		u.Name == other.Email &&
		u.OnlineName == other.OnlineName &&
		u.ReferenceID == other.ReferenceID &&
		(u.Reference == nil && other.Reference == nil ||
			u.Reference != nil && other.Reference != nil && u.Reference.Equal(*other.Reference)) &&
		u.Flags == other.Flags &&
		u.Discord == other.Discord &&
		u.Twitch == other.Twitch &&
		u.Password == other.Password &&
		slices.Equal(u.Nicknames, other.Nicknames) &&
		(u.WhitelistEntry == nil && other.WhitelistEntry == nil ||
			u.WhitelistEntry != nil && other.WhitelistEntry != nil && u.WhitelistEntry.Equal(*other.WhitelistEntry))
}

type UserFlags int

const (
	UserFlagAdmin UserFlags = 1 << iota
)

func (u User) IsAdmin() bool {
	return u.isFlagged(UserFlagAdmin)
}

func (u *User) SetAdmin(state bool) {
	u.setFlag(UserFlagAdmin, state)
}

func (u User) isFlagged(flag UserFlags) bool {
	return u.Flags&flag != 0
}

func (u *User) setFlag(flag UserFlags, state bool) {
	if state {
		u.Flags |= flag
	} else {
		u.Flags &= ^flag
	}
}
