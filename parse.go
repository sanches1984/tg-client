package tg_client

import (
	"fmt"
	tgbotapi "github.com/Syfaro/telegram-bot-api"
	"reflect"
	"strings"
)

func (c *Client) parseMessage(update tgbotapi.Update) IncomingMessage {
	msg := IncomingMessage{}

	if update.CallbackQuery != nil {
		msg.ID = update.CallbackQuery.Message.MessageID
		msg.UserID = update.CallbackQuery.From.ID
		msg.Login = update.CallbackQuery.From.UserName
		msg.UserName = fmt.Sprintf("%s %s", update.CallbackQuery.From.FirstName, update.CallbackQuery.From.LastName)
		msg.ChatID = update.CallbackQuery.Message.Chat.ID
		msg.Type = MessageCallback
		msg.Callback = NewCallback(update.CallbackQuery.Data)
		if reflect.TypeOf(update.CallbackQuery.Message.Text).Kind() == reflect.String {
			msg.Message = strings.TrimSpace(update.CallbackQuery.Message.Text)
		}

	} else if update.Message != nil {
		msg.ID = update.Message.MessageID
		msg.UserID = update.Message.From.ID
		msg.Login = update.Message.From.UserName
		msg.UserName = fmt.Sprintf("%s %s", update.Message.From.FirstName, update.Message.From.LastName)
		msg.ChatID = update.Message.Chat.ID
		if reflect.TypeOf(update.Message.Text).Kind() == reflect.String {
			msg.Message = strings.TrimSpace(update.Message.Text)
		}

		if msg.IsCommand() {
			msg.Type = MessageCommand
			delete(c.waitingMessage, msg.UserID)
		} else if v, ok := c.waitingMessage[msg.UserID]; ok {
			msg.Type = MessageResponse
			msg.Callback.Type = v.Type
			msg.Callback.ItemID = v.Value
			msg.Callback.Value = getMention(update)
		} else if update.Message.SuccessfulPayment != nil {
			msg.Type = MessagePayment
			msg.Callback.Type = CallbackPaymentCharge
			msg.Callback.Value = update.Message.SuccessfulPayment.InvoicePayload
			msg.Payment = &PaymentInfo{
				TelegramChargeID: update.Message.SuccessfulPayment.TelegramPaymentChargeID,
				ProviderChargeID: update.Message.SuccessfulPayment.ProviderPaymentChargeID,
				Amount:           update.Message.SuccessfulPayment.TotalAmount,
				Currency:         update.Message.SuccessfulPayment.Currency,
			}
		}
	} else if update.PreCheckoutQuery != nil {
		msg.Type = MessagePayment
		msg.UserID = update.PreCheckoutQuery.From.ID
		msg.Callback.Type = CallbackPaymentCheckout
		msg.Callback.Value = update.PreCheckoutQuery.InvoicePayload
		msg.Payment = &PaymentInfo{
			CheckoutID: update.PreCheckoutQuery.ID,
			Amount:     update.PreCheckoutQuery.TotalAmount,
			Currency:   update.PreCheckoutQuery.Currency,
		}
		if update.PreCheckoutQuery.OrderInfo != nil {
			msg.Payment.Email = update.PreCheckoutQuery.OrderInfo.Email
		}
		delete(c.waitingMessage, msg.UserID)
	}

	if v, ok := c.lastBotMessage[msg.UserID]; ok && msg.UserID != 0 {
		msg.LastMessage = &v
		if msg.Type == MessagePayment && msg.ChatID == 0 {
			msg.ChatID = v.ChatID
		}
	}

	return msg
}

func getMention(update tgbotapi.Update) string {
	if reflect.TypeOf(update.Message.Text).Kind() != reflect.String {
		return ""
	}

	msg := strings.TrimSpace(update.Message.Text)
	if msg != "" && update.Message.Entities != nil && len(*update.Message.Entities) > 0 &&
		(*update.Message.Entities)[0].Type == entityTypeMention {
		return strings.Replace(msg, "@", "", 1)
	}

	return ""
}
