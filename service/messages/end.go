package messages

type End struct{}

func (*End) New() Message{
	return new(End)
}

func (*End) GetType() MessageType { return END }

func (*End) Unflatten(d [][]byte) {}

func (*End) Flatten() [][]byte {
	return [][]byte{}
}

func init(){
	registerMessage(new(End))
}
