package admin_test

import (
	"context"
	"testing"
	"time"

	"github.com/IcaroAguiar/central-do-jogo/internal/domain"
	"github.com/IcaroAguiar/central-do-jogo/internal/features/admin"
	"github.com/IcaroAguiar/central-do-jogo/internal/jobs"
)

type memSessions struct {
	enabled bool
	user    *domain.User
	baseURL string
}

func (m *memSessions) Enabled() bool { return m.enabled }
func (m *memSessions) CurrentUser(context.Context, string) (*domain.User, error) {
	return m.user, nil
}
func (m *memSessions) PublicBaseURL() string { return m.baseURL }

type memHealth struct{ rows []jobs.SourceHealth }

func (m *memHealth) List(context.Context) ([]jobs.SourceHealth, error) { return m.rows, nil }

type memMatches struct {
	bySlug  map[string]*domain.MatchRecord
	risk    []domain.MatchRecord
	updates []string
}

func (m *memMatches) GetBySlug(_ context.Context, slug string) (*domain.MatchRecord, error) {
	return m.bySlug[slug], nil
}
func (m *memMatches) ListAtRisk(context.Context, int) ([]domain.MatchRecord, error) {
	return m.risk, nil
}
func (m *memMatches) UpdateSurfaceState(_ context.Context, matchID domain.ID, surface string, state domain.AvailabilityState, _ time.Time) error {
	m.updates = append(m.updates, matchID.String()+":"+surface+":"+string(state))
	if rec := m.bySlug["flamengo-x-vasco"]; rec != nil {
		switch surface {
		case "broadcast":
			rec.BroadcastState = state
		case "lineup":
			rec.LineupState = state
		case "news":
			rec.NewsState = state
		}
	}
	return nil
}

type memOverrides struct{ saved []domain.MatchOverride }

func (m *memOverrides) Save(_ context.Context, o domain.MatchOverride) (*domain.MatchOverride, error) {
	m.saved = append(m.saved, o)
	return &o, nil
}

type memAudit struct{ events []domain.AuditEvent }

func (m *memAudit) Append(_ context.Context, e domain.AuditEvent) (*domain.AuditEvent, error) {
	e.ID = int64(len(m.events) + 1)
	m.events = append(m.events, e)
	return &e, nil
}
func (m *memAudit) ListRecent(context.Context, string, string, int) ([]domain.AuditEvent, error) {
	return m.events, nil
}

func TestApplyMatchActionRequiresMaintainer(t *testing.T) {
	t.Parallel()
	svc := admin.NewService(
		&memSessions{enabled: true, user: &domain.User{ID: "u1", Role: domain.RoleUser}},
		&memHealth{}, &memMatches{}, &memOverrides{}, &memAudit{}, time.Now,
	)
	err := svc.ApplyMatchAction(context.Background(), "tok", "x", admin.MatchAction{
		Action: admin.ActionConfirm, Surface: "broadcast", Reason: "ok",
	})
	if err != admin.ErrForbidden {
		t.Fatalf("err = %v", err)
	}
}

func TestApplyMatchActionConfirmAudits(t *testing.T) {
	t.Parallel()
	rec := &domain.MatchRecord{
		Match: domain.Match{ID: "match_1", Slug: "flamengo-x-vasco", BroadcastState: domain.AvailabilityDivergent},
	}
	matches := &memMatches{bySlug: map[string]*domain.MatchRecord{"flamengo-x-vasco": rec}}
	audit := &memAudit{}
	svc := admin.NewService(
		&memSessions{enabled: true, user: &domain.User{ID: "u1", Email: "ops@example.com", Role: domain.RoleMaintainer}, baseURL: "http://127.0.0.1:8080"},
		&memHealth{}, matches, &memOverrides{}, audit, time.Now,
	)
	if err := svc.ApplyMatchAction(context.Background(), "tok", "flamengo-x-vasco", admin.MatchAction{
		Action: admin.ActionConfirm, Surface: "broadcast", Reason: "verified on TV",
	}); err != nil {
		t.Fatal(err)
	}
	if len(matches.updates) != 1 || matches.updates[0] != "match_1:broadcast:available" {
		t.Fatalf("updates = %#v", matches.updates)
	}
	if len(audit.events) != 1 || audit.events[0].Actor != "ops@example.com" {
		t.Fatalf("audit = %#v", audit.events)
	}
}
