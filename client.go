package tg_client

import (
	"context"
	"errors"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sanches1984/tg-client/payment"
	"log"
	"strings"
	"sync"
)

const timeout = 60

type HandleFunc func(ctx context.Context, msg *IncomingMessage) []OutgoingMessage

type PaymentClient interface {
	Send(payment *payment.Payment) error
	Complete(checkoutID string, err error) error
}

type Client struct {
	api      *tgbotapi.BotAPI
	updateCh tgbotapi.UpdatesChannel
	payments PaymentClient

	waitingMessage sync.Map
	lastBotMessage sync.Map

	middlewares       []Middleware
	handlers          map[string]HandleFunc
	prepareFn         HandleFunc
	callbackFn        HandleFunc
	paymentCheckoutFn HandleFunc
	paymentChargeFn   HandleFunc
	messageFn         HandleFunc
	defaultFn         HandleFunc
}

func New(token string, mw ...Middleware) (*Client, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = timeout

	return &Client{
		api:         api,
		updateCh:    api.GetUpdatesChan(u),
		handlers:    make(map[string]HandleFunc),
		middlewares: mw,
	}, nil
}

func (c *Client) InitPayments(paymentToken string, withFiscal bool) {
	c.payments = payment.New(c.api, paymentToken, withFiscal)
}

func (c *Client) HandlePrepareMessage(handleFn HandleFunc) {
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

	fn := c.processMessageFn()
	for _, mw := range c.middlewares {
		fn = mw(fn)
	}
	outMsg := fn(ctx, msg)

	for _, m := range outMsg {
		if err := c.SendMessage(&m); err != nil {
			log.Println("send message error, chat_id=", m.ChatID)
		}
	}
}

func (c *Client) processMessageFn() HandleFunc {
	return func(ctx context.Context, msg *IncomingMessage) []OutgoingMessage {
		var outMsg []OutgoingMessage
		if c.prepareFn != nil {
			outMsg = c.prepareFn(ctx, msg)
		}
		if len(outMsg) == 0 {
			outMsg = c.getOutgoingMessages(ctx, msg)
		}
		return outMsg
	}
}

func (c *Client) getOutgoingMessages(ctx context.Context, msg *IncomingMessage) []OutgoingMessage {
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

func (c *Client) SendPayment(payment *payment.Payment) error {
	if c.payments == nil {
		return errors.New("payments not initialized")
	}
	return c.payments.Send(payment)
}

func (c *Client) CompletePayment(checkoutID string, err error) error {
	if c.payments == nil {
		return errors.New("payments not initialized")
	}
	return c.payments.Complete(checkoutID, err)
}

func (c *Client) createMessage(msg *OutgoingMessage) error {
	var m tgbotapi.Message
	var err error
	if msg.File != nil {
		m, err = c.api.Send(tgbotapi.NewDocument(msg.ChatID, *msg.File))
	} else {
		tgMsg := tgbotapi.NewMessage(msg.ChatID, msg.Message)
		tgMsg.ReplyMarkup = msg.Markup
		if msg.Formatted {
			tgMsg.ParseMode = tgbotapi.ModeMarkdown
		}
		m, err = c.api.Send(tgMsg)
	}
	if err != nil {
		return err
	}

	msg.ID = m.MessageID
	if msg.File == nil {
		c.lastBotMessage.Store(msg.UserID, msg)
	}
	return nil
}

func (c *Client) editMessage(msg *OutgoingMessage) error {
	tgMsg := tgbotapi.NewEditMessageText(msg.ChatID, msg.ReplyMessageID, msg.Message)
	if msg.Formatted {
		tgMsg.ParseMode = tgbotapi.ModeMarkdown
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
