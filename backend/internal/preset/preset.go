// Package preset defines the built-in government photo presets plus the
// "custom" path where width/height/max_kb come from the request.
package preset

import "fmt"

type Preset struct {
	Name   string `json:"name"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	MaxKB  int    `json:"max_kb"`
}

type Custom struct {
	Width  int
	Height int
	MaxKB  int
}

const (
	OCSC       = "ocsc"
	Passport   = "passport"
	Teacher    = "teacher"
	CustomName = "custom"
)

var builtins = []Preset{
	{Name: OCSC, Width: 200, Height: 230, MaxKB: 100},
	{Name: Passport, Width: 500, Height: 500, MaxKB: 200},
	{Name: Teacher, Width: 300, Height: 400, MaxKB: 200},
}

func All() []Preset {
	out := make([]Preset, len(builtins))
	copy(out, builtins)
	return out
}

func Lookup(name string) (Preset, bool) {
	for _, p := range builtins {
		if p.Name == name {
			return p, true
		}
	}
	return Preset{}, false
}

func IsValidName(name string) bool {
	if name == CustomName {
		return true
	}
	_, ok := Lookup(name)
	return ok
}

func ValidateCustom(c Custom) error {
	if c.Width <= 0 || c.Height <= 0 {
		return fmt.Errorf("width and height must be positive integers")
	}
	if c.MaxKB <= 0 {
		return fmt.Errorf("max_kb must be a positive integer")
	}
	return nil
}
