package push_test

import (
	"context"
	"testing"
	"time"

	"github.com/IcaroAguiar/central-do-jogo/internal/domain"
	"github.com/IcaroAguiar/central-do-jogo/internal/features/push"
)

func TestSubscribeRejectsNonAllowlistedEndpoints(t *testing.T) {
	t.Parallel()
	repo := newMemPush()
	svc := push.NewService(
		&memSessions{enabled: true, user: &domain.User{ID: "u1"}, baseURL: "http://localhost"},
		repo,
		true,
		"pub",
		time.Now,
	)
	cases := []string{
		"http://fcm.googleapis.com/fcm/send/x",
		"https://evil.example/push",
		"https://169.254.169.254/latest/meta-data",
		"https://127.0.0.1/push",
		"https://localhost/push",
	}
	for _, endpoint := range cases {
		_, err := svc.Subscribe(context.Background(), "tok", push.SubscribeInput{
			Endpoint: endpoint, P256dh: "k", Auth: "a",
		})
		if err == nil {
			t.Fatalf("expected reject for %s", endpoint)
		}
	}
	_, err := svc.Subscribe(context.Background(), "tok", push.SubscribeInput{
		Endpoint: "https://fcm.googleapis.com/fcm/send/abc", P256dh: "k", Auth: "a",
	})
	if err != nil {
		t.Fatal(err)
	}
}
