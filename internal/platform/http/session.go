package httpplatform

// SessionCookieName is the HttpOnly cookie carrying the opaque session token.
// Features that need the current user must read this cookie and resolve it via
// a SessionResolver port — they must not import features/auth.
const SessionCookieName = "cdj_session"
