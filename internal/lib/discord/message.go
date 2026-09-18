package discord

import (
	"fmt"
	"strings"

	"github.com/Chance093/roomie-bills/internal/db"
	"github.com/Chance093/roomie-bills/internal/lib/plaid"
	"github.com/Chance093/roomie-bills/internal/types"
)

func (c *Client) SendHostedLink(roomie, hostedLink string) error {
	var b strings.Builder
	fmt.Fprintf(&b, "A link has been requested for %s.\n", roomie)
	fmt.Fprintf(&b, "Plaid link: %s", hostedLink)

	if _, err := c.ChannelMessageSend(c.channelId, b.String()); err != nil {
		return err
	}

	fmt.Println("Sent hosted link to discord channel")

	return nil
}

type SplitBill struct {
	plaid.Bill
	Split float64
}

func (c *Client) SendBills(bills []types.SplitBill) error {
	var b strings.Builder
	b.WriteString("```")
	b.WriteString("New Bills:\n\n")

	for _, bill := range bills {
		fmt.Fprintf(&b, "#️⃣ Bill ID: %d\n", bill.Id) // first space is an emoji
		fmt.Fprintf(&b, "📋 New Bill: %s\n", bill.Payee)
		fmt.Fprintf(&b, "📅 Date: %s\n", bill.Date)
		fmt.Fprintf(&b, "💰 Total: $%.2f\n", bill.Total)
		fmt.Fprintf(&b, "👤 Each roommate owes %s: $%.2f\n", bill.Roomie, bill.Split)
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
		fmt.Fprintf(&b, "#️⃣ Bill ID: %d\n", bill.Id) // first space is an emoji
		fmt.Fprintf(&b, "📋 Bill Name: %s\n", bill.Name)
		fmt.Fprintf(&b, "💰 Total: $%.2f\n", bill.Amount)
		b.WriteString("😠 ")

		for i, payer := range bill.Payers {
			if len(bill.Payers) == 1 {
				fmt.Fprintf(&b, "%s ", payer)
			} else {
				if i == len(bill.Payers)-1 {
					fmt.Fprintf(&b, "and %s ", payer)
				} else {
					if len(bill.Payers) == 2 {
						fmt.Fprintf(&b, "%s ", payer)
					} else {
						fmt.Fprintf(&b, "%s, ", payer)
					}
				}
			}
		}

		if len(bill.Payers) == 1 {
			fmt.Fprintf(&b, "still owes %s money!\n", bill.Payee)
		} else {
			fmt.Fprintf(&b, "still owe %s money!\n", bill.Payee)
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
