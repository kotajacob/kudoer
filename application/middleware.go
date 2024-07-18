// License: AGPL-3.0-only
// (c) 2024 Dakota Walsh <kota@nilsu.org>
package application

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"

	"git.sr.ht/~kota/kudoer/ui"
	"github.com/justinas/nosurf"
)

// FromHash removes the hash from static files so they can be served from the
// embeded files (which do not contain a hash in the name).
//
// Additionally, the cache-control header is set to an extremely high value so
// as to keep these assets caches as long as the browser is willing.
func (app *application) FromHash(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := ui.FromHash(r.URL.Path)
		rp := ui.FromHash(r.URL.RawPath)

		w.Header().Set("cache-control", "max-age=31557600, immutable")

		r2 := new(http.Request)
		*r2 = *r
		r2.URL = new(url.URL)
		*r2.URL = *r.URL
		r2.URL.Path = p
		r2.URL.RawPath = rp
		next.ServeHTTP(w, r2)
	})
}

// noSurf sets a random CSRF token as a cookie.
// This is then checked in potentially vulnerable forms against a random embeded
// field value to prevent cross-site request forgery attacks.
func noSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)
	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   true,
	})
	return csrfHandler
}

// cspNonce securely generates a 128bit base64 encoded number.
func cspNonce() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	return base64.RawStdEncoding.EncodeToString(b), err
}

// secureHeaders is a middleware which adds strict CSP and other headers.
// A CSP nonce is stored in the request's context which can be retrieved with
// the nonce helper function.
func (app *application) secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nonce, err := cspNonce()
		if err != nil {
			app.serverError(w, err)
			return
		}
		w.Header().Set(
			"Content-Security-Policy",
			"default-src 'none'; script-src 'nonce-"+
				nonce+"'; style-src 'nonce-"+
				nonce+"'; img-src 'self' https: data:",
		)
		w.Header().Set("Referrer-Policy", "origin-when-cross-origin")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "deny")
		w.Header().Set("X-XSS-Protection", "0")
		r = r.WithContext(context.WithValue(r.Context(), "nonce", nonce))

		next.ServeHTTP(w, r)
	})
}

// nonce retrieves a stored nonce string from a request's context.
func nonce(c context.Context) string {
	if val, ok := c.Value("nonce").(string); ok {
		return val
	}
	return ""
}

// logRequest is a middleware that prints each request to the info log.
func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.infoLog.Printf(
			"%s - %s %s %s",
			r.RemoteAddr,
			r.Proto,
			r.Method,
			r.URL.RequestURI(),
		)
		next.ServeHTTP(w, r)
	})
}

// recoverPanic is a middleware which recovers from a panic and logs the error.
func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.serverError(w, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// requireAuthentication is a middleware which redirects a user to a login page
// if they are not authenticated.
func (app *application) requireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if app.authenticated(r) == "" {
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
			return
		}
		w.Header().Add("Cache-Control", "no-store")

		next.ServeHTTP(w, r)
	})
}
