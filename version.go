package thingiverseio

import "fmt"

// Version represents a library version according to the semantic version scheme.
type Version struct {
	Major, Minor, Fix int
}

func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Fix)
}

// CurrentVersion is the current version of the library.
var CurrentVersion = Version{0, 1, 0}
