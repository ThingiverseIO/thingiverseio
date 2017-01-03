package messages

//go:generate stringer -type=CallType

type CallType int

const (
	CALL CallType = iota
	CALLALL
	TRIGGER
	TRIGGERALL
)
