package db

import (
	"context"
	"database/sql"
	"fmt"
)

type Store interface {
	// interface composition: Store embeds Querier (in Go, vs inheritance)
	Querier
	TransferTx(ctx context.Context, arg CreateTransferParams) (result TransferTxResult, err error)
}

type SQLStore struct {
	// struct composition: SQLStore embeds Queries (in Go, vs inheritance)
	*Queries
	db *sql.DB
}

func NewStore(db *sql.DB) Store {
	return &SQLStore{
		Queries: New(db),
		db:      db,
	}
}

// execTx runs a function within a database transaction
// This function is unexported (lowercase), so it can only be called from within the db package
func (store *SQLStore) executeTransaction(ctx context.Context, innerFunction func(*Queries) error) error {
	transaction, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	queries := New(transaction)
	txError := innerFunction(queries)
	if txError != nil {
		// if there's an error, rollback the transaction
		if rollbackError := transaction.Rollback(); rollbackError != nil {
			return fmt.Errorf("tx error: %v, rollback error: %v", txError, rollbackError)
		}
		return txError
	}

	return transaction.Commit()

}

type TransferTxResult struct {
	Transfer    Transfer `json: "from_account_id"`
	FromAccount Account  `json: "from_account"`
	ToAccount   Account  `json: "to_account"`
	FromEntry   Entry    `json: "from_entry"`
	ToEntry     Entry    `json: "to_entry"`
}

// txKey is a custom key to store the transaction name in the context
// It shouldn't be a string to avoid collisions with other context keys
var txKey = struct{}{}

// TransferTx performs a money transfer from one account to the other
// It creates a transfer record, add account entries, and update accounts' balance within a single database transaction
func (store *SQLStore) TransferTx(ctx context.Context, arg CreateTransferParams) (result TransferTxResult, err error) {
	txErr := store.executeTransaction(ctx, func(queries *Queries) error {
		// used for debugging: txName := ctx.Value(txKey)
		result.Transfer, err = queries.CreateTransfer(ctx, CreateTransferParams{
			FromAccountID: arg.FromAccountID,
			ToAccountID:   arg.ToAccountID,
			Amount:        arg.Amount,
		})
		if err != nil {
			return err
		}

		result.FromEntry, err = queries.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.FromAccountID,
			Amount:    -arg.Amount,
		})
		if err != nil {
			return err
		}

		result.ToEntry, err = queries.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.ToAccountID,
			Amount:    arg.Amount,
		})
		if err != nil {
			return err
		}

		// update in the same ID order to avoid deadlocks
		if arg.FromAccountID.Int64 < arg.ToAccountID.Int64 {
			result.FromAccount, result.ToAccount, err = moveMoney(
				ctx, queries, arg.FromAccountID.Int64, -arg.Amount, arg.ToAccountID.Int64, +arg.Amount,
			)
		} else {
			result.ToAccount, result.FromAccount, err = moveMoney(
				ctx, queries, arg.ToAccountID.Int64, +arg.Amount, arg.FromAccountID.Int64, -arg.Amount,
			)
		}

		return nil
	})

	return result, txErr
}

func moveMoney(
	ctx context.Context,
	q *Queries,
	accountID1 int64,
	amount1 int64,
	accountID2 int64,
	amount2 int64,
) (account1 Account, account2 Account, err error) {
	account1, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		ID:     accountID1,
		Amount: amount1,
	})
	if err != nil {
		return
	}

	account2, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		ID:     accountID2,
		Amount: amount2,
	})
	if err != nil {
		return
	}

	return
}
