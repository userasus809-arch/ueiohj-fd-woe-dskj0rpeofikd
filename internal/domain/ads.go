package domain

import (
	"errors"
	"time"
)

const (
	MaxAdCampaignTitleLength      = 128
	MaxAdCampaignMessageBytes     = 1024
	MaxAdCampaignButtonTextLength = 64
	MaxAdCampaignSponsorInfoLen   = 128
	MaxAdCampaignURLLength        = 512
	MaxAdCampaignTargetChannels   = 500
)

// AdCampaign is an admin-managed sponsored message shown to channel
// viewers via messages.getSponsoredMessages. It intentionally has no
// billing/auction model yet (see the MVP scope note in the RPC layer);
// campaigns are simply active/inactive and, optionally, scoped to a set
// of channels.
type AdCampaign struct {
	ID       int64
	Title    string // internal admin label, never shown to viewers
	Message  string
	Entities []MessageEntity

	ButtonText  string
	URL         string
	SponsorInfo string

	Active bool
	// TargetChannelIDs, if non-empty, restricts the campaign to specific
	// channels. Empty means "eligible on every channel".
	TargetChannelIDs []int64

	ImpressionCount int64 // times served by getSponsoredMessages
	ViewCount       int64 // confirmed client-side views
	ClickCount      int64

	CreatedBy string
	CreatedAt time.Time
	UpdatedAt time.Time
}

var (
	ErrAdCampaignInvalid  = errors.New("ad campaign invalid")
	ErrAdCampaignNotFound = errors.New("ad campaign not found")
)

func (c AdCampaign) Validate() error {
	if c.Title == "" || len(c.Title) > MaxAdCampaignTitleLength {
		return ErrAdCampaignInvalid
	}
	if c.Message == "" || len(c.Message) > MaxAdCampaignMessageBytes {
		return ErrAdCampaignInvalid
	}
	if len(c.ButtonText) > MaxAdCampaignButtonTextLength {
		return ErrAdCampaignInvalid
	}
	if c.URL == "" || len(c.URL) > MaxAdCampaignURLLength {
		return ErrAdCampaignInvalid
	}
	if len(c.SponsorInfo) > MaxAdCampaignSponsorInfoLen {
		return ErrAdCampaignInvalid
	}
	if len(c.TargetChannelIDs) > MaxAdCampaignTargetChannels {
		return ErrAdCampaignInvalid
	}
	for _, id := range c.TargetChannelIDs {
		if id <= 0 {
			return ErrAdCampaignInvalid
		}
	}
	return nil
}
