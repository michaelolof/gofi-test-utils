package middleware

import (
	"encoding/json"
	"log/slog"

	"github.com/michaelolof/gofi"

	"example.com/projectsvc/internal/apperr"
	"example.com/projectsvc/internal/http/schemas"
)

// ErrorHandler returns the router's error handler. Every error a handler or
// middleware returns ends up here: it is logged with its ID and written as
// a schemas.ErrBody.
func ErrorHandler(log *slog.Logger) func(err error, c gofi.Context) {
	return func(err error, c gofi.Context) {
		e := apperr.From(err)
		status := e.HTTPStatus()

		level := slog.LevelWarn
		if status >= 500 {
			level = slog.LevelError
		}
		log.Log(c.Context(), level, "request failed",
			"error_id", e.ID,
			"status", status,
			"method", c.Method(),
			"path", c.Path(),
			"request_id", GetRequestID(c),
			"error", err.Error(),
		)

		body, _ := json.Marshal(schemas.ErrBody{
			Status:  "error",
			ErrorID: e.ID,
			Message: e.PublicMessage(),
		})
		c.Writer().Header().Set("Content-Type", "application/json")
		_ = c.SendBytes(status, body)
	}
}
