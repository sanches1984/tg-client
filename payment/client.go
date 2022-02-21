package payment

import (
	"errors"
	tgbotapi "github.com/Syfaro/telegram-bot-api"
	"net/url"
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
	if c.withFiscal {
		return c.sendPaymentWithFiscal(payment)
	}

	msg, err := c.api.Send(payment.Invoice(c.paymentToken))
	if err != nil {
		return err
	}

	payment.MessageID = msg.MessageID
	return nil
}

func (c *Client) Complete(checkoutID string, err error) error {
	v := url.Values{}
	v.Add("pre_checkout_query_id", checkoutID)
	v.Add("ok", strconv.FormatBool(err == nil))
	if err != nil {
		v.Add("error_message", err.Error())
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

func (c *Client) sendPaymentWithFiscal(p *Payment) error {
	vals := p.Values()
	vals.Add("provider_token", c.paymentToken)

	resp, err := c.api.MakeRequest("sendInvoice", vals)
	if err != nil {
		return err
	}
	if !resp.Ok {
		return errors.New(resp.Description)
	}
	return nil
}
