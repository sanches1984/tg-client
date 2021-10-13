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

		if update.Message.Photo != nil {
			msg.FileURL, _ = c.api.GetFileDirectURL((*update.Message.Photo)[len(*update.Message.Photo)-1].FileID)
		} else if update.Message.Document != nil {
			msg.FileURL, _ = c.api.GetFileDirectURL(update.Message.Document.FileID)
		}

		if msg.IsCommand() {
			msg.Type = MessageCommand
			c.waitingMessage.Delete(msg.UserID)
		} else if v, ok := c.waitingMessage.Load(msg.UserID); ok {
			msg.Type = MessageResponse
			msg.Callback = &Callback{Value: getMention(update)}
			if v.(*WaitData) != nil {
				msg.Callback.Type = v.(*WaitData).Type
				msg.Callback.ItemID = v.(*WaitData).Value
			}

		} else if update.Message.SuccessfulPayment != nil {
			msg.Type = MessagePaymentCharge
			msg.Payment = &PaymentInfo{
				TelegramChargeID: update.Message.SuccessfulPayment.TelegramPaymentChargeID,
				ProviderChargeID: update.Message.SuccessfulPayment.ProviderPaymentChargeID,
				Amount:           update.Message.SuccessfulPayment.TotalAmount,
				Currency:         update.Message.SuccessfulPayment.Currency,
				Payload:          update.Message.SuccessfulPayment.InvoicePayload,
			}
		} else {
			msg.Type = MessageText
		}
	} else if update.PreCheckoutQuery != nil {
		msg.Type = MessagePaymentCheckout
		msg.UserID = update.PreCheckoutQuery.From.ID
		msg.Payment = &PaymentInfo{
			CheckoutID: update.PreCheckoutQuery.ID,
			Amount:     update.PreCheckoutQuery.TotalAmount,
			Currency:   update.PreCheckoutQuery.Currency,
			Payload:    update.PreCheckoutQuery.InvoicePayload,
		}
		if update.PreCheckoutQuery.OrderInfo != nil {
			msg.Payment.Email = update.PreCheckoutQuery.OrderInfo.Email
		}
		c.waitingMessage.Delete(msg.UserID)
	}

	if v, ok := c.lastBotMessage.Load(msg.UserID); ok && msg.UserID != 0 {
		if v.(*OutgoingMessage) != nil {
			msg.LastMessage = v.(*OutgoingMessage)
			if (msg.Type == MessagePaymentCheckout || msg.Type == MessagePaymentCharge) && msg.ChatID == 0 {
				msg.ChatID = v.(*OutgoingMessage).ChatID
			}
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
