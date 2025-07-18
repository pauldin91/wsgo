package internal

type Message struct {
	Sender   string
	Receiver string
	Content  []byte
}

func NewMessage(content []byte, sender, receiver string) Message {
	return Message{
		Content:  content,
		Sender:   sender,
		Receiver: receiver,
	}
}
