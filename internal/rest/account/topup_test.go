package account_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTopUp(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Run("with no previous balance", func(t *testing.T) {
			db, onClose := setupDb(t)
			defer onClose()

			// Prepare DB State
			accRow := db.QueryRow("INSERT INTO accounts DEFAULT VALUES RETURNING id")
			var accId int64
			assert.NoError(t, accRow.Scan(&accId))

			// Set Expectations
			expectedBalance := 250

			w := httptest.NewRecorder()
			r, _ := http.NewRequest("POST", fmt.Sprintf("/account/%d/topup", accId), bytes.NewBuffer([]byte(fmt.Sprintf("{\"amount\":%d}", expectedBalance))))
			r.Header.Set("Content-Type", "application/json")

			runApplication(t, db, w, r)

			// Assertions
			assert.Equal(t, http.StatusOK, w.Code)

			// Confirm DB state
			accRow = db.QueryRow("SELECT balance FROM accounts WHERE ID=$1", accId)
			var balance int
			assert.NoError(t, accRow.Scan(&balance))
			assert.Equal(t, expectedBalance, balance)
		})
		t.Run("with previous balance", func(t *testing.T) {
			db, onClose := setupDb(t)
			defer onClose()

			// Prepare DB State
			initialBalance := 50
			accRow := db.QueryRow("INSERT INTO accounts (balance) VALUES ($1) RETURNING id", initialBalance)
			var accId int64
			assert.NoError(t, accRow.Scan(&accId))

			// States
			addedBalance := 250
			w := httptest.NewRecorder()
			r, _ := http.NewRequest("POST", fmt.Sprintf("/account/%d/topup", accId), bytes.NewBuffer([]byte(fmt.Sprintf("{\"amount\":%d}", addedBalance))))
			r.Header.Set("Content-Type", "application/json")

			runApplication(t, db, w, r)

			// Assertions
			assert.Equal(t, http.StatusOK, w.Code)

			// Confirm DB state
			accRow = db.QueryRow("SELECT balance FROM accounts WHERE ID=$1", accId)
			var balance int
			assert.NoError(t, accRow.Scan(&balance))
			assert.Equal(t, initialBalance+addedBalance, balance)
		})
	})
	t.Run("fail", func(t *testing.T) {
		t.Run("account does not exist", func(t *testing.T) {
			db, onClose := setupDb(t)
			defer onClose()

			// Set Expectations
			expectedBalance := 250

			w := httptest.NewRecorder()
			r, _ := http.NewRequest("POST", fmt.Sprintf("/account/%d/topup", 0), bytes.NewBuffer([]byte(fmt.Sprintf("{\"amount\":%d}", expectedBalance))))
			r.Header.Set("Content-Type", "application/json")

			runApplication(t, db, w, r)
			assert.Equal(t, http.StatusNotFound, w.Code)
		})
		t.Run("bad request body", func(t *testing.T) {
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
					body:               bytes.NewBuffer([]byte("{\"amount\":0}")),
					expectedStatusCode: http.StatusUnprocessableEntity,
				},
			}

			for testName, test := range testCases {
				t.Run(testName, func(t *testing.T) {
					db, onClose := setupDb(t)
					defer onClose()

					accRow := db.QueryRow("INSERT INTO accounts DEFAULT VALUES RETURNING id")
					var accId int64
					assert.NoError(t, accRow.Scan(&accId))

					w := httptest.NewRecorder()
					r, _ := http.NewRequest("POST", fmt.Sprintf("/account/%d/topup", accId), test.body)
					r.Header.Set("Content-Type", "application/json")

					runApplication(t, db, w, r)
					assert.Equal(t, test.expectedStatusCode, w.Code)
				})
			}
		})
	})
}
