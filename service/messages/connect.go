package messages

type Connect struct{}

func (*Connect) New() Message{
	return new(Connect)
}

func (*Connect) GetType() MessageType { return CONNECT }

func (*Connect) Unflatten(d []string) {}

func (*Connect) Flatten() [][]byte {
	return [][]byte{}
}

func init(){
	registerMessage(new(Connect))
}
