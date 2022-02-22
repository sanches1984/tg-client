//+build payment

package tg_client

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"testing"
)

func TestSendFile(t *testing.T) {
	client, err := New("token")
	require.NoError(t, err)
	data, err := ioutil.ReadFile("README.md")
	require.NoError(t, err)

	err = client.SendMessage(&OutgoingMessage{
		Type:   MessageDefault,
		ChatID: 151524701,
		UserID: 151524701,
		File: &tgbotapi.FileBytes{
			Name:  "file.txt",
			Bytes: data,
		},
	})
	require.NoError(t, err)
}
