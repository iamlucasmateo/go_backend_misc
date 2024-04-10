package api

import (
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

type createUserResponse struct {
	Username string `json:"username" binding:"required,alphanum"`
	FullName string `json:"full_name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
}

func createUserResponseFromUser(user *db.User) createUserResponse {
	return createUserResponse{
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

	}
	userResponse := createUserResponseFromUser(&user)
	ginCtx.JSON(http.StatusCreated, userResponse)
}
