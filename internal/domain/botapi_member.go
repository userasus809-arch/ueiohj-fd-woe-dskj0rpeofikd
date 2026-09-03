package domain

// BotAPIChatMember is the protocol-neutral projection behind getChatMember
// and getChatAdministrators. Status follows the official Bot API values:
// "creator", "administrator", "member", "restricted", "left", "kicked".
type BotAPIChatMember struct {
	Status      string
	UserID      int64
	Username    string
	FirstName   string
	LastName    string
	IsAnonymous bool
	CustomTitle string
	// UntilDate is set for "restricted"/"kicked" members with a temporary
	// ban (0 means permanent).
	UntilDate int

	// Administrator/creator rights (meaningful when Status is
	// "administrator" or "creator").
	CanManageChat       bool
	CanDeleteMessages   bool
	CanManageVideoChats bool
	CanRestrictMembers  bool
	CanPromoteMembers   bool
	CanChangeInfo       bool
	CanInviteUsers      bool
	CanPostMessages     bool
	CanEditMessages     bool
	CanPinMessages      bool
	CanManageTopics     bool

	// Member permissions (meaningful when Status is "member" or
	// "restricted"); true means the member IS allowed to do this, i.e.
	// already inverted from the internal banned-rights representation.
	IsMember              bool
	CanSendMessages       bool
	CanSendPolls          bool
	CanSendOtherMessages  bool
	CanAddWebPagePreviews bool
	CanPinMessagesMember  bool
	CanManageTopicsMember bool
}
