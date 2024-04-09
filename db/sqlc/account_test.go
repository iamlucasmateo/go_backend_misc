package db

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/go_backend_misc/util"
	"github.com/stretchr/testify/require"
)

func createRandomAccount(userSuffix string) (Account, User, CreateAccountParams, error) {
	user, _, _ := createRandomUser(userSuffix)
	arg := CreateAccountParams{
		Owner:    user.Username,
		Balance:  util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}

	account, err := testQueries.CreateAccount(context.Background(), arg)

	return account, user, arg, err
}

func TestCreateAccount(t *testing.T) {
	account, user, createAccountParams, err := createRandomAccount("_test_create_account")
	if err != nil {
		t.Errorf("error creating account: %v", err)
	}
	// require.NoError checks if the error is nil
	require.NoError(t, err)
	require.NotEmpty(t, account)

	if createAccountParams.Owner != account.Owner {
		t.Errorf("unexpected owner: %s", account.Owner)
	}
	// etc
	require.Equal(t, createAccountParams.Owner, account.Owner)
	require.Equal(t, user.Username, account.Owner)
	require.Equal(t, createAccountParams.Balance, account.Balance)
	require.Equal(t, createAccountParams.Currency, account.Currency)

	require.NotZero(t, account.ID)
	require.NotZero(t, account.CreatedAt)
}

func TestGetAccount(t *testing.T) {
	account, user, _, _ := createRandomAccount("_test_get_account")
	retrievedAccount, err := testQueries.GetAccount(context.Background(), account.ID)
	require.NoError(t, err)
	require.Equal(t, account.ID, retrievedAccount.ID)
	require.Equal(t, account.Owner, retrievedAccount.Owner)
	require.Equal(t, account.Owner, user.Username)
	require.Equal(t, account.Owner, user.Username)
	require.Equal(t, account.Balance, retrievedAccount.Balance)
}

func TestUpdateAccount(t *testing.T) {
	account, user, _, _ := createRandomAccount("_test_update_account")
	arg := UpdateAccountParams{
		ID:      account.ID,
		Balance: util.RandomMoney(),
	}

	updatedAccount, err := testQueries.UpdateAccount(context.Background(), arg)
	require.NoError(t, err)
	require.Equal(t, account.ID, updatedAccount.ID)
	require.Equal(t, account.Owner, updatedAccount.Owner)
	require.Equal(t, account.Owner, user.Username)
	require.Equal(t, arg.Balance, updatedAccount.Balance)
	require.Equal(t, account.Currency, updatedAccount.Currency)
}

func TestDeleteAccount(t *testing.T) {
	account, _, _, _ := createRandomAccount("_test_delete_account")
	err := testQueries.DeleteAccount(context.Background(), account.ID)
	require.NoError(t, err)
	retrievedAccount, err := testQueries.GetAccount(context.Background(), account.ID)
	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, retrievedAccount)
}

func TestListAccounts(t *testing.T) {
	for i := 0; i < 10; i++ {
		userSuffix := "_test_list_accounts_" + fmt.Sprintf("%v", i)
		createRandomAccount(userSuffix)
	}

	arg := ListAccountsParams{
		Limit:  5,
		Offset: 5,
	}

	accounts, err := testQueries.ListAccounts(context.Background(), arg)
	require.NoError(t, err)
	require.Equal(t, len(accounts), 5)

	for _, account := range accounts {
		require.NotEmpty(t, account)
	}
}
