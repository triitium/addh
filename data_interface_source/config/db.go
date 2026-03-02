package config

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

func ConnectDatabases() (map[string]*sql.DB, error) {
	dbURL := os.Getenv("POSTGRES_URL")
	dbUser := os.Getenv("POSTGRES_USER")
	dbPassword := os.Getenv("POSTGRES_PASSWORD")

	dbSet := map[string]struct{}{}
	for _, cfg := range DynamicMappings.Mappings {
		dbSet[cfg.Database] = struct{}{}
	}

	dbs := map[string]*sql.DB{}
	for dbName := range dbSet {
		connStr := fmt.Sprintf(
			"postgres://%s:%s@%s/%s?sslmode=disable",
			dbUser,
			dbPassword,
			dbURL,
			dbName,
		)
		dbConn, err := sql.Open("postgres", connStr)
		if err != nil {
			return nil, err
		}
		dbs[dbName] = dbConn
	}

	return dbs, nil
}
