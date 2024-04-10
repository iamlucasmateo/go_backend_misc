package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"

	db "github.com/go_backend_misc/db/sqlc"
)

type Server struct {
	store  db.Store
	router *gin.Engine
}

type ServerStatus struct {
	Message string
}

func NewServer(store db.Store) *Server {
	server := &Server{store: store}
	router := gin.Default()

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("currency", validCurrency)
	}

	router.GET("/status", server.status)

	router.POST("/account", server.createAccount)
	router.GET("/account/:id", server.getAccount)
	router.GET("/accounts/", server.listAccounts)

	router.POST("/transfer", server.createTransfer)

	router.POST("/user", server.createUser)

	server.router = router

	return server
}

func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}

func errorMessageResponse(message string) gin.H {
	return gin.H{"error": message}
}

func (server *Server) status(ginCtx *gin.Context) {
	serverStatus := &ServerStatus{Message: "OK"}
	ginCtx.JSON(http.StatusOK, serverStatus)

}
