// Package schemas holds the schema blocks shared by handlers: header blocks,
// the success and error response wrappers, and response body types.
package schemas

// AuthHeader is the request header block for authenticated routes.
type AuthHeader struct {
	Authorization string `json:"authorization" validate:"required" description:"Bearer token." example:"Bearer eyJhbGciOi..."`
}

// SuccessSchema is the response block for a successful response carrying T.
type SuccessSchema[T any] struct {
	Body SuccessBody[T]
}

// SuccessBody is the success response payload.
type SuccessBody[T any] struct {
	Status string `json:"status" validate:"required" default:"success"`
	Data   T      `json:"data" validate:"required"`
}

// ErrSchema is the response block for every error response. Declare it as
// the Err field of every route schema.
type ErrSchema struct {
	Body ErrBody
}

// ErrBody is the error response payload written by the router's error handler.
type ErrBody struct {
	Status  string `json:"status" validate:"required" default:"error"`
	ErrorID int64  `json:"error_id" description:"Unique ID of the code path that produced the error. Quote it when reporting a problem."`
	Message string `json:"message" validate:"required"`
}
