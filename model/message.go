package model

type Message struct {
	Sender   string `json:"sender"`
	Receiver string `json:"receiver"`
	Content  []byte `json:"content"`
}

func NewMessage(content []byte, sender, receiver string) Message {
	return Message{
		Content:  content,
		Sender:   sender,
		Receiver: receiver,
	}
}
