// Package brasilia provides the product timezone (CON-008): America/Sao_Paulo
// as a fixed UTC-3 offset. Brazil has not observed DST since 2019, so a fixed
// zone avoids depending on tzdata in distroless images.
package brasilia

import "time"

// Zone is Brasília time (UTC-3).
var Zone = time.FixedZone("-03:00", -3*60*60)

// Format formats t in Brasília as dd/mm/yyyy HH:MM. Empty when t is nil.
func Format(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.In(Zone).Format("02/01/2006 15:04")
}

// FormatWithLabel appends the explicit product timezone label.
func FormatWithLabel(t *time.Time) string {
	formatted := Format(t)
	if formatted == "" {
		return ""
	}
	return formatted + " (Horário de Brasília)"
}
