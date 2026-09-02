package domain

// BotAPIChatInfo is the protocol-neutral projection behind the Bot API
// getChat method. Type follows the official values: "private", "channel",
// or "supergroup" (this deployment has no separate small-group chat kind:
// every group is a megagroup/supergroup).
type BotAPIChatInfo struct {
	ID        int64
	Type      string
	Title     string
	Username  string
	FirstName string
	LastName  string
}
