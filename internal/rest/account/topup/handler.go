package topup

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/pkg/errors"

	"github.com/ktsivkov/su-exc/internal/account"
)

func Handler(requestParser RequestParser, finder account.Finder, topUpper account.TopUpper, logger *slog.Logger) http.HandlerFunc {
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

		targetAccount, err := finder.FindById(ctx, req.Target)
		if err != nil {
			if errors.Is(err, account.ErrDoesNotExist) {
				logger.WarnContext(ctx, "account id not found", "id", req.Target)
				w.WriteHeader(http.StatusNotFound)
				if _, err := fmt.Fprintf(w, "target account with id=%d does not exist", req.Target); err != nil {
					logger.ErrorContext(ctx, "cannot write bytes to client", "error", err)
				}
				return
			}

			logger.ErrorContext(ctx, "account existence check failed", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			if _, err := fmt.Fprint(w, "could not process the request"); err != nil {
				logger.ErrorContext(ctx, "cannot write bytes to client", "error", err)
			}
			return
		}

		if err := topUpper.TopUp(ctx, targetAccount, req.Data.Amount); err != nil {
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
