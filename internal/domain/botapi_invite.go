package domain

// BotAPIChatInviteLink is the protocol-neutral projection behind
// createChatInviteLink/editChatInviteLink/revokeChatInviteLink.
type BotAPIChatInviteLink struct {
	InviteLink              string
	CreatorUserID           int64
	CreatesJoinRequest      bool
	IsPrimary               bool
	IsRevoked               bool
	Name                    string
	ExpireDate              int
	MemberLimit             int
	PendingJoinRequestCount int
}

// BotAPIInviteLinkParams carries createChatInviteLink/editChatInviteLink
// parameters. The Has* flags distinguish "the caller explicitly set this
// field" from "the caller omitted it" so editChatInviteLink can leave
// unspecified fields unchanged, matching the official Bot API contract.
type BotAPIInviteLinkParams struct {
	Name                  string
	HasName               bool
	ExpireDate            int
	HasExpireDate         bool
	MemberLimit           int
	HasMemberLimit        bool
	CreatesJoinRequest    bool
	HasCreatesJoinRequest bool
}
