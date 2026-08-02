package render

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestRenderHomeEscapesClubNamesAndInitialData(t *testing.T) {
	t.Parallel()
	r, err := New()
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	maliciousName := `<script>alert("pwn")</script>`
	page := HomePage{
		Meta:        Meta{Title: "Central do Jogo", Description: "desc", CanonicalURL: "/", OGType: "website"},
		Clubs:       []ClubLink{{Slug: "evil", Name: maliciousName}},
		InitialData: map[string]any{"payload": `</script><script>alert(1)</script>`},
	}

	rec := httptest.NewRecorder()
	if err := r.RenderHome(rec, page); err != nil {
		t.Fatalf("RenderHome() error: %v", err)
	}
	body := rec.Body.String()

	if strings.Contains(body, "<script>alert(\"pwn\")</script>") {
		t.Fatal("club name was rendered unescaped, XSS risk in HTML body")
	}
	if !strings.Contains(body, "&lt;script&gt;alert(&#34;pwn&#34;)&lt;/script&gt;") {
		t.Fatalf("expected HTML-escaped club name in body, got: %s", body)
	}

	assertInitialDataEscaped(t, body)
}

func TestRenderClubEscapesXSSInNotFoundAndFoundStates(t *testing.T) {
	t.Parallel()
	r, err := New()
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	t.Run("found", func(t *testing.T) {
		t.Parallel()
		page := ClubPage{
			Meta: Meta{Title: `"><img src=x onerror=alert(1)>`, CanonicalURL: "/clubes/evil", OGType: "website"},
			Club: ClubViewModel{Name: `<b onmouseover=alert(1)>Evil FC</b>`, ShortName: "Evil"},
			Matches: []MatchLink{
				{Slug: "m1", HomeClubName: `<script>alert('h')</script>`, AwayClubName: "Away", KickoffState: "published"},
			},
			InitialData: map[string]any{"note": "</script><svg onload=alert(1)>"},
		}
		rec := httptest.NewRecorder()
		if err := r.RenderClub(rec, 200, page); err != nil {
			t.Fatalf("RenderClub() error: %v", err)
		}
		body := rec.Body.String()
		assertNoRawScriptInjection(t, body)
		assertInitialDataEscaped(t, body)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		page := ClubPage{
			Meta:        Meta{Title: "Clube não encontrado", CanonicalURL: "/clubes/<script>x</script>", OGType: "website"},
			NotFound:    true,
			InitialData: map[string]any{"notFound": true},
		}
		rec := httptest.NewRecorder()
		if err := r.RenderClub(rec, 404, page); err != nil {
			t.Fatalf("RenderClub() error: %v", err)
		}
		if rec.Code != 404 {
			t.Fatalf("status = %d, want 404", rec.Code)
		}
		body := rec.Body.String()
		if !strings.Contains(body, "não encontrado") {
			t.Fatalf("expected not-found copy in body, got: %s", body)
		}
		assertNoRawScriptInjection(t, body)
	})
}

func TestRenderMatchEscapesNestedUserContent(t *testing.T) {
	t.Parallel()
	r, err := New()
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	page := MatchPage{
		Meta: Meta{Title: "Match", CanonicalURL: "/jogos/evil", OGType: "article"},
		Match: MatchViewModel{
			HomeClubName:   `<script>alert('home')</script>`,
			AwayClubName:   "Away",
			BroadcastState: "available",
			LineupState:    "available",
			NewsState:      "available",
			Broadcasts: []BroadcastViewModel{
				{Channel: `"><script>alert('bcast')</script>`, Access: "free", Confidence: "high", Source: "src"},
			},
			Lineups: []LineupViewModel{
				{
					SideLabel: "Home FC",
					Coach:     `<img src=x onerror=alert('coach')>`,
					Players: []LineupPlayerViewModel{
						{ShirtNumber: "1", Name: `</li><script>alert('player')</script>`, IsStarter: true},
					},
				},
			},
			News: []NewsViewModel{
				{Title: `<script>alert('news')</script>`, URL: `javascript:alert(1)`, Source: "src"},
			},
		},
		InitialData: map[string]any{"raw": "</script><script>alert('data')</script>"},
	}

	rec := httptest.NewRecorder()
	if err := r.RenderMatch(rec, 200, page); err != nil {
		t.Fatalf("RenderMatch() error: %v", err)
	}
	body := rec.Body.String()
	assertNoRawScriptInjection(t, body)
	assertInitialDataEscaped(t, body)

	// html/template neutralizes javascript: URLs in href attributes.
	if strings.Contains(body, `href="javascript:alert(1)"`) {
		t.Fatal("javascript: URL was not neutralized in href attribute")
	}
}

