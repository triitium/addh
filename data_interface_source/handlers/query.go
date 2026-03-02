package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

func QueryHandler(dbs map[string]*sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dbName := r.URL.Query().Get("database")
		tableName := r.URL.Query().Get("table")
		maxRows := r.URL.Query().Get("max")

		if dbName == "" || tableName == "" || maxRows == "" {
			http.Error(w, "Missing required parameters: database, table, or max", http.StatusBadRequest)
			return
		}

		db, ok := dbs[dbName]
		if !ok {
			http.Error(w, "Unknown database", http.StatusBadRequest)
			return
		}

		filter := r.URL.Query().Get("filter")
		order := r.URL.Query().Get("order_by")
		columnsCSV := r.URL.Query().Get("columns")
		opsCSV := r.URL.Query().Get("ops")
		valuesJSON := r.URL.Query().Get("values")
		fromStr := r.URL.Query().Get("from")
		toStr := r.URL.Query().Get("to")
		orderBy := r.URL.Query().Get("orderBy")
		orderDir := strings.ToUpper(r.URL.Query().Get("orderDir"))
		offset := r.URL.Query().Get("offset")

		var columns []string
		if columnsCSV != "" {
			columns = strings.Split(columnsCSV, ",")
		}

		var ops []string
		if opsCSV != "" {
			ops = strings.Split(opsCSV, ",")
		}

		var values []interface{}
		if valuesJSON != "" {
			if err := json.Unmarshal([]byte(valuesJSON), &values); err != nil {
				http.Error(w, "Invalid values parameter: "+err.Error(), http.StatusBadRequest)
				return
			}
		}

		var fromTime, toTime *time.Time
		if fromStr != "" {
			t, err := time.Parse(time.RFC3339, fromStr)
			if err != nil {
				http.Error(w, "Invalid from timestamp", http.StatusBadRequest)
				return
			}
			fromTime = &t
		}
		if toStr != "" {
			t, err := time.Parse(time.RFC3339, toStr)
			if err != nil {
				http.Error(w, "Invalid to timestamp", http.StatusBadRequest)
				return
			}
			toTime = &t
		}

		selectClause := "*"
		if len(columns) > 0 {
			selectClause = strings.Join(columns, ", ")
		}
		query := fmt.Sprintf("SELECT %s FROM %s", selectClause, tableName)

		var whereClauses []string
		args := []interface{}{}

		for i, col := range columns {
			if i < len(ops) && i < len(values) {
				whereClauses = append(whereClauses, fmt.Sprintf("%s %s $%d", col, ops[i], len(args)+1))
				args = append(args, values[i])
			}
		}
		if fromTime != nil {
			whereClauses = append(whereClauses, fmt.Sprintf("timestamp >= $%d", len(args)+1))
			args = append(args, *fromTime)
		}
		if toTime != nil {
			whereClauses = append(whereClauses, fmt.Sprintf("timestamp <= $%d", len(args)+1))
			args = append(args, *toTime)
		}
		if filter != "" {
			whereClauses = append(whereClauses, filter)
		}

		if len(whereClauses) > 0 {
			query += " WHERE " + strings.Join(whereClauses, " AND ")
		}

		if orderBy != "" {
			if orderDir != "ASC" && orderDir != "DESC" {
				orderDir = "ASC"
			}
			query += fmt.Sprintf(" ORDER BY %s %s", orderBy, orderDir)
		} else if order != "" {
			query += " ORDER BY " + order
		}

		query += fmt.Sprintf(" LIMIT %s", maxRows)
		if offset != "" {
			query += fmt.Sprintf(" OFFSET %s", offset)
		}

		rows, err := db.Query(query, args...)
		if err != nil {
			http.Error(w, "Query failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		cols, err := rows.Columns()
		if err != nil {
			http.Error(w, "Failed to get columns: "+err.Error(), http.StatusInternalServerError)
			return
		}

		results := []map[string]interface{}{}
		for rows.Next() {
			colsPointers := make([]interface{}, len(cols))
			colsValues := make([]interface{}, len(cols))
			for i := range colsPointers {
				colsPointers[i] = &colsValues[i]
			}
			if err := rows.Scan(colsPointers...); err != nil {
				http.Error(w, "Row scan failed: "+err.Error(), http.StatusInternalServerError)
				return
			}
			rowMap := map[string]interface{}{}
			for i, colName := range cols {
				val := colsValues[i]
				if b, ok := val.([]byte); ok {
					rowMap[colName] = string(b)
				} else {
					rowMap[colName] = val
				}
			}
			results = append(results, rowMap)
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(results); err != nil {
			log.Printf("Failed to encode JSON: %v", err)
		}
	}
}

func ListTablesHandler(dbs map[string]*sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dbName := r.URL.Query().Get("database")
		if dbName == "" {
			http.Error(w, "Missing database parameter", http.StatusBadRequest)
			return
		}

		db, ok := dbs[dbName]
		if !ok {
			http.Error(w, "Unknown database", http.StatusBadRequest)
			return
		}

		query := "SELECT tablename FROM pg_catalog.pg_tables WHERE schemaname='public'"
		rows, err := db.Query(query)
		if err != nil {
			http.Error(w, "Query failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		tables := []string{}
		for rows.Next() {
			var t string
			if err := rows.Scan(&t); err != nil {
				http.Error(w, "Row scan failed: "+err.Error(), http.StatusInternalServerError)
				return
			}
			tables = append(tables, t)
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(tables); err != nil {
			log.Printf("Failed to encode JSON: %v", err)
		}
	}
}
