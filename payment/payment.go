package payment

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Payment struct {
	MessageID   int
	ChatID      int64
	UserID      int64
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

func (p Payment) InvoiceWithFiscal(paymentToken string) tgbotapi.InvoiceConfig {
	invoice := p.Invoice(paymentToken)
	invoice.NeedEmail = true
	invoice.SendEmailToProvider = true
	invoice.ProviderData = NewProviderData(p.Amount, p.Description).ToJSON()
	return invoice
}

func (p Payment) Invoice(paymentToken string) tgbotapi.InvoiceConfig {
	prices := []tgbotapi.LabeledPrice{{Label: currencyRUBInfo, Amount: p.Amount}}
	return tgbotapi.NewInvoice(p.ChatID, p.Title, p.Description, p.Payload, paymentToken, "", currencyRUB, prices)
}
