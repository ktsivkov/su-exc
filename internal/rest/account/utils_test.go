package account_test

import (
	"database/sql"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	_ "github.com/lib/pq"

	"github.com/stretchr/testify/assert"

	"github.com/ktsivkov/su-exc/internal/account"
	"github.com/ktsivkov/su-exc/internal/rest"
)

func runApplication(t *testing.T, db *sql.DB, w *httptest.ResponseRecorder, r *http.Request) {
	fileName := "create_account.out"
	file, err := os.Create(fileName)
	assert.NoError(t, err)
	defer func(t *testing.T, fileName string) {
		assert.NoError(t, os.Remove(fileName))
	}(t, fileName)

	logger := slog.New(slog.NewJSONHandler(file, nil))
	accountRepo, err := account.NewRepository(db)
	assert.NoError(t, err)
	rest.ApiRouter(accountRepo, logger).ServeHTTP(w, r)
}

func setupDb(t *testing.T) (*sql.DB, func()) {
	db, err := sql.Open("postgres", os.Getenv("POSTGRES_URI_TEST"))
	assert.NoError(t, err)

	// Truncate accounts table
	_, err = db.Exec("TRUNCATE TABLE accounts RESTART IDENTITY")
	assert.NoError(t, err)

	return db, func() {
		assert.NoError(t, db.Close())
	}
}
