// Package cbf_official_site parses CBF schedule metadata from JSON fixture
// representations of official announcement pages.
package cbf_official_site

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"github.com/IcaroAguiar/central-do-jogo/internal/sources"
)

const (
	sourceID      = "cbf_official_site"
	parserVersion = "1.0.0"
)

// cbfMeta represents the redacted metadata fixture structure.
type cbfMeta struct {
	SourceID  string `json:"source_id"`
	SourceURL string `json:"source_url"`
	Observed  struct {
		AnnouncesTable bool     `json:"announces_basic_table_pdf"`
		Round1Dates    []string `json:"round1_dates"`
	} `json:"observed"`
}

// Adapter implements sources.Adapter for CBF official site schedule metadata.
type Adapter struct{}

func (a *Adapter) SourceID() string { return sourceID }

func (a *Adapter) Parse(_ context.Context, raw []byte) (*sources.Observation, error) {
	hash := fmt.Sprintf("%x", sha256.Sum256(raw))

	var meta cbfMeta
	if err := json.Unmarshal(raw, &meta); err != nil {
		return nil, fmt.Errorf("cbf_official_site: failed to parse meta JSON: %w", err)
	}

	if meta.SourceID != "" && meta.SourceID != sourceID {
		return nil, fmt.Errorf("cbf_official_site: source_id mismatch: got %q", meta.SourceID)
	}

	obs := &sources.Observation{
		SourceID:      sourceID,
		DataType:      sources.DataTypeSchedule,
		ObservedAt:    time.Now().UTC(),
		ParserVersion: parserVersion,
		ContentHash:   hash,
		RawRef:        meta.SourceURL,
	}

	for _, dateStr := range meta.Observed.Round1Dates {
		t, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			continue
		}
		obs.Schedules = append(obs.Schedules, sources.ScheduleEntry{
			Competition: "Brasileiro Serie A",
			Round:       "Round 1",
			KickoffAt:   &t,
		})
	}

	if len(obs.Schedules) == 0 && !meta.Observed.AnnouncesTable {
		return nil, fmt.Errorf("cbf_official_site: no schedule data found (fail-closed)")
	}

	return obs, nil
}
