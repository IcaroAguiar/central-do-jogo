// Package gazetaesportiva_rss parses news links from Gazeta Esportiva RSS feed.
package gazetaesportiva_rss

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/xml"
	"fmt"
	"time"

	"github.com/IcaroAguiar/central-do-jogo/internal/sources"
)

const (
	sourceID      = "gazetaesportiva_rss"
	parserVersion = "1.0.0"
)

type rssItem struct {
	Title   string `xml:"title"`
	Link    string `xml:"link"`
	PubDate string `xml:"pubDate"`
}

type rssChannel struct {
	Items []rssItem `xml:"item"`
}

type rssFeed struct {
	Channel rssChannel `xml:"channel"`
}

// Adapter implements sources.Adapter for Gazeta Esportiva RSS feed.
type Adapter struct{}

func (a *Adapter) SourceID() string { return sourceID }

func (a *Adapter) Parse(_ context.Context, raw []byte, observedAt time.Time) (*sources.Observation, error) {
	if observedAt.IsZero() {
		return nil, fmt.Errorf("gazetaesportiva_rss: observedAt must not be zero")
	}

	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 {
		return nil, fmt.Errorf("gazetaesportiva_rss: empty input (fail-closed)")
	}

	hash := fmt.Sprintf("%x", sha256.Sum256(raw))

	var feed rssFeed
	if err := xml.Unmarshal(raw, &feed); err != nil {
		return nil, fmt.Errorf("gazetaesportiva_rss: failed to parse RSS XML: %w", err)
	}

	obs := &sources.Observation{
		SourceID:      sourceID,
		DataType:      sources.DataTypeNews,
		ObservedAt:    observedAt.UTC(),
		ParserVersion: parserVersion,
		ContentHash:   hash,
		RawRef:        "https://www.gazetaesportiva.com/feed/",
	}

	for _, item := range feed.Channel.Items {
		if item.Title == "" || item.Link == "" {
			continue
		}

		pubAt := observedAt.UTC()
		if item.PubDate != "" {
			if t, err := time.Parse(time.RFC1123Z, item.PubDate); err == nil {
				pubAt = t
			} else if t, err := time.Parse(time.RFC1123, item.PubDate); err == nil {
				pubAt = t
			}
		}

		obs.NewsLinks = append(obs.NewsLinks, sources.NewsLinkEntry{
			Title:       item.Title,
			URL:         item.Link,
			PublishedAt: pubAt,
		})
	}

	if len(obs.NewsLinks) == 0 {
		return nil, fmt.Errorf("gazetaesportiva_rss: non-empty input produced zero news entries (fail-closed)")
	}

	return obs, nil
}
