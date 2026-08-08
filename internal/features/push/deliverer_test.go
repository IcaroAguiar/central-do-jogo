package push_test

import (
	"context"
	"crypto/ecdh"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/IcaroAguiar/central-do-jogo/internal/domain"
	"github.com/IcaroAguiar/central-do-jogo/internal/features/push"
	webpush "github.com/SherClockHolmes/webpush-go"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) Do(req *http.Request) (*http.Response, error) { return f(req) }

func mustVAPIDKeys(t *testing.T) (publicKey, privateKey string) {
	t.Helper()
	priv, pub, err := webpush.GenerateVAPIDKeys()
	if err != nil {
		t.Fatal(err)
	}
	return pub, priv
}

func mustSubscriptionKeys(t *testing.T) (p256dh, auth string) {
	t.Helper()
	priv, err := ecdh.P256().GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	p256dh = base64.RawURLEncoding.EncodeToString(priv.PublicKey().Bytes())
	authBytes := make([]byte, 16)
	if _, err := rand.Read(authBytes); err != nil {
		t.Fatal(err)
	}
	auth = base64.RawURLEncoding.EncodeToString(authBytes)
	return p256dh, auth
}

func TestDelivererForConfigFallsBackToStubWhenDisabled(t *testing.T) {
	t.Parallel()
	d, err := push.DelivererForConfig(false, "", "", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	result := d.Deliver(context.Background(), domain.PushSubscription{}, []byte(`{}`))
	if !result.Accepted || result.Err != nil || result.Gone {
		t.Fatalf("stub result = %+v", result)
	}
}

func TestVAPIDDelivererMapsStatusCodes(t *testing.T) {
	t.Parallel()
	pub, priv := mustVAPIDKeys(t)
	p256dh, auth := mustSubscriptionKeys(t)

	cases := []struct {
		name   string
		status int
		want   push.DeliveryResult
	}{
		{name: "created", status: http.StatusCreated, want: push.DeliveryResult{Accepted: true}},
		{name: "ok", status: http.StatusOK, want: push.DeliveryResult{Accepted: true}},
		{name: "gone", status: http.StatusGone, want: push.DeliveryResult{Gone: true}},
		{name: "not_found", status: http.StatusNotFound, want: push.DeliveryResult{Gone: true}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			client := roundTripFunc(func(req *http.Request) (*http.Response, error) {
				if req.Method != http.MethodPost {
					t.Fatalf("method = %s", req.Method)
				}
				if !strings.HasPrefix(req.URL.String(), "https://fcm.googleapis.com/") {
					t.Fatalf("endpoint = %s", req.URL)
				}
				return &http.Response{
					StatusCode: tc.status,
					Body:       io.NopCloser(strings.NewReader("")),
					Header:     make(http.Header),
					Request:    req,
				}, nil
			})
			d, err := push.NewVAPIDDeliverer(pub, priv, "mailto:ops@example.com", client)
			if err != nil {
				t.Fatal(err)
			}
			got := d.Deliver(context.Background(), domain.PushSubscription{
				Endpoint: "https://fcm.googleapis.com/fcm/send/endpoint",
				P256dh:   p256dh,
				Auth:     auth,
			}, []byte(`{"title":"t"}`))
			if got.Accepted != tc.want.Accepted || got.Gone != tc.want.Gone {
				t.Fatalf("got %+v want %+v", got, tc.want)
			}
			if (tc.want.Accepted || tc.want.Gone) && got.Err != nil {
				t.Fatalf("unexpected err: %v", got.Err)
			}
		})
	}
}

func TestVAPIDDelivererRejectsDisallowedEndpointWithoutHTTP(t *testing.T) {
	t.Parallel()
	pub, priv := mustVAPIDKeys(t)
	p256dh, auth := mustSubscriptionKeys(t)
	called := false
	client := roundTripFunc(func(*http.Request) (*http.Response, error) {
		called = true
		return nil, errors.New("should not call network")
	})
	d, err := push.NewVAPIDDeliverer(pub, priv, "mailto:ops@example.com", client)
	if err != nil {
		t.Fatal(err)
	}
	got := d.Deliver(context.Background(), domain.PushSubscription{
		Endpoint: "https://evil.example/push",
		P256dh:   p256dh,
		Auth:     auth,
	}, []byte(`{"title":"t"}`))
	if called {
		t.Fatal("disallowed endpoint must not hit HTTP client")
	}
	if got.Err == nil || got.Accepted || got.Gone {
		t.Fatalf("got %+v", got)
	}
}

func TestVAPIDDelivererMapsServerError(t *testing.T) {
	t.Parallel()
	pub, priv := mustVAPIDKeys(t)
	p256dh, auth := mustSubscriptionKeys(t)
	client := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusBadGateway,
			Body:       io.NopCloser(strings.NewReader("upstream")),
			Header:     make(http.Header),
			Request:    req,
		}, nil
	})
	d, err := push.NewVAPIDDeliverer(pub, priv, "mailto:ops@example.com", client)
	if err != nil {
		t.Fatal(err)
	}
	got := d.Deliver(context.Background(), domain.PushSubscription{
		Endpoint: "https://fcm.googleapis.com/fcm/send/endpoint",
		P256dh:   p256dh,
		Auth:     auth,
	}, []byte(`{"title":"t"}`))
	if got.Accepted || got.Gone || got.Err == nil {
		t.Fatalf("got %+v", got)
	}
	if !strings.Contains(got.Err.Error(), "502") {
		t.Fatalf("err = %v", got.Err)
	}
}

func TestVAPIDDelivererRequiresKeysAndSubject(t *testing.T) {
	t.Parallel()
	_, err := push.NewVAPIDDeliverer("", "priv", "mailto:x@y.z", nil)
	if !errors.Is(err, push.ErrPushDisabled) {
		t.Fatalf("err = %v", err)
	}
	pub, priv := mustVAPIDKeys(t)
	_, err = push.NewVAPIDDeliverer(pub, priv, "", nil)
	if !errors.Is(err, push.ErrPushDisabled) {
		t.Fatalf("err = %v", err)
	}
}
