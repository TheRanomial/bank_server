package api

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	db "github.com/TheRanomial/bank_server/db/sqlc"
	"github.com/TheRanomial/bank_server/token"
	"github.com/gin-gonic/gin"
)

type TransferRequest struct {
	FromAccountID int64  `json:"from_account_id" binding:"required,min=1"`
	ToAccountID   int64  `json:"to_account_id" binding:"required,min=1"`
	Amount        int64  `json:"amount" binding:"required,gt=0"`
	Currency      string `json:"currency" binding:"required,currency"`
}

func (s *Server) CreateTransfer(ctx *gin.Context) {
    req := TransferRequest{}

    if err := ctx.ShouldBindJSON(&req); err != nil {
        ctx.JSON(http.StatusBadRequest, errorResponse(err))
        return
    }

	FromAccount,valid:=s.ValidAccount(ctx,req.FromAccountID,req.Currency)
	if !valid {
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if FromAccount.Owner != authPayload.Username {
		err := errors.New("from account doesn't belong to the authenticated user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	_,valid=s.ValidAccount(ctx,req.ToAccountID,req.Currency)
	if !valid {
		return
	}
	
    arg := db.TransferTxParams{
        FromAccountID: req.FromAccountID,
        ToAccountID:   req.ToAccountID,
        Amount:        req.Amount,
    }

    account, err := s.store.TransferTx(ctx, arg)
    if err != nil {
        ctx.JSON(http.StatusBadRequest, errorResponse(err))
        return
    }
    ctx.JSON(http.StatusOK, account)
}


func (s *Server) ValidAccount(ctx *gin.Context,accountid int64, currency string) (db.Account,bool) {

	account,err:=s.store.GetAccount(ctx,accountid)
	if err != nil {
        if err == sql.ErrNoRows {
            ctx.JSON(http.StatusNotFound, errorResponse(err))
            return account,false
        }
        ctx.JSON(http.StatusInternalServerError, errorResponse(err))
        return account,false
    }

	if account.Currency!=currency{
		err:=fmt.Errorf("account %d currency mismatch: %s vs %s ",account.ID,account.Currency,currency)
		ctx.JSON(http.StatusBadRequest,errorResponse(err))
		return account,false
	}

	return account,true
}


