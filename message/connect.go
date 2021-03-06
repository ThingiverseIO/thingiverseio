package message

type Connect struct{}

func (*Connect) New() Message{
	return new(Connect)
}

func (*Connect) GetType() Type { return CONNECT }

func (*Connect) Unflatten(d [][]byte) {}

func (*Connect) Flatten() [][]byte {
	return [][]byte{}
}

func init(){
	registerMessage(new(Connect))
}
