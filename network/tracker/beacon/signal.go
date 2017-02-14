package beacon

//go:generate event_generator -t Signal

type Signal struct {
	SenderIp []byte
	Data     []byte
}
