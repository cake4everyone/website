package database

import (
	"website/types"

	"gorm.io/gorm"
)

type MarkerPoint struct {
	X int `gorm:";NOT NULL"`
	Z int `gorm:";NOT NULL"`
}

type Marker struct {
	gorm.Model
	Users  []*WhitelistEntry `gorm:"many2many:user_markers"`
	Point1 MarkerPoint       `gorm:"embedded;embeddedPrefix:point1_"`
	Point2 MarkerPoint       `gorm:"embedded;embeddedPrefix:point2_"`
	Color  types.Color
}

// Overlaps returns true if the two markers overlap.
// Assuming point 1 is the top-left and point 2 is the bottom-right.
func (m Marker) Overlaps(other Marker) bool {
	return m.Point1.X <= other.Point2.X &&
		m.Point2.X >= other.Point1.X &&
		m.Point1.Z <= other.Point2.Z &&
		m.Point2.Z >= other.Point1.Z
}
