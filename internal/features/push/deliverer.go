package push

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/IcaroAguiar/central-do-jogo/internal/domain"
	webpush "github.com/SherClockHolmes/webpush-go"
)

// VAPIDDeliverer sends encrypted Web Push notifications (REQ-011 / REQ-025).
type VAPIDDeliverer struct {
	publicKey  string
	privateKey string
	subject    string
	ttl        int
	client     webpush.HTTPClient
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
		subject = "mailto:ops@centraldojogo.local"
	}
	return &VAPIDDeliverer{
		publicKey:  publicKey,
		privateKey: privateKey,
		subject:    subject,
		ttl:        60,
		client:     client,
	}, nil
}

// NewDeliverer returns a VAPID deliverer when keys are present, otherwise StubDeliverer.
func NewDeliverer(publicKey, privateKey, subject string, client webpush.HTTPClient) Deliverer {
	d, err := NewVAPIDDeliverer(publicKey, privateKey, subject, client)
	if err != nil {
		return StubDeliverer{}
	}
	return d
}

// Deliver implements Deliverer against the subscription's push service endpoint.
func (d *VAPIDDeliverer) Deliver(ctx context.Context, sub domain.PushSubscription, payload []byte) DeliveryResult {
	if d == nil {
		return DeliveryResult{Err: ErrPushDisabled}
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

	switch {
	case resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK:
		return DeliveryResult{Accepted: true}
	case resp.StatusCode == http.StatusGone || resp.StatusCode == http.StatusNotFound:
		return DeliveryResult{Gone: true}
	default:
		return DeliveryResult{
			Err: fmt.Errorf("webpush status %d", resp.StatusCode),
		}
	}
}

var _ Deliverer = (*VAPIDDeliverer)(nil)
