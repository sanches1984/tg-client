package payment

import (
	"encoding/json"
	"fmt"
)

const currencyRUB = "RUB"
const currencyRUBInfo = "руб."

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

type Prices []Price

type Price struct {
	Label  string `json:"label"`
	Amount int    `json:"amount"`
}

func NewProviderData(amount int, description string) ProviderData {
	return ProviderData{
		Receipt: Receipt{
			Items: []ReceiptItem{
				{
					Description: description,
					Quantity:    "1.00",
					VatCode:     1,
					Amount: ReceiptItemAmount{
						Value:    fmt.Sprintf("%d.00", amount/100),
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

func NewPrices(amount int) Prices {
	return []Price{{Label: currencyRUBInfo, Amount: amount}}
}

func (p Prices) ToJSON() string {
	data, _ := json.Marshal(p)
	return string(data)
}
