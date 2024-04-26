package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	mockdb "github.com/go_backend_misc/db/mock"
	db "github.com/go_backend_misc/db/sqlc"
	"github.com/go_backend_misc/token"
	"github.com/go_backend_misc/util"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type getAccountTestCase struct {
	name          string
	accountID     int64
	setupAuth     func(request *http.Request, tokenMaker token.TokenMaker)
	buildStubs    func(store *mockdb.MockStore)
	checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
}

func TestGetAccountAPI(t *testing.T) {
	user, _ := randomUser(t)
	account := randomAccount(user.Username)
	testCases := []getAccountTestCase{
		{
			name:      "OK",
			accountID: account.ID,
			setupAuth: func(request *http.Request, tokenMaker token.TokenMaker) {
				addAuthorization(t, request, tokenMaker, user.Username)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccount(t, recorder.Body, &account)
			},
		},
		{
			name:      "NotFound",
			accountID: account.ID,
			setupAuth: func(request *http.Request, tokenMaker token.TokenMaker) {
				addAuthorization(t, request, tokenMaker, user.Username)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:      "InternalError",
			accountID: account.ID,
			setupAuth: func(request *http.Request, tokenMaker token.TokenMaker) {
				addAuthorization(t, request, tokenMaker, user.Username)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:      "InvalidID",
			accountID: 0,
			setupAuth: func(request *http.Request, tokenMaker token.TokenMaker) {
				addAuthorization(t, request, tokenMaker, user.Username)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			testCase.buildStubs(store)

			// start server
			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()
			url := fmt.Sprintf("/account/%d", testCase.accountID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			testCase.setupAuth(request, server.tokenMaker)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			testCase.checkResponse(t, recorder)
		})
	}
}

func randomAccount(owner string) db.Account {
	return db.Account{
		ID:       util.RandomInt(1, 1000),
		Owner:    owner,
		Currency: util.RandomCurrency(),
		Balance:  util.RandomMoney(),
	}
}

func requireBodyMatchAccount(t *testing.T, body *bytes.Buffer, account *db.Account) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var receivedAccount db.Account
	err = json.Unmarshal(data, &receivedAccount)
	require.NoError(t, err)
	require.Equal(t, receivedAccount.ID, account.ID)
	require.Equal(t, receivedAccount.Currency, account.Currency)
	require.Equal(t, receivedAccount.Balance, account.Balance)
}
