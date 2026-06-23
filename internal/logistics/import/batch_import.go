package imp

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type TrackingEntry struct {
	ID             string `json:"id"`
	OrderID        string `json:"order_id,omitempty"`
	TrackingNumber string `json:"tracking_number"`
	Carrier        string `json:"carrier"`
	Status         string `json:"status"`
}

type BatchImporter struct{}

func NewBatchImporter() *BatchImporter { return &BatchImporter{} }

func (bi *BatchImporter) ImportCSV(ctx context.Context, reader io.Reader) ([]*TrackingEntry, int, error) {
	csvReader := csv.NewReader(reader)
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, 0, fmt.Errorf("read csv: %w", err)
	}

	var entries []*TrackingEntry
	imported := 0

	for i, row := range records {
		if i == 0 && isHeader(row) {
			continue // skip header
		}
		if len(row) < 2 {
			continue
		}

		entry := &TrackingEntry{
			ID:             uuid.New().String(),
			TrackingNumber: strings.TrimSpace(row[0]),
			Carrier:        strings.TrimSpace(row[1]),
			Status:         "pending",
		}
		if len(row) >= 3 {
			entry.OrderID = strings.TrimSpace(row[2])
		}

		entries = append(entries, entry)
		imported++
	}

	log.Info().Int("total", imported).Msg("tracking numbers imported")
	return entries, imported, nil
}

func (bi *BatchImporter) ImportJSON(ctx context.Context, reader io.Reader) ([]*TrackingEntry, int, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, 0, err
	}

	var entries []*TrackingEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, 0, fmt.Errorf("parse json: %w", err)
	}

	for _, e := range entries {
		e.ID = uuid.New().String()
		if e.Status == "" {
			e.Status = "pending"
		}
	}

	log.Info().Int("count", len(entries)).Msg("tracking entries imported from JSON")
	return entries, len(entries), nil
}

func isHeader(row []string) bool {
	first := strings.ToLower(strings.TrimSpace(row[0]))
	return first == "tracking_number" || first == "tracking" || first == "tn" ||
		first == "物流单号" || first == "运单号"
}
