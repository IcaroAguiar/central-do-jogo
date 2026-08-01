// Package cbf_match_center parses lineup data from CBF match center
// escalação JSON fixtures (redacted representations of the live page).
package cbf_match_center

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"github.com/IcaroAguiar/central-do-jogo/internal/domain"
	"github.com/IcaroAguiar/central-do-jogo/internal/sources"
)

const (
	sourceID      = "cbf_match_center"
	parserVersion = "1.0.0"
)

type playerEntry struct {
	ShirtNumber string `json:"shirt_number"`
	Name        string `json:"name"`
}

type cbfLineupFixture struct {
	SourceID           string        `json:"source_id"`
	SampleURL          string        `json:"sample_url"`
	HomeTeam           string        `json:"home_team"`
	AwayTeam           string        `json:"away_team"`
	HomeStartingSample []playerEntry `json:"home_starting_sample"`
	AwayStartingSample []playerEntry `json:"away_starting_sample"`
	BenchSample        []playerEntry `json:"bench_sample"`
	Formation          *string       `json:"formation"`
	HeadCoach          *string       `json:"head_coach"`
}

// Adapter implements sources.Adapter for CBF match center lineups.
type Adapter struct{}

func (a *Adapter) SourceID() string { return sourceID }

func (a *Adapter) Parse(_ context.Context, raw []byte) (*sources.Observation, error) {
	hash := fmt.Sprintf("%x", sha256.Sum256(raw))

	var fixture cbfLineupFixture
	if err := json.Unmarshal(raw, &fixture); err != nil {
		return nil, fmt.Errorf("cbf_match_center: failed to parse JSON: %w", err)
	}

	if fixture.SourceID != "" && fixture.SourceID != sourceID {
		return nil, fmt.Errorf("cbf_match_center: source_id mismatch: got %q", fixture.SourceID)
	}

	obs := &sources.Observation{
		SourceID:      sourceID,
		DataType:      sources.DataTypeLineup,
		ObservedAt:    time.Now().UTC(),
		ParserVersion: parserVersion,
		ContentHash:   hash,
		RawRef:        fixture.SampleURL,
	}

	var formation string
	if fixture.Formation != nil {
		formation = *fixture.Formation
	}
	var coach string
	if fixture.HeadCoach != nil {
		coach = *fixture.HeadCoach
	}

	if len(fixture.HomeStartingSample) > 0 {
		homePlayers := make([]domain.LineupPlayer, 0, len(fixture.HomeStartingSample))
		for _, p := range fixture.HomeStartingSample {
			homePlayers = append(homePlayers, domain.LineupPlayer{
				ShirtNumber: p.ShirtNumber,
				Name:        p.Name,
				IsStarter:   true,
			})
		}
		obs.Lineups = append(obs.Lineups, sources.LineupEntry{
			HomeTeam:  fixture.HomeTeam,
			AwayTeam:  fixture.AwayTeam,
			Side:      domain.LineupHome,
			Formation: formation,
			Coach:     coach,
			Players:   homePlayers,
			Official:  true,
		})
	}

	if len(fixture.AwayStartingSample) > 0 {
		awayPlayers := make([]domain.LineupPlayer, 0, len(fixture.AwayStartingSample))
		for _, p := range fixture.AwayStartingSample {
			awayPlayers = append(awayPlayers, domain.LineupPlayer{
				ShirtNumber: p.ShirtNumber,
				Name:        p.Name,
				IsStarter:   true,
			})
		}
		obs.Lineups = append(obs.Lineups, sources.LineupEntry{
			HomeTeam:  fixture.HomeTeam,
			AwayTeam:  fixture.AwayTeam,
			Side:      domain.LineupAway,
			Formation: formation,
			Coach:     coach,
			Players:   awayPlayers,
			Official:  true,
		})
	}

	if len(obs.Lineups) == 0 {
		return nil, fmt.Errorf("cbf_match_center: no lineup data found (fail-closed)")
	}

	return obs, nil
}
