package api

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/go_backend_misc/db/sqlc"
	"github.com/go_backend_misc/token"
	"github.com/lib/pq"
)

type createAccountRequest struct {
	Currency string `json:"currency" binding:"required,currency"`
}

func (server *Server) createAccount(ginCtx *gin.Context) {
	var req createAccountRequest
	if err := ginCtx.ShouldBindJSON(&req); err != nil {
		ginCtx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	authPayload := ginCtx.MustGet(authorizationPayloadKey).(*token.Payload)
	arg := db.CreateAccountParams{
		Owner:    authPayload.Username,
		Currency: req.Currency,
		Balance:  0,
	}

	account, err := server.store.CreateAccount(ginCtx, arg)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			// Find the name of the error: log.Println(pqErr.Code.Name())
			switch pqErr.Code.Name() {
			case "foreign_key_violation":
				msg := "Owner does not exist"
				ginCtx.JSON(http.StatusConflict, errorMessageResponse(msg))
				return
			case "unique_violation":
				msg := "Account with that owner and currency already exists"
				ginCtx.JSON(http.StatusConflict, errorMessageResponse(msg))
				return
			}
		}
		ginCtx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ginCtx.JSON(http.StatusOK, account)

}

type getAccountParams struct {
	Id int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) getAccount(ctx *gin.Context) {
	var req getAccountParams
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	account, err := server.store.GetAccount(ctx, req.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
		} else {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		}
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if account.Owner != authPayload.Username {
		err := errors.New("account doesn't belong to the authenticated user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, account)

}

type listAccountsQueryParams struct {
	Offset   int32 `form:"offset" binding:"required,min=0"`
	PageSize int32 `form:"page_size" binding:"required,min=1,max=20"`
}

func (server *Server) listAccounts(ctx *gin.Context) {
	var req listAccountsQueryParams
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	listAccountParams := db.ListAccountsByUsernameParams{
		Owner:  authPayload.Username,
		Limit:  req.PageSize,
		Offset: req.Offset,
	}
	accounts, err := server.store.ListAccountsByUsername(ctx, listAccountParams)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
		} else {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		}
		return
	}

	ctx.JSON(http.StatusOK, accounts)

}
