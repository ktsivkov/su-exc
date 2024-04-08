package rest

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"github.com/ktsivkov/su-exc/internal/account"
	"github.com/ktsivkov/su-exc/internal/lifecycle"
	"github.com/ktsivkov/su-exc/internal/rest/account/create"
	"github.com/ktsivkov/su-exc/internal/rest/account/topup"
	"github.com/ktsivkov/su-exc/internal/rest/account/transfer"
)

func Boot(ctx context.Context, dbUri string, port int, shutdownGracePeriod time.Duration) error {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// Dependencies
	db, err := sql.Open("postgres", dbUri)
	if err != nil {
		logger.Error("cannot create database connection", "error", err)
		return errors.Wrap(err, "cannot create database connection")
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("cannot close database connection", "error", err)
		}
	}()

	accountRepo, err := account.NewRepository(db)
	if err != nil {
		logger.Error("cannot initialize account repository", "error", err)
		return errors.Wrap(err, "cannot initialize account repository")
	}

	router := ApiRouter(accountRepo, logger)

	addr := fmt.Sprintf(":%d", port)
	srv := &http.Server{
		Handler: router,
		Addr:    addr,

		WriteTimeout: 30 * time.Second,
		ReadTimeout:  30 * time.Second,
	}

	// Graceful shutdown
	gCtx, stop := lifecycle.GetGracefulContext(ctx, func(ctx context.Context, cause error) {
		tCtx, shutdown := context.WithTimeout(ctx, shutdownGracePeriod)
		defer shutdown()

		logger.Info("HTTP server graceful shutdown begins...", "cause", cause)

		if err := srv.Shutdown(tCtx); err != nil {
			logger.Error("HTTP server graceful shutdown failed!", "error", err)
			return
		}

		logger.Info("HTTP server graceful shutdown finished successfully.")
	})

	go func(cancelFunc context.CancelFunc) {
		logger.Info("Application starting...", "addr", addr)
		if err := srv.ListenAndServe(); err != nil {
			defer cancelFunc()
			logger.Info("HTTP server enters graceful shutdown mode...")
			if !errors.Is(err, http.ErrServerClosed) {
				logger.Error("HTTP server failed!", "error", err)
				return
			}
		}
	}(stop)

	defer stop()
	<-gCtx.Done()

	return nil
}

func ApiRouter(accountRepo *account.Repository, logger *slog.Logger) *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/accounts", create.Handler(accountRepo, logger)).Methods(http.MethodPost)
	router.HandleFunc("/account/{id:[0-9]+}/topup", topup.Handler(topup.GetRequestParser(), accountRepo, accountRepo, logger)).Methods(http.MethodPost)
	router.HandleFunc("/account/{id:[0-9]+}/transfer", transfer.Handler(transfer.GetRequestParser(), accountRepo, accountRepo, logger)).Methods(http.MethodPost)
	return router
}
