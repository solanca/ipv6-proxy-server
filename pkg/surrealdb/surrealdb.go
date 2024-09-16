package surrealdb

import (
	"atlas/pkg/logging"
	"os"

	surreal "github.com/surrealdb/surrealdb.go"
)

var (
	// Database is the database for the server.
	Database = New()
)

func New() surreal.DB {
	db, err := surreal.New("ws://161.129.154.58:8000/rpc")
	if err != nil {
		logging.Logger.Fatal().
			Err(err).
			Msg("Failed to initialize to database connection")
		os.Exit(0)
	}

	_, err = db.Signin(map[string]interface{}{
		"user": "root",
		"pass": "M7M__j.PrHT@TZdUcD+Bzg",
	})
	if err != nil {
		logging.Logger.Fatal().
			Err(err).
			Msg("Failed to initialize to database connection")
		os.Exit(0)
	}

	_, err = db.Use("snowproxies", "snowproxies")
	if err != nil {
		logging.Logger.Fatal().
			Err(err).
			Msg("Failed to initialize to database connection")
		os.Exit(0)
	}

	return *db
}
