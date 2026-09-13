package discord

import (
	"fmt"
	"strings"

	"github.com/Chance093/roomie-bills/internal/db"
	"github.com/Chance093/roomie-bills/internal/lib/plaid"
	"github.com/Chance093/roomie-bills/internal/types"
)

func (c *Client) SendHostedLink(roomie, hostedLink string) error {
	messageOne := fmt.Sprintf("A link has been requested for %s.\n", roomie)
	messageTwo := fmt.Sprintf("Plaid link: %s", hostedLink)
	finalMessage := messageOne + messageTwo

	if _, err := c.ChannelMessageSend(c.channelId, finalMessage); err != nil {
		return err
	}

	fmt.Println("Sent hosted link to discord channel")

	return nil
}

type SplitBill struct {
	plaid.Bill
	Split float64
}

// TODO: combine this with send no bills message
func (c *Client) SendBills(bills []types.SplitBill) error {
	var b strings.Builder
	b.WriteString("```")
	b.WriteString("New Bills:\n\n")

	for _, bill := range bills {
		// TODO: fix help messages
		b.WriteString(fmt.Sprintf("#️⃣ Bill ID: %d\n", bill.Id)) // first space is an emoji
		b.WriteString(fmt.Sprintf("📋 New Bill: %s\n", bill.Payee))
		b.WriteString(fmt.Sprintf("📅 Date: %s\n", bill.Date))
		b.WriteString(fmt.Sprintf("💰 Total: $%.2f\n", bill.Total))
		b.WriteString(fmt.Sprintf("👤 Each roommate owes %s: $%.2f\n", bill.Roomie, bill.Split))
		b.WriteString("\n")
	}

	b.WriteString("Type command /paid in the channel followed by the bill id once you have paid back your roomie!\n")
	b.WriteString("```")

	if _, err := c.ChannelMessageSend(c.channelId, b.String()); err != nil {
		return err
	}

	return nil
}

func (c *Client) SendOutstandingBills(outstandingBills []db.OutstandingBill) error {
	if len(outstandingBills) == 0 {
		message := "```All previous bills caught up :)```"

		if _, err := c.ChannelMessageSend(c.channelId, message); err != nil {
			return err
		}

		return nil
	}

	var b strings.Builder
	b.WriteString("```")
	b.WriteString("Outstanding Bills:\n\n")

	for _, bill := range outstandingBills {
		// TODO: fix help messages
		b.WriteString(fmt.Sprintf("#️⃣ Bill ID: %d\n", bill.Id)) // first space is an emoji
		b.WriteString(fmt.Sprintf("📋 Bill Name: %s\n", bill.Name))
		b.WriteString(fmt.Sprintf("💰 Total: $%.2f\n", bill.Amount))
		b.WriteString("😠 ")

		for i, payer := range bill.Payers {
			if len(bill.Payers) == 1 {
				b.WriteString(fmt.Sprintf("%s ", payer))
			} else {
				if i == len(bill.Payers)-1 {
					b.WriteString(fmt.Sprintf("and %s ", payer))
				} else {
					if len(bill.Payers) == 2 {
						b.WriteString(fmt.Sprintf("%s ", payer))
					} else {
						b.WriteString(fmt.Sprintf("%s, ", payer))
					}
				}
			}
		}

		if len(bill.Payers) == 1 {
			b.WriteString(fmt.Sprintf("still owes %s money!\n", bill.Payee))
		} else {
			b.WriteString(fmt.Sprintf("still owe %s money!\n", bill.Payee))
		}
		b.WriteString("\n")
	}

	b.WriteString("Type command /paid in the channel followed by the bill id once you have paid back your roomie!\n")
	b.WriteString("```")

	if _, err := c.ChannelMessageSend(c.channelId, b.String()); err != nil {
		return err
	}

	return nil
}

func (c *Client) SendNoNewBillsMessage() error {
	message := "```No new bills :)```"

	if _, err := c.ChannelMessageSend(c.channelId, message); err != nil {
		return err
	}

	return nil
}
