package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"

	db "github.com/go_backend_misc/db/sqlc"
	"github.com/go_backend_misc/token"
	"github.com/go_backend_misc/util"
)

type Server struct {
	config     util.Config
	store      db.Store
	tokenMaker token.TokenMaker
	router     *gin.Engine
}

type ServerStatus struct {
	Message string
}

func NewServer(config util.Config, store db.Store) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}
	server := &Server{
		config:     config,
		store:      store,
		tokenMaker: tokenMaker,
	}

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("currency", validCurrency)
	}

	server.setupRouter()

	return server, nil
}

func (server *Server) setupRouter() {
	router := gin.Default()

	router.GET("/status", server.status)
	router.POST("/user", server.createUser)
	router.POST("/user/login", server.loginUser)

	authRoutes := router.Group("/").Use(authMiddleware(server.tokenMaker))

	authRoutes.POST("/account", server.createAccount)
	authRoutes.GET("/account/:id", server.getAccount)
	authRoutes.GET("/accounts/", server.listAccounts)

	authRoutes.POST("/transfer", server.createTransfer)

	server.router = router
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
