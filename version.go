package thingiverseio

import "fmt"

type Version struct {
	Major, Minor, Fix int
}

func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Fix)
}

var CurrentVersion = Version{0, 0, 1}
