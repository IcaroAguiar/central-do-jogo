// Package auth implements configurable OAuth login, server-side sessions,
// and maintainer allowlist grants (REQ-017, REQ-018, TASK-027).
package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/IcaroAguiar/central-do-jogo/internal/domain"
)

const (
	ProviderGoogle = "google"

	// SessionCookieName is the HttpOnly cookie carrying the opaque session token.
	SessionCookieName = "cdj_session"
	// OAuthStateCookieName carries the signed OAuth CSRF state.
	OAuthStateCookieName = "cdj_oauth_state"

	oauthStateTTL = 10 * time.Minute
)

// ErrAuthDisabled is returned when OAuth credentials are not configured.
var ErrAuthDisabled = errors.New("auth disabled")

// ErrInvalidState is returned when the OAuth state cookie/query mismatch.
var ErrInvalidState = errors.New("invalid oauth state")

// ErrEmailUnverified is returned when the IdP identity lacks a verified email.
var ErrEmailUnverified = errors.New("email not verified")

// Identity is a verified account identity from an OAuth provider.
type Identity struct {
	Provider      string
	Subject       string
	Email         string
	EmailVerified bool
	DisplayName   string
}

// Provider exchanges OAuth authorization codes for verified identities.
type Provider interface {
	Name() string
	AuthCodeURL(state string) string
	Exchange(ctx context.Context, code string) (Identity, error)
}

// UserRepository persists users and sessions.
type UserRepository interface {
	GetByID(ctx context.Context, id domain.ID) (*domain.User, error)
	UpsertByProviderSubject(ctx context.Context, user domain.User, now time.Time) (*domain.User, error)
	CreateSession(ctx context.Context, sess domain.Session) error
	GetSessionByTokenHash(ctx context.Context, tokenHash string, now time.Time) (*domain.Session, error)
	RevokeSession(ctx context.Context, tokenHash string, now time.Time) error
}

// Config holds runtime auth settings already validated by config.Load.
type Config struct {
	Enabled           bool
	SessionSecret     []byte
	SessionTTL        time.Duration
	CookieSecure      bool
	PublicBaseURL     string
	MaintainerEmails  map[string]struct{}
	PostLoginRedirect string
}

// Service orchestrates OAuth login and session lifecycle without owning HTTP
// transport (cookies stay in handlers).
type Service struct {
	users    UserRepository
	provider Provider
	cfg      Config
	now      func() time.Time
}

// NewService creates an auth service. provider may be nil when auth is disabled.
func NewService(users UserRepository, provider Provider, cfg Config, now func() time.Time) *Service {
	if now == nil {
		now = time.Now
	}
	return &Service{users: users, provider: provider, cfg: cfg, now: now}
}

// Enabled reports whether OAuth login is configured.
func (s *Service) Enabled() bool {
	return s.cfg.Enabled && s.provider != nil
}

// CookieSecure reports whether auth cookies must set the Secure flag.
func (s *Service) CookieSecure() bool { return s.cfg.CookieSecure }

// PublicBaseURL returns the configured public origin used for CSRF checks.
func (s *Service) PublicBaseURL() string { return s.cfg.PublicBaseURL }

// SessionTTL returns how long a new session cookie should live.
func (s *Service) SessionTTL() time.Duration { return s.cfg.SessionTTL }

// OAuthStateTTL returns the OAuth state cookie lifetime.
func (s *Service) OAuthStateTTL() time.Duration { return oauthStateTTL }

// StartLoginResult is the transport-agnostic output of StartLogin.
type StartLoginResult struct {
	AuthURL           string
	SignedStateCookie string
}

// StartLogin builds the IdP authorization URL and a signed OAuth state value.
func (s *Service) StartLogin() (StartLoginResult, error) {
	if !s.Enabled() {
		return StartLoginResult{}, ErrAuthDisabled
	}
	state, err := randomToken(24)
	if err != nil {
		return StartLoginResult{}, err
	}
	signed, err := s.signState(state)
	if err != nil {
		return StartLoginResult{}, err
	}
	return StartLoginResult{
		AuthURL:           s.provider.AuthCodeURL(state),
		SignedStateCookie: signed,
	}, nil
}

// LoginResult is the transport-agnostic output of CompleteLogin.
type LoginResult struct {
	User         *domain.User
	SessionToken string
	ExpiresAt    time.Time
}

