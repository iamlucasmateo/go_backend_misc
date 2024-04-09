package db

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func runTransferTxTests(
	t *testing.T,
	err error,
	fromAccount *Account,
	toAccount *Account,
	transferResult TransferTxResult,
	amount int64,
	store Store,
) {
	require.NoError(t, err)

	transfer := transferResult.Transfer
	require.NotEmpty(t, transfer)
	require.Equal(t, fromAccount.ID, transfer.FromAccountID.Int64)
	require.Equal(t, toAccount.ID, transfer.ToAccountID.Int64)
	require.Equal(t, amount, transfer.Amount)
	require.NotZero(t, transfer.ID)
	require.NotZero(t, transfer.CreatedAt)

	_, err = store.GetTransfer(context.Background(), transfer.ID)
	require.NoError(t, err)

	// check entries
	fromEntry := transferResult.FromEntry
	require.NotEmpty(t, fromEntry)
	require.Equal(t, fromAccount.ID, fromEntry.AccountID.Int64)
	require.Equal(t, -amount, fromEntry.Amount)
	require.NotZero(t, fromEntry.ID)
	require.NotZero(t, fromEntry.CreatedAt)

	_, err = store.GetEntry(context.Background(), fromEntry.ID)
	require.NoError(t, err)
}

func TestTransferTx(t *testing.T) {
	store := NewStore(testDB)
	fromAccountTest, _, _, _ := createRandomAccount("_test_transfer_tx_1")
	toAccountTest, _, _, _ := createRandomAccount("_test_transfer_tx_2")
	fromId := sql.NullInt64{
		Int64: fromAccountTest.ID,
		Valid: true,
	}
	toId := sql.NullInt64{
		Int64: toAccountTest.ID,
		Valid: true,
	}
	var amount int64 = 10
	result, err := store.TransferTx(context.Background(), CreateTransferParams{
		FromAccountID: fromId,
		ToAccountID:   toId,
		Amount:        amount,
	})

	runTransferTxTests(t, err, &fromAccountTest, &toAccountTest, result, amount, store)

	dbAccountFrom, err := store.GetAccount(context.Background(), fromAccountTest.ID)
	require.NoError(t, err)
	dbAccountTo, err := store.GetAccount(context.Background(), toAccountTest.ID)
	require.NoError(t, err)
	require.Equal(t, fromAccountTest.Balance-amount, dbAccountFrom.Balance)
	require.Equal(t, toAccountTest.Balance+amount, dbAccountTo.Balance)
}

func TestTransferTxConcurrent(t *testing.T) {
	store := NewStore(testDB)
	accountFromTest, _, _, _ := createRandomAccount("_test_transfer_tx_1")
	accountToTest, _, _, _ := createRandomAccount("_test_transfer_tx_2")
	fromId := sql.NullInt64{
		Int64: accountFromTest.ID,
		Valid: true,
	}
	toId := sql.NullInt64{
		Int64: accountToTest.ID,
		Valid: true,
	}

	// run n concurrent transfer transactions
	n := 5
	amount := int64(10)

	errs := make(chan error)
	results := make(chan TransferTxResult)

	for i := 0; i < n; i++ {
		txName := fmt.Sprintf("tx %d", i)
		go func() {
			ctx := context.WithValue(context.Background(), txKey, txName)
			result, err := store.TransferTx(ctx, CreateTransferParams{
				FromAccountID: fromId,
				ToAccountID:   toId,
				Amount:        amount,
			})
			errs <- err
			results <- result
		}()
	}

	// check results
	for i := 0; i < n; i++ {
		err := <-errs
		result := <-results
		runTransferTxTests(t, err, &accountFromTest, &accountToTest, result, amount, store)
	}

	// check final updated balances
	dbAccountFrom, err := store.GetAccount(context.Background(), accountFromTest.ID)
	require.NoError(t, err)
	dbAccountTo, err := store.GetAccount(context.Background(), accountToTest.ID)
	require.NoError(t, err)
	require.Equal(t, accountFromTest.Balance-(amount*int64(n)), dbAccountFrom.Balance)
	require.Equal(t, accountToTest.Balance+(amount*int64(n)), dbAccountTo.Balance)
}

func TestTransferTxConcurrentCrossAmounts(t *testing.T) {
	store := NewStore(testDB)
	account1, _, _, _ := createRandomAccount("_test_transfer_tx_1")
	account2, _, _, _ := createRandomAccount("_test_transfer_tx_2")

	// run n concurrent transfer transactions
	n := 10
	amount := int64(10)

	errs := make(chan error)
	results := make(chan TransferTxResult)

	for i := 0; i < n; i++ {
		fromId := sql.NullInt64{
			Int64: account1.ID,
			Valid: true,
		}
		toId := sql.NullInt64{
			Int64: account2.ID,
			Valid: true,
		}
		// half the transactions will switch the from and to accounts
		if i >= 5 {
			fromId.Int64 = account2.ID
			toId.Int64 = account1.ID
		}
		go func() {
			result, err := store.TransferTx(context.Background(), CreateTransferParams{
				FromAccountID: fromId,
				ToAccountID:   toId,
				Amount:        amount,
			})
			errs <- err
			results <- result
		}()
	}

	// check results
	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)
	}

	// check final updated balances
	dbAccount1, err := store.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)
	dbAccount2, err := store.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)
	require.Equal(t, account1.Balance, dbAccount1.Balance)
	require.Equal(t, account2.Balance, dbAccount2.Balance)
}
