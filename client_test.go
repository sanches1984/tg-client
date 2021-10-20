//+build payment

package tg_client

import (
	tgbotapi "github.com/Syfaro/telegram-bot-api"
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
