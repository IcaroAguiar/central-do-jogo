package api

// SSR page discriminators embedded in #initial-data (PAT-004).
// Values must stay identical to the OpenAPI SSRPage enum and web/src/lib/pages.ts.
const (
	PageHome  = "home"
	PageClub  = "club"
	PageMatch = "match"
)

// AllSSRPages lists every valid SSR page ID in stable order.
func AllSSRPages() []string {
	return []string{PageHome, PageClub, PageMatch}
}

// ValidSSRPage reports whether page is a known SSR discriminator.
func ValidSSRPage(page string) bool {
	switch page {
	case PageHome, PageClub, PageMatch:
		return true
	default:
		return false
	}
}
