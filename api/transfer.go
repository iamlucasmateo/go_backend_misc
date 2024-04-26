package api

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/go_backend_misc/db/sqlc"
	"github.com/go_backend_misc/token"
)

type transferRequest struct {
	FromAccountID int64  `json:"from_account_id" binding:"required,min=1"`
	ToAccountID   int64  `json:"to_account_id" binding:"required,min=1"`
	Amount        int64  `json:"amount" binding:"required,gt=0"`
	Currency      string `json:"currency" binding:"required,currency"`
}

func (server *Server) createTransfer(ginCtx *gin.Context) {
	var req transferRequest
	if err := ginCtx.ShouldBindJSON(&req); err != nil {
		ginCtx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fromAccount, isValid := server.validAccount(ginCtx, req.FromAccountID, req.Currency)
	if !isValid {
		return
	}

	if _, isValid := server.validAccount(ginCtx, req.ToAccountID, req.Currency); !isValid {
		return
	}

	authPayload := ginCtx.MustGet(authorizationPayloadKey).(*token.Payload)
	if fromAccount.Owner != authPayload.Username {
		ginCtx.JSON(http.StatusUnauthorized, errorResponse(errors.New("unauthorized user")))
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

func (server *Server) validAccount(ctx *gin.Context, accountID int64, currency string) (account db.Account, isValid bool) {
	account, err := server.store.GetAccount(ctx, accountID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return account, false
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return account, false
	}

	if account.Currency != currency {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "account currency mismatch"})
		return account, false
	}

	return account, true

}
