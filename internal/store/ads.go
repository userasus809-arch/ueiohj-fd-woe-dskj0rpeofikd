package store

import (
	"context"
	"time"

	"telesrv/internal/domain"
)

// AdCampaignStore persists ad campaigns and serves the selection query
// messages.getSponsoredMessages needs (an active campaign eligible for a
// given channel). Counters are incremented via dedicated atomic methods
// rather than full-row updates so concurrent impressions never lose
// increments to a stale read-modify-write.
type AdCampaignStore interface {
	CreateAdCampaign(ctx context.Context, campaign domain.AdCampaign) (domain.AdCampaign, error)
	UpdateAdCampaign(ctx context.Context, campaign domain.AdCampaign) (domain.AdCampaign, error)
	GetAdCampaign(ctx context.Context, id int64) (domain.AdCampaign, bool, error)
	ListAdCampaigns(ctx context.Context, limit, offset int) ([]domain.AdCampaign, error)
	DeleteAdCampaign(ctx context.Context, id int64) (bool, error)

	// SelectActiveAdCampaign returns one active campaign eligible for
	// channelID (untargeted, or explicitly targeting it), or found=false
	// if none is eligible. Implementations rotate among eligible
	// campaigns (e.g. least-recently-served first) rather than always
	// returning the same one.
	SelectActiveAdCampaign(ctx context.Context, channelID int64, now time.Time) (domain.AdCampaign, bool, error)

	IncrementAdCampaignImpressions(ctx context.Context, id int64) error
	IncrementAdCampaignViews(ctx context.Context, id int64) error
	IncrementAdCampaignClicks(ctx context.Context, id int64) error
}
