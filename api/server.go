package api

import (
	"fmt"

	db "github.com/TheRanomial/bank_server/db/sqlc"
	"github.com/TheRanomial/bank_server/token"
	"github.com/TheRanomial/bank_server/util"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

type Server struct{
	config util.Config
	store  db.Store
	tokenMaker token.Maker
	router *gin.Engine
}

func NewServer(config util.Config,store db.Store) (*Server,error){

	tokenMaker,err:=token.NewPasetoMaker(config.TokenSymmetricKey)
	if err!=nil{
		return nil,fmt.Errorf("cannot create token maker: %w",err)
	}

	server:=&Server{
		config: config,
		store: store,
		tokenMaker: tokenMaker,
	}

	if v,ok:=binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("currency",validCurrency)
	}

	server.setupRouter()
	return server,nil
}

func (server *Server) setupRouter(){
	router:=gin.Default()
	authRoutes:=router.Group("/").Use(authMiddleware(server.tokenMaker))

	router.POST("/users",server.CreateUser)
	router.POST("/users/login",server.LoginUser)


	authRoutes.POST("/accounts", server.CreateAccount)
    authRoutes.GET("/accounts/:id", server.Getaccount)
    authRoutes.GET("/accounts", server.ListAccounts)
    authRoutes.POST("/transfers", server.CreateTransfer)

	server.router=router
}

func (s *Server) Start(address string) error {
	return s.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}


