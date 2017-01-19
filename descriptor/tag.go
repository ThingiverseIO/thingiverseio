package descriptor

import (
	"fmt"
	"strings"
)

type Tag struct {
	Key   string
	Value string
}

func (t Tag) String() string {
	if t.Value != "" {
		return fmt.Sprintf("%s:%s", t.Key, t.Value)
	}
	return t.Key
}

func (t *Tag) Scan(tag string) {
	split := strings.Split(tag, ":")
	t.Key = split[0]
	if len(split) > 1 {
		t.Value = split[1]
	}
}
