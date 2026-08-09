// Package httpapi holds shared public JSON DTOs and SSR page IDs that mirror
// api/openapi.yaml. Feature packages must reuse these types instead of
// declaring parallel ClubRef / CompetitionRef / page string literals.
package httpapi

import "github.com/IcaroAguiar/central-do-jogo/internal/domain"

// ClubRef is the OpenAPI ClubRef schema (slug, name, shortName).
type ClubRef struct {
	Slug      string `json:"slug"`
	Name      string `json:"name"`
	ShortName string `json:"shortName"`
}

// CompetitionRef is the OpenAPI CompetitionRef schema.
type CompetitionRef struct {
	Slug   string `json:"slug"`
	Name   string `json:"name"`
	Season int    `json:"season"`
}

// ClubRefFromClub maps a domain club into the public ClubRef DTO.
func ClubRefFromClub(c domain.Club) ClubRef {
	return ClubRef{Slug: c.Slug, Name: c.Name, ShortName: c.ShortName}
}

// ClubRefFromParts builds a ClubRef from already-resolved public fields
// (for example domain ClubSummary slug/name/shortName without leaking IDs).
func ClubRefFromParts(slug, name, shortName string) ClubRef {
	return ClubRef{Slug: slug, Name: name, ShortName: shortName}
}
