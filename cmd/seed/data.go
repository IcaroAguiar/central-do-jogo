package main

import (
	"time"

	"github.com/IcaroAguiar/central-do-jogo/internal/domain"
)

// Deterministic identifiers for the manual seed source. Real adapters
// register their own sources; this one exists only so seeded evidence rows
// satisfy the foreign key constraints without depending on ingest (out of
// scope for this seed).
const (
	seedSeason            = 2026
	seedCompetitionSlug   = "brasileirao"
	seedCompetitionName   = "Campeonato Brasileiro Série A"
	seedSourceID          = "manual_seed"
	seedSourceDisplayName = "Central do Jogo (dados de exemplo)"
	seedSourceHomeURL     = "https://www.centraldojogo.com.br"
)

type clubSeed struct {
	Slug      string
	Name      string
	ShortName string
	Aliases   []string
}

// serieAClubs lists the 20 Serie A 2026 clubs supported by this seed (REQ-002).
var serieAClubs = []clubSeed{
	{Slug: "atletico-mineiro", Name: "Clube Atlético Mineiro", ShortName: "Atlético-MG", Aliases: []string{"Galo"}},
	{Slug: "bahia", Name: "Esporte Clube Bahia", ShortName: "Bahia", Aliases: []string{"Tricolor de Aço"}},
	{Slug: "botafogo", Name: "Botafogo de Futebol e Regatas", ShortName: "Botafogo", Aliases: []string{"Fogão"}},
	{Slug: "bragantino", Name: "Red Bull Bragantino", ShortName: "Bragantino", Aliases: []string{"Massa Bruta"}},
	{Slug: "corinthians", Name: "Sport Club Corinthians Paulista", ShortName: "Corinthians", Aliases: []string{"Timão"}},
	{Slug: "cruzeiro", Name: "Cruzeiro Esporte Clube", ShortName: "Cruzeiro", Aliases: []string{"Raposa"}},
	{Slug: "flamengo", Name: "Clube de Regatas do Flamengo", ShortName: "Flamengo", Aliases: []string{"Mengão", "Fla"}},
	{Slug: "fluminense", Name: "Fluminense Football Club", ShortName: "Fluminense", Aliases: []string{"Flu", "Tricolor das Laranjeiras"}},
	{Slug: "fortaleza", Name: "Fortaleza Esporte Clube", ShortName: "Fortaleza", Aliases: []string{"Leão do Pici"}},
	{Slug: "gremio", Name: "Grêmio Foot-Ball Porto Alegrense", ShortName: "Grêmio", Aliases: []string{"Imortal Tricolor"}},
	{Slug: "internacional", Name: "Sport Club Internacional", ShortName: "Internacional", Aliases: []string{"Inter", "Colorado"}},
	{Slug: "juventude", Name: "Esporte Clube Juventude", ShortName: "Juventude", Aliases: []string{"Papo"}},
	{Slug: "mirassol", Name: "Mirassol Futebol Clube", ShortName: "Mirassol", Aliases: []string{"Leão"}},
	{Slug: "palmeiras", Name: "Sociedade Esportiva Palmeiras", ShortName: "Palmeiras", Aliases: []string{"Verdão"}},
	{Slug: "santos", Name: "Santos Futebol Clube", ShortName: "Santos", Aliases: []string{"Peixe"}},
	{Slug: "sao-paulo", Name: "São Paulo Futebol Clube", ShortName: "São Paulo", Aliases: []string{"Tricolor Paulista"}},
	{Slug: "sport", Name: "Sport Club do Recife", ShortName: "Sport", Aliases: []string{"Leão da Ilha"}},
	{Slug: "vasco", Name: "Club de Regatas Vasco da Gama", ShortName: "Vasco", Aliases: []string{"Gigante da Colina"}},
	{Slug: "vitoria", Name: "Esporte Clube Vitória", ShortName: "Vitória", Aliases: []string{"Leão da Barra"}},
	{Slug: "ceara", Name: "Ceará Sporting Club", ShortName: "Ceará", Aliases: []string{"Vozão"}},
}

