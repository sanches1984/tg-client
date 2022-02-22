package payment

import (
	"errors"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"strconv"
)

type Client struct {
	api          *tgbotapi.BotAPI
	paymentToken string
	withFiscal   bool
}

func New(api *tgbotapi.BotAPI, paymentToken string, withFiscal bool) *Client {
	return &Client{
		api:          api,
		paymentToken: paymentToken,
		withFiscal:   withFiscal,
	}
}

func (c *Client) Send(payment *Payment) error {
	var invoice tgbotapi.InvoiceConfig
	if c.withFiscal {
		invoice = payment.InvoiceWithFiscal(c.paymentToken)
	} else {
		invoice = payment.Invoice(c.paymentToken)
	}

	msg, err := c.api.Send(invoice)
	if err != nil {
		return err
	}

	payment.MessageID = msg.MessageID
	return nil
}

func (c *Client) Complete(checkoutID string, err error) error {
	v := map[string]string{}
	v["pre_checkout_query_id"] = checkoutID
	v["ok"] = strconv.FormatBool(err == nil)
	if err != nil {
		v["error_message"] = err.Error()
	}

	resp, err := c.api.MakeRequest("answerPreCheckoutQuery", v)
	if err != nil {
		return err
	}
	if !resp.Ok {
		return errors.New(resp.Description)
	}
	return nil
}
