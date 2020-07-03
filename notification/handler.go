package notification

type Message struct {
	Text string
}

type CommandProvider interface {
	SendMessage(data *Message) error
}
