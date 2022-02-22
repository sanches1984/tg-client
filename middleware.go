package tg_client

import (
	"context"
	"github.com/rs/zerolog"
)

type Middleware func(handleFunc HandleFunc) HandleFunc

func NewLoggerMiddleware(logger zerolog.Logger) Middleware {
	return func(handleFunc HandleFunc) HandleFunc {
		return func(ctx context.Context, msg *IncomingMessage) []OutgoingMessage {
			if msg.Callback != nil {
				logger.Debug().Int64("user_id", msg.UserID).Str("callback", msg.Callback.Type).
					Str("value", msg.Callback.Value).Str("message", msg.Message).Msg("incoming callback")
			} else if msg.Payment != nil {
				logger.Debug().Int64("user_id", msg.UserID).Str("checkout_id", msg.Payment.CheckoutID).Msg("incoming payment")
			} else {
				logger.Debug().Int64("user_id", msg.UserID).Str("message", msg.Message).Msg("incoming message")
			}

			outMsg := handleFunc(ctx, msg)

			for _, m := range outMsg {
				logger.Debug().Int64("user_id", m.UserID).Str("msg", m.Message).Msg("outgoing message")
			}

			return outMsg
		}
	}
}
