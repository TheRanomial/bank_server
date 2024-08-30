package gapi

import (
	"fmt"

	db "github.com/TheRanomial/bank_server/db/sqlc"
	"github.com/TheRanomial/bank_server/pb"
	"github.com/TheRanomial/bank_server/token"
	"github.com/TheRanomial/bank_server/util"
	"github.com/TheRanomial/bank_server/worker"
)

type Server struct{
	pb.UnimplementedSimpleBankServer
	config util.Config
	store  db.Store
	tokenMaker token.Maker
	taskDistributor worker.TaskDistributor
}

func NewServer(config util.Config,store db.Store,taskdistributor worker.TaskDistributor) (*Server,error){

	tokenMaker,err:=token.NewPasetoMaker(config.TokenSymmetricKey)
	if err!=nil{
		return nil,fmt.Errorf("cannot create token maker: %w",err)
	}

	server:=&Server{
		config: config,
		store: store,
		tokenMaker: tokenMaker,
		taskDistributor: taskdistributor,
	}
	return server,nil
}

