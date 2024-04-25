package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/go_backend_misc/db/sqlc"
	"github.com/go_backend_misc/util"
	"github.com/lib/pq"
)

type createUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=6"`
	FullName string `json:"full_name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
}

type userResponse struct {
	Username string `json:"username" binding:"required,alphanum"`
	FullName string `json:"full_name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
}

func createUserResponseFromUser(user *db.User) userResponse {
	return userResponse{
		Username: user.Username,
		FullName: user.FullName,
		Email:    user.Email,
	}
}

func (server *Server) createUser(ginCtx *gin.Context) {
	var req createUserRequest
	if err := ginCtx.ShouldBindJSON(&req); err != nil {
		ginCtx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		msg := "Error while processing password"
		ginCtx.JSON(http.StatusInternalServerError, errorMessageResponse(msg))
	}
	arg := db.CreateUserParams{
		Username:       req.Username,
		HashedPassword: hashedPassword,
		FullName:       req.FullName,
		Email:          req.Email,
	}
	user, err := server.store.CreateUser(ginCtx, arg)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			// Find the name of the error: log.Println(pqErr.Code.Name())
			var msg string
			switch pqErr.Code.Name() {
			case "unique_violation":
				if pqErr.Constraint == "users_pkey" {
					msg = "Username already exists"
					ginCtx.JSON(http.StatusConflict, errorMessageResponse(msg))
					return
				} else if pqErr.Constraint == "users_email_key" {
					msg = "Email already exists"
					ginCtx.JSON(http.StatusConflict, errorMessageResponse(msg))
					return
				} else {
					msg = "There was an error creating the user"
					ginCtx.JSON(http.StatusInternalServerError, errorMessageResponse(msg))
					return
				}

			default:
				msg = "There was an error creating the user"
				ginCtx.JSON(http.StatusInternalServerError, errorMessageResponse(msg))
				return
			}
		}

		msg := "There was an error creating the user"
		ginCtx.JSON(http.StatusInternalServerError, errorMessageResponse(msg))
		return
	}
	userResponse := createUserResponseFromUser(&user)
	ginCtx.JSON(http.StatusCreated, userResponse)
}

type loginUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=6"`
}

type loginUserResponse struct {
	AccessToken string `json:"token"`
	User        userResponse
}

func (server *Server) loginUser(ctx *gin.Context) {
	var request loginUserRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	user, err := server.store.GetUserByUsername(ctx, request.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorMessageResponse("invalid credentials"))
			return
		}
		// unexpected DB error
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
	}

	err = util.CheckPassword(request.Password, user.HashedPassword)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, errorMessageResponse("invalid credentials"))
		return
	}

	accessToken, err := server.tokenMaker.CreateToken(user.Username, server.config.AccessTokenDuration)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	response := loginUserResponse{
		AccessToken: accessToken,
		User:        createUserResponseFromUser(&user),
	}
	ctx.JSON(http.StatusOK, response)
}
