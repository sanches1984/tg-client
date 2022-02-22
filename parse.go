package tg_client

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sanches1984/tg-client/payment"
	"reflect"
	"strconv"
	"strings"
)

func (c *Client) parseMessage(update tgbotapi.Update) *IncomingMessage {
	msg := &IncomingMessage{}

	if update.CallbackQuery != nil {
		msg = parseCallback(update)
	} else if update.PreCheckoutQuery != nil {
		msg = parseCheckoutQuery(update)
		c.waitingMessage.Delete(msg.UserID)
	} else if update.Message != nil {
		msg = parseUserMessage(update)
		msg.FileURL, _ = c.api.GetFileDirectURL(msg.FileID)

		if msg.IsCommand() {
			msg.Type = MessageCommand
			c.waitingMessage.Delete(msg.UserID)
		} else if v, ok := c.waitingMessage.Load(msg.UserID); ok {
			msg.Type = MessageResponse
			msg.Callback = getCallback(update, v.(*WaitData))
		} else if update.Message.SuccessfulPayment != nil {
			msg.Type = MessagePaymentCharge
			msg.Payment = getPaymentInfo(update)
		} else {
			msg.Type = MessageText
		}
	}

	if v, ok := c.lastBotMessage.Load(msg.UserID); ok && msg.UserID != 0 {
		if v.(*OutgoingMessage) != nil {
			msg.LastMessage = v.(*OutgoingMessage)
			if msg.IsPayment() && msg.ChatID == 0 {
				msg.ChatID = v.(*OutgoingMessage).ChatID
			}
		}
	}

	return msg
}

func parseCallback(update tgbotapi.Update) *IncomingMessage {
	msg := &IncomingMessage{
		ID:       update.CallbackQuery.Message.MessageID,
		Type:     MessageCallback,
		UserID:   update.CallbackQuery.From.ID,
		Login:    update.CallbackQuery.From.UserName,
		UserName: fmt.Sprintf("%s %s", update.CallbackQuery.From.FirstName, update.CallbackQuery.From.LastName),
		ChatID:   update.CallbackQuery.Message.Chat.ID,
		Callback: NewCallback(update.CallbackQuery.Data),
	}

	if reflect.TypeOf(update.CallbackQuery.Message.Text).Kind() == reflect.String {
		msg.Message = strings.TrimSpace(update.CallbackQuery.Message.Text)
	}
	return msg
}

func parseCheckoutQuery(update tgbotapi.Update) *IncomingMessage {
	msg := &IncomingMessage{
		Type:   MessagePaymentCheckout,
		UserID: update.PreCheckoutQuery.From.ID,
		Payment: &payment.PaymentInfo{
			CheckoutID: update.PreCheckoutQuery.ID,
			Amount:     update.PreCheckoutQuery.TotalAmount,
			Currency:   update.PreCheckoutQuery.Currency,
			Payload:    update.PreCheckoutQuery.InvoicePayload,
		},
	}
	if update.PreCheckoutQuery.OrderInfo != nil {
		msg.Payment.Email = update.PreCheckoutQuery.OrderInfo.Email
	}
	return msg
}

func parseUserMessage(update tgbotapi.Update) *IncomingMessage {
	msg := &IncomingMessage{
		ID:       update.Message.MessageID,
		UserID:   update.Message.From.ID,
		Login:    update.Message.From.UserName,
		UserName: fmt.Sprintf("%s %s", update.Message.From.FirstName, update.Message.From.LastName),
		ChatID:   update.Message.Chat.ID,
	}
	if reflect.TypeOf(update.Message.Text).Kind() == reflect.String {
		msg.Message = strings.TrimSpace(update.Message.Text)
	}

	if len(update.Message.Photo) > 0 {
		// get first image
		msg.FileID = update.Message.Photo[0].FileID
	} else if update.Message.Document != nil {
		msg.FileID = update.Message.Document.FileID
	}

	return msg
}

func getPaymentInfo(update tgbotapi.Update) *payment.PaymentInfo {
	return &payment.PaymentInfo{
		TelegramChargeID: update.Message.SuccessfulPayment.TelegramPaymentChargeID,
		ProviderChargeID: update.Message.SuccessfulPayment.ProviderPaymentChargeID,
		Amount:           update.Message.SuccessfulPayment.TotalAmount,
		Currency:         update.Message.SuccessfulPayment.Currency,
		Payload:          update.Message.SuccessfulPayment.InvoicePayload,
	}
}

func getCallback(update tgbotapi.Update, wd *WaitData) *Callback {
	callback := &Callback{Value: getUserIDMention(update)}
	if wd != nil {
		callback.Type = wd.Type
		callback.ItemID = wd.Value
	}
	return callback
}

func getUserIDMention(update tgbotapi.Update) string {
	if reflect.TypeOf(update.Message.Text).Kind() != reflect.String {
		return ""
	}

	if len(update.Message.Entities) > 0 && update.Message.Entities[0].IsMention() {
		return strconv.FormatInt(update.Message.Entities[0].User.ID, 10)
	}

	return ""
}
