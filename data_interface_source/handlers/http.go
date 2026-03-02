package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"data_interface-api/config"
)

func DynamicDataHandler(dbs map[string]*sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var payload map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}

		apiKeyVal, ok := payload["api_key"]
		if !ok {
			http.Error(w, "Missing API key", http.StatusBadRequest)
			return
		}
		apiKey := fmt.Sprintf("%v", apiKeyVal)

		cfg, ok := config.DynamicMappings.Mappings[apiKey]
		if !ok {
			http.Error(w, "Invalid API key", http.StatusUnauthorized)
			return
		}

		db, ok := dbs[cfg.Database]
		if !ok {
			http.Error(w, "Unknown database", http.StatusBadRequest)
			return
		}

		tableName := cfg.Table
		if strings.Contains(tableName, "%s") {
			tableName = fmt.Sprintf(tableName, time.Now().Format("2006_01_02_15_04_05"))
		}

		if cfg.CreateIfMissing {
			if len(cfg.Columns) > 0 {
				var cols []string
				for k, t := range cfg.Columns {
					cols = append(cols, fmt.Sprintf("%s %s", k, t))
				}
				cols = append(cols, "timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP")
				query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s);", tableName, strings.Join(cols, ", "))
				if _, err := db.Exec(query); err != nil {
					http.Error(w, "Table creation failed: "+err.Error(), http.StatusInternalServerError)
					return
				}
			}
		}

		var data map[string]interface{}
		if content, ok := payload["content"].(map[string]interface{}); ok {
			data = content
		} else {
			data = payload
		}

		columns := []string{}
		values := []interface{}{}
		placeholders := []string{}
		i := 1
		for k, v := range data {
			if k == "api_key" {
				continue
			}
			columns = append(columns, k)
			values = append(values, v)
			placeholders = append(placeholders, fmt.Sprintf("$%d", i))
			i++
		}

		if len(columns) == 0 {
			http.Error(w, "No data fields found", http.StatusBadRequest)
			return
		}

		query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", tableName,
			strings.Join(columns, ","), strings.Join(placeholders, ","))
		if _, err := db.Exec(query, values...); err != nil {
			http.Error(w, "Insert failed: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Data inserted dynamically"))
	}
}
