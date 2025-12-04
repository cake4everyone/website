package types

import (
	"database/sql/driver"
	"fmt"
	"strconv"
)

type Color struct {
	R uint8 `json:"r"`
	G uint8 `json:"g"`
	B uint8 `json:"b"`
	A uint8 `json:"a"`
}

type ColorInt int

func (c ColorInt) ToColor() Color {
	return Color{
		A: uint8(c >> 24),
		R: uint8(c >> 16),
		G: uint8(c >> 8),
		B: uint8(c),
	}
}

func (c Color) ToInt() ColorInt {
	return ColorInt(int(c.A)<<24 | int(c.R)<<16 | int(c.G)<<8 | int(c.B))
}

// Scan implements the [sql.Scanner] interface for Color
func (c *Color) Scan(src any) error {
	switch src := src.(type) {
	case nil:
		c = &Color{}
		return nil
	case string:
		val, err := strconv.ParseUint(src, 16, 32)
		*c = ColorInt(int32(val)).ToColor()
		return err
	case []byte:
		return c.Scan(string(src))
	default:
		return fmt.Errorf("cannot scan '%[1]v' (value of type %[1]T) into Color", src)
	}
}

// Value implements the [driver.Valuer] interface for Color
func (c Color) Value() (driver.Value, error) {
	return fmt.Sprintf("%08x", c.ToInt()), nil
}

func (c Color) String() string {
	return fmt.Sprintf("%02x%02x%02x%02x", c.R, c.G, c.B, c.A)
}