type broadcastSeed struct {
	Channel     string
	Platform    string
	Access      domain.AccessType
	Region      string
	OfficialURL string
	Confidence  domain.ConfidenceLevel
}

type lineupPlayerSeed struct {
	ShirtNumber string
	Name        string
	IsStarter   bool
}

type lineupSeed struct {
	Side      domain.LineupSide
	Formation string
	Coach     string
	Players   []lineupPlayerSeed
	Official  bool
	Suffix    string // disambiguates deterministic IDs when a side has more than one claim (divergent case)
}

type newsSeed struct {
	Title       string
	URL         string
	HoursBefore int // hours before kickoff the article was published
}

// matchSeed describes one seeded fixture plus its broadcast/lineup/news
// surfaces, deliberately spanning the REQ-004 kickoff states and REQ-010
// availability states so the public journeys have realistic gaps to render.
type matchSeed struct {
	Slug     string
	HomeSlug string
	AwaySlug string
	Round    string
	Venue    string

	// KickoffOffset is added to the seed run time to compute KickoffAt.
	// Nil means the kickoff time is not yet known (indefinite state).
	KickoffOffset *time.Duration
	KickoffState  domain.KickoffState

	Broadcasts         []broadcastSeed
	BroadcastState     domain.AvailabilityState
	BroadcastAttempted bool

	Lineups         []lineupSeed
	LineupState     domain.AvailabilityState
	LineupAttempted bool

	News          []newsSeed
	NewsState     domain.AvailabilityState
	NewsAttempted bool
}

func hours(h int) *time.Duration {
	d := time.Duration(h) * time.Hour
	return &d
}

