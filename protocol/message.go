package protocol

import "fmt"

type Message struct {
	Sender   string `json:"sender"`
	Receiver string `json:"receiver"`
	Content  []byte `json:"content"`
}

func NewMessage(content []byte, sender, receiver string) (Message, error) {
	if sender == "" {
		return Message{}, fmt.Errorf("sender cannot be empty")
	}
	if receiver == "" {
		return Message{}, fmt.Errorf("receiver cannot be empty")
	}
	if len(content) == 0 {
		return Message{}, fmt.Errorf("content cannot be empty")
	}
	return Message{
		Content:  content,
		Sender:   sender,
		Receiver: receiver,
	}, nil
}

func (m Message) Validate() error {
	if m.Sender == "" {
		return fmt.Errorf("sender cannot be empty")
	}
	if m.Receiver == "" {
		return fmt.Errorf("receiver cannot be empty")
	}
	if len(m.Content) == 0 {
		return fmt.Errorf("content cannot be empty")
	}
	return nil
}
