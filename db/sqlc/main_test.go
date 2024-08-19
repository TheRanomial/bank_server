package db

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/TheRanomial/bank_server/util"
	"github.com/jackc/pgx/v5/pgxpool"
)


var testQueries *Queries
var testStore Store
var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
    var err error
    config,_:=util.LoadConfig("../..")
    testPool, err = pgxpool.New(context.Background(), config.DBSource)
    if err != nil {
        log.Fatal("Cannot connect to db:", err)
    }
    defer testPool.Close()

    testQueries = New(testPool)
    testStore = NewStore(testPool)  
    os.Exit(m.Run())
}