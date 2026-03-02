package main

import (
	"log"
	"net/http"
	"os"

	"data_interface-api/config"
	"data_interface-api/handlers"
)

func main() {
    configPath := os.Getenv("CONFIG_PATH")
    if configPath == "" {
        configPath = "/interface/config/dynamic_mapping.json"
    }

	err := config.LoadDynamicConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load dynamic config: %v", err)
	}
	log.Printf("Loaded API keys: %+v\n", config.DynamicMappings.Mappings)

	dbs, err := config.ConnectDatabases()
	if err != nil {
		log.Fatalf("Failed to connect to databases: %v", err)
	}
	defer func() {
		for _, db := range dbs {
			db.Close()
		}
	}()

	http.HandleFunc("/datastream/input/entrypoint", handlers.RecoverMiddleware(handlers.DynamicDataHandler(dbs)))
	http.HandleFunc("/datastream/output/generic", handlers.RecoverMiddleware(handlers.QueryHandler(dbs)))
	http.HandleFunc("/datastream/output/tables", handlers.RecoverMiddleware(handlers.ListTablesHandler(dbs)))

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	addr := ":" + port
	log.Printf("Datainterface API Go server running on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
