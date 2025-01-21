package auth

// ContextKey is a custom type for context keys to avoid collisions
type ContextKey string

const (
	// ClaimsKey is the context key for JWT claims
	ClaimsKey ContextKey = "claims"
)
