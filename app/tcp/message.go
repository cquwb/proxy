package tcp

type Message struct {
	t int //消息类型
	k int //第几条连接
}

func NewMessage(t, k int) *Message {
	return &Message{t, k}
}

func (m *Message) GetT() int {
	return m.t
}

func (m *Message) GetK() int {
	return m.k
}
