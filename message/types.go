package message

//go:generate stringer -type=Type

// Type represents a message constant.
type Type int

// Message type constants
const (
	HELLO Type = iota
	HELLOOK
	DOHAVE
	HAVE
	CONNECT
	REQUEST
	RESULT
	STARTLISTEN
	STOPLISTEN
	GETPROPERTY
	SETPROPERTY
	STARTOBSERVE
	STOPOBSERVE
	END
	REJECT
	MOCK
)

func Is(t Type) MessageFilter {
	return func(d Message) bool {
		return d.GetType() == t
	}
}
