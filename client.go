package tg_client

import (
	"context"
	"errors"
	tgbotapi "github.com/Syfaro/telegram-bot-api"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net/url"
	"strconv"
	"strings"
	"sync"
)

const timeout = 60

type HandleFunc func(ctx context.Context, msg IncomingMessage) []OutgoingMessage

type Client struct {
	api          *tgbotapi.BotAPI
	updateCh     tgbotapi.UpdatesChannel
	token        string
	paymentToken string
	withFiscal   bool

	waitingMessage sync.Map
	lastBotMessage sync.Map

	handlers          map[string]HandleFunc
	prepareFn         func(ctx context.Context, msg *IncomingMessage) []OutgoingMessage
	callbackFn        HandleFunc
	paymentCheckoutFn HandleFunc
	paymentChargeFn   HandleFunc
	messageFn         HandleFunc
	defaultFn         HandleFunc

	logger zerolog.Logger
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
		handlers: make(map[string]HandleFunc),
		logger:   log.Logger,
	}, nil
}

func (c *Client) WithPayments(paymentToken string, withFiscal bool) *Client {
	c.paymentToken = paymentToken
	c.withFiscal = withFiscal
	return c
}

func (c *Client) WithLogger(logger zerolog.Logger) *Client {
	c.logger = logger
	return c
}

func (c *Client) HandlePrepareMessage(handleFn func(ctx context.Context, msg *IncomingMessage) []OutgoingMessage) {
	c.prepareFn = handleFn
}

func (c *Client) HandleCommand(name string, handleFn HandleFunc) {
	c.handlers[strings.ToLower(name)] = handleFn
}

func (c *Client) HandleCallback(handleFn HandleFunc) {
	c.callbackFn = handleFn
}

func (c *Client) HandleMessage(handleFn HandleFunc) {
	c.messageFn = handleFn
}

func (c *Client) HandlePayment(checkoutPaymentFn HandleFunc, chargePaymentFn HandleFunc) {
	c.paymentCheckoutFn = checkoutPaymentFn
	c.paymentChargeFn = chargePaymentFn
}

func (c *Client) HandleDefault(handleFn HandleFunc) {
	c.defaultFn = handleFn
}

func (c *Client) Listen(ctx context.Context) {
	for update := range c.updateCh {
		go c.processMessage(ctx, update)
	}
}

func (c *Client) processMessage(ctx context.Context, update tgbotapi.Update) {
	msg := c.parseMessage(update)
	if msg.Callback != nil {
		c.logger.Debug().Int("user_id", msg.UserID).Str("callback", msg.Callback.Type).
			Str("value", msg.Callback.Value).Str("message", msg.Message).Msg("incoming callback")
	} else {
		c.logger.Debug().Int("user_id", msg.UserID).Str("message", msg.Message).Msg("incoming message")
	}

	var outMsg []OutgoingMessage
	if c.prepareFn != nil {
		outMsg = c.prepareFn(ctx, &msg)
	}

	if len(outMsg) == 0 {
		outMsg = c.getOutgoingMessages(ctx, msg)
	}

	for _, m := range outMsg {
		c.logger.Debug().Int("user_id", m.UserID).Str("msg", m.Message).Msg("outgoing message")
		if err := c.SendMessage(&m); err != nil {
			c.logger.Error().Err(err).Int("user_id", m.UserID).Str("msg", m.Message).Msg("send message error")
		}
	}
}

func (c *Client) getOutgoingMessages(ctx context.Context, msg IncomingMessage) []OutgoingMessage {
	switch msg.Type {
	case MessageCommand:
		if fn, ok := c.handlers[strings.ToLower(msg.Message)]; ok && fn != nil {
			return fn(ctx, msg)
		}
	case MessageCallback:
		if c.callbackFn != nil {
			return c.callbackFn(ctx, msg)
		}
	case MessagePaymentCheckout:
		if c.paymentCheckoutFn != nil {
			return c.paymentCheckoutFn(ctx, msg)
		}
	case MessagePaymentCharge:
		if c.paymentChargeFn != nil {
			return c.paymentChargeFn(ctx, msg)
		}
	case MessageResponse, MessageText:
		if c.messageFn != nil {
			return c.messageFn(ctx, msg)
		}
	default:
		if c.defaultFn != nil {
			return c.messageFn(ctx, msg)
		}
	}
	return nil
}

func (c *Client) SendMessage(msg *OutgoingMessage) error {
	if msg.WaitData != nil {
		c.waitingMessage.Store(msg.UserID, msg.WaitData)
	}

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
