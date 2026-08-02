package clubs

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/IcaroAguiar/central-do-jogo/internal/platform/logging"
	"github.com/IcaroAguiar/central-do-jogo/internal/platform/render"
)

// NewHomeSSRHandler builds the GET "/" SSR handler listing supported clubs.
func NewHomeSSRHandler(svc *Service, renderer *render.Renderer, baseURL string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		items, err := svc.ListClubs(r.Context())
		if err != nil {
			logging.FromContext(r.Context()).Error("ssr home: list clubs", slog.String("error", err.Error()))
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		clubs := make([]render.ClubLink, 0, len(items))
		for _, c := range items {
			clubs = append(clubs, render.ClubLink{Slug: c.Slug, Name: c.Name})
		}

		page := render.HomePage{
			Meta: render.Meta{
				Title:        "Central do Jogo",
				Description:  "Onde assistir, escalações oficiais e notícias do futebol brasileiro, com proveniência e confiança explícitas.",
				CanonicalURL: render.CanonicalURL(baseURL, "/"),
				OGType:       "website",
			},
			Clubs:       clubs,
			InitialData: map[string]any{"page": "home", "clubs": items},
		}
		if err := renderer.RenderHome(w, page); err != nil {
			logging.FromContext(r.Context()).Error("ssr home: render", slog.String("error", err.Error()))
		}
	})
}

// clubMatchesLimit caps how many upcoming season matches appear on the SSR club page.
const clubMatchesLimit = 10

// NewClubSSRHandler builds the GET "/clubes/{slug}" SSR handler.
func NewClubSSRHandler(svc *Service, renderer *render.Renderer, baseURL string, now func() time.Time) http.Handler {
	if now == nil {
		now = time.Now
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("slug")
		logger := logging.FromContext(r.Context())

		detail, err := svc.GetDetail(r.Context(), slug)
		if err != nil {
			logger.Error("ssr club: get detail", slog.String("error", err.Error()))
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if detail == nil {
			page := render.ClubPage{
				Meta: render.Meta{
					Title:        "Clube não encontrado — Central do Jogo",
					CanonicalURL: render.CanonicalURL(baseURL, "/clubes/"+slug),
					OGType:       "website",
				},
				NotFound:    true,
				InitialData: map[string]any{"page": "club", "notFound": true},
			}
			if err := renderer.RenderClub(w, http.StatusNotFound, page); err != nil {
				logger.Error("ssr club: render not found", slog.String("error", err.Error()))
			}
			return
		}

		matchesResp, err := svc.GetMatches(r.Context(), slug, RangeWeek, now().Year())
		if err != nil {
			logger.Error("ssr club: get matches", slog.String("error", err.Error()))
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		var matchLinks []render.MatchLink
		if matchesResp != nil {
			if len(matchesResp.Matches) > clubMatchesLimit {
				matchesResp.Matches = matchesResp.Matches[:clubMatchesLimit]
			}
			for _, m := range matchesResp.Matches {
				matchLinks = append(matchLinks, render.MatchLink{
					Slug:         m.Slug,
					HomeClubName: m.HomeClub.Name,
					AwayClubName: m.AwayClub.Name,
					KickoffAt:    m.KickoffAt,
					KickoffState: m.KickoffState,
				})
			}
		}

		page := render.ClubPage{
			Meta: render.Meta{
				Title:        detail.Name + " — Central do Jogo",
				Description:  "Agenda, transmissões, escalações e notícias de " + detail.Name + ".",
				CanonicalURL: render.CanonicalURL(baseURL, "/clubes/"+detail.Slug),
				OGType:       "website",
			},
			Club:        render.ClubViewModel{Name: detail.Name, ShortName: detail.ShortName},
			Matches:     matchLinks,
			InitialData: map[string]any{"page": "club", "club": detail, "matches": matchesResp},
		}
		if err := renderer.RenderClub(w, http.StatusOK, page); err != nil {
			logger.Error("ssr club: render", slog.String("error", err.Error()))
		}
	})
}
