package db

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	dbDriver = "postgres"
	dbSource = "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable"
)

var testQueries *Queries
var testStore *Store
var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
    var err error
    testPool, err = pgxpool.New(context.Background(), dbSource)
    if err != nil {
        log.Fatal("Cannot connect to db:", err)
    }
    defer testPool.Close()

    testQueries = New(testPool)
    testStore = NewStore(testPool)  
    os.Exit(m.Run())
}