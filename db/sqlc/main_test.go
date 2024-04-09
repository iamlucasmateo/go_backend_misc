package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	// requires the _ identifier to avoid the "imported and not used" error
	"github.com/go_backend_misc/util"
	_ "github.com/lib/pq"
)

var testQueries *Queries
var testDB *sql.DB
var testUser User

func TestMain(m *testing.M) {
	config, err := util.LoadConfig("../..")
	if err != nil {
		log.Fatalf("cannot load config: %v", err)
	}
	testDB, err = sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatalf("cannot connect to db: %v", err)
	}

	testQueries = New(testDB)

	// to reuse a user
	// testUser, _, _ = createRandomUser("_test_create_account")

	os.Exit(m.Run())
}
