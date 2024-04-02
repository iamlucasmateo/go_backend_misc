package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	db "github.com/go_backend_misc/db/sqlc"
)

type Server struct {
	store  *db.Store
	router *gin.Engine
}

type ServerStatus struct {
	Message string
}

func NewServer(store *db.Store) *Server {
	server := &Server{store: store}
	router := gin.Default()

	router.GET("/status", server.status)

	router.POST("/account", server.createAccount)
	router.GET("/account/:id", server.getAccount)
	router.GET("/accounts/", server.listAccounts)

	server.router = router

	return server
}

func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}

func (server *Server) status(ginCtx *gin.Context) {
	serverStatus := &ServerStatus{Message: "OK"}
	ginCtx.JSON(http.StatusOK, serverStatus)

}