// seedMatches covers 12 fixtures across two rounds, using all 20 seeded
// clubs at least once and exercising every KickoffState and
// AvailabilityState value.
var seedMatches = []matchSeed{
	{
		Slug: "flamengo-x-corinthians-2026-r1", HomeSlug: "flamengo", AwaySlug: "corinthians",
		Round: "Rodada 1", Venue: "Maracanã", KickoffOffset: hours(24), KickoffState: domain.KickoffPublished,
		Broadcasts: []broadcastSeed{
			{Channel: "TV Globo", Platform: "", Access: domain.AccessFree, Region: "Nacional", OfficialURL: "https://globoplay.globo.com", Confidence: domain.ConfidenceHigh},
			{Channel: "Premiere", Platform: "Globoplay", Access: domain.AccessSubscription, Region: "Nacional", OfficialURL: "https://globoplay.globo.com/premiere", Confidence: domain.ConfidenceHigh},
		},
		BroadcastState: domain.AvailabilityAvailable,
		Lineups: []lineupSeed{
			{Side: domain.LineupHome, Formation: "4-3-3", Coach: "Técnico Rubro-Negro", Official: true, Players: []lineupPlayerSeed{
				{ShirtNumber: "1", Name: "Goleiro Titular", IsStarter: true},
				{ShirtNumber: "9", Name: "Centroavante Titular", IsStarter: true},
				{ShirtNumber: "20", Name: "Reserva de Ataque", IsStarter: false},
			}},
			{Side: domain.LineupAway, Formation: "4-4-2", Coach: "Técnico Alvinegro", Official: true, Players: []lineupPlayerSeed{
				{ShirtNumber: "1", Name: "Goleiro Corinthiano", IsStarter: true},
				{ShirtNumber: "10", Name: "Meia Titular", IsStarter: true},
			}},
		},
		LineupState: domain.AvailabilityAvailable,
		News: []newsSeed{
			{Title: "Escalações confirmadas para o clássico", URL: "https://exemplo-noticias.com.br/flamengo-corinthians-escalacoes", HoursBefore: 3},
			{Title: "Onde assistir Flamengo x Corinthians", URL: "https://exemplo-noticias.com.br/flamengo-corinthians-onde-assistir", HoursBefore: 30},
			{Title: "Retrospecto do confronto direto", URL: "https://exemplo-noticias.com.br/flamengo-corinthians-retrospecto", HoursBefore: 48},
		},
		NewsState: domain.AvailabilityAvailable,
	},
	{
		Slug: "palmeiras-x-sao-paulo-2026-r1", HomeSlug: "palmeiras", AwaySlug: "sao-paulo",
		Round: "Rodada 1", Venue: "Allianz Parque", KickoffOffset: hours(30), KickoffState: domain.KickoffPublished,
		Broadcasts: []broadcastSeed{
			{Channel: "Premiere", Platform: "Globoplay", Access: domain.AccessSubscription, Region: "Nacional", OfficialURL: "https://globoplay.globo.com/premiere", Confidence: domain.ConfidenceMedium},
		},
		BroadcastState:  domain.AvailabilityAvailable,
		LineupState:     domain.AvailabilityAwaitingPublication,
		LineupAttempted: true,
		News: []newsSeed{
			{Title: "Onde assistir ao Choque-Rei", URL: "https://exemplo-noticias.com.br/palmeiras-sao-paulo-onde-assistir", HoursBefore: 20},
			{Title: "Desfalques dos dois lados", URL: "https://exemplo-noticias.com.br/palmeiras-sao-paulo-desfalques", HoursBefore: 40},
		},
		NewsState: domain.AvailabilityAvailable,
	},
	{
		Slug: "gremio-x-internacional-2026-r1", HomeSlug: "gremio", AwaySlug: "internacional",
		Round: "Rodada 1", Venue: "Arena do Grêmio", KickoffOffset: hours(48), KickoffState: domain.KickoffPublished,
		Broadcasts: []broadcastSeed{
			{Channel: "SporTV", Platform: "Globoplay", Access: domain.AccessSubscription, Region: "Sul", OfficialURL: "https://globoplay.globo.com/sportv", Confidence: domain.ConfidenceMedium},
			{Channel: "Grêmio TV", Platform: "YouTube", Access: domain.AccessFree, Region: "Regional", OfficialURL: "", Confidence: domain.ConfidenceLow},
		},
		BroadcastState: domain.AvailabilityDivergent,
		Lineups: []lineupSeed{
			{Side: domain.LineupHome, Formation: "4-2-3-1", Coach: "Técnico Tricolor", Official: true, Players: []lineupPlayerSeed{
				{ShirtNumber: "1", Name: "Goleiro Gremista", IsStarter: true},
			}},
			{Side: domain.LineupAway, Formation: "4-3-3", Coach: "Técnico Colorado", Official: true, Players: []lineupPlayerSeed{
				{ShirtNumber: "1", Name: "Goleiro Colorado", IsStarter: true},
			}},
		},
		LineupState:   domain.AvailabilityAvailable,
		NewsState:     domain.AvailabilityAwaitingPublication,
		NewsAttempted: true,
	},
	{
		Slug: "cruzeiro-x-atletico-mineiro-2026-r1", HomeSlug: "cruzeiro", AwaySlug: "atletico-mineiro",
		Round: "Rodada 1", Venue: "Mineirão", KickoffOffset: hours(52), KickoffState: domain.KickoffPublished,
		BroadcastState:     domain.AvailabilityNotFound,
		BroadcastAttempted: true,
		Lineups: []lineupSeed{
			{Side: domain.LineupHome, Formation: "4-4-2", Coach: "Técnico Celeste", Official: true, Players: []lineupPlayerSeed{
				{ShirtNumber: "1", Name: "Goleiro Celeste", IsStarter: true},
			}},
			{Side: domain.LineupAway, Formation: "4-3-3", Coach: "Técnico Galo", Official: true, Players: []lineupPlayerSeed{
				{ShirtNumber: "1", Name: "Goleiro do Galo", IsStarter: true},
			}},
		},
		LineupState: domain.AvailabilityAvailable,
		News: []newsSeed{
			{Title: "Tudo sobre o Clássico Mineiro", URL: "https://exemplo-noticias.com.br/cruzeiro-atletico-mineiro", HoursBefore: 24},
		},
		NewsState: domain.AvailabilityAvailable,
	},
	{
		Slug: "botafogo-x-fluminense-2026-r1", HomeSlug: "botafogo", AwaySlug: "fluminense",
		Round: "Rodada 1", Venue: "Estádio Nilton Santos", KickoffOffset: hours(72), KickoffState: domain.KickoffPublished,
		Broadcasts: []broadcastSeed{
			{Channel: "Bandsports", Platform: "", Access: domain.AccessSubscription, Region: "Rio de Janeiro", OfficialURL: "", Confidence: domain.ConfidenceLow},
		},
		BroadcastState:  domain.AvailabilityAvailable,
		LineupState:     domain.AvailabilityNotFound,
		LineupAttempted: true,
		NewsState:       domain.AvailabilityNotFound,
		NewsAttempted:   true,
	},
	{
		Slug: "vasco-x-fortaleza-2026-r1", HomeSlug: "vasco", AwaySlug: "fortaleza",
		Round: "Rodada 1", Venue: "Estádio São Januário", KickoffOffset: hours(76), KickoffState: domain.KickoffPublished,
		BroadcastState: domain.AvailabilityNoCoverage,
		LineupState:    domain.AvailabilityNoCoverage,
		NewsState:      domain.AvailabilityNoCoverage,
	},
	{
		Slug: "bahia-x-sport-2026-r1", HomeSlug: "bahia", AwaySlug: "sport",
		Round: "Rodada 1", Venue: "Arena Fonte Nova", KickoffOffset: hours(96), KickoffState: domain.KickoffPublished,
		Broadcasts: []broadcastSeed{
			{Channel: "SporTV", Platform: "Globoplay", Access: domain.AccessSubscription, Region: "Nordeste", OfficialURL: "https://globoplay.globo.com/sportv", Confidence: domain.ConfidenceMedium},
		},
		BroadcastState: domain.AvailabilityAvailable,
		Lineups: []lineupSeed{
			{Side: domain.LineupHome, Formation: "4-3-3", Coach: "Técnico Tricolor Baiano", Official: true, Players: []lineupPlayerSeed{
				{ShirtNumber: "1", Name: "Goleiro Baiano", IsStarter: true},
			}},
			{Side: domain.LineupAway, Formation: "4-4-2", Coach: "Técnico Rubro-Negro Pernambucano", Official: true, Players: []lineupPlayerSeed{
				{ShirtNumber: "1", Name: "Goleiro Pernambucano", IsStarter: true},
			}},
		},
		LineupState: domain.AvailabilityAvailable,
		News: []newsSeed{
			{Title: "Onde assistir Bahia x Sport", URL: "https://exemplo-noticias.com.br/bahia-sport-onde-assistir", HoursBefore: 12},
		},
		NewsState: domain.AvailabilityAvailable,
	},
	{
		Slug: "bragantino-x-juventude-2026-r1", HomeSlug: "bragantino", AwaySlug: "juventude",
		Round: "Rodada 1", Venue: "Estádio Nabi Abi Chedid", KickoffOffset: hours(100), KickoffState: domain.KickoffPublished,
		BroadcastState:     domain.AvailabilityAwaitingPublication,
		BroadcastAttempted: true,
		LineupState:        domain.AvailabilityAwaitingPublication,
		LineupAttempted:    true,
		News: []newsSeed{
			{Title: "Onde assistir Bragantino x Juventude", URL: "https://exemplo-noticias.com.br/bragantino-juventude", HoursBefore: 18},
		},
		NewsState: domain.AvailabilityAvailable,
	},
	{
		Slug: "mirassol-x-ceara-2026-r1", HomeSlug: "mirassol", AwaySlug: "ceara",
		Round: "Rodada 1", Venue: "Estádio Municipal José Maria de Campos Maia", KickoffOffset: hours(104), KickoffState: domain.KickoffPublished,
		Broadcasts: []broadcastSeed{
			{Channel: "Canal do YouTube da Liga", Platform: "YouTube", Access: domain.AccessFree, Region: "Nacional", OfficialURL: "", Confidence: domain.ConfidenceLow},
		},
		BroadcastState: domain.AvailabilityAvailable,
		Lineups: []lineupSeed{
			{Side: domain.LineupHome, Formation: "4-2-3-1", Coach: "Técnico Leão", Official: true, Suffix: "1", Players: []lineupPlayerSeed{
				{ShirtNumber: "1", Name: "Goleiro Mirassolense", IsStarter: true},
			}},
			{Side: domain.LineupHome, Formation: "4-3-3", Coach: "Técnico Leão", Official: false, Suffix: "2", Players: []lineupPlayerSeed{
				{ShirtNumber: "1", Name: "Goleiro Mirassolense", IsStarter: true},
			}},
			{Side: domain.LineupAway, Formation: "4-4-2", Coach: "Técnico Vozão", Official: true, Suffix: "1", Players: []lineupPlayerSeed{
				{ShirtNumber: "1", Name: "Goleiro Cearense", IsStarter: true},
			}},
		},
		LineupState: domain.AvailabilityDivergent,
		News: []newsSeed{
			{Title: "Onde assistir Mirassol x Ceará", URL: "https://exemplo-noticias.com.br/mirassol-ceara", HoursBefore: 8},
		},
		NewsState: domain.AvailabilityAvailable,
	},
	{
		Slug: "vitoria-x-santos-2026-r1", HomeSlug: "vitoria", AwaySlug: "santos",
		Round: "Rodada 1", Venue: "Estádio Manoel Barradas", KickoffOffset: hours(120), KickoffState: domain.KickoffPublished,
		Broadcasts: []broadcastSeed{
			{Channel: "TV Globo", Platform: "", Access: domain.AccessFree, Region: "Nordeste", OfficialURL: "https://globoplay.globo.com", Confidence: domain.ConfidenceHigh},
		},
		BroadcastState: domain.AvailabilityAvailable,
		Lineups: []lineupSeed{
			{Side: domain.LineupHome, Formation: "4-3-3", Coach: "Técnico Rubro-Negro Baiano", Official: true, Players: []lineupPlayerSeed{
				{ShirtNumber: "1", Name: "Goleiro Vitorioso", IsStarter: true},
			}},
			{Side: domain.LineupAway, Formation: "4-2-3-1", Coach: "Técnico Alvinegro Praiano", Official: true, Players: []lineupPlayerSeed{
				{ShirtNumber: "1", Name: "Goleiro Santista", IsStarter: true},
			}},
		},
		LineupState: domain.AvailabilityAvailable,
		NewsState:   domain.AvailabilityDivergent,
	},
	{
		Slug: "flamengo-x-palmeiras-2026-r2", HomeSlug: "flamengo", AwaySlug: "palmeiras",
		Round: "Rodada 2", Venue: "Maracanã", KickoffOffset: nil, KickoffState: domain.KickoffIndefinite,
		BroadcastState:     domain.AvailabilityAwaitingPublication,
		BroadcastAttempted: true,
		LineupState:        domain.AvailabilityAwaitingPublication,
		LineupAttempted:    true,
		NewsState:          domain.AvailabilityAwaitingPublication,
		NewsAttempted:      true,
	},
	{
		Slug: "corinthians-x-sao-paulo-2026-r2", HomeSlug: "corinthians", AwaySlug: "sao-paulo",
		Round: "Rodada 2", Venue: "Neo Química Arena", KickoffOffset: hours(200), KickoffState: domain.KickoffChanged,
		Broadcasts: []broadcastSeed{
			{Channel: "Premiere", Platform: "Globoplay", Access: domain.AccessSubscription, Region: "Nacional", OfficialURL: "https://globoplay.globo.com/premiere", Confidence: domain.ConfidenceHigh},
		},
		BroadcastState:  domain.AvailabilityAvailable,
		LineupState:     domain.AvailabilityAwaitingPublication,
		LineupAttempted: true,
		News: []newsSeed{
			{Title: "Horário do Majestoso foi alterado", URL: "https://exemplo-noticias.com.br/corinthians-sao-paulo-horario-alterado", HoursBefore: 6},
		},
		NewsState: domain.AvailabilityAvailable,
	},
}