func TestRenderMatchUsesBrasiliaAndPTBRLabels(t *testing.T) {
	t.Parallel()
	r, err := New()
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	kickoff := time.Date(2026, 8, 2, 21, 0, 0, 0, time.UTC) // 18:00 BRT
	attempt := kickoff.Add(-time.Hour)
	page := MatchPage{
		Meta: Meta{Title: "Match", CanonicalURL: "/jogos/a", OGType: "article"},
		Match: MatchViewModel{
			HomeClubName:           "Home",
			AwayClubName:           "Away",
			KickoffAt:              &kickoff,
			BroadcastState:         "awaiting_publication",
			LineupState:            "available",
			NewsState:              "no_coverage",
			BroadcastLastAttemptAt: &attempt,
		},
	}
	rec := httptest.NewRecorder()
	if err := r.RenderMatch(rec, 200, page); err != nil {
		t.Fatalf("RenderMatch() error: %v", err)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "02/08/2026 18:00 (Horário de Brasília)") {
		t.Fatalf("expected Brasília kickoff label, body=%s", body)
	}
	if !strings.Contains(body, "Aguardando divulgação oficial") {
		t.Fatalf("expected pt-BR availability label, body=%s", body)
	}
	if strings.Contains(body, " UTC") {
		t.Fatal("SSR must not render UTC product timestamps (CON-008)")
	}
}

// assertNoRawScriptInjection fails if any executable <script> tag other than
// our own #initial-data element and the hardcoded PAT-004 app-shell module
// script (both fixed, non-user-controlled template strings, see base.tmpl),
// or any live HTML tag carrying an inline event handler, appears verbatim
// (i.e. unescaped) in the rendered body.
func assertNoRawScriptInjection(t *testing.T, body string) {
	t.Helper()
	scriptCount := strings.Count(body, "<script")
	expectedCount := strings.Count(body, `<script id="initial-data"`) + strings.Count(body, `<script type="module" src="/app.js">`)
	if scriptCount != expectedCount {
		t.Fatalf("found %d <script> tags but only %d are expected (initial-data + app-shell); body: %s", scriptCount, expectedCount, body)
	}
	// html/template escapes "<" to "&lt;", so a raw "<tag ...on*=" sequence
	// can only appear if user content broke out of its escaped context.
	live := []string{"<img ", "<svg ", "<b onmouseover"}
	for _, tag := range live {
		if strings.Contains(body, tag) {
			t.Fatalf("found unescaped live tag %q in body: %s", tag, body)
		}
	}
}

// assertInitialDataEscaped verifies the initial-data script content cannot
// break out of its own <script> element (the classic </script> JSON XSS).
func assertInitialDataEscaped(t *testing.T, body string) {
	t.Helper()
	start := strings.Index(body, `<script id="initial-data" type="application/json">`)
	if start == -1 {
		t.Fatalf("initial-data script tag not found in body: %s", body)
	}
	rest := body[start+len(`<script id="initial-data" type="application/json">`):]
	end := strings.Index(rest, "</script>")
	if end == -1 {
		t.Fatalf("initial-data script tag never closes: %s", body)
	}
	payload := rest[:end]
	if strings.Contains(payload, "</script") {
		t.Fatalf("initial-data payload contains an unescaped closing script tag: %s", payload)
	}
	if strings.Contains(payload, "<script") {
		t.Fatalf("initial-data payload contains an unescaped opening script tag: %s", payload)
	}
	var decoded any
	if err := json.Unmarshal([]byte(payload), &decoded); err != nil {
		t.Fatalf("initial-data payload is not valid JSON: %v; payload=%s", err, payload)
	}
}
