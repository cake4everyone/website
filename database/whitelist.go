package database

import (
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
const NICKNAME_SEPARATOR = "\u0000"

func (names Nicknames) String() string {
	return strings.Join(names, ", ")
}

// Set sets the nicknames from a single string.
// It supports multiple nicknames separated by commas.
func (names *Nicknames) Set(new string) {
	*names = strings.Split(new, ",")
	for i, n := range *names {
		n = strings.TrimSpace(n)
		if n == "" {
			*names = append((*names)[:i], (*names)[i+1:]...)
			continue
		}
		(*names)[i] = n
	}
	slices.Sort(*names)
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
	ID          int       `gorm:"NOT NULL"`
	User        *User     `gorm:"foreignKey:ID"`
	UUID        uuid.UUID `gorm:"primaryKey"`
	Name        string
	Nicknames   Nicknames
	ReferenceID int             `gorm:"column:reference"`
	Reference   *WhitelistEntry `gorm:"foreignKey:ReferenceID;references:ID"`
	DiscordID   string          `gorm:"unique"`
	Markers     []*Marker       `gorm:"many2many:user_markers"`
	Flags       int             `gorm:"NOT NULL;default:0"`
}

// TableName implements [schema.Tabler] interface.
// It specifies the table name used by GORM.
func (WhitelistEntry) TableName() string {
	return "whitelist"
}

func (e *WhitelistEntry) Equal(other WhitelistEntry) bool {
	return e.ID == other.ID &&
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
)

func (e WhitelistEntry) IsActive() bool {
	return e.Flags&FlagActive != 0
}

func (e *WhitelistEntry) SetActive(state bool) {
	if state {
		e.Flags |= FlagActive
	} else {
		e.Flags &= ^FlagActive
	}
}

func (e WhitelistEntry) IsAdmin() bool {
	return e.Flags&FlagAdmin != 0
}
