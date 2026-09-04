package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"telesrv/internal/domain"
	"telesrv/internal/store"
	"telesrv/internal/store/postgres/sqlcgen"
)

type AdCampaignStore struct {
	db sqlcgen.DBTX
}

func NewAdCampaignStore(db sqlcgen.DBTX) *AdCampaignStore {
	return &AdCampaignStore{db: db}
}

var _ store.AdCampaignStore = (*AdCampaignStore)(nil)

const adCampaignColumns = `
id, title, message, entities, button_text, url, sponsor_info, active,
target_channel_ids, impression_count, view_count, click_count,
created_by, created_at, updated_at`

func scanAdCampaign(row pgx.Row, c *domain.AdCampaign) error {
	var entitiesJSON string
	if err := row.Scan(
		&c.ID, &c.Title, &c.Message, &entitiesJSON, &c.ButtonText, &c.URL, &c.SponsorInfo, &c.Active,
		&c.TargetChannelIDs, &c.ImpressionCount, &c.ViewCount, &c.ClickCount,
		&c.CreatedBy, &c.CreatedAt, &c.UpdatedAt,
	); err != nil {
		return err
	}
	entities, err := decodeMessageEntities(entitiesJSON)
	if err != nil {
		return fmt.Errorf("decode ad campaign entities: %w", err)
	}
	c.Entities = entities
	return nil
}

func (s *AdCampaignStore) CreateAdCampaign(ctx context.Context, campaign domain.AdCampaign) (domain.AdCampaign, error) {
	if err := campaign.Validate(); err != nil {
		return domain.AdCampaign{}, err
	}
	entitiesJSON, err := encodeMessageEntities(campaign.Entities)
	if err != nil {
		return domain.AdCampaign{}, fmt.Errorf("encode ad campaign entities: %w", err)
	}
	var out domain.AdCampaign
	row := s.db.QueryRow(ctx, `
INSERT INTO ad_campaigns (
  title, message, entities, button_text, url, sponsor_info, active,
  target_channel_ids, created_by
) VALUES ($1, $2, $3::jsonb, $4, $5, $6, $7, $8, $9)
RETURNING `+adCampaignColumns,
		campaign.Title, campaign.Message, string(entitiesJSON), campaign.ButtonText, campaign.URL,
		campaign.SponsorInfo, campaign.Active, campaign.TargetChannelIDs, campaign.CreatedBy,
	)
	if err := scanAdCampaign(row, &out); err != nil {
		return domain.AdCampaign{}, fmt.Errorf("insert ad campaign: %w", err)
	}
	return out, nil
}

func (s *AdCampaignStore) UpdateAdCampaign(ctx context.Context, campaign domain.AdCampaign) (domain.AdCampaign, error) {
	if campaign.ID <= 0 {
		return domain.AdCampaign{}, domain.ErrAdCampaignNotFound
	}
	if err := campaign.Validate(); err != nil {
		return domain.AdCampaign{}, err
	}
	entitiesJSON, err := encodeMessageEntities(campaign.Entities)
	if err != nil {
		return domain.AdCampaign{}, fmt.Errorf("encode ad campaign entities: %w", err)
	}
	var out domain.AdCampaign
	row := s.db.QueryRow(ctx, `
UPDATE ad_campaigns SET
  title = $2, message = $3, entities = $4::jsonb, button_text = $5, url = $6,
  sponsor_info = $7, active = $8, target_channel_ids = $9, updated_at = now()
WHERE id = $1
RETURNING `+adCampaignColumns,
		campaign.ID, campaign.Title, campaign.Message, string(entitiesJSON), campaign.ButtonText,
		campaign.URL, campaign.SponsorInfo, campaign.Active, campaign.TargetChannelIDs,
	)
	if err := scanAdCampaign(row, &out); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.AdCampaign{}, domain.ErrAdCampaignNotFound
		}
		return domain.AdCampaign{}, fmt.Errorf("update ad campaign: %w", err)
	}
	return out, nil
}

func (s *AdCampaignStore) GetAdCampaign(ctx context.Context, id int64) (domain.AdCampaign, bool, error) {
	var out domain.AdCampaign
	row := s.db.QueryRow(ctx, `SELECT `+adCampaignColumns+` FROM ad_campaigns WHERE id = $1`, id)
	if err := scanAdCampaign(row, &out); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.AdCampaign{}, false, nil
		}
		return domain.AdCampaign{}, false, fmt.Errorf("get ad campaign: %w", err)
	}
	return out, true, nil
}

func (s *AdCampaignStore) ListAdCampaigns(ctx context.Context, limit, offset int) ([]domain.AdCampaign, error) {
	if limit <= 0 || limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}
	rows, err := s.db.Query(ctx, `
SELECT `+adCampaignColumns+`
FROM ad_campaigns
ORDER BY id DESC
LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list ad campaigns: %w", err)
	}
	defer rows.Close()
	var out []domain.AdCampaign
	for rows.Next() {
		var c domain.AdCampaign
		if err := scanAdCampaign(rows, &c); err != nil {
			return nil, fmt.Errorf("scan ad campaign: %w", err)
		}
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list ad campaigns: %w", err)
	}
	return out, nil
}

func (s *AdCampaignStore) DeleteAdCampaign(ctx context.Context, id int64) (bool, error) {
	tag, err := s.db.Exec(ctx, `DELETE FROM ad_campaigns WHERE id = $1`, id)
	if err != nil {
		return false, fmt.Errorf("delete ad campaign: %w", err)
	}
	return tag.RowsAffected() > 0, nil
}

func (s *AdCampaignStore) SelectActiveAdCampaign(ctx context.Context, channelID int64, _ time.Time) (domain.AdCampaign, bool, error) {
	var out domain.AdCampaign
	row := s.db.QueryRow(ctx, `
SELECT `+adCampaignColumns+`
FROM ad_campaigns
WHERE active AND (target_channel_ids = '{}' OR $1 = ANY(target_channel_ids))
ORDER BY impression_count ASC, id ASC
LIMIT 1`, channelID)
	if err := scanAdCampaign(row, &out); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.AdCampaign{}, false, nil
		}
		return domain.AdCampaign{}, false, fmt.Errorf("select active ad campaign: %w", err)
	}
	return out, true, nil
}

func (s *AdCampaignStore) IncrementAdCampaignImpressions(ctx context.Context, id int64) error {
	if _, err := s.db.Exec(ctx, `UPDATE ad_campaigns SET impression_count = impression_count + 1 WHERE id = $1`, id); err != nil {
		return fmt.Errorf("increment ad campaign impressions: %w", err)
	}
	return nil
}

func (s *AdCampaignStore) IncrementAdCampaignViews(ctx context.Context, id int64) error {
	if _, err := s.db.Exec(ctx, `UPDATE ad_campaigns SET view_count = view_count + 1 WHERE id = $1`, id); err != nil {
		return fmt.Errorf("increment ad campaign views: %w", err)
	}
	return nil
}

func (s *AdCampaignStore) IncrementAdCampaignClicks(ctx context.Context, id int64) error {
	if _, err := s.db.Exec(ctx, `UPDATE ad_campaigns SET click_count = click_count + 1 WHERE id = $1`, id); err != nil {
		return fmt.Errorf("increment ad campaign clicks: %w", err)
	}
	return nil
}
