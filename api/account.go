package api

import (
	"database/sql"
	"fmt"

	"net/http"
	"strconv"

	db "github.com/TheRanomial/bank_server/db/sqlc"
	"github.com/TheRanomial/bank_server/token"
	"github.com/gin-gonic/gin"
)

type CreateAccountRequest struct {
	Balance  	int64  `json:"balance" binding:"required"`
	Currency 	string `json:"currency" binding:"required,oneof=USD INR GBP EUR"`
}

func (s *Server) CreateAccount(ctx *gin.Context) {
	req:=CreateAccountRequest{}

	if err:=ctx.ShouldBindJSON(&req); err!=nil {
		ctx.JSON(http.StatusBadRequest,errorResponse(err))
		return
	}

	authPayload:=ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	arg:=db.CreateAccountParams{
		Owner: authPayload.Username,
		Balance: req.Balance,
		Currency: req.Currency,
	}

	account,err:=s.store.CreateAccount(ctx,arg)
	if err != nil {
		errCode := db.ErrorCode(err)
		if errCode == db.ForeignKeyViolation || errCode == db.UniqueViolation {
			ctx.JSON(http.StatusForbidden, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	
	ctx.JSON(http.StatusOK,account)
}

func (s *Server) Getaccount(ctx *gin.Context){

	idparam:=ctx.Param("id")
	id, err := strconv.ParseInt(idparam, 10, 64)
	if err!=nil{
		ctx.JSON(http.StatusBadRequest,errorResponse(err))
		return
	}

	account,err:=s.store.GetAccount(ctx,id)
	if err != nil {
        if err == sql.ErrNoRows {
            ctx.JSON(http.StatusNotFound, errorResponse(err))
            return
        }
        ctx.JSON(http.StatusInternalServerError, errorResponse(err))
        return
    }

	authPayload:= ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if account.Owner != authPayload.Username {
		err = fmt.Errorf("account %d does not belong to the authenticated user", account.ID)
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK,account)
}

type ListAccountsRequest struct {
	PageID		int32   `form:"page_id" binding:"required,min=1"`
	PageSize    int32   `form:"page_size" binding:"required,min=5,max=10"`
}

func (s *Server) ListAccounts(ctx *gin.Context){

	req:=ListAccountsRequest{}
	if err:=ctx.ShouldBindQuery(&req); err!=nil {
		ctx.JSON(http.StatusBadRequest,errorResponse(err))
		return
	}

	authPayload:= ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	arg:=db.ListAccountsParams{
		Owner: authPayload.Username,
		Limit: req.PageSize,
		Offset: (req.PageID-1)*req.PageSize,
	}

	accounts,err:=s.store.ListAccounts(ctx,arg)
	if err!=nil{
		if err==sql.ErrNoRows{
			ctx.JSON(http.StatusNotFound,errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError,errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK,accounts)
}