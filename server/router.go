package server

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog/v2"

	oapimiddleware "github.com/oapi-codegen/nethttp-middleware"
)

func NewRouter() (http.Handler, error) {
	openAPISpec, err := loadOpenAPISpec()
	if err != nil {
		return nil, err
	}

	router := chi.NewRouter()
	router.Use(
		// Logging middleware
		// Includes recoverer middleware
		httplog.RequestLogger(
			httplog.NewLogger(
				"t4g",
				httplog.Options{
					LogLevel:       slog.LevelDebug,
					RequestHeaders: true,
					Concise:        true,
				},
			),
		),

		// Request validation middleware
		oapimiddleware.OapiRequestValidatorWithOptions(
			openAPISpec,
			&oapimiddleware.Options{
				SilenceServersWarning: true,
			},
		),
	)

	// Create route handler for OpenAPI routes
	openAPIHandler := NewStrictHandler(NewServer(), nil)

	return HandlerFromMux(openAPIHandler, router), nil
}

func loadOpenAPISpec() (*openapi3.T, error) {
	// Load openapi spec
	spec, err := GetSwagger()
	if err != nil {
		return nil, fmt.Errorf("failed to load openapi spec: %w", err)
	}

	// Set server endpoint to the route. This ensures request validation
	// doesn't fail in the validation middleware.
	// See: https://github.com/deepmap/oapi-codegen/issues/1123
	spec.Servers = openapi3.Servers{&openapi3.Server{URL: "/"}}

	return spec, nil
}
