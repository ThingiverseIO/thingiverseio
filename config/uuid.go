package config

import uuid "github.com/nu7hatch/gouuid"

//go:generate event_generator -t UUID

type UUID string

func NewUUID() UUID {
	id, _ := uuid.NewV4()
	return UUID(id.String())
}

func (u UUID) isSet() bool {
	return len(u) != 0
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
