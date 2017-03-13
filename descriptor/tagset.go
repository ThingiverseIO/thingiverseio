package descriptor

import "strings"

type Tagset map[string]string

func (ts Tagset) Add(t Tag) {
	ts[t.Key] = t.Value
}

func (ts Tagset) Merge(nts Tagset) {
	for k, v := range nts {
		ts[k] = v
	}
}

func (ts Tagset) Has(t Tag) (has bool) {
	var val string
	if val, has = ts[t.Key]; has {
		has = t.Value == val
	}
	return
}

func (ts Tagset) GetFirst() (t Tag) {
	for k, v := range ts {
		t = Tag{
			Key:   k,
			Value: v,
		}
		return
	}
	return
}

func (ts Tagset) AsArray() (t []Tag) {
	for k, v := range ts {
		t = append(t, Tag{Key: k, Value: v})
	}
	return
}

func (ts Tagset) String() (s string) {
	var tags []string
	for _, t := range ts.AsArray() {
		tags = append(tags, t.String())
	}

	s = strings.Join(tags, ", ")
	return
}
