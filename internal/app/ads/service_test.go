package ads

import (
	"context"
	"testing"
	"time"

	"telesrv/internal/domain"
)

type fakeAdCampaignStore struct {
	campaigns map[int64]domain.AdCampaign
	nextID    int64
}

func newFakeAdCampaignStore() *fakeAdCampaignStore {
	return &fakeAdCampaignStore{campaigns: map[int64]domain.AdCampaign{}}
}

func (f *fakeAdCampaignStore) CreateAdCampaign(_ context.Context, c domain.AdCampaign) (domain.AdCampaign, error) {
	if err := c.Validate(); err != nil {
		return domain.AdCampaign{}, err
	}
	f.nextID++
	c.ID = f.nextID
	c.CreatedAt = time.Unix(1700000000, 0)
	c.UpdatedAt = c.CreatedAt
	f.campaigns[c.ID] = c
	return c, nil
}

func (f *fakeAdCampaignStore) UpdateAdCampaign(_ context.Context, c domain.AdCampaign) (domain.AdCampaign, error) {
	existing, ok := f.campaigns[c.ID]
	if !ok {
		return domain.AdCampaign{}, domain.ErrAdCampaignNotFound
	}
	if err := c.Validate(); err != nil {
		return domain.AdCampaign{}, err
	}
	c.ImpressionCount, c.ViewCount, c.ClickCount = existing.ImpressionCount, existing.ViewCount, existing.ClickCount
	c.CreatedAt = existing.CreatedAt
	c.UpdatedAt = time.Unix(1700000100, 0)
	f.campaigns[c.ID] = c
	return c, nil
}

func (f *fakeAdCampaignStore) GetAdCampaign(_ context.Context, id int64) (domain.AdCampaign, bool, error) {
	c, ok := f.campaigns[id]
	return c, ok, nil
}

func (f *fakeAdCampaignStore) ListAdCampaigns(_ context.Context, _, _ int) ([]domain.AdCampaign, error) {
	out := make([]domain.AdCampaign, 0, len(f.campaigns))
	for _, c := range f.campaigns {
		out = append(out, c)
	}
	return out, nil
}

func (f *fakeAdCampaignStore) DeleteAdCampaign(_ context.Context, id int64) (bool, error) {
	if _, ok := f.campaigns[id]; !ok {
		return false, nil
	}
	delete(f.campaigns, id)
	return true, nil
}

func (f *fakeAdCampaignStore) SelectActiveAdCampaign(_ context.Context, channelID int64, _ time.Time) (domain.AdCampaign, bool, error) {
	var best *domain.AdCampaign
	for id := range f.campaigns {
		c := f.campaigns[id]
		if !c.Active {
			continue
		}
		if len(c.TargetChannelIDs) > 0 {
			eligible := false
			for _, target := range c.TargetChannelIDs {
				if target == channelID {
					eligible = true
					break
				}
			}
			if !eligible {
				continue
			}
		}
		if best == nil || c.ImpressionCount < best.ImpressionCount {
			cc := c
			best = &cc
		}
	}
	if best == nil {
		return domain.AdCampaign{}, false, nil
	}
	return *best, true, nil
}

func (f *fakeAdCampaignStore) IncrementAdCampaignImpressions(_ context.Context, id int64) error {
	c := f.campaigns[id]
	c.ImpressionCount++
	f.campaigns[id] = c
	return nil
}

func (f *fakeAdCampaignStore) IncrementAdCampaignViews(_ context.Context, id int64) error {
	c := f.campaigns[id]
	c.ViewCount++
	f.campaigns[id] = c
	return nil
}

func (f *fakeAdCampaignStore) IncrementAdCampaignClicks(_ context.Context, id int64) error {
	c := f.campaigns[id]
	c.ClickCount++
	f.campaigns[id] = c
	return nil
}

func validCampaign() domain.AdCampaign {
	return domain.AdCampaign{
		Title:   "Spring promo",
		Message: "Check out our new product!",
		URL:     "https://example.com",
		Active:  true,
	}
}

func TestCreateCampaignTrimsAndValidates(t *testing.T) {
	svc := NewService(newFakeAdCampaignStore(), nil)
	campaign := validCampaign()
	campaign.Title = "  Spring promo  "

	created, err := svc.CreateCampaign(context.Background(), campaign)
	if err != nil {
		t.Fatalf("CreateCampaign: %v", err)
	}
	if created.ID == 0 || created.Title != "Spring promo" {
		t.Fatalf("created = %#v", created)
	}

	if _, err := svc.CreateCampaign(context.Background(), domain.AdCampaign{}); err == nil {
		t.Fatalf("expected validation error for empty campaign")
	}
}

