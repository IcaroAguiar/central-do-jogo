package domain

// ClubSummary is a lightweight club projection embedded in match records.
type ClubSummary struct {
	ID        ID
	Slug      string
	Name      string
	ShortName string
}

// CompetitionSummary is a lightweight competition projection embedded in match records.
type CompetitionSummary struct {
	ID     ID
	Slug   string
	Name   string
	Season int
}

// MatchRecord is a match joined with its home/away club and competition summaries.
type MatchRecord struct {
	Match
	HomeClub    ClubSummary
	AwayClub    ClubSummary
	Competition CompetitionSummary
}

// BroadcastRecord is a broadcast joined with its evidence source display name.
type BroadcastRecord struct {
	Broadcast
	SourceDisplayName string
}

// NewsRecord is a news link joined with its source display name.
type NewsRecord struct {
	NewsLink
	SourceDisplayName string
}
