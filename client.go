package tg_client

import (
	"context"
	"errors"
	tgbotapi "github.com/Syfaro/telegram-bot-api"
	"net/url"
	"strconv"
	"sync"
)

const timeout = 60

type Client struct {
	api          *tgbotapi.BotAPI
	updateCh     tgbotapi.UpdatesChannel
	token        string
	paymentToken string
	withFiscal   bool

	waitingMessage sync.Map
	lastBotMessage sync.Map
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
		token:    token,
		updateCh: updateChannel,
	}, nil
}

func (c *Client) WithPayments(paymentToken string, withFiscal bool) *Client {
	c.paymentToken = paymentToken
	c.withFiscal = withFiscal
	return c
}

func (c *Client) Listen(ctx context.Context, processFn func(ctx context.Context, msg IncomingMessage)) {
	for update := range c.updateCh {
		processFn(ctx, c.parseMessage(update))
	}
}

func (c *Client) SendMessage(msg *OutgoingMessage) error {
	switch msg.Type {
	case MessageDelete:
		return c.deleteMessage(msg)
	case MessageEdit:
		return c.editMessage(msg)
	default:
		return c.createMessage(msg)
	}
}

func (c *Client) SendPayment(p *Payment) error {
	if c.paymentToken == "" {
		return errors.New("payment token not set")
	}
	if c.withFiscal {
		return c.sendPaymentWithFiscal(p)
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

func (c *Client) createMessage(msg *OutgoingMessage) error {
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
	c.lastBotMessage.Store(msg.UserID, msg)
	return nil
}

func (c *Client) editMessage(msg *OutgoingMessage) error {
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

func (c *Client) deleteMessage(msg *OutgoingMessage) error {
	tgMsg := tgbotapi.NewDeleteMessage(msg.ChatID, msg.ReplyMessageID)
	_, err := c.api.Send(tgMsg)
	return err
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
