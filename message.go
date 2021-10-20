package tg_client

import tgbotapi "github.com/Syfaro/telegram-bot-api"

type IncomingMessageType string
type OutgoingMessageType string
type WaitMessageType string

const (
	MessageCommand         IncomingMessageType = "command"
	MessageCallback        IncomingMessageType = "callback"
	MessageResponse        IncomingMessageType = "response"
	MessagePaymentCheckout IncomingMessageType = "payment-checkout"
	MessagePaymentCharge   IncomingMessageType = "payment-charge"
	MessageText            IncomingMessageType = "text"

	MessageDefault OutgoingMessageType = "default"
	MessageEdit    OutgoingMessageType = "edit"
	MessageDelete  OutgoingMessageType = "delete"

	entityTypeMention = "mention"
	parseModeMarkdown = "markdown"
	currencyRUB       = "RUB"
	currencyRUBInfo   = "руб."
)

type IncomingMessage struct {
	ID          int
	UserID      int
	Type        IncomingMessageType
	Login       string
	UserName    string
	ChatID      int64
	Message     string
	FileURL     string
	Callback    *Callback
	Payment     *PaymentInfo
	LastMessage *OutgoingMessage
	User        interface{}
}

type OutgoingMessage struct {
	ID             int
	Type           OutgoingMessageType
	ChatID         int64
	UserID         int
	Message        string
	Markup         interface{}
	WaitData       *WaitData
	ReplyMessageID int
	Formatted      bool
	File           *tgbotapi.FileBytes
}

type WaitData struct {
	Type  string
	Value int64
}

func (m IncomingMessage) IsCommand() bool {
	return len(m.Message) > 1 && m.Message[:1] == "/"
}
