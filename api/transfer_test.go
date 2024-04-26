package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	mockdb "github.com/go_backend_misc/db/mock"
	db "github.com/go_backend_misc/db/sqlc"
	"github.com/go_backend_misc/token"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestCreateTransfer(t *testing.T) {
	type transferTestCase struct {
		name          string
		body          gin.H
		setupAuth     func(request *http.Request, tokenMaker token.TokenMaker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}

	invalidBody := transferTestCase{
		name: "Invalid JSON body",
		body: gin.H{
			"from_account_id": 123,
			"to_account_id":   456,
			"amount":          100,
		},
		buildStubs: func(store *mockdb.MockStore) {},
		setupAuth: func(request *http.Request, tokenMaker token.TokenMaker) {
			addAuthorization(t, request, tokenMaker, "testUser")
		},
		checkResponse: func(recorder *httptest.ResponseRecorder) {
			require.Equal(t, http.StatusBadRequest, recorder.Code)
			var content map[string]any
			json.Unmarshal(recorder.Body.Bytes(), &content)
			expected := "Key: 'transferRequest.Currency' Error:Field validation for 'Currency' failed on the 'required' tag"
			require.Contains(t, content["error"], expected)
		},
	}

	noFromAccount := transferTestCase{
		name: "No FromAccount",
		body: gin.H{
			"from_account_id": 123,
			"to_account_id":   456,
			"amount":          100,
			"currency":        "USD",
		},
		setupAuth: func(request *http.Request, tokenMaker token.TokenMaker) {
			addAuthorization(t, request, tokenMaker, "testUser")
		},
		buildStubs: func(store *mockdb.MockStore) {
			store.EXPECT().
				GetAccount(gomock.Any(), int64(123)).
				Times(1).
				Return(db.Account{}, sql.ErrNoRows)
		},
		checkResponse: func(recorder *httptest.ResponseRecorder) {
			require.Equal(t, http.StatusNotFound, recorder.Code)
			var content map[string]any
			json.Unmarshal(recorder.Body.Bytes(), &content)
			expected := "sql: no rows in result set"
			require.Equal(t, content["error"], expected)
		},
	}

	sqlError := transferTestCase{
		name: "No FromAccount",
		body: gin.H{
			"from_account_id": 123,
			"to_account_id":   456,
			"amount":          100,
			"currency":        "USD",
		},
		setupAuth: func(request *http.Request, tokenMaker token.TokenMaker) {
			addAuthorization(t, request, tokenMaker, "testUser")
		},
		buildStubs: func(store *mockdb.MockStore) {
			store.EXPECT().
				GetAccount(gomock.Any(), int64(123)).
				Times(1).
				Return(db.Account{}, sql.ErrConnDone)
		},
		checkResponse: func(recorder *httptest.ResponseRecorder) {
			require.Equal(t, http.StatusInternalServerError, recorder.Code)
			var content map[string]any
			json.Unmarshal(recorder.Body.Bytes(), &content)
			expected := "sql: connection is already closed"
			require.Equal(t, content["error"], expected)
		},
	}

	currencyMismatch := transferTestCase{
		name: "No FromAccount",
		body: gin.H{
			"from_account_id": 123,
			"to_account_id":   456,
			"amount":          100,
			"currency":        "USD",
		},
		setupAuth: func(request *http.Request, tokenMaker token.TokenMaker) {
			addAuthorization(t, request, tokenMaker, "testUser")
		},
		buildStubs: func(store *mockdb.MockStore) {
			fromAccount := db.Account{
				ID:        123,
				Owner:     "test_owner",
				Balance:   100,
				Currency:  "CAD",
				CreatedAt: time.Now(),
			}
			store.EXPECT().
				GetAccount(gomock.Any(), int64(123)).
				Times(1).
				Return(fromAccount, nil)
		},
		checkResponse: func(recorder *httptest.ResponseRecorder) {
			require.Equal(t, http.StatusBadRequest, recorder.Code)
			var content map[string]any
			json.Unmarshal(recorder.Body.Bytes(), &content)
			expected := "account currency mismatch"
			require.Equal(t, content["error"], expected)
		},
	}

	okCase := transferTestCase{
		name: "OK case",
		body: gin.H{
			"from_account_id": 123,
			"to_account_id":   456,
			"amount":          100,
			"currency":        "USD",
		},
		setupAuth: func(request *http.Request, tokenMaker token.TokenMaker) {
			fromAccount, _ := getAccounts()
			addAuthorization(t, request, tokenMaker, fromAccount.Owner)
		},
		buildStubs: func(store *mockdb.MockStore) {
			fromAccount, toAccount := getAccounts()
			store.EXPECT().
				GetAccount(gomock.Any(), int64(fromAccount.ID)).
				Times(1).
				Return(fromAccount, nil)

			store.EXPECT().
				GetAccount(gomock.Any(), int64(toAccount.ID)).
				Times(1).
				Return(toAccount, nil)

			tranferResult := getOkTransferResult(fromAccount, toAccount)

			expectedArg := db.CreateTransferParams{
				FromAccountID: db.Int64ToSqlInt64(fromAccount.ID),
				ToAccountID:   db.Int64ToSqlInt64(toAccount.ID),
				Amount:        100,
			}

			store.EXPECT().
				TransferTx(gomock.Any(), expectedArg).
				Times(1).
				Return(*tranferResult, nil)

		},
		checkResponse: func(recorder *httptest.ResponseRecorder) {
			require.Equal(t, http.StatusOK, recorder.Code)
			var content db.TransferTxResult
			json.Unmarshal(recorder.Body.Bytes(), &content)
			fromAccount, toAccount := getAccounts()
			require.Equal(t, content, *getOkTransferResult(fromAccount, toAccount))
		},
	}

	noToAccount := transferTestCase{
		name: "No ToAccount",
		body: gin.H{
			"from_account_id": 123,
			"to_account_id":   456,
			"amount":          100,
			"currency":        "USD",
		},
		setupAuth: func(request *http.Request, tokenMaker token.TokenMaker) {
			addAuthorization(t, request, tokenMaker, "testUser")
		},
		buildStubs: func(store *mockdb.MockStore) {
			fromAccount := db.Account{
				ID:        123,
				Owner:     "test_owner",
				Balance:   100,
				Currency:  "USD",
				CreatedAt: time.Now(),
			}
			store.EXPECT().
				GetAccount(gomock.Any(), int64(123)).
				Times(1).
				Return(fromAccount, nil)

			store.EXPECT().
				GetAccount(gomock.Any(), int64(456)).
				Times(1).
				Return(db.Account{}, sql.ErrNoRows)
		},
		checkResponse: func(recorder *httptest.ResponseRecorder) {
			require.Equal(t, http.StatusNotFound, recorder.Code)
			var content map[string]any
			json.Unmarshal(recorder.Body.Bytes(), &content)
			expected := "sql: no rows in result set"
			require.Equal(t, content["error"], expected)
		},
	}

	testCases := []transferTestCase{
		invalidBody, noFromAccount, sqlError,
		currencyMismatch, okCase, noToAccount,
	}

	for _, testCase := range testCases {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		store := mockdb.NewMockStore(ctrl)
		testCase.buildStubs(store)

		server := newTestServer(t, store)
		recorder := httptest.NewRecorder()

		data, err := json.Marshal(testCase.body)
		require.NoError(t, err)
		url := "/transfer"

		request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
		require.NoError(t, err)
		testCase.setupAuth(request, server.tokenMaker)
		server.router.ServeHTTP(recorder, request)
		testCase.checkResponse(recorder)
	}
}

func getOkTransferResult(fromAccount db.Account, toAccount db.Account) *db.TransferTxResult {
	transferResult := db.TransferTxResult{
		Transfer: db.Transfer{
			ID:            1,
			FromAccountID: sql.NullInt64{Int64: fromAccount.ID},
			ToAccountID:   sql.NullInt64{Int64: toAccount.ID},
			Amount:        100,
		},
		FromAccount: fromAccount,
		ToAccount:   toAccount,
		FromEntry: db.Entry{
			ID:        1,
			AccountID: sql.NullInt64{Int64: fromAccount.ID},
			Amount:    -100,
		},
		ToEntry: db.Entry{
			ID:        1,
			AccountID: sql.NullInt64{Int64: toAccount.ID},
			Amount:    100,
		},
	}

	return &transferResult
}

func getAccounts() (fromAccount db.Account, toAccount db.Account) {
	fromAccount = db.Account{
		ID:       123,
		Owner:    "test_owner",
		Balance:  300,
		Currency: "USD",
	}
	toAccount = db.Account{
		ID:       456,
		Owner:    "test_owner",
		Balance:  100,
		Currency: "USD",
	}

	return fromAccount, toAccount
}
