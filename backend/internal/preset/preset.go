// Package preset defines the built-in government photo presets (OCSC,
// passport, teacher) plus the "custom" path where width/height/max_kb come
// from the request instead of a lookup table.
package preset

import "fmt"

// Preset describes a fixed target size and max file weight for a named
// government form.
type Preset struct {
	Name   string `json:"name"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	MaxKB  int    `json:"max_kb"`
}

// Custom holds the caller-supplied dimensions/size cap for preset=custom.
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

// builtins is the ordered set of non-custom presets. Order matters for the
// /api/v1/presets response, which lists them in this order.
var builtins = []Preset{
	{Name: OCSC, Width: 200, Height: 230, MaxKB: 100},
	{Name: Passport, Width: 500, Height: 500, MaxKB: 200},
	{Name: Teacher, Width: 300, Height: 400, MaxKB: 200},
}

// All returns the list of built-in presets (does not include "custom").
func All() []Preset {
	out := make([]Preset, len(builtins))
	copy(out, builtins)
	return out
}

// Lookup returns the built-in preset by name. It does not resolve "custom" —
// callers must build a Custom value from the request for that case.
func Lookup(name string) (Preset, bool) {
	for _, p := range builtins {
		if p.Name == name {
			return p, true
		}
	}
	return Preset{}, false
}

// IsValidName reports whether name is one of the four accepted preset
// values (the three built-ins plus "custom").
func IsValidName(name string) bool {
	if name == CustomName {
		return true
	}
	_, ok := Lookup(name)
	return ok
}

// ValidateCustom checks that a custom preset's dimensions/size cap are
// sane positive values.
func ValidateCustom(c Custom) error {
	if c.Width <= 0 || c.Height <= 0 {
		return fmt.Errorf("width and height must be positive integers")
	}
	if c.MaxKB <= 0 {
		return fmt.Errorf("max_kb must be a positive integer")
	}
	return nil
}
