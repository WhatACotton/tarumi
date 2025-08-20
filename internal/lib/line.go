package lib

type ISendMessage interface {
	SendMessage(message string) error
}
