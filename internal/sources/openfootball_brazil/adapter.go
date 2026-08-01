// Package openfootball_brazil parses Football.TXT schedule files from
// the openfootball/brazil GitHub repository.
package openfootball_brazil

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/IcaroAguiar/central-do-jogo/internal/sources"
)

const (
	sourceID      = "openfootball_brazil"
	parserVersion = "1.0.0"
)

var (
	matchdayRe    = regexp.MustCompile(`^▪ Matchday (\d+)`)
	dateRe        = regexp.MustCompile(`^\s{2}(\w{3} \w{3} \d{1,2} \d{4})`)
	matchTimeRe   = regexp.MustCompile(`^\s+(\d{1,2}:\d{2})\s+(.+?)\s+v\s+(.+?)\s{2,}`)
	matchNoTimeRe = regexp.MustCompile(`^\s{4,}([A-Z].+?)\s+v\s+(.+?)\s{2,}`)
)

// Adapter implements sources.Adapter for openfootball Football.TXT files.
type Adapter struct{}

func (a *Adapter) SourceID() string { return sourceID }

func (a *Adapter) Parse(_ context.Context, raw []byte) (*sources.Observation, error) {
	hash := fmt.Sprintf("%x", sha256.Sum256(raw))

	obs := &sources.Observation{
		SourceID:      sourceID,
		DataType:      sources.DataTypeSchedule,
		ObservedAt:    time.Now().UTC(),
		ParserVersion: parserVersion,
		ContentHash:   hash,
		RawRef:        "openfootball/brazil Football.TXT",
	}

	scanner := bufio.NewScanner(bytes.NewReader(raw))
	var currentRound string
	var currentDate string
	var lastTime string

	for scanner.Scan() {
		line := scanner.Text()

		if m := matchdayRe.FindStringSubmatch(line); m != nil {
			currentRound = "Matchday " + m[1]
			continue
		}

		if m := dateRe.FindStringSubmatch(line); m != nil {
			currentDate = m[1]
			lastTime = ""
			continue
		}

		if m := matchTimeRe.FindStringSubmatch(line); m != nil {
			lastTime = m[1]
			homeTeam := strings.TrimSpace(m[2])
			awayTeam := strings.TrimSpace(m[3])

			entry := sources.ScheduleEntry{
				HomeTeam:    homeTeam,
				AwayTeam:    awayTeam,
				Round:       currentRound,
				Competition: "Brasileiro Serie A",
			}

			if currentDate != "" && lastTime != "" {
				if t, err := parseKickoff(currentDate, lastTime); err == nil {
					entry.KickoffAt = &t
				}
			}

			obs.Schedules = append(obs.Schedules, entry)
			continue
		}

		if m := matchNoTimeRe.FindStringSubmatch(line); m != nil {
			homeTeam := strings.TrimSpace(m[1])
			awayTeam := strings.TrimSpace(m[2])

			entry := sources.ScheduleEntry{
				HomeTeam:    homeTeam,
				AwayTeam:    awayTeam,
				Round:       currentRound,
				Competition: "Brasileiro Serie A",
			}

			if currentDate != "" && lastTime != "" {
				if t, err := parseKickoff(currentDate, lastTime); err == nil {
					entry.KickoffAt = &t
				}
			}

			obs.Schedules = append(obs.Schedules, entry)
		}
	}

	return obs, nil
}

func parseKickoff(dateStr, timeStr string) (time.Time, error) {
	loc, _ := time.LoadLocation("America/Sao_Paulo")
	full := dateStr + " " + timeStr
	t, err := time.ParseInLocation("Mon Jan 2 2006 15:04", full, loc)
	if err != nil {
		return time.Time{}, err
	}
	return t.UTC(), nil
}
