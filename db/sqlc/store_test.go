package db

import (
	"context"
	"database/sql"
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
	store *Store,
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

	dbAccountFrom, err := store.GetAccount(context.Background(), fromAccount.ID)
	require.NoError(t, err)
	dbAccountTo, err := store.GetAccount(context.Background(), toAccount.ID)
	require.NoError(t, err)
	require.Equal(t, fromAccount.Balance-amount, dbAccountFrom.Balance)
	require.Equal(t, toAccount.Balance+amount, dbAccountTo.Balance)
}

func TestTransferTx(t *testing.T) {
	store := NewStore(testDB)
	fromAccountTest, _, _ := createRandomAccount(t, "_test_transfer_tx_1")
	toAccountTest, _, _ := createRandomAccount(t, "_test_transfer_tx_2")
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
}

func NoTestTransferTxConcurrent(t *testing.T) {
	store := NewStore(testDB)
	testAccount1, _, _ := createRandomAccount(t, "_test_transfer_tx_1")
	testAccount2, _, _ := createRandomAccount(t, "_test_transfer_tx_2")
	id1 := sql.NullInt64{
		Int64: testAccount1.ID,
		Valid: true,
	}
	id2 := sql.NullInt64{
		Int64: testAccount2.ID,
		Valid: true,
	}

	// run n concurrent transfer transactions
	n := 5
	amount := int64(10)

	errs := make(chan error)
	results := make(chan TransferTxResult)

	for i := 0; i < n; i++ {
		go func() {
			result, err := store.TransferTx(context.Background(), CreateTransferParams{
				FromAccountID: id1,
				ToAccountID:   id2,
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
		runTransferTxTests(t, err, &testAccount1, &testAccount2, result, amount, store)

	}
}
