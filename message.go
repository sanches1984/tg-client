package tg_client

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sanches1984/tg-client/payment"
)

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
)

type IncomingMessage struct {
	ID          int
	UserID      int64
	Type        IncomingMessageType
	Login       string
	UserName    string
	ChatID      int64
	Message     string
	FileID      string
	FileURL     string
	Callback    *Callback
	Payment     *payment.PaymentInfo
	LastMessage *OutgoingMessage
	User        interface{}
}

type OutgoingMessage struct {
	ID             int
	Type           OutgoingMessageType
	ChatID         int64
	UserID         int64
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

func (m IncomingMessage) IsPayment() bool {
	return m.Type == MessagePaymentCheckout || m.Type == MessagePaymentCharge
}
