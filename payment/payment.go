package payment

import (
	tgbotapi "github.com/Syfaro/telegram-bot-api"
	"net/url"
	"strconv"
)

type Payment struct {
	MessageID   int
	ChatID      int64
	UserID      int
	Title       string
	Description string
	Amount      int
	Payload     string
}

type PaymentInfo struct {
	CheckoutID       string
	TelegramChargeID string
	ProviderChargeID string
	Amount           int
	Currency         string
	Email            string
	Payload          string
}

func (p Payment) Values() url.Values {
	v := url.Values{}
	v.Add("chat_id", strconv.FormatInt(p.ChatID, 10))
	v.Add("title", p.Title)
	v.Add("description", p.Description)
	v.Add("payload", p.Payload)
	v.Add("currency", currencyRUB)
	v.Add("prices", NewPrices(p.Amount).ToJSON())
	v.Add("need_email", "true")
	v.Add("send_email_to_provider", "true")
	v.Add("provider_data", NewProviderData(p.Amount, p.Description).ToJSON())
	return v
}

func (p Payment) Invoice(paymentToken string) tgbotapi.InvoiceConfig {
	prices := []tgbotapi.LabeledPrice{{Label: currencyRUBInfo, Amount: p.Amount}}
	return tgbotapi.NewInvoice(p.ChatID, p.Title, p.Description, p.Payload, paymentToken, "", currencyRUB, &prices)
}
