package database

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"slices"
	"strings"

	"github.com/google/uuid"
	_ "gorm.io/gorm/schema"
)

// Nicknames is a custom type representing a list of nicknames and how to
// serialize them for the database.
type Nicknames []string

// NICKNAME_SEPARATOR is the separator used to join multiple nicknames into a
// single string.
const NICKNAME_SEPARATOR = "|"

func (names Nicknames) String() string {
	return strings.Join(names, ", ")
}

// Set sets the nicknames from a slice of strings.
// It sorts and trims given nicknames.
func (names *Nicknames) Set(new []string) {
	*names = make(Nicknames, 0, len(new))
	for _, n := range new {
		n = strings.TrimSpace(n)
		if n == "" {
			continue
		}
		*names = append(*names, n)
	}
	slices.SortFunc(*names, func(a, b string) int {
		return strings.Compare(strings.ToLower(a), strings.ToLower(b))
	})
}

// Scan implements the [sql.Scanner] interface for Nicknames.
// It supports scanning into a slice of nickname strings.
func (names *Nicknames) Scan(src any) error {
	switch src := src.(type) {
	case nil:
		return nil
	case string:
		*names = strings.Split(src, NICKNAME_SEPARATOR)
		return nil
	case []byte:
		*names = strings.Split(string(src), NICKNAME_SEPARATOR)
		return nil
	default:
		return fmt.Errorf("cannot scan '%[1]v' (value of type %[1]T) into Nicknames", src)
	}
}

// Value implements the [driver.Valuer] interface for Nicknames.
// It supports serializing a slice of nickname strings into a single string.
func (names Nicknames) Value() (driver.Value, error) {
	return strings.Join(names, NICKNAME_SEPARATOR), nil
}

type WhitelistEntry struct {
	UserID      uint      `gorm:"column:id;unique;NOT NULL"`
	User        *User     `gorm:"foreignKey:UserID"`
	UUID        uuid.UUID `gorm:"primaryKey"`
	Name        string
	Nicknames   Nicknames
	ReferenceID uint            `gorm:"column:reference"`
	Reference   *WhitelistEntry `gorm:"foreignKey:ReferenceID;references:UserID"`
	DiscordID   sql.NullString  `gorm:"unique"`
	Markers     []*Marker       `gorm:"many2many:user_markers"`
	Flags       int             `gorm:"NOT NULL;default:0"`
}

// TableName implements [schema.Tabler] interface.
// It specifies the table name used by GORM.
func (WhitelistEntry) TableName() string {
	return "whitelist"
}

func (e *WhitelistEntry) Equal(other WhitelistEntry) bool {
	return e.UserID == other.UserID &&
		e.UUID == other.UUID &&
		e.Name == other.Name &&
		e.ReferenceID == other.ReferenceID &&
		e.DiscordID == other.DiscordID &&
		slices.Equal(e.Nicknames, other.Nicknames) &&
		slices.Equal(e.Markers, other.Markers)
}

const (
	FlagActive = 1 << iota
	FlagAdmin

	FlagAll = FlagActive | FlagAdmin
)

func (e WhitelistEntry) IsActive() bool {
	return e.isFlagged(FlagActive)
}

func (e *WhitelistEntry) SetActive(state bool) {
	e.setFlag(FlagActive, state)
}

func (e WhitelistEntry) IsAdmin() bool {
	return e.isFlagged(FlagAdmin)
}

func (e *WhitelistEntry) SetAdmin(state bool) {
	e.setFlag(FlagAdmin, state)
}

func (e WhitelistEntry) isFlagged(flag int) bool {
	return e.Flags&flag != 0
}

func (e *WhitelistEntry) setFlag(flag int, state bool) {
	if state {
		e.Flags |= flag
	} else {
		e.Flags &= ^flag
	}
}
