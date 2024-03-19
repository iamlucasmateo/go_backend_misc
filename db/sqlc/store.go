package db

import (
	"context"
	"database/sql"
	"fmt"
)

type Store struct {
	// composition: Store embeds Queries (in Go, vs inheritance)
	*Queries
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		Queries: New(db),
		db:      db,
	}
}

// execTx runs a function within a database transaction
// This function is unexported (lowercase), so it can only be called from within the db package
func (store *Store) executeTransaction(ctx context.Context, innerFunction func(*Queries) error) error {
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

// TransferTx performs a money transfer from one account to the other
// It creates a transfer record, add account entries, and update accounts' balance within a single database transaction
func (store *Store) TransferTx(ctx context.Context, arg CreateTransferParams) (result TransferTxResult, err error) {
	txErr := store.executeTransaction(ctx, func(queries *Queries) error {
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

		fromAccount, err := queries.GetAccount(ctx, arg.FromAccountID.Int64)
		if err != nil {
			return err
		}
		result.FromAccount, err = queries.UpdateAccount(ctx, UpdateAccountParams{
			ID:      arg.FromAccountID.Int64,
			Balance: fromAccount.Balance - arg.Amount,
		})
		if err != nil {
			return err
		}
		toAccount, err := queries.GetAccount(ctx, arg.ToAccountID.Int64)
		if err != nil {
			return err
		}
		result.ToAccount, err = queries.UpdateAccount(ctx, UpdateAccountParams{
			ID:      arg.ToAccountID.Int64,
			Balance: toAccount.Balance + arg.Amount,
		})
		if err != nil {
			return err
		}

		return nil
	})

	return result, txErr
}
