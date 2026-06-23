package sync

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
)

type TrackingSync struct {
	bSiteAPIBase string
	bSiteAPIKey  string
}

func NewTrackingSync(bSiteAPIBase, bSiteAPIKey string) *TrackingSync {
	return &TrackingSync{
		bSiteAPIBase: bSiteAPIBase,
		bSiteAPIKey:  bSiteAPIKey,
	}
}

func (s *TrackingSync) SyncToBSite(ctx context.Context, orderID, trackingNumber, carrier, bSiteDomain string) error {
	log.Info().
		Str("order", orderID).
		Str("tracking", trackingNumber).
		Str("carrier", carrier).
		Str("b_site", bSiteDomain).
		Msg("syncing tracking to B-site")

	// In production:
	// 1. POST to B-site WooCommerce REST API
	// 2. Update order status to "shipped"
	// 3. Add tracking number to order meta
	// 4. Send customer notification email

	time.Sleep(100 * time.Millisecond) // simulate API call
	return nil
}

func (s *TrackingSync) BatchSync(ctx context.Context, entries []TrackingSyncEntry) (int, int, error) {
	synced := 0
	failed := 0

	for _, entry := range entries {
		if err := s.SyncToBSite(ctx, entry.OrderID, entry.TrackingNumber, entry.Carrier, entry.BSiteDomain); err != nil {
			log.Error().Err(err).Str("order", entry.OrderID).Msg("sync failed")
			failed++
			continue
		}
		synced++
	}

	log.Info().Int("synced", synced).Int("failed", failed).Msg("batch sync completed")
	return synced, failed, nil
}

type TrackingSyncEntry struct {
	OrderID        string
	TrackingNumber string
	Carrier        string
	BSiteDomain    string
}
