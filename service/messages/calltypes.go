package messages

//go:generate stringer -type=CallType

type CallType int

const (
	ONE2ONE CallType = iota
	ONE2MANY
	MANY2ONE
	MANY2MANY
)
