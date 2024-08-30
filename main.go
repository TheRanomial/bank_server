package main

import (
	"context"
	"log"
	"net"
	"net/http"

	"github.com/TheRanomial/bank_server/api"
	db "github.com/TheRanomial/bank_server/db/sqlc"
	"github.com/TheRanomial/bank_server/gapi"
	"github.com/TheRanomial/bank_server/mail"
	"github.com/TheRanomial/bank_server/pb"
	"github.com/TheRanomial/bank_server/util"
	"github.com/TheRanomial/bank_server/worker"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
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

	runDBmigrations(config.MigrationURL,config.DBSource)

    store:=db.NewStore(conn)

	opts:=asynq.RedisClientOpt{
		Addr: config.RedisAddress,
	}

	taskDistributor:=worker.NewRedisTaskDistributor(opts)
	go runTaskProcessor(opts,store,config)
	go RunGatewayServer(config,store,taskDistributor)
    RunGrpcServer(config,store,taskDistributor)
}

func runDBmigrations(migrationURL string,dbSource string){

	migration,err:=migrate.New(migrationURL,dbSource)
	if err!=nil {
		log.Fatal("Could not migrate due to ",err)
	}

	if err:=migration.Up(); err!=nil && err!=migrate.ErrNoChange{
		log.Fatal("Could not apply migrations ",err)
	}
	log.Println("DB Migrate successfully")
}

func runTaskProcessor(redisOpt asynq.RedisClientOpt,store db.Store,config util.Config,){
	mailer := mail.NewGmailSender(config.EmailSenderName, config.EmailSenderAddress, config.EmailSenderPassword)
	taskProcessor:=worker.NewRedisTaskProcessor(redisOpt,store,mailer)
	log.Println("start task processor")
	err:=taskProcessor.Start()
	if err!=nil {
		log.Fatal(err)
	}
}

func RunGrpcServer(config util.Config, store db.Store,taskdistributor worker.TaskDistributor){

	server,err:=gapi.NewServer(config,store,taskdistributor)
	if err!=nil{
		log.Fatal("Cannot create server: ",err)
	}

	gprcLogger := grpc.UnaryInterceptor(gapi.GrpcLogger)
	grpcServer:=grpc.NewServer(gprcLogger)
	pb.RegisterSimpleBankServer(grpcServer,server)
	reflection.Register(grpcServer)

	listener,err:=net.Listen("tcp",config.GRPCServerAddress)
	if err!=nil{
		log.Fatal("Cannot create server: ",err)
	}

	log.Printf("Starting gRPC server at %s",listener.Addr().String())
	err=grpcServer.Serve(listener)
	if err!=nil{
		log.Fatal("Cannot start server: ",err)
	}
}

func RunGatewayServer(config util.Config, store db.Store,taskdistributor worker.TaskDistributor){

	//our server
	server,err:=gapi.NewServer(config,store,taskdistributor)
	if err!=nil{
		log.Fatal("Cannot create server: ",err)
	}

	ctx,cancel:=context.WithCancel(context.Background())
	defer cancel()

	//grpc server
	grpcMux:=runtime.NewServeMux()
	err=pb.RegisterSimpleBankHandlerServer(ctx,grpcMux,server)
	if err!=nil{
		log.Fatal(err)
	}

	//rerouting to grpc
	mux:=http.NewServeMux()
	mux.Handle("/",grpcMux)

	listener,err:=net.Listen("tcp",config.HTTPServerAddress)
	if err!=nil{
		log.Fatal("Cannot create server: ",err)
	}

	log.Printf("Starting gRPC server at %s",listener.Addr().String())
	err=http.Serve(listener,mux)
	if err!=nil{
		log.Fatal("Cannot start HTTP gateway server: ",err)
	}

}

func runGinServer(config util.Config, store db.Store) {
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create server")
	}

	err = server.Start(config.HTTPServerAddress)
	if err != nil {
		log.Fatal("cannot start server")
	}
}