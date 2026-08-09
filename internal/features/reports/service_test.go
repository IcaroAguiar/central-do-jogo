package reports_test

import (
	"context"
	"testing"
	"time"

	"github.com/IcaroAguiar/central-do-jogo/internal/domain"
	"github.com/IcaroAguiar/central-do-jogo/internal/features/reports"
)

type memSessions struct {
	enabled bool
	user    *domain.User
}

func (m *memSessions) Enabled() bool { return m.enabled }
func (m *memSessions) CurrentUser(context.Context, string) (*domain.User, error) {
	return m.user, nil
}
func (m *memSessions) PublicBaseURL() string { return "http://127.0.0.1:8080" }

type memReports struct {
	rows []domain.Report
}

func (m *memReports) Insert(_ context.Context, report domain.Report) error {
	m.rows = append(m.rows, report)
	return nil
}
func (m *memReports) ListOpen(context.Context, int) ([]domain.Report, error) {
	var out []domain.Report
	for _, r := range m.rows {
		if r.Status == reports.StatusOpen {
			out = append(out, r)
		}
	}
	return out, nil
}
func (m *memReports) MarkReviewed(_ context.Context, id domain.ID, status string, now time.Time) error {
	for i := range m.rows {
		if m.rows[i].ID == id {
			m.rows[i].Status = status
			m.rows[i].ReviewedAt = &now
			return nil
		}
	}
	return nil
}

func TestCreateDoesNotRequireAuth(t *testing.T) {
	t.Parallel()
	repo := &memReports{}
	svc := reports.NewService(&memSessions{enabled: false}, repo, time.Now)
	if err := svc.Create(context.Background(), reports.CreateInput{
		ContextType: "match", ContextSlug: "flamengo-x-vasco", Message: "canal errado",
	}); err != nil {
		t.Fatal(err)
	}
	if len(repo.rows) != 1 || repo.rows[0].Status != reports.StatusOpen {
		t.Fatalf("rows = %#v", repo.rows)
	}
}

func TestListOpenRequiresMaintainer(t *testing.T) {
	t.Parallel()
	svc := reports.NewService(
		&memSessions{enabled: true, user: &domain.User{Role: domain.RoleUser}},
		&memReports{},
		time.Now,
	)
	if _, err := svc.ListOpen(context.Background(), "tok"); err != reports.ErrForbidden {
		t.Fatalf("err = %v", err)
	}
}
