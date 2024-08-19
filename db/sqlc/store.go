package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store interface {
	CreateAccount(ctx context.Context, arg CreateAccountParams) (Account, error)
	CreateEntry(ctx context.Context, arg CreateEntryParams) (Entry, error)
	CreateTransfer(ctx context.Context, arg CreateTransferParams) (Transfer, error)
	CreateUser(ctx context.Context, arg CreateUserParams) (User, error)
	DeleteAccount(ctx context.Context, id int64) error
	GetAccount(ctx context.Context, id int64) (Account, error)
	GetAccountForUpdate(ctx context.Context, id int64) (Account, error)
	GetEntry(ctx context.Context, id int64) (Entry, error)
	GetTransfer(ctx context.Context, id int64) (Transfer, error)
	GetUser(ctx context.Context, username string) (User, error)
	ListAccounts(ctx context.Context, arg ListAccountsParams) ([]Account, error)
	ListEntries(ctx context.Context, arg ListEntriesParams) ([]Entry, error)
	ListTransfers(ctx context.Context, arg ListTransfersParams) ([]Transfer, error)
	UpdateAccount(ctx context.Context, arg UpdateAccountParams) error
	UpdateAccountBalance(ctx context.Context, arg UpdateAccountBalanceParams) (Account, error)
	TransferTx(ctx context.Context,arg TransferTxParams) (TransferTxResult,error)
}

type SQLStore struct {
	*Queries
	db *pgxpool.Pool
}

func NewStore(db *pgxpool.Pool) Store{
	return &SQLStore{
		db:db,
		Queries: New(db),
	}
}

func (s *SQLStore) ExecTx(ctx context.Context,fn func(*Queries) error) error {
	tx,err:=s.db.BeginTx(ctx,pgx.TxOptions{})

	if err!=nil {
		return err
	}

	q:= New(tx)
	err= fn(q)

	if err!=nil{
		if rberr:=tx.Rollback(ctx); rberr!=nil {
			return fmt.Errorf("tx error: %v and rb error: %v",rberr,err)
		}
		return err
	}
	return tx.Commit(ctx)
}

type TransferTxParams struct {
	FromAccountID	int64	`json:"from_account_id"`
	ToAccountID     int64	`json:"to_account_id"`
	Amount			int64  	`json:"amount"`
}

type TransferTxResult struct {
	Transfer	  Transfer
	FromAccount	  Account
	ToAccount	  Account
	FromEntry	  Entry
	ToEntry		  Entry
}

func (store *SQLStore) TransferTx(ctx context.Context,arg TransferTxParams) (TransferTxResult,error){
	var result TransferTxResult

	err:=store.ExecTx(ctx,func (q *Queries)error {
		var err error

		fromAccount, err := q.GetAccount(ctx, arg.FromAccountID)
		if err != nil {
			return err
		}

		if fromAccount.Balance < arg.Amount {
			return fmt.Errorf("insufficient balance in account %d", arg.FromAccountID)
		}

		result.Transfer,err=q.CreateTransfer(ctx,CreateTransferParams{
			FromAccountID: arg.FromAccountID,
			ToAccountID: arg.ToAccountID,
			Amount: arg.Amount,
		})

		if err!=nil{
			return err
		}

		result.FromEntry,err=q.CreateEntry(ctx,CreateEntryParams{
			AccountID: arg.FromAccountID,
			Amount: -arg.Amount,
		})
		if err!=nil{
			return err
		}

		result.ToEntry,err=q.CreateEntry(ctx,CreateEntryParams{
			AccountID: arg.ToAccountID,
			Amount:arg.Amount,
		})
		if err!=nil{
			return err
		}

		if arg.FromAccountID<arg.ToAccountID{
			result.FromAccount,err=q.UpdateAccountBalance(ctx,UpdateAccountBalanceParams{
				ID:     arg.FromAccountID,
				Amount: -arg.Amount,
			})
			if err!=nil{
				return err
			}
			result.ToAccount,err=q.UpdateAccountBalance(ctx,UpdateAccountBalanceParams{
				ID:     arg.ToAccountID,
				Amount: arg.Amount,
			})
			if err!=nil{
				return err
			}
		} else {
			result.ToAccount,err=q.UpdateAccountBalance(ctx,UpdateAccountBalanceParams{
				ID:     arg.ToAccountID,
				Amount: arg.Amount,
			})
			if err!=nil{
				return err
			}

			result.FromAccount,err=q.UpdateAccountBalance(ctx,UpdateAccountBalanceParams{
				ID:     arg.FromAccountID,
				Amount: -arg.Amount,
			})
			if err!=nil{
				return err
			}
		}
		return nil
	})
	return result,err
}

