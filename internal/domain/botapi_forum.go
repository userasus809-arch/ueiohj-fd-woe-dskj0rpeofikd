package domain

// BotAPIForumTopic is the protocol-neutral projection behind
// createForumTopic (and the ForumTopic object it returns).
// MessageThreadID doubles as the topic's identifying message id, matching
// the official API's own message_thread_id convention.
type BotAPIForumTopic struct {
	MessageThreadID   int
	Name              string
	IconColor         int
	IconCustomEmojiID int64
}
