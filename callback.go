package tg_client

import (
	"strconv"
	"strings"
)

const (
	CallbackPaymentNew      = "payment-new"
	CallbackPaymentCheckout = "payment-checkout"
	CallbackPaymentCharge   = "payment-charge"
	CallbackPaymentError    = "payment-error"
)

type Callback struct {
	Type   string
	Value  string
	ItemID int64
}

func NewCallback(data string) Callback {
	rows := strings.Split(data, "_")
	c := Callback{
		Type: rows[0],
	}

	if len(rows) >= 3 {
		c.ItemID, _ = strconv.ParseInt(rows[2], 10, 64)
	}
	if len(rows) >= 2 {
		c.Value = rows[1]
	}
	return c
}
