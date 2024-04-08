package account_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTransfer(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		srcAccBalanceInitial := 200
		transferAmount := 100

		db, onClose := setupDb(t)
		defer onClose()

		srcAccRow := db.QueryRow("INSERT INTO accounts (balance) VALUES ($1) RETURNING id", srcAccBalanceInitial)
		var srcAccId int64
		assert.NoError(t, srcAccRow.Scan(&srcAccId))

		targetAccRow := db.QueryRow("INSERT INTO accounts DEFAULT VALUES RETURNING id")
		var targetAccId int64
		assert.NoError(t, targetAccRow.Scan(&targetAccId))

		reqBody := map[string]any{
			"target": targetAccId,
			"amount": transferAmount,
		}
		reqBodyJsonBytes, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		r, _ := http.NewRequest("POST", fmt.Sprintf("/account/%d/transfer", srcAccId), bytes.NewBuffer(reqBodyJsonBytes))
		r.Header.Set("Content-Type", "application/json")

		runApplication(t, db, w, r)
		assert.Equal(t, http.StatusOK, w.Code)

		srcAccRow = db.QueryRow("SELECT balance FROM accounts WHERE id=$1", srcAccId)
		var srcAccBalance int
		assert.NoError(t, srcAccRow.Scan(&srcAccBalance))

		targetAccRow = db.QueryRow("SELECT balance FROM accounts WHERE id=$1", targetAccId)
		var targetAccBalance int
		assert.NoError(t, targetAccRow.Scan(&targetAccBalance))

		assert.Equal(t, transferAmount, targetAccBalance)
		assert.Equal(t, srcAccBalanceInitial-transferAmount, srcAccBalance)
	})

	t.Run("fail", func(t *testing.T) {
		t.Run("insufficient balance", func(t *testing.T) {
			srcAccBalanceInitial := 200
			transferAmount := 300

			db, onClose := setupDb(t)
			defer onClose()

			srcAccRow := db.QueryRow("INSERT INTO accounts (balance) VALUES ($1) RETURNING id", srcAccBalanceInitial)
			var srcAccId int64
			assert.NoError(t, srcAccRow.Scan(&srcAccId))

			targetAccRow := db.QueryRow("INSERT INTO accounts DEFAULT VALUES RETURNING id")
			var targetAccId int64
			assert.NoError(t, targetAccRow.Scan(&targetAccId))

			reqBody := map[string]any{
				"target": targetAccId,
				"amount": transferAmount,
			}
			reqBodyJsonBytes, _ := json.Marshal(reqBody)

			w := httptest.NewRecorder()
			r, _ := http.NewRequest("POST", fmt.Sprintf("/account/%d/transfer", srcAccId), bytes.NewBuffer(reqBodyJsonBytes))
			r.Header.Set("Content-Type", "application/json")

			runApplication(t, db, w, r)
			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
		t.Run("source account does not exist", func(t *testing.T) {
			transferAmount := 300

			db, onClose := setupDb(t)
			defer onClose()

			var srcAccId int64

			targetAccRow := db.QueryRow("INSERT INTO accounts DEFAULT VALUES RETURNING id")
			var targetAccId int64
			assert.NoError(t, targetAccRow.Scan(&targetAccId))

			reqBody := map[string]any{
				"target": targetAccId,
				"amount": transferAmount,
			}
			reqBodyJsonBytes, _ := json.Marshal(reqBody)

			w := httptest.NewRecorder()
			r, _ := http.NewRequest("POST", fmt.Sprintf("/account/%d/transfer", srcAccId), bytes.NewBuffer(reqBodyJsonBytes))
			r.Header.Set("Content-Type", "application/json")

			runApplication(t, db, w, r)
			assert.Equal(t, http.StatusNotFound, w.Code)
		})
		t.Run("target account does not exist", func(t *testing.T) {
			srcAccBalanceInitial := 200
			transferAmount := 300

			db, onClose := setupDb(t)
			defer onClose()

			srcAccRow := db.QueryRow("INSERT INTO accounts (balance) VALUES ($1) RETURNING id", srcAccBalanceInitial)
			var srcAccId int64
			assert.NoError(t, srcAccRow.Scan(&srcAccId))

			var targetAccId int64

			reqBody := map[string]any{
				"target": targetAccId,
				"amount": transferAmount,
			}
			reqBodyJsonBytes, _ := json.Marshal(reqBody)

			w := httptest.NewRecorder()
			r, _ := http.NewRequest("POST", fmt.Sprintf("/account/%d/transfer", srcAccId), bytes.NewBuffer(reqBodyJsonBytes))
			r.Header.Set("Content-Type", "application/json")

			runApplication(t, db, w, r)
			assert.Equal(t, http.StatusNotFound, w.Code)
		})
		t.Run("bad request body", func(t *testing.T) {
			db, onClose := setupDb(t)
			defer onClose()

			srcAccRow := db.QueryRow("INSERT INTO accounts DEFAULT VALUES RETURNING id")
			var srcAccId int64
			assert.NoError(t, srcAccRow.Scan(&srcAccId))

			targetAccRow := db.QueryRow("INSERT INTO accounts DEFAULT VALUES RETURNING id")
			var targetAccId int64
			assert.NoError(t, targetAccRow.Scan(&targetAccId))

			type testCase struct {
				body               io.Reader
				expectedStatusCode int
			}
			testCases := map[string]testCase{
				"no request body": {
					body:               http.NoBody,
					expectedStatusCode: http.StatusBadRequest,
				},
				"empty request body": {
					body:               bytes.NewBuffer([]byte("null")),
					expectedStatusCode: http.StatusUnprocessableEntity,
				},
				"invalid json request body": {
					body:               bytes.NewBuffer([]byte("{\"}")),
					expectedStatusCode: http.StatusBadRequest,
				},
				"invalid amount input": {
					body:               bytes.NewBuffer([]byte(fmt.Sprintf("{\"amount\":0, \"target\": %d}", targetAccId))),
					expectedStatusCode: http.StatusUnprocessableEntity,
				},
			}
			for testName, test := range testCases {
				t.Run(testName, func(t *testing.T) {
					w := httptest.NewRecorder()
					r, _ := http.NewRequest("POST", fmt.Sprintf("/account/%d/transfer", srcAccId), test.body)
					r.Header.Set("Content-Type", "application/json")

					runApplication(t, db, w, r)
					assert.Equal(t, test.expectedStatusCode, w.Code)
				})
			}
		})
	})
}
