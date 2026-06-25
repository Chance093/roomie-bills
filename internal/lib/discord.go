package lib

import "fmt"

type DiscordClient struct {
	TOKEN string
}

func NewDiscordClient(token string) DiscordClient {
	return DiscordClient{TOKEN: token}
}

func (dc *DiscordClient) PostMessage(message string) {
	dc.formatMessage()
	fmt.Println("printing message to discord")
}

func (dc *DiscordClient) formatMessage() {
	fmt.Println("formatting message")
}
