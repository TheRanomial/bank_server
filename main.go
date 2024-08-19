package main

import (
	"context"
	"log"

	"github.com/TheRanomial/bank_server/api"
	db "github.com/TheRanomial/bank_server/db/sqlc"
	"github.com/TheRanomial/bank_server/util"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
)

func main(){

	config,err:=util.LoadConfig(".")
	if err!=nil{
		log.Fatal("Cannot load configs")
	}
	
	conn, err:= pgxpool.New(context.Background(), config.DBSource)
    if err != nil {
        log.Fatal("Cannot connect to db:", err)
    }
    defer conn.Close()

    store:=db.NewStore(conn)
    server,err:=api.NewServer(config,store)
	if err!=nil {
		log.Fatal("Cannot start the server",err)
	}
    
	err=server.Start(config.ServerAddress)
	if err!=nil {
		log.Fatal("Cannot start the server",err)
	}
}