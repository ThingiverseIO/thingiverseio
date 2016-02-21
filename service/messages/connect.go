package messages

type Connect struct{}

func (*Connect) GetType() MessageType { return CONNECT }

func (*Connect) Unflatten(d []string) {}

func (*Connect) Flatten() [][]byte {
	return [][]byte{}
}
