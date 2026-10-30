// Package apperr is the service's error type. Every error a handler or
// middleware returns is wrapped here with a unique error ID and a Kind; the
// router's error handler turns it into a response.
package apperr

import (
	"errors"
	"net/http"

	"github.com/michaelolof/gofi"
)

// Kind classifies an error and decides its HTTP status.
type Kind int

const (
	Internal Kind = iota
	InvalidInput
	Unauthorized
	Forbidden
	NotFound
	Conflict
)

// Status returns the HTTP status code for k.
func (k Kind) Status() int {
	switch k {
	case InvalidInput:
		return http.StatusBadRequest
	case Unauthorized:
		return http.StatusUnauthorized
	case Forbidden:
		return http.StatusForbidden
	case NotFound:
		return http.StatusNotFound
	case Conflict:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

// Error is an error with a unique ID, a Kind and an optional public message.
type Error struct {
	// ID is a unique integer literal written at the call site: the Unix time
	// in milliseconds when the line was written. Searching the codebase for
	// it finds the exact line that produced the error.
	ID   int64
	Kind Kind
	// Msg is safe to show to clients. Empty means "use the default for Kind".
	Msg string
	Err error

	status int // set only for errors converted from *gofi.HTTPError
}

// New returns an error that has no underlying cause.
func New(id int64, kind Kind, msg string) *Error {
	return &Error{ID: id, Kind: kind, Msg: msg}
}

// Wrap returns err wrapped with an ID and a Kind.
func Wrap(id int64, kind Kind, err error) *Error {
	return &Error{ID: id, Kind: kind, Err: err}
}

// WithMsg sets the message shown to clients.
func (e *Error) WithMsg(msg string) *Error {
	e.Msg = msg
	return e
}

func (e *Error) Error() string {
	switch {
	case e.Err != nil && e.Msg != "":
		return e.Msg + ": " + e.Err.Error()
	case e.Err != nil:
		return e.Err.Error()
	default:
		return e.Msg
	}
}

func (e *Error) Unwrap() error { return e.Err }

// HTTPStatus returns the response status code for e.
func (e *Error) HTTPStatus() int {
	if e.status != 0 {
		return e.status
	}
	return e.Kind.Status()
}

// PublicMessage returns the message that is safe to send to clients.
func (e *Error) PublicMessage() string {
	if e.Msg != "" {
		return e.Msg
	}
	// Gofi's request validation messages name the field and the failed rule
	// and contain no server state, so they are safe to show.
	var verr gofi.ValidationError
	if e.Kind == InvalidInput && errors.As(e.Err, &verr) {
		return verr.Error()
	}
	return http.StatusText(e.HTTPStatus())
}

// From converts any error into an *Error. Errors that were not wrapped are
// treated as internal errors with ID 0, except Gofi's own errors: request
// validation failures and *gofi.HTTPError values (for example a 413 when the
// body is too large) keep their meaning.
func From(err error) *Error {
	var e *Error
	if errors.As(err, &e) {
		return e
	}
	var verr gofi.ValidationError
	if errors.As(err, &verr) {
		return Wrap(0, InvalidInput, err)
	}
	var herr *gofi.HTTPError
	if errors.As(err, &herr) {
		e := Wrap(0, kindForStatus(herr.Code), err)
		e.status = herr.Code
		return e
	}
	return Wrap(0, Internal, err)
}

func kindForStatus(code int) Kind {
	switch code {
	case http.StatusBadRequest:
		return InvalidInput
	case http.StatusUnauthorized:
		return Unauthorized
	case http.StatusForbidden:
		return Forbidden
	case http.StatusNotFound:
		return NotFound
	case http.StatusConflict:
		return Conflict
	default:
		return Internal
	}
}
