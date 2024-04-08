package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/go_backend_misc/db/sqlc"
)

type transferRequest struct {
	FromAccountID int64  `json:"from_account_id" binding:"required,min=1"`
	ToAccountID   int64  `json:"to_account_id" binding:"required,min=1"`
	Amount        int64  `json:"amount" binding:"required,gt=0"`
	Currency      string `json:"currency" binding:"required,oneof=USD EUR CAD"`
}

func (server *Server) createTransfer(ginCtx *gin.Context) {
	var req transferRequest
	if err := ginCtx.ShouldBindJSON(&req); err != nil {
		ginCtx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !server.validAccount(ginCtx, req.FromAccountID, req.Currency) {
		return
	}

	if !server.validAccount(ginCtx, req.ToAccountID, req.Currency) {
		return
	}

	arg := db.CreateTransferParams{
		FromAccountID: db.Int64ToSqlInt64(req.FromAccountID),
		ToAccountID:   db.Int64ToSqlInt64(req.ToAccountID),
		Amount:        req.Amount,
	}

	transferResult, err := server.store.TransferTx(ginCtx, arg)
	if err != nil {
		ginCtx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ginCtx.JSON(http.StatusOK, transferResult)

}

func (server *Server) validAccount(ctx *gin.Context, accountID int64, currency string) bool {
	account, err := server.store.GetAccount(ctx, accountID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return false
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return false
	}

	if account.Currency != currency {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "account currency mismatch"})
		return false
	}

	return true

}
