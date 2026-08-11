// Package api provides support for the web layer of the application.
package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/ardanlabs/practical-ai-gcuk-2026/app/sdk/errs"
)

// Handler represents a handler function that must be supplied for a route.
type Handler func(ctx context.Context, r *http.Request) (any, error)

// Statuser is implemented by any value that knows its own HTTP status code.
// Errors use it to select the response status; handler results use it to
// return a status other than 200 OK.
type Statuser interface {
	HTTPStatus() int
}

// StatusData pairs a handler result with an explicit HTTP status code. A nil
// Data field results in a response with the chosen status and no body.
type StatusData struct {
	Data any
	Code int
}

// NewStatus constructs a StatusData so a handler can control the status code
// of a successful response, for example 201 Created or 204 No Content.
func NewStatus(code int, data any) StatusData {
	return StatusData{
		Data: data,
		Code: code,
	}
}

// HTTPStatus implements the Statuser interface.
func (sd StatusData) HTTPStatus() int {
	return sd.Code
}

// Wrap adapts a Handler into an http.HandlerFunc, taking care of encoding the
// response and any error that comes back from the handler.
func Wrap(log *slog.Logger, handler Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		data, err := handler(ctx, r)
		if err != nil {
			log.ErrorContext(ctx, "request failed", "path", r.URL.Path, "method", r.Method, "err", err)
			respondError(ctx, log, w, err)

			return
		}

		statusCode := http.StatusOK

		switch v := data.(type) {
		case StatusData:
			statusCode = v.Code
			data = v.Data

		case Statuser:
			statusCode = v.HTTPStatus()
		}

		if err := Respond(w, statusCode, data); err != nil {
			log.ErrorContext(ctx, "respond failed", "path", r.URL.Path, "err", err)
		}
	}
}

// Respond writes the data as JSON with the specified status code. A nil data
// value results in a response with no content.
func Respond(w http.ResponseWriter, statusCode int, data any) error {
	if data == nil {
		w.WriteHeader(statusCode)

		return nil
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if _, err := w.Write(jsonData); err != nil {
		return fmt.Errorf("write: %w", err)
	}

	return nil
}

func respondError(ctx context.Context, log *slog.Logger, w http.ResponseWriter, err error) {
	statusCode := http.StatusInternalServerError
	data := any(errs.Error{Message: http.StatusText(http.StatusInternalServerError)})

	var st Statuser
	if errors.As(err, &st) {
		statusCode = st.HTTPStatus()
		data = err
	}

	if err := Respond(w, statusCode, data); err != nil {
		log.ErrorContext(ctx, "respond error failed", "err", err)
	}
}
