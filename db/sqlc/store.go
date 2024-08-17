package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

//we place queries in store to extend functionalities like transactions
type Store struct {
	*Queries
	db *pgxpool.Pool
}

//this creates a new store 
func NewStore(db *pgxpool.Pool) *Store{
	return &Store{
		db:db,
		Queries: New(db),
	}
}

//execTx
func (s *Store) ExecTx(ctx context.Context,fn func(*Queries) error) error {
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

func (store *Store) TransferTx(ctx context.Context,arg TransferTxParams) (TransferTxResult,error){
	var result TransferTxResult

	err:=store.ExecTx(ctx,func (q *Queries)error {
		var err error
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
			Amount: arg.Amount,
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