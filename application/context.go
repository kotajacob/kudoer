package application

// ContextKey is a custom type for keys kudoer adds to an http context.
//
// This is needed to avoid collisions with other packages that modify the http
// context.
type ContextKey string

const (
	ContextKeyUsername ContextKey = "username"
	ContextKeyNonce    ContextKey = "nonce"
)
