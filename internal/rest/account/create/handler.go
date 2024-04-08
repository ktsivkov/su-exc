package create

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/ktsivkov/su-exc/internal/account"
)

func Handler(creator account.Creator, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		id, err := creator.Create(ctx)
		if err != nil {
			logger.Error("account creation failed", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		if _, err := fmt.Fprintf(w, "%d", id); err != nil {
			logger.Error("could not write bytes to client", "error", err)
		}
	}
}
