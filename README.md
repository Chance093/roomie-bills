# Roomie Bills

## TODO
- [x] Save account info to db using bgjobs
- [x] Create plaid funcs to obtain transactions from all accounts
- [x] Create funcs to parse through transactions and find bills
- [x] Create funcs to split bills 4 ways
- [x] Save bills to db
- [x] Create discord package
- [ ] Convert cron jobs to use message queue (for retries)
- [ ] Use decimal package to track money for bills
- [ ] Give proper info to bills discord message (bill id, roomie)
- [ ] Create cron job to send weekly reminders for previous bills
- [ ] Create api endpoint that marks a bill paid by a roomie
- [ ] Set up discord command that hits above api endpoint
- [ ] Do all todos
### Message Format

Your format is a good start, but I'd expand it a bit to make it more actionable:

```
📋 New Bill — Cox Internet
📅 June 25, 2026
💰 Total: $90.00
👤 Each roommate owes Chance: $22.50

React with ✅ when you've paid Chance back!
— Kane | Alex | Madison
```

Tagging the roommates directly (e.g. @Kane) ensures they get a notification even if they don't check the channel often.

### Tracking Payments — Reactions vs. Something Better

Reactions are simple but have a real problem: anyone can react, there's no timestamp, and it's easy to accidentally remove a reaction.

A better approach would be a /paid slash command the bot listens for:

- A roommate types /paid Cox Internet June
- The bot updates the original message or posts a reply confirming who has paid and who hasn't
- The bot keeps an internal record with timestamps

This gives you an actual audit trail rather than relying on emoji state.

### End-of-Month Summary

Have the bot auto-post a summary at the end of each month listing any unpaid balances — essentially a nudge for anyone who forgot. Something like:
```
📊 End of Month Summary — June 2026

Still outstanding:
• Alex owes Chance $22.50 (Cox Internet)
• Kane owes Chance $54.55 (Electric)

Use /paid to mark these resolved!
```

### Workflow

#### Posting a bill
1. runs cronjob every saturday to check for any new bills 
2. Request for all transactions in the past 1-2 weeks from plaid api
3. Find all transactions that are bills (cox, electric, gas, water, rent)
4. Compare those transactions with what we have in db to see if any of them have already been accounted for 
5. Take unaccounted transactions and split bill 4 ways (electric will be done a bit differently)
6. Create a formatted discord message for each bill
7. Have slack bot post discord message to discord channel
8. Save transactions to db (total, split, bill type, payee, etc.)

#### Marking a bill paid
1. User types /paid (bill)
2. Sends a post request to server 
3. Server marks the bill as paid in the db

#### End of month summary
1. cron job runs every saturday
2. Checks database for all unpaid bills 
3. Creates a formatted message showing which bills haven't been paid by who
4. Have slack bot post discord message to discord channel
