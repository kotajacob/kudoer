package application

const (
	// SessionKeyUsername is the key holding a string of the session's
	// authenticated username.
	SessionKeyUsername = "authenticatedUsername"

	// SessionKeyAdmin is the key holding a boolean that reports if the
	// session's logged in user is an admin account.
	SessionKeyAdmin = "isAdmin"
)

// ContextKey is a custom type for keys kudoer adds to an http context.
//
// This is needed to avoid collisions with other packages that modify the http
// context.
type ContextKey string

const (
	ContextKeyUsername ContextKey = "username"
	ContextKeyNonce    ContextKey = "nonce"
)

