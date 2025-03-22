package mid

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/bobg/errors"
)

// Session is the type of a session stored in a [SessionStore].
type Session interface {
	// CSRFKey is a persistent random bytestring that can be used for CSRF protection.
	CSRFKey() [sha256.Size]byte

	// Active is true when the session is created and false after it is canceled (via SessionStore.Cancel).
	Active() bool

	// Exp is the expiration time of the session.
	Exp() time.Time
}

const csrfNonceLen = 16

// CSRFToken generates a new token containing a random nonce hashed with this session's CSRF key.
// It can be used to protect against CSRF attacks.
// Resources served by the application (e.g. HTML pages) should include a CSRF token.
// State-changing requests to the application that rely on a Session for authentication
// should require the caller to supply a valid CSRF token.
// Validity can be checked with CSRFCheck.
// For more on this topic see https://en.wikipedia.org/wiki/Cross-site_request_forgery.
func CSRFToken(s Session) (string, error) {
	var buf [csrfNonceLen + sha256.Size]byte
	_, err := rand.Read(buf[:csrfNonceLen])
	if err != nil {
		return "", errors.Wrap(err, "generating random nonce")
	}
	sum, err := csrfSum(s, buf[:])
	if err != nil {
		return "", err
	}
	copy(buf[csrfNonceLen:], sum)
	return base64.RawURLEncoding.EncodeToString(buf[:]), nil
}

// ErrCSRF is the error produced when an invalid CSRF token is presented to [CSRFCheck].
var ErrCSRF = errors.New("CSRF check failed")

// CSRFCheck checks a CSRF token against a session for validity.
// TODO: check s is active and unexpired?
func CSRFCheck(s Session, inp string) error {
	got, err := base64.RawURLEncoding.DecodeString(inp)
	if err != nil {
		return errors.Wrap(err, "decoding base64")
	}
	if len(got) != csrfNonceLen+sha256.Size {
		return ErrCSRF
	}
	want, err := csrfSum(s, got)
	if err != nil {
		return err
	}
	if !hmac.Equal(got[csrfNonceLen:], want) {
		return ErrCSRF
	}
	return nil
}

func csrfSum(s Session, inp []byte) ([]byte, error) {
	csrfKey := s.CSRFKey()
	h := hmac.New(sha256.New, csrfKey[:])
	_, err := h.Write(inp[:csrfNonceLen])
	if err != nil {
		return nil, errors.Wrap(err, "computing HMAC")
	}
	return h.Sum(nil), nil
}

// ErrNoSession is the error produced by [SessionStore.Get] when no matching session is found.
var ErrNoSession = errors.New("no session")

// SessionStore is persistent storage for session objects.
type SessionStore interface {
	// Get gets the session with the given key.
	// If no such session is found, it returns [ErrNoSession].
	Get(context.Context, string) (Session, error)

	// Cancel cancels the session with the given unique key.
	// If the session does not exist, or is already canceled or expired,
	// this function silently succeeds.
	Cancel(context.Context, string) error
}

// GetSession checks for a session cookie in a given HTTP request
// and gets the corresponding session from the store.
func GetSession(ctx context.Context, store SessionStore, cookieName string, req *http.Request) (Session, error) {
	cookie, err := req.Cookie(cookieName)
	if err != nil {
		return nil, errors.Wrap(err, "getting session cookie")
	}
	return store.Get(ctx, cookie.Value)
}

// IsNoSession tests whether the given error is either [ErrNoSession] or [http.ErrNoCookie].
func IsNoSession(err error) bool {
	return errors.Is(err, http.ErrNoCookie) || errors.Is(err, ErrNoSession)
}

// SessionHandler is an [http.Handler] middleware wrapper.
// It checks the incoming request for a session in the given store.
// If one is found, the request's context is decorated with the session.
// It can be retrieved by the next handler with [ContextSession].
// If an active, unexpired session is not found, a 403 Forbidden error is returned.
func SessionHandler(store SessionStore, cookieName string, next http.Handler) http.Handler {
	return Err(func(w http.ResponseWriter, req *http.Request) error {
		ctx := req.Context()
		s, err := GetSession(ctx, store, cookieName, req)
		if IsNoSession(err) {
			return CodeErr{C: http.StatusForbidden, Err: err}
		}
		if err != nil {
			return errors.Wrap(err, "getting session")
		}
		if !s.Active() || s.Exp().Before(time.Now()) {
			return CodeErr{C: http.StatusForbidden, Err: fmt.Errorf("session inactive or expired")}
		}
		ctx = context.WithValue(ctx, sessKeyType{}, s)
		req = req.WithContext(ctx)
		next.ServeHTTP(w, req)
		return nil
	})
}

type sessKeyType struct{}

// ContextSession returns the [Session] associated with a context (by [SessionHandler]), if there is one.
// If there isn't, this returns nil.
func ContextSession(ctx context.Context) Session {
	s, _ := ctx.Value(sessKeyType{}).(Session)
	return s
}
