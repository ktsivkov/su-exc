package transfer

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/pkg/errors"

	"github.com/ktsivkov/su-exc/internal/account"
)

func Handler(requestParser RequestParser, finder account.Finder, transferer account.Transferrer, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		req, err := requestParser(ctx, r)
		if err != nil {
			logger.WarnContext(ctx, "cannot parse request", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			if _, err := fmt.Fprintf(w, "request parsing error: %s", err); err != nil {
				logger.ErrorContext(ctx, "cannot write bytes to client", "error", err)
			}
			return
		}

		if err := req.Validate(); err != nil {
			logger.WarnContext(ctx, "invalid request", "error", err)
			w.WriteHeader(http.StatusUnprocessableEntity)
			if _, err := fmt.Fprintf(w, "request validation error: %s", err); err != nil {
				logger.ErrorContext(ctx, "cannot write bytes to client", "error", err)
			}
			return
		}

		sourceAccount, err := finder.FindById(ctx, req.Source)
		if err != nil {
			if errors.Is(err, account.ErrDoesNotExist) {
				logger.WarnContext(ctx, "source account does not exist", "id", req.Source)
				w.WriteHeader(http.StatusNotFound)
				if _, err := fmt.Fprintf(w, "source account with id=%d does not exist", req.Source); err != nil {
					logger.ErrorContext(ctx, "cannot write bytes to http response", "error", err)
				}
				return
			}

			logger.ErrorContext(ctx, "source account existence check failed", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			if _, err := fmt.Fprint(w, "could not process the request"); err != nil {
				logger.ErrorContext(ctx, "cannot write bytes to client", "error", err)
			}
			return
		}

		targetAccount, err := finder.FindById(ctx, req.Data.Target)
		if err != nil {
			if errors.Is(err, account.ErrDoesNotExist) {
				logger.WarnContext(ctx, "target account does not exist", "id", req.Data.Target)
				w.WriteHeader(http.StatusNotFound)
				if _, err := fmt.Fprintf(w, "target account with id=%d does not exist", req.Data.Target); err != nil {
					logger.ErrorContext(ctx, "cannot write bytes to http response", "error", err)
				}
				return
			}

			logger.ErrorContext(ctx, "target account existence check failed", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			if _, err := fmt.Fprint(w, "could not process the request"); err != nil {
				logger.ErrorContext(ctx, "cannot write bytes to client", "error", err)
			}
			return
		}

		if err := transferer.Transfer(ctx, sourceAccount, targetAccount, req.Data.Amount); err != nil {
			if errors.Is(err, account.ErrInsufficientBalance) {
				w.WriteHeader(http.StatusBadRequest)
				if _, err := fmt.Fprint(w, err); err != nil {
					logger.ErrorContext(ctx, "cannot write bytes to client", "error", err)
				}
				return
			}

			logger.ErrorContext(ctx, "account top-up failed", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			if _, err := fmt.Fprint(w, "could not process the request"); err != nil {
				logger.ErrorContext(ctx, "cannot write bytes to client", "error", err)
			}
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
