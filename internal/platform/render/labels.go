package render

import (
	"fmt"

	"github.com/IcaroAguiar/central-do-jogo/internal/domain"
)

// AvailabilityLabel returns pt-BR copy for REQ-010 gap states (SSR parity with
// web/src/lib/availability.ts).
func AvailabilityLabel(state string) string {
	switch domain.AvailabilityState(state) {
	case domain.AvailabilityAvailable:
		return "Disponível"
	case domain.AvailabilityAwaitingPublication:
		return "Aguardando divulgação oficial"
	case domain.AvailabilityNotFound:
		return "Não encontrado nas fontes monitoradas"
	case domain.AvailabilityDivergent:
		return "Fontes divergentes"
	case domain.AvailabilityNoCoverage:
		return "Sem cobertura para esta competição"
	default:
		return state
	}
}

// AccessLabel returns pt-BR copy for broadcast access.
func AccessLabel(access string) string {
	switch domain.AccessType(access) {
	case domain.AccessFree:
		return "Gratuito"
	case domain.AccessSubscription:
		return "Assinatura"
	case domain.AccessUnknown:
		return "Acesso a confirmar"
	default:
		return access
	}
}

// ConfidenceLabel returns pt-BR copy for confidence bands.
func ConfidenceLabel(level string) string {
	switch domain.ConfidenceLevel(level) {
	case domain.ConfidenceHigh:
		return "Confiança alta"
	case domain.ConfidenceMedium:
		return "Confiança média"
	case domain.ConfidenceLow:
		return "Confiança baixa"
	default:
		return level
	}
}

// KickoffStateLabel returns pt-BR copy for kickoff states.
func KickoffStateLabel(state string) string {
	switch domain.KickoffState(state) {
	case domain.KickoffPublished:
		return "publicada"
	case domain.KickoffIndefinite:
		return "indefinida"
	case domain.KickoffChanged:
		return "alterada"
	default:
		return state
	}
}

// LastAttemptSentence renders the REQ-010 last-attempt clause in pt-BR.
func LastAttemptSentence(formatted string) string {
	if formatted == "" {
		return "Nenhuma tentativa registrada ainda."
	}
	return fmt.Sprintf("Última tentativa em %s.", formatted)
}
