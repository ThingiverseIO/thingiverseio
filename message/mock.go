package message

type Mock struct {
	Data [][]byte
}

func (*Mock) New() Message {
	return new(Mock)
}

func (*Mock) GetType() Type { return MOCK }

func (m *Mock) Unflatten(d [][]byte) {
	m.Data = d
}

func (m *Mock) Flatten() [][]byte {
	return m.Data
}

func init() {
	registerMessage(&Mock{})
}
