package push

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/IcaroAguiar/central-do-jogo/internal/domain"
	webpush "github.com/SherClockHolmes/webpush-go"
)

const defaultPushHTTPTimeout = 20 * time.Second

// VAPIDDeliverer sends encrypted Web Push notifications (REQ-011 / REQ-025).
type VAPIDDeliverer struct {
	publicKey  string
	privateKey string
	subject    string
	ttl        int
	client     webpush.HTTPClient
}

// DelivererForConfig returns StubDeliverer when push is disabled, otherwise a
// fail-closed VAPID deliverer (worker and smoke CLI share this wiring).
func DelivererForConfig(enabled bool, publicKey, privateKey, subject string, client webpush.HTTPClient) (Deliverer, error) {
	if !enabled {
		return StubDeliverer{}, nil
	}
	return NewVAPIDDeliverer(publicKey, privateKey, subject, client)
}

// NewVAPIDDeliverer builds a real push-network deliverer. subject should be a
// mailto: or https: contact URI embedded in the VAPID JWT.
func NewVAPIDDeliverer(publicKey, privateKey, subject string, client webpush.HTTPClient) (*VAPIDDeliverer, error) {
	publicKey = strings.TrimSpace(publicKey)
	privateKey = strings.TrimSpace(privateKey)
	subject = strings.TrimSpace(subject)
	if publicKey == "" || privateKey == "" {
		return nil, fmt.Errorf("%w: VAPID keys required", ErrPushDisabled)
	}
	if subject == "" {
		return nil, fmt.Errorf("%w: VAPID subject required", ErrPushDisabled)
	}
	if client == nil {
		client = defaultPushHTTPClient()
	}
	return &VAPIDDeliverer{
		publicKey:  publicKey,
		privateKey: privateKey,
		subject:    subject,
		ttl:        60,
		client:     client,
	}, nil
}

func defaultPushHTTPClient() *http.Client {
	return &http.Client{
		Timeout: defaultPushHTTPTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 3 {
				return errors.New("too many push redirects")
			}
			if _, err := validatePushEndpointURL(req.URL.String()); err != nil {
				return fmt.Errorf("push redirect blocked: %w", err)
			}
			return nil
		},
	}
}

// Deliver implements Deliverer against the subscription's push service endpoint.
func (d *VAPIDDeliverer) Deliver(ctx context.Context, sub domain.PushSubscription, payload []byte) DeliveryResult {
	if d == nil {
		return DeliveryResult{Err: ErrPushDisabled}
	}
	if _, err := validatePushEndpointURL(sub.Endpoint); err != nil {
		return DeliveryResult{Err: err}
	}
	subscription := &webpush.Subscription{
		Endpoint: sub.Endpoint,
		Keys: webpush.Keys{
			Auth:   sub.Auth,
			P256dh: sub.P256dh,
		},
	}
	opts := &webpush.Options{
		HTTPClient:      d.client,
		Subscriber:      d.subject,
		TTL:             d.ttl,
		VAPIDPublicKey:  d.publicKey,
		VAPIDPrivateKey: d.privateKey,
	}
	resp, err := webpush.SendNotificationWithContext(ctx, payload, subscription, opts)
	if err != nil {
		return DeliveryResult{Err: err}
	}
	defer func() { _ = resp.Body.Close() }()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 64<<10))

	switch resp.StatusCode {
	case http.StatusCreated, http.StatusOK:
		return DeliveryResult{Accepted: true}
	case http.StatusGone, http.StatusNotFound:
		return DeliveryResult{Gone: true}
	default:
		return DeliveryResult{
			Err: fmt.Errorf("webpush status %d", resp.StatusCode),
		}
	}
}

var _ Deliverer = (*VAPIDDeliverer)(nil)