func TestSelectSponsoredMessageRotatesAndCountsImpressions(t *testing.T) {
	store := newFakeAdCampaignStore()
	svc := NewService(store, nil)
	ctx := context.Background()

	a, err := svc.CreateCampaign(ctx, validCampaign())
	if err != nil {
		t.Fatalf("create a: %v", err)
	}
	untargeted := validCampaign()
	untargeted.Title = "Untargeted"
	b, err := svc.CreateCampaign(ctx, untargeted)
	if err != nil {
		t.Fatalf("create b: %v", err)
	}

	selected, found, err := svc.SelectSponsoredMessage(ctx, 555)
	if err != nil || !found {
		t.Fatalf("select: found=%v err=%v", found, err)
	}
	if selected.ID != a.ID && selected.ID != b.ID {
		t.Fatalf("unexpected campaign selected: %#v", selected)
	}

	stored, _, _ := store.GetAdCampaign(ctx, selected.ID)
	if stored.ImpressionCount != 1 {
		t.Fatalf("impression count = %d, want 1", stored.ImpressionCount)
	}
}

func TestSelectSponsoredMessageRespectsChannelTargeting(t *testing.T) {
	store := newFakeAdCampaignStore()
	svc := NewService(store, nil)
	ctx := context.Background()

	targeted := validCampaign()
	targeted.TargetChannelIDs = []int64{999}
	created, err := svc.CreateCampaign(ctx, targeted)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if _, found, err := svc.SelectSponsoredMessage(ctx, 111); err != nil || found {
		t.Fatalf("expected no eligible campaign for untargeted channel, found=%v err=%v", found, err)
	}
	selected, found, err := svc.SelectSponsoredMessage(ctx, 999)
	if err != nil || !found || selected.ID != created.ID {
		t.Fatalf("expected targeted campaign for channel 999: found=%v err=%v selected=%#v", found, err, selected)
	}
}

func TestSelectSponsoredMessageNoneEligible(t *testing.T) {
	svc := NewService(newFakeAdCampaignStore(), nil)
	if _, found, err := svc.SelectSponsoredMessage(context.Background(), 1); err != nil || found {
		t.Fatalf("expected no campaigns: found=%v err=%v", found, err)
	}
}

func TestRecordViewAndClickIncrementCounters(t *testing.T) {
	store := newFakeAdCampaignStore()
	svc := NewService(store, nil)
	ctx := context.Background()

	created, err := svc.CreateCampaign(ctx, validCampaign())
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := svc.RecordView(ctx, created.ID); err != nil {
		t.Fatalf("record view: %v", err)
	}
	if err := svc.RecordClick(ctx, created.ID); err != nil {
		t.Fatalf("record click: %v", err)
	}
	stored, _, _ := store.GetAdCampaign(ctx, created.ID)
	if stored.ViewCount != 1 || stored.ClickCount != 1 {
		t.Fatalf("stored = %#v", stored)
	}
}

func TestUpdateCampaignPreservesCounters(t *testing.T) {
	store := newFakeAdCampaignStore()
	svc := NewService(store, nil)
	ctx := context.Background()

	created, err := svc.CreateCampaign(ctx, validCampaign())
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := store.IncrementAdCampaignImpressions(ctx, created.ID); err != nil {
		t.Fatalf("increment: %v", err)
	}

	updated := created
	updated.Title = "Renamed"
	result, err := svc.UpdateCampaign(ctx, updated)
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if result.Title != "Renamed" || result.ImpressionCount != 1 {
		t.Fatalf("result = %#v", result)
	}
}

func TestDeleteCampaign(t *testing.T) {
	store := newFakeAdCampaignStore()
	svc := NewService(store, nil)
	ctx := context.Background()

	created, err := svc.CreateCampaign(ctx, validCampaign())
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	deleted, err := svc.DeleteCampaign(ctx, created.ID)
	if err != nil || !deleted {
		t.Fatalf("delete: deleted=%v err=%v", deleted, err)
	}
	if _, found, _ := svc.GetCampaign(ctx, created.ID); found {
		t.Fatalf("campaign should be gone")
	}
}
