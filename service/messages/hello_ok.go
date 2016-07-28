package messages

type HelloOk struct {
}

func (*HelloOk) New() Message{
	return new(HelloOk)
}

func (*HelloOk) GetType() MessageType { return HELLOOK }

func (h *HelloOk) Unflatten(d []string) {
}

func (h *HelloOk) Flatten() [][]byte {
	return [][]byte{}
}

func init(){
	registerMessage(new(HelloOk))
}
