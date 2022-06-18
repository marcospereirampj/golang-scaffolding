package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"go.uber.org/fx"
)

type (
	HTTPHandler interface {
		Method() string
		Pattern() string

		http.Handler
	}

	RouterParams struct {
		fx.In
		Handlers []HTTPHandler `group:"handlers"`
	}
)

func NewHTTPRouter(params RouterParams) *chi.Mux {
	router := chi.NewRouter()

	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "healthy")
	})

	router.Route("/project-name", func(r chi.Router) {
		for _, handler := range params.Handlers {
			r.Route(handler.Pattern(), func(subRoute chi.Router) {
				subRoute.Method(handler.Method(), "/", handler)
			})
		}
	})

	return router
}

func StartHTTPServer(lifecycle fx.Lifecycle, router *chi.Mux) {
	server := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  5000 * time.Second,
		WriteTimeout: 5000 * time.Second,
	}

	lifecycle.Append(fx.Hook{
		OnStart: func(context.Context) error {
			fmt.Println("Starting server...")
			go server.ListenAndServe()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return server.Shutdown(ctx)
		},
	})
}

func Serve() {
	ServerDependencies := fx.Provide(
		NewHTTPRouter)

	app := fx.New(
		fx.Options(
			ServerDependencies,
		),
		fx.Invoke(StartHTTPServer),
	)

	app.Run()
}
