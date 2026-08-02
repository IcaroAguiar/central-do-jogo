// Package render implements PAT-004 server-side HTML rendering: Go renders
// semantic HTML, indexable metadata, and an escaped initial-data payload for
// progressive enhancement by the React PWA. Renderer is decoupled from the
// domain/store layers; callers (internal/features/*) map their own data into
// the view models declared here.
package render

import (
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/IcaroAguiar/central-do-jogo/internal/platform/brasilia"
	servertemplates "github.com/IcaroAguiar/central-do-jogo/web/server-templates"
)

// CanonicalURL joins baseURL (scheme+host, no trailing slash, may be empty)
// with an absolute path to build a canonical/OG URL. When baseURL is empty,
// a root-relative URL is returned.
func CanonicalURL(baseURL, path string) string {
	return baseURL + path
}

// Meta carries the indexable metadata shared by every page (REQ-015).
type Meta struct {
	Title        string
	Description  string
	CanonicalURL string
	OGType       string
}

// Renderer executes the embedded SSR templates.
type Renderer struct {
	tmpl *template.Template
}

// New parses the embedded server templates. Parsing happens once at startup;
// a parse failure is a programming error and should abort boot.
func New() (*Renderer, error) {
	funcs := template.FuncMap{
		"formatBrasilia":      brasilia.FormatWithLabel,
		"availabilityLabel":   AvailabilityLabel,
		"accessLabel":         AccessLabel,
		"confidenceLabel":     ConfidenceLabel,
		"kickoffStateLabel":   KickoffStateLabel,
		"lastAttemptSentence": LastAttemptSentence,
	}
	tmpl, err := template.New("").Funcs(funcs).ParseFS(servertemplates.FS, "*.tmpl")
	if err != nil {
		return nil, fmt.Errorf("parse server templates: %w", err)
	}
	return &Renderer{tmpl: tmpl}, nil
}

func (r *Renderer) render(w http.ResponseWriter, status int, name string, data any) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if err := r.tmpl.ExecuteTemplate(w, name, data); err != nil {
		return fmt.Errorf("execute template %s: %w", name, err)
	}
	return nil
}

// ClubLink is a minimal club reference used for index links (home page).
type ClubLink struct {
	Slug string
	Name string
}

// HomePage is the view model for the "/" SSR page.
type HomePage struct {
	Meta        Meta
	Clubs       []ClubLink
	InitialData any
}

// RenderHome writes the home page.
func (r *Renderer) RenderHome(w http.ResponseWriter, page HomePage) error {
	return r.render(w, http.StatusOK, "home", page)
}

// MatchLink is a minimal match reference used inside a club's agenda list.
type MatchLink struct {
	Slug         string
	HomeClubName string
	AwayClubName string
	KickoffAt    *time.Time
	KickoffState string
}

// ClubViewModel is the club-specific content for the club SSR page.
type ClubViewModel struct {
	Name      string
	ShortName string
}

// ClubPage is the view model for the "/clubes/{slug}" SSR page.
type ClubPage struct {
	Meta        Meta
	NotFound    bool
	Club        ClubViewModel
	Matches     []MatchLink
	InitialData any
}

// RenderClub writes the club page with the given HTTP status (200 or 404).
func (r *Renderer) RenderClub(w http.ResponseWriter, status int, page ClubPage) error {
	return r.render(w, status, "club", page)
}

// BroadcastViewModel is one broadcast entry on the match SSR page.
type BroadcastViewModel struct {
	Channel    string
	Platform   string
	Access     string
	Confidence string
	Source     string
}

// LineupPlayerViewModel is one player entry on the match SSR page.
type LineupPlayerViewModel struct {
	ShirtNumber string
	Name        string
	IsStarter   bool
}

// LineupViewModel is one club's lineup on the match SSR page.
type LineupViewModel struct {
	SideLabel string
	Formation string
	Coach     string
	Players   []LineupPlayerViewModel
}

// NewsViewModel is one related news card on the match SSR page.
type NewsViewModel struct {
	Title  string
	URL    string
	Source string
}

// MatchViewModel is the match-specific content for the match SSR page.
type MatchViewModel struct {
	HomeClubName           string
	AwayClubName           string
	Round                  string
	Venue                  string
	KickoffAt              *time.Time
	KickoffState           string
	BroadcastState         string
	LineupState            string
	NewsState              string
	BroadcastLastAttemptAt *time.Time
	LineupLastAttemptAt    *time.Time
	NewsLastAttemptAt      *time.Time
	Broadcasts             []BroadcastViewModel
	Lineups                []LineupViewModel
	News                   []NewsViewModel
}

// MatchPage is the view model for the "/jogos/{slug}" SSR page.
type MatchPage struct {
	Meta        Meta
	NotFound    bool
	Match       MatchViewModel
	InitialData any
}

// RenderMatch writes the match page with the given HTTP status (200 or 404).
func (r *Renderer) RenderMatch(w http.ResponseWriter, status int, page MatchPage) error {
	return r.render(w, status, "match", page)
}
