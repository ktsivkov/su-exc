package account_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreate(t *testing.T) {
	t.Run("create account", func(t *testing.T) {
		db, onClose := setupDb(t)
		defer onClose()

		w := httptest.NewRecorder()
		r, err := http.NewRequest("POST", "/accounts", nil)
		assert.NoError(t, err)
		r.Header.Set("Content-Type", "application/json")

		runApplication(t, db, w, r)

		assert.Equal(t, http.StatusCreated, w.Code)
		row := db.QueryRow("SELECT id FROM accounts WHERE id=$1", w.Body.String())
		assert.NoError(t, row.Err())
	})
}
