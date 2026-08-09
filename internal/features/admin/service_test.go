package admin_test

import (
	"context"
	"testing"
	"time"

	"github.com/IcaroAguiar/central-do-jogo/internal/domain"
	"github.com/IcaroAguiar/central-do-jogo/internal/features/admin"
)

type memGate struct {
	enabled bool
	user    *domain.User
	baseURL string
	err     error
}

func (m *memGate) Enabled() bool { return m.enabled }
func (m *memGate) RequireMaintainer(context.Context, string) (*domain.User, error) {
	if !m.enabled {
		return nil, domain.ErrAuthDisabled
	}
	if m.err != nil {
		return nil, m.err
	}
	if m.user == nil {
		return nil, domain.ErrUnauthorized
	}
	if m.user.Role != domain.RoleMaintainer {
		return nil, domain.ErrForbidden
	}
	return m.user, nil
}
func (m *memGate) PublicBaseURL() string { return m.baseURL }

type memHealth struct{ rows []domain.SourceHealth }

func (m *memHealth) List(context.Context) ([]domain.SourceHealth, error) { return m.rows, nil }

type memMatches struct {
	bySlug map[string]*domain.MatchRecord
	risk   []domain.MatchRecord
}

func (m *memMatches) GetBySlug(_ context.Context, slug string) (*domain.MatchRecord, error) {
	return m.bySlug[slug], nil
}
func (m *memMatches) ListAtRisk(context.Context, int) ([]domain.MatchRecord, error) {
	return m.risk, nil
}

type memWriter struct {
	applies []domain.MatchActionApply
	err     error
}

func (m *memWriter) ApplyMatchAction(_ context.Context, in domain.MatchActionApply) error {
	if m.err != nil {
		return m.err
	}
	m.applies = append(m.applies, in)
	return nil
}

type memAudit struct{ events []domain.AuditEvent }

func (m *memAudit) ListRecent(context.Context, string, string, int) ([]domain.AuditEvent, error) {
	return m.events, nil
}

func TestApplyMatchActionRequiresMaintainer(t *testing.T) {
	t.Parallel()
	svc := admin.NewService(
		&memGate{enabled: true, user: &domain.User{ID: "u1", Role: domain.RoleUser}},
		&memHealth{}, &memMatches{}, &memWriter{}, &memAudit{}, time.Now,
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
	writer := &memWriter{}
	svc := admin.NewService(
		&memGate{enabled: true, user: &domain.User{ID: "u1", Email: "ops@example.com", Role: domain.RoleMaintainer}, baseURL: "http://127.0.0.1:8080"},
		&memHealth{}, matches, writer, &memAudit{}, time.Now,
	)
	if err := svc.ApplyMatchAction(context.Background(), "tok", "flamengo-x-vasco", admin.MatchAction{
		Action: admin.ActionConfirm, Surface: "broadcast", Reason: "verified on TV",
	}); err != nil {
		t.Fatal(err)
	}
	if len(writer.applies) != 1 {
		t.Fatalf("applies = %#v", writer.applies)
	}
	got := writer.applies[0]
	if got.MatchID != "match_1" || got.Surface != "broadcast" || got.AfterState != domain.AvailabilityAvailable {
		t.Fatalf("apply = %#v", got)
	}
	if got.Override != nil {
		t.Fatalf("unexpected override: %#v", got.Override)
	}
	if got.Audit.Actor != "ops@example.com" || got.Audit.Action != admin.ActionConfirm {
		t.Fatalf("audit = %#v", got.Audit)
	}
}

func TestApplyMatchActionCorrectUsesWriterTransactionBundle(t *testing.T) {
	t.Parallel()
	rec := &domain.MatchRecord{
		Match: domain.Match{ID: "match_1", Slug: "flamengo-x-vasco", LineupState: domain.AvailabilityNotFound},
	}
	writer := &memWriter{}
	svc := admin.NewService(
		&memGate{enabled: true, user: &domain.User{ID: "u1", Email: "ops@example.com", Role: domain.RoleMaintainer}},
		&memHealth{},
		&memMatches{bySlug: map[string]*domain.MatchRecord{"flamengo-x-vasco": rec}},
		writer,
		&memAudit{},
		time.Now,
	)
	if err := svc.ApplyMatchAction(context.Background(), "tok", "flamengo-x-vasco", admin.MatchAction{
		Action: admin.ActionCorrect, Surface: "lineup", Reason: "club posted XI", Value: "available",
	}); err != nil {
		t.Fatal(err)
	}
	if len(writer.applies) != 1 || writer.applies[0].Override == nil {
		t.Fatalf("applies = %#v", writer.applies)
	}
	if writer.applies[0].Override.DataType != "lineup" {
		t.Fatalf("override = %#v", writer.applies[0].Override)
	}
}

func TestListSourceHealthUsesMaintainerGate(t *testing.T) {
	t.Parallel()
	svc := admin.NewService(
		&memGate{enabled: true, user: &domain.User{ID: "u1", Role: domain.RoleMaintainer}},
		&memHealth{rows: []domain.SourceHealth{{SourceID: "cbf_match_center"}}},
		&memMatches{}, &memWriter{}, &memAudit{}, time.Now,
	)
	rows, err := svc.ListSourceHealth(context.Background(), "tok")
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].SourceID != "cbf_match_center" {
		t.Fatalf("rows = %#v", rows)
	}
}
