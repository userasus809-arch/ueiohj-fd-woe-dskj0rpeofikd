package ads

import (
	"context"
	"strings"
	"time"

	"go.uber.org/zap"

	"telesrv/internal/domain"
	"telesrv/internal/store"
)

// Service is the admin- and RPC-facing entry point for sponsored messages:
// campaign CRUD for the admin panel, and campaign selection/counters for
// messages.getSponsoredMessages and its view/click follow-ups.
type Service struct {
	store store.AdCampaignStore
	log   *zap.Logger
	clock func() time.Time
}

func NewService(st store.AdCampaignStore, log *zap.Logger) *Service {
	if log == nil {
		log = zap.NewNop()
	}
	return &Service{store: st, log: log, clock: time.Now}
}

func (s *Service) Ready() bool { return s != nil && s.store != nil }

// CreateCampaign is the admin panel's entry point. It trims/normalizes
// input and defers full validation to domain.AdCampaign.Validate via the
// store.
func (s *Service) CreateCampaign(ctx context.Context, campaign domain.AdCampaign) (domain.AdCampaign, error) {
	if s == nil || s.store == nil {
		return domain.AdCampaign{}, domain.ErrAdCampaignInvalid
	}
	campaign.Title = strings.TrimSpace(campaign.Title)
	campaign.Message = strings.TrimSpace(campaign.Message)
	campaign.URL = strings.TrimSpace(campaign.URL)
	campaign.ButtonText = strings.TrimSpace(campaign.ButtonText)
	campaign.SponsorInfo = strings.TrimSpace(campaign.SponsorInfo)
	campaign.ID = 0
	campaign.ImpressionCount, campaign.ViewCount, campaign.ClickCount = 0, 0, 0
	if err := campaign.Validate(); err != nil {
		return domain.AdCampaign{}, err
	}
	return s.store.CreateAdCampaign(ctx, campaign)
}

// UpdateCampaign replaces a campaign's editable fields. Counters cannot be
// set this way; the store only ever mutates them via the Increment* calls.
func (s *Service) UpdateCampaign(ctx context.Context, campaign domain.AdCampaign) (domain.AdCampaign, error) {
	if s == nil || s.store == nil {
		return domain.AdCampaign{}, domain.ErrAdCampaignInvalid
	}
	if campaign.ID <= 0 {
		return domain.AdCampaign{}, domain.ErrAdCampaignNotFound
	}
	campaign.Title = strings.TrimSpace(campaign.Title)
	campaign.Message = strings.TrimSpace(campaign.Message)
	campaign.URL = strings.TrimSpace(campaign.URL)
	campaign.ButtonText = strings.TrimSpace(campaign.ButtonText)
	campaign.SponsorInfo = strings.TrimSpace(campaign.SponsorInfo)
	if err := campaign.Validate(); err != nil {
		return domain.AdCampaign{}, err
	}
	return s.store.UpdateAdCampaign(ctx, campaign)
}

func (s *Service) GetCampaign(ctx context.Context, id int64) (domain.AdCampaign, bool, error) {
	if s == nil || s.store == nil {
		return domain.AdCampaign{}, false, nil
	}
	return s.store.GetAdCampaign(ctx, id)
}

func (s *Service) ListCampaigns(ctx context.Context, limit, offset int) ([]domain.AdCampaign, error) {
	if s == nil || s.store == nil {
		return nil, nil
	}
	return s.store.ListAdCampaigns(ctx, limit, offset)
}

func (s *Service) DeleteCampaign(ctx context.Context, id int64) (bool, error) {
	if s == nil || s.store == nil {
		return false, domain.ErrAdCampaignNotFound
	}
	return s.store.DeleteAdCampaign(ctx, id)
}

// SelectSponsoredMessage picks an eligible active campaign for channelID
// (or found=false if none is eligible) and records that it was served.
// The impression counter is best-effort: a failed increment logs and does
// not stop the ad from being shown, since losing a single counter tick
// matters far less than failing to serve an already-selected ad.
func (s *Service) SelectSponsoredMessage(ctx context.Context, channelID int64) (domain.AdCampaign, bool, error) {
	if s == nil || s.store == nil {
		return domain.AdCampaign{}, false, nil
	}
	campaign, found, err := s.store.SelectActiveAdCampaign(ctx, channelID, s.clock())
	if err != nil {
		return domain.AdCampaign{}, false, err
	}
	if !found {
		return domain.AdCampaign{}, false, nil
	}
	if err := s.store.IncrementAdCampaignImpressions(ctx, campaign.ID); err != nil {
		s.log.Warn("increment ad campaign impressions", zap.Int64("campaign_id", campaign.ID), zap.Error(err))
	}
	return campaign, true, nil
}

// RecordView backs messages.viewSponsoredMessage's server-side counter.
func (s *Service) RecordView(ctx context.Context, campaignID int64) error {
	if s == nil || s.store == nil || campaignID <= 0 {
		return nil
	}
	return s.store.IncrementAdCampaignViews(ctx, campaignID)
}

// RecordClick backs messages.clickSponsoredMessage's server-side counter.
func (s *Service) RecordClick(ctx context.Context, campaignID int64) error {
	if s == nil || s.store == nil || campaignID <= 0 {
		return nil
	}
	return s.store.IncrementAdCampaignClicks(ctx, campaignID)
}
