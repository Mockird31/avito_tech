package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Mockird31/avito_tech/config"
	"github.com/Mockird31/avito_tech/pkg/logger"
	"github.com/Mockird31/avito_tech/pkg/postgres"
	"github.com/gorilla/mux"
	"go.uber.org/zap"

	appRouter "github.com/Mockird31/avito_tech/internal/app/router"
	"github.com/Mockird31/avito_tech/internal/middleware"
)

func Run(cfg *config.Config) {
	logger, err := logger.NewZapLogger()
	if err != nil {
		logger.Error("Error creating logger:", zap.Error(err))
		return
	}

	postgresConn, err := postgres.ConnectPostgres(cfg.Postgres)
	if err != nil {
		logger.Error("failed to connect to postgres:", zap.Error(err))
		return
	}
	defer func() {
		if err := postgresConn.Close(); err != nil {
			logger.Error("Error closing Postgres:", zap.Error(err))
		}
	}()

	err = postgres.RunMigrations(cfg.Postgres)
	if err != nil {
		logger.Error("Error running migrations:", zap.Error(err))
		return
	}

	r := mux.NewRouter()

	r.Use(middleware.LoggerMiddleware(logger))

	appRouter.TeamRouter(r, postgresConn)
	appRouter.UserRouter(r, postgresConn)
	appRouter.PullRequestRouter(r, postgresConn)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: r,
	}

	err = srv.ListenAndServe()
	if err != nil {
		logger.Error("Error starting server:", zap.Error(err))
		return
	}

	shutDown := make(chan os.Signal, 1)
	signal.Notify(shutDown, syscall.SIGINT, syscall.SIGTERM)
	<-shutDown

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = srv.Shutdown(ctx)
	if err != nil {
		logger.Error("Error shutting down server:", zap.Error(err))
		os.Exit(1)
	}
}
