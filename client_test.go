package tg_client

import (
	"context"
	"fmt"
	"testing"
)

func TestSuccess(t *testing.T) {
	client, _ := New("1926987692:AAETyEmB8gfLzMdOQnOg8pfT3yymp1kQn4I")
	client.Listen(context.Background(), func(ctx context.Context, msg IncomingMessage) {
		fmt.Println(msg.FileURL)
	})
}
