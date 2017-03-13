package uuid

import uuid "github.com/nu7hatch/gouuid"

//go:generate event_generator -t UUID

type UUID string

func New() UUID {
	id, _ := uuid.NewV4()
	return UUID(id.String())
}


func Empty() UUID {
	return UUID("")
}

func (u UUID) IsEmpty() bool {
	return u == Empty()
}

func (u UUID) String() (s string) {
	if len(u) <= 5 {
		return string(u)
	}
	return string(u[:5])
}

func (u UUID) FullString() string {
	return string(u)
}