// CompleteLogin validates state, exchanges the code, upserts the user with
// allowlist-derived role, and creates a server-side session.
func (s *Service) CompleteLogin(ctx context.Context, code, state, signedStateCookie string) (LoginResult, error) {
	if !s.Enabled() {
		return LoginResult{}, ErrAuthDisabled
	}
	if err := s.verifyState(signedStateCookie, state); err != nil {
		return LoginResult{}, err
	}

	identity, err := s.provider.Exchange(ctx, code)
	if err != nil {
		return LoginResult{}, fmt.Errorf("exchange code: %w", err)
	}
	if !identity.EmailVerified || strings.TrimSpace(identity.Email) == "" {
		return LoginResult{}, ErrEmailUnverified
	}

	now := s.now().UTC()
	userID, err := domain.NewID("usr_")
	if err != nil {
		return LoginResult{}, err
	}
	role := domain.RoleUser
	if s.isMaintainer(identity.Email) {
		role = domain.RoleMaintainer
	}
	user, err := s.users.UpsertByProviderSubject(ctx, domain.User{
		ID:              userID,
		Provider:        identity.Provider,
		ProviderSubject: identity.Subject,
		Email:           strings.ToLower(strings.TrimSpace(identity.Email)),
		DisplayName:     strings.TrimSpace(identity.DisplayName),
		Role:            role,
	}, now)
	if err != nil {
		return LoginResult{}, err
	}

	rawToken, err := randomToken(32)
	if err != nil {
		return LoginResult{}, err
	}
	sessID, err := domain.NewID("ses_")
	if err != nil {
		return LoginResult{}, err
	}
	expires := now.Add(s.cfg.SessionTTL)
	sess := domain.Session{
		ID:        sessID,
		UserID:    user.ID,
		TokenHash: hashToken(rawToken),
		ExpiresAt: expires,
		CreatedAt: now,
	}
	if err := s.users.CreateSession(ctx, sess); err != nil {
		return LoginResult{}, err
	}
	return LoginResult{User: user, SessionToken: rawToken, ExpiresAt: expires}, nil
}

// CurrentUser resolves a raw session token into a user, or nil when anonymous.
// Effective maintainer status is recomputed from the allowlist on every read
// so removals take effect without waiting for re-login (REQ-018).
func (s *Service) CurrentUser(ctx context.Context, sessionToken string) (*domain.User, error) {
	if !s.Enabled() || sessionToken == "" {
		return nil, nil
	}
	sess, err := s.users.GetSessionByTokenHash(ctx, hashToken(sessionToken), s.now().UTC())
	if err != nil {
		return nil, err
	}
	if sess == nil {
		return nil, nil
	}
	user, err := s.users.GetByID(ctx, sess.UserID)
	if err != nil || user == nil {
		return user, err
	}
	effective := domain.RoleUser
	if s.isMaintainer(user.Email) {
		effective = domain.RoleMaintainer
	}
	user.Role = effective
	return user, nil
}

// Logout revokes the session identified by the raw token when present.
func (s *Service) Logout(ctx context.Context, sessionToken string) error {
	if !s.Enabled() || sessionToken == "" {
		return nil
	}
	return s.users.RevokeSession(ctx, hashToken(sessionToken), s.now().UTC())
}

// PostLoginRedirect returns the safe relative path after OAuth success.
func (s *Service) PostLoginRedirect() string {
	if s.cfg.PostLoginRedirect != "" {
		return s.cfg.PostLoginRedirect
	}
	return "/"
}

func (s *Service) isMaintainer(email string) bool {
	_, ok := s.cfg.MaintainerEmails[strings.ToLower(strings.TrimSpace(email))]
	return ok
}

func (s *Service) signState(state string) (string, error) {
	mac := hmac.New(sha256.New, s.cfg.SessionSecret)
	exp := s.now().UTC().Add(oauthStateTTL).Unix()
	payload := fmt.Sprintf("%s.%d", state, exp)
	_, _ = mac.Write([]byte(payload))
	sig := hex.EncodeToString(mac.Sum(nil))
	return base64.RawURLEncoding.EncodeToString([]byte(payload + "." + sig)), nil
}

func (s *Service) verifyState(signedCookie, state string) error {
	if signedCookie == "" {
		return ErrInvalidState
	}
	raw, err := base64.RawURLEncoding.DecodeString(signedCookie)
	if err != nil {
		return ErrInvalidState
	}
	parts := strings.Split(string(raw), ".")
	if len(parts) != 3 {
		return ErrInvalidState
	}
	cookieState, expRaw, sig := parts[0], parts[1], parts[2]
	if !hmac.Equal([]byte(cookieState), []byte(state)) {
		return ErrInvalidState
	}
	exp, err := parseUnix(expRaw)
	if err != nil || s.now().UTC().Unix() > exp {
		return ErrInvalidState
	}
	mac := hmac.New(sha256.New, s.cfg.SessionSecret)
	_, _ = mac.Write([]byte(cookieState + "." + expRaw))
	want := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(want), []byte(sig)) {
		return ErrInvalidState
	}
	return nil
}

func parseUnix(raw string) (int64, error) {
	var n int64
	for _, ch := range raw {
		if ch < '0' || ch > '9' {
			return 0, fmt.Errorf("invalid unix")
		}
		n = n*10 + int64(ch-'0')
	}
	return n, nil
}

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func randomToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("random token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// SafeRelativePath accepts only root-relative paths without scheme/host tricks.
func SafeRelativePath(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || !strings.HasPrefix(raw, "/") || strings.HasPrefix(raw, "//") {
		return "/"
	}
	if strings.ContainsAny(raw, "\r\n\\") {
		return "/"
	}
	if u, err := url.Parse(raw); err != nil || u.IsAbs() || u.Host != "" {
		return "/"
	}
	return raw
}
