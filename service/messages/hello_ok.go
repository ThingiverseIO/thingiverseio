package messages

type HelloOk struct {
}

func (*HelloOk) GetType() MessageType { return HELLO_OK }

func (h *HelloOk) Unflatten(d []string) {
}

func (h *HelloOk) Flatten() [][]byte {
	return [][]byte{}
}
