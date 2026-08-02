package matches

import (
	"log/slog"
	"net/http"

	"github.com/IcaroAguiar/central-do-jogo/internal/api"
	"github.com/IcaroAguiar/central-do-jogo/internal/platform/logging"
	"github.com/IcaroAguiar/central-do-jogo/internal/platform/render"
)

// NewSSRHandler builds the GET "/jogos/{slug}" SSR handler.
func NewSSRHandler(svc *Service, renderer *render.Renderer, baseURL string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("slug")
		logger := logging.FromContext(r.Context())

		detail, err := svc.GetDetail(r.Context(), slug)
		if err != nil {
			logger.Error("ssr match: get detail", slog.String("error", err.Error()))
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if detail == nil {
			page := render.MatchPage{
				Meta: render.Meta{
					Title:        "Partida não encontrada — Central do Jogo",
					CanonicalURL: render.CanonicalURL(baseURL, "/jogos/"+slug),
					OGType:       "website",
				},
				NotFound:    true,
				InitialData: map[string]any{"page": api.PageMatch, "notFound": true},
			}
			if err := renderer.RenderMatch(w, http.StatusNotFound, page); err != nil {
				logger.Error("ssr match: render not found", slog.String("error", err.Error()))
			}
			return
		}

		page := render.MatchPage{
			Meta: render.Meta{
				Title:        detail.HomeClub.Name + " x " + detail.AwayClub.Name + " — Central do Jogo",
				Description:  "Onde assistir, escalações e notícias de " + detail.HomeClub.Name + " x " + detail.AwayClub.Name + ".",
				CanonicalURL: render.CanonicalURL(baseURL, "/jogos/"+detail.Slug),
				OGType:       "article",
			},
			Match:       toMatchViewModel(detail),
			InitialData: map[string]any{"page": api.PageMatch, "match": detail},
		}
		if err := renderer.RenderMatch(w, http.StatusOK, page); err != nil {
			logger.Error("ssr match: render", slog.String("error", err.Error()))
		}
	})
}

func toMatchViewModel(d *Detail) render.MatchViewModel {
	broadcasts := make([]render.BroadcastViewModel, 0, len(d.Broadcasts))
	for _, b := range d.Broadcasts {
		broadcasts = append(broadcasts, render.BroadcastViewModel{
			Channel:    b.Channel,
			Platform:   b.Platform,
			Access:     b.Access,
			Confidence: b.Confidence,
			Source:     b.Source,
		})
	}

	lineups := make([]render.LineupViewModel, 0, len(d.Lineups))
	for _, l := range d.Lineups {
		players := make([]render.LineupPlayerViewModel, 0, len(l.Players))
		for _, p := range l.Players {
			players = append(players, render.LineupPlayerViewModel{
				ShirtNumber: p.ShirtNumber,
				Name:        p.Name,
				IsStarter:   p.IsStarter,
			})
		}
		sideLabel := d.AwayClub.Name
		if l.Side == "home" {
			sideLabel = d.HomeClub.Name
		}
		lineups = append(lineups, render.LineupViewModel{
			SideLabel: sideLabel,
			Formation: l.Formation,
			Coach:     l.Coach,
			Players:   players,
		})
	}

	news := make([]render.NewsViewModel, 0, len(d.News))
	for _, n := range d.News {
		news = append(news, render.NewsViewModel{Title: n.Title, URL: n.URL, Source: n.Source})
	}

	return render.MatchViewModel{
		HomeClubName:           d.HomeClub.Name,
		AwayClubName:           d.AwayClub.Name,
		Round:                  d.Round,
		Venue:                  d.Venue,
		KickoffAt:              d.KickoffAt,
		KickoffState:           d.KickoffState,
		BroadcastState:         d.BroadcastState,
		LineupState:            d.LineupState,
		NewsState:              d.NewsState,
		BroadcastLastAttemptAt: d.BroadcastLastAttemptAt,
		LineupLastAttemptAt:    d.LineupLastAttemptAt,
		NewsLastAttemptAt:      d.NewsLastAttemptAt,
		Broadcasts:             broadcasts,
		Lineups:                lineups,
		News:                   news,
	}
}
