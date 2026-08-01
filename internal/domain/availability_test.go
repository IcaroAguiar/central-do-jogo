package domain

import "testing"

func TestAvailabilityStateValid(t *testing.T) {
	t.Parallel()

	cases := []struct {
		state AvailabilityState
		want  bool
	}{
		{AvailabilityAvailable, true},
		{AvailabilityAwaitingPublication, true},
		{AvailabilityNotFound, true},
		{AvailabilityDivergent, true},
		{AvailabilityNoCoverage, true},
		{AvailabilityState("bogus"), false},
		{AvailabilityState(""), false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(string(tc.state), func(t *testing.T) {
			t.Parallel()
			if got := tc.state.Valid(); got != tc.want {
				t.Fatalf("Valid() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestKickoffStateValid(t *testing.T) {
	t.Parallel()

	if !KickoffPublished.Valid() {
		t.Fatal("published should be valid")
	}
	if KickoffState("later").Valid() {
		t.Fatal("unknown kickoff state must be invalid")
	}
}
