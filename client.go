package tg_client

import (
	"context"
	"errors"
	tgbotapi "github.com/Syfaro/telegram-bot-api"
	"net/url"
	"strconv"
)

const timeout = 60

type Client struct {
	api          *tgbotapi.BotAPI
	updateCh     tgbotapi.UpdatesChannel
	paymentToken string

	waitingMessage map[int]WaitData
	lastBotMessage map[int]OutgoingMessage
}

func New(token string) (*Client, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = timeout

	updateChannel, err := api.GetUpdatesChan(u)
	if err != nil {
		return nil, err
	}

	return &Client{
		api:      api,
		updateCh: updateChannel,
	}, nil
}

func (c *Client) WithPayments(paymentToken string) *Client {
	c.paymentToken = paymentToken
	return c
}

func (c *Client) Listen(ctx context.Context, processFn func(ctx context.Context, msg IncomingMessage)) {
	for update := range c.updateCh {
		processFn(ctx, c.parseMessage(update))
	}
}

func (c *Client) SendMessage(msg *OutgoingMessage) error {
	tgMsg := tgbotapi.NewMessage(msg.ChatID, msg.Message)
	tgMsg.ReplyMarkup = msg.Markup
	if msg.Formatted {
		tgMsg.ParseMode = parseModeMarkdown
	}

	m, err := c.api.Send(tgMsg)
	if err != nil {
		return err
	}
	msg.ID = m.MessageID
	return nil
}

func (c *Client) EditMessage(msg *OutgoingMessage) error {
	tgMsg := tgbotapi.NewEditMessageText(msg.ChatID, msg.ReplyMessageID, msg.Message)
	if msg.Formatted {
		tgMsg.ParseMode = parseModeMarkdown
	}
	m, err := c.api.Send(tgMsg)
	if err != nil {
		return err
	}
	msg.ID = m.MessageID
	return nil
}

func (c *Client) DeleteMessage(chatID int64, msgID int) error {
	msg := tgbotapi.NewDeleteMessage(chatID, msgID)
	_, err := c.api.Send(msg)
	return err
}

func (c *Client) SendPayment(p *Payment) error {
	if c.paymentToken == "" {
		return errors.New("payment token not set")
	}

	prices := []tgbotapi.LabeledPrice{{Label: "руб.", Amount: p.Amount}}
	invoice := tgbotapi.NewInvoice(p.ChatID, p.Title, p.Description, p.Payload, c.paymentToken, "", currencyRUB, &prices)
	msg, err := c.api.Send(invoice)
	if err != nil {
		return err
	}

	p.MessageID = msg.MessageID
	return nil
}

func (c *Client) SendPaymentWithFiscal(p *Payment) error {
	if c.paymentToken == "" {
		return errors.New("payment token not set")
	}

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

func (c *Client) CompletePayment(checkoutID string, err error) error {
	if c.paymentToken == "" {
		return errors.New("payment token not set")
	}

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
