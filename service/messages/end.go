package messages

type End struct{}

func (*End) GetType() MessageType { return END }

func (*End) Unflatten(d []string) {}

func (*End) Flatten() [][]byte {
	return [][]byte{}
}
