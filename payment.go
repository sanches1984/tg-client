package tg_client

import (
	"encoding/json"
	"fmt"
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

type ProviderData struct {
	Receipt Receipt `json:"receipt"`
}

type Receipt struct {
	Items []ReceiptItem `json:"items"`
}

type ReceiptItem struct {
	Description string            `json:"description"`
	Quantity    string            `json:"quantity"`
	Amount      ReceiptItemAmount `json:"amount"`
	VatCode     int               `json:"vat_code"`
}

type ReceiptItemAmount struct {
	Value    string `json:"value"`
	Currency string `json:"currency"`
}

type Price struct {
	Label  string `json:"label"`
	Amount int    `json:"amount"`
}

type PaymentInfo struct {
	CheckoutID       string
	TelegramChargeID string
	ProviderChargeID string
	Amount           int
	Currency         string
	Email            string
}

func NewProviderData(p Payment) ProviderData {
	return ProviderData{
		Receipt: Receipt{
			Items: []ReceiptItem{
				{
					Description: p.Title,
					Quantity:    "1.00",
					VatCode:     1,
					Amount: ReceiptItemAmount{
						Value:    fmt.Sprintf("%d.00", p.Amount/100),
						Currency: currencyRUB,
					},
				},
			},
		},
	}
}

func (pd ProviderData) ToJSON() string {
	data, _ := json.Marshal(pd)
	return string(data)
}

func (p Payment) Values() url.Values {
	v := url.Values{}
	v.Add("chat_id", strconv.FormatInt(p.ChatID, 10))
	v.Add("title", p.Title)
	v.Add("description", p.Description)
	v.Add("payload", p.Payload)
	v.Add("currency", currencyRUB)

	prices := []Price{{Label: currencyRUBInfo, Amount: p.Amount}}
	dataPrices, _ := json.Marshal(prices)
	v.Add("prices", string(dataPrices))
	v.Add("need_email", "true")
	v.Add("send_email_to_provider", "true")

	v.Add("provider_data", NewProviderData(p).ToJSON())
	return v
}
