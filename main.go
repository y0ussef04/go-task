package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	_ "github.com/go-sql-driver/mysql"
)

// Global variables
var (
	currentDatabase string
	dbMutex         sync.RWMutex
)

// Response structures
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Error   string      `json:"error,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type ColumnInfo struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type Record map[string]interface{}

// dbConn connects to MySQL server with current database if set
func dbConn() *sql.DB {
	dbMutex.RLock()
	dbName := currentDatabase
	dbMutex.RUnlock()

	connStr := "root:secret@tcp(mysql:3306)/"
	if dbName != "" {
		connStr = fmt.Sprintf("root:secret@tcp(mysql:3306)/%s", dbName)
	}

	db, err := sql.Open("mysql", connStr)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

// CreateDatabase creates a new database with the given name
func CreateDatabase(db *sql.DB, dbName string) error {
	query := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", dbName)
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create database %s: %v", dbName, err)
	}
	fmt.Printf("✅ Database '%s' created successfully\n", dbName)
	return nil
}

// UseDatabase switches to the specified database
func UseDatabase(db *sql.DB, dbName string) error {
	query := fmt.Sprintf("USE %s", dbName)
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to use database %s: %v", dbName, err)
	}

	// Update global current database
	dbMutex.Lock()
	currentDatabase = dbName
	dbMutex.Unlock()

	fmt.Printf("✅ Switched to database '%s'\n", dbName)
	return nil
}

// ListDatabases returns a list of all available databases
func ListDatabases(db *sql.DB) ([]string, error) {
	query := "SHOW DATABASES"
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list databases: %v", err)
	}
	defer rows.Close()

	var databases []string
	for rows.Next() {
		var dbName string
		if err := rows.Scan(&dbName); err != nil {
			return nil, fmt.Errorf("failed to scan database name: %v", err)
		}
		databases = append(databases, dbName)
	}
	return databases, nil
}

// ListTables returns a list of all tables in the current database
func ListTables(db *sql.DB) ([]string, error) {
	query := "SHOW TABLES"
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list tables: %v", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, fmt.Errorf("failed to scan table name: %v", err)
		}
		tables = append(tables, tableName)
	}
	return tables, nil
}

// CreateTable creates a table with the given name and columns
func CreateTable(db *sql.DB, tableName string, columns map[string]string) error {
	if len(columns) == 0 {
		return fmt.Errorf("no columns provided for table %s", tableName)
	}

	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (", tableName)
	cols := make([]string, 0, len(columns))
	for col, typ := range columns {
		cols = append(cols, fmt.Sprintf("%s %s", col, typ))
	}
	query += strings.Join(cols, ", ") + ")"

	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create table %s: %v", tableName, err)
	}
	fmt.Printf("✅ Table '%s' created successfully\n", tableName)
	return nil
}

// AddColumn adds a new column to an existing table
func AddColumn(db *sql.DB, tableName string, columnName string, dataType string) error {
	query := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", tableName, columnName, dataType)
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to add column %s to table %s: %v", columnName, tableName, err)
	}
	fmt.Printf("✅ Column '%s' added to table '%s'\n", columnName, tableName)
	return nil
}

// GetTableColumns returns the list of column names for a table
func GetTableColumns(db *sql.DB, tableName string) ([]string, error) {
	query := fmt.Sprintf("SHOW COLUMNS FROM %s", tableName)
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get columns for table %s: %v", tableName, err)
	}
	defer rows.Close()

	var columns []string
	for rows.Next() {
		var field, typ, null, key, defaultVal, extra sql.NullString
		err := rows.Scan(&field, &typ, &null, &key, &defaultVal, &extra)
		if err != nil {
			return nil, err
		}
		if field.Valid {
			columns = append(columns, field.String)
		}
	}
	return columns, nil
}

// GetColumnTypes returns a map of column name to data type for a table
func GetColumnTypes(db *sql.DB, tableName string) (map[string]string, error) {
	query := fmt.Sprintf("SHOW COLUMNS FROM %s", tableName)
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get column types for table %s: %v", tableName, err)
	}
	defer rows.Close()

	columnTypes := make(map[string]string)
	for rows.Next() {
		var field, typ, null, key, defaultVal, extra sql.NullString
		err := rows.Scan(&field, &typ, &null, &key, &defaultVal, &extra)
		if err != nil {
			return nil, err
		}
		if field.Valid && typ.Valid {
			columnTypes[field.String] = typ.String
		}
	}
	return columnTypes, nil
}

// Insert inserts a new record into the table
func Insert(db *sql.DB, tableName string, data map[string]interface{}) error {
	if len(data) == 0 {
		return fmt.Errorf("no data provided for insert")
	}

	columns := make([]string, 0, len(data))
	placeholders := make([]string, 0, len(data))
	values := make([]interface{}, 0, len(data))

	for col, val := range data {
		columns = append(columns, col)
		placeholders = append(placeholders, "?")
		values = append(values, val)
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		tableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))

	_, err := db.Exec(query, values...)
	if err != nil {
		return fmt.Errorf("failed to insert into table %s: %v", tableName, err)
	}
	fmt.Printf("✅ Record inserted into '%s'\n", tableName)
	return nil
}

// Select retrieves all records from the table
func Select(db *sql.DB, tableName string) ([]Record, error) {
	columns, err := GetTableColumns(db, tableName)
	if err != nil {
		return nil, err
	}

	if len(columns) == 0 {
		return []Record{}, nil
	}

	query := fmt.Sprintf("SELECT %s FROM %s", strings.Join(columns, ", "), tableName)
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to select from table %s: %v", tableName, err)
	}
	defer rows.Close()

	var records []Record
	for rows.Next() {
		values := make([]interface{}, len(columns))
		scanArgs := make([]interface{}, len(columns))
		for i := range values {
			scanArgs[i] = &values[i]
		}

		err := rows.Scan(scanArgs...)
		if err != nil {
			return nil, err
		}

		record := make(Record)
		for i, col := range columns {
			if values[i] == nil {
				record[col] = nil
			} else {
				record[col] = values[i]
			}
		}
		records = append(records, record)
	}
	return records, nil
}

// Update updates records in the table based on condition
func Update(db *sql.DB, tableName string, data map[string]interface{}, condition string) error {
	if len(data) == 0 {
		return fmt.Errorf("no data provided for update")
	}

	setParts := make([]string, 0, len(data))
	values := make([]interface{}, 0, len(data))

	for col, val := range data {
		setParts = append(setParts, fmt.Sprintf("%s = ?", col))
		values = append(values, val)
	}

	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s",
		tableName,
		strings.Join(setParts, ", "),
		condition)

	_, err := db.Exec(query, values...)
	if err != nil {
		return fmt.Errorf("failed to update table %s: %v", tableName, err)
	}
	fmt.Printf("✅ Records updated in '%s'\n", tableName)
	return nil
}

// Delete deletes records from the table based on condition
func Delete(db *sql.DB, tableName string, condition string) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE %s", tableName, condition)
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to delete from table %s: %v", tableName, err)
	}
	fmt.Printf("✅ Records deleted from '%s'\n", tableName)
	return nil
}

// Create sample RealEstate database
func createSampleTables(db *sql.DB) error {
	// Create database
	err := CreateDatabase(db, "RealEstate")
	if err != nil {
		return err
	}

	err = UseDatabase(db, "RealEstate")
	if err != nil {
		return err
	}

	// Create Campaign table
	err = CreateTable(db, "Campaign", map[string]string{
		"camp_id":   "INT AUTO_INCREMENT PRIMARY KEY",
		"Camp_name": "VARCHAR(255)",
		"Cost":      "FLOAT",
	})
	if err != nil {
		return err
	}

	// Create Agent table
	err = CreateTable(db, "Agent", map[string]string{
		"Agent_id":      "INT AUTO_INCREMENT PRIMARY KEY",
		"Agent_name":    "VARCHAR(255)",
		"Agent_salary":  "FLOAT",
		"Agent_Address": "VARCHAR(255)",
	})
	if err != nil {
		return err
	}

	// Create Properties table
	err = CreateTable(db, "Properties", map[string]string{
		"Prop_id":       "INT AUTO_INCREMENT PRIMARY KEY",
		"Prop_location": "VARCHAR(255)",
	})
	if err != nil {
		return err
	}

	return nil
}

// API Handlers

// sendJSONResponse sends a JSON response
func sendJSONResponse(w http.ResponseWriter, status int, response APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}

// createDatabaseHandler handles POST /api/databases
func createDatabaseHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			sendJSONResponse(w, 405, APIResponse{Success: false, Error: "Method not allowed"})
			return
		}

		var req struct {
			Name string `json:"name"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sendJSONResponse(w, 400, APIResponse{Success: false, Error: "Invalid JSON"})
			return
		}

		if req.Name == "" {
			sendJSONResponse(w, 400, APIResponse{Success: false, Error: "Database name is required"})
			return
		}

		err := CreateDatabase(db, req.Name)
		if err != nil {
			sendJSONResponse(w, 500, APIResponse{Success: false, Error: err.Error()})
			return
		}

		sendJSONResponse(w, 200, APIResponse{
			Success: true,
			Message: fmt.Sprintf("Database '%s' created successfully", req.Name),
		})
	}
}

// useDatabaseHandler handles POST /api/databases/use
func useDatabaseHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			sendJSONResponse(w, 405, APIResponse{Success: false, Error: "Method not allowed"})
			return
		}

		var req struct {
			Name string `json:"name"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sendJSONResponse(w, 400, APIResponse{Success: false, Error: "Invalid JSON"})
			return
		}

		if req.Name == "" {
			sendJSONResponse(w, 400, APIResponse{Success: false, Error: "Database name is required"})
			return
		}

		err := UseDatabase(db, req.Name)
		if err != nil {
			sendJSONResponse(w, 500, APIResponse{Success: false, Error: err.Error()})
			return
		}

		sendJSONResponse(w, 200, APIResponse{
			Success: true,
			Message: fmt.Sprintf("Switched to database '%s'", req.Name),
		})
	}
}

// listDatabasesHandler handles GET /api/databases/list
func listDatabasesHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			sendJSONResponse(w, 405, APIResponse{Success: false, Error: "Method not allowed"})
			return
		}

		databases, err := ListDatabases(db)
		if err != nil {
			sendJSONResponse(w, 500, APIResponse{Success: false, Error: err.Error()})
			return
		}

		sendJSONResponse(w, 200, APIResponse{
			Success: true,
			Data:    databases,
		})
	}
}

// listTablesHandler handles GET /api/tables/list
func listTablesHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			sendJSONResponse(w, 405, APIResponse{Success: false, Error: "Method not allowed"})
			return
		}

		tables, err := ListTables(db)
		if err != nil {
			sendJSONResponse(w, 500, APIResponse{Success: false, Error: err.Error()})
			return
		}

		sendJSONResponse(w, 200, APIResponse{
			Success: true,
			Data:    tables,
		})
	}
}

// createTableHandler handles POST /api/tables
func createTableHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			sendJSONResponse(w, 405, APIResponse{Success: false, Error: "Method not allowed"})
			return
		}

		var req struct {
			Name    string            `json:"name"`
			Columns map[string]string `json:"columns"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sendJSONResponse(w, 400, APIResponse{Success: false, Error: "Invalid JSON"})
			return
		}

		if req.Name == "" {
			sendJSONResponse(w, 400, APIResponse{Success: false, Error: "Table name is required"})
			return
		}

		err := CreateTable(db, req.Name, req.Columns)
		if err != nil {
			sendJSONResponse(w, 500, APIResponse{Success: false, Error: err.Error()})
			return
		}

		sendJSONResponse(w, 200, APIResponse{
			Success: true,
			Message: fmt.Sprintf("Table '%s' created successfully", req.Name),
		})
	}
}

// addColumnHandler handles POST /api/tables/{tableName}/columns
func addColumnHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			sendJSONResponse(w, 405, APIResponse{Success: false, Error: "Method not allowed"})
			return
		}

		tableName := strings.TrimPrefix(r.URL.Path, "/api/tables/")
		tableName = strings.TrimSuffix(tableName, "/columns")

		var req struct {
			Name string `json:"name"`
			Type string `json:"type"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sendJSONResponse(w, 400, APIResponse{Success: false, Error: "Invalid JSON"})
			return
		}

		if req.Name == "" || req.Type == "" {
			sendJSONResponse(w, 400, APIResponse{Success: false, Error: "Column name and type are required"})
			return
		}

		err := AddColumn(db, tableName, req.Name, req.Type)
		if err != nil {
			sendJSONResponse(w, 500, APIResponse{Success: false, Error: err.Error()})
			return
		}

		sendJSONResponse(w, 200, APIResponse{
			Success: true,
			Message: fmt.Sprintf("Column '%s' added to table '%s'", req.Name, tableName),
		})
	}
}

// getTableSchemaHandler handles GET /api/tables/{tableName}/schema
func getTableSchemaHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			sendJSONResponse(w, 405, APIResponse{Success: false, Error: "Method not allowed"})
			return
		}

		tableName := strings.TrimPrefix(r.URL.Path, "/api/tables/")
		tableName = strings.TrimSuffix(tableName, "/schema")

		columnTypes, err := GetColumnTypes(db, tableName)
		if err != nil {
			sendJSONResponse(w, 500, APIResponse{Success: false, Error: err.Error()})
			return
		}

		var columns []ColumnInfo
		for name, typ := range columnTypes {
			columns = append(columns, ColumnInfo{Name: name, Type: typ})
		}

		sendJSONResponse(w, 200, APIResponse{
			Success: true,
			Data:    map[string]interface{}{"columns": columns},
		})
	}
}

// insertRecordHandler handles POST /api/tables/{tableName}/records
func insertRecordHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			sendJSONResponse(w, 405, APIResponse{Success: false, Error: "Method not allowed"})
			return
		}

		tableName := strings.TrimPrefix(r.URL.Path, "/api/tables/")
		tableName = strings.TrimSuffix(tableName, "/records")

		var data map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			sendJSONResponse(w, 400, APIResponse{Success: false, Error: "Invalid JSON"})
			return
		}

		err := Insert(db, tableName, data)
		if err != nil {
			sendJSONResponse(w, 500, APIResponse{Success: false, Error: err.Error()})
			return
		}

		sendJSONResponse(w, 200, APIResponse{
			Success: true,
			Message: fmt.Sprintf("Record inserted into '%s'", tableName),
		})
	}
}

// getRecordsHandler handles GET /api/tables/{tableName}/records
func getRecordsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			sendJSONResponse(w, 405, APIResponse{Success: false, Error: "Method not allowed"})
			return
		}

		tableName := strings.TrimPrefix(r.URL.Path, "/api/tables/")
		tableName = strings.TrimSuffix(tableName, "/records")

		records, err := Select(db, tableName)
		if err != nil {
			sendJSONResponse(w, 500, APIResponse{Success: false, Error: err.Error()})
			return
		}

		sendJSONResponse(w, 200, APIResponse{
			Success: true,
			Data:    map[string]interface{}{"records": records},
		})
	}
}

// updateRecordHandler handles PUT /api/tables/{tableName}/records
func updateRecordHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			sendJSONResponse(w, 405, APIResponse{Success: false, Error: "Method not allowed"})
			return
		}

		tableName := strings.TrimPrefix(r.URL.Path, "/api/tables/")
		tableName = strings.TrimSuffix(tableName, "/records")

		var req struct {
			Data      map[string]interface{} `json:"data"`
			Condition string                `json:"condition"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sendJSONResponse(w, 400, APIResponse{Success: false, Error: "Invalid JSON"})
			return
		}

		if req.Condition == "" {
			sendJSONResponse(w, 400, APIResponse{Success: false, Error: "Condition is required"})
			return
		}

		err := Update(db, tableName, req.Data, req.Condition)
		if err != nil {
			sendJSONResponse(w, 500, APIResponse{Success: false, Error: err.Error()})
			return
		}

		sendJSONResponse(w, 200, APIResponse{
			Success: true,
			Message: fmt.Sprintf("Records updated in '%s'", tableName),
		})
	}
}

// deleteRecordHandler handles DELETE /api/tables/{tableName}/records
func deleteRecordHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			sendJSONResponse(w, 405, APIResponse{Success: false, Error: "Method not allowed"})
			return
		}

		tableName := strings.TrimPrefix(r.URL.Path, "/api/tables/")
		tableName = strings.TrimSuffix(tableName, "/records")

		var req struct {
			Condition string `json:"condition"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sendJSONResponse(w, 400, APIResponse{Success: false, Error: "Invalid JSON"})
			return
		}

		if req.Condition == "" {
			sendJSONResponse(w, 400, APIResponse{Success: false, Error: "Condition is required"})
			return
		}

		err := Delete(db, tableName, req.Condition)
		if err != nil {
			sendJSONResponse(w, 500, APIResponse{Success: false, Error: err.Error()})
			return
		}

		sendJSONResponse(w, 200, APIResponse{
			Success: true,
			Message: fmt.Sprintf("Records deleted from '%s'", tableName),
		})
	}
}

// createSampleHandler handles POST /api/sample
func createSampleHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			sendJSONResponse(w, 405, APIResponse{Success: false, Error: "Method not allowed"})
			return
		}

		err := createSampleTables(db)
		if err != nil {
			sendJSONResponse(w, 500, APIResponse{Success: false, Error: err.Error()})
			return
		}

		sendJSONResponse(w, 200, APIResponse{
			Success: true,
			Message: "RealEstate database with sample tables created successfully",
		})
	}
}

func main() {
	fmt.Println("🚀 Starting Dynamic Database Management System...")
	db := dbConn()
	defer db.Close()

	// Serve static files
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Serve index.html for root
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.ServeFile(w, r, "static/index.html")
		} else {
			http.NotFound(w, r)
		}
	})

	// API routes
	http.HandleFunc("/api/databases", createDatabaseHandler(db))
	http.HandleFunc("/api/tables", createTableHandler(db))
	http.HandleFunc("/api/sample", createSampleHandler(db))

	// Dynamic routes for databases and tables (must come after specific routes)
	http.HandleFunc("/api/databases/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("🔍 Handling request: %s %s\n", r.Method, r.URL.Path)
		path := strings.TrimPrefix(r.URL.Path, "/api/databases/")
		fmt.Printf("🔍 Path after trim: %s\n", path)
		parts := strings.Split(path, "/")
		fmt.Printf("🔍 Path parts: %v\n", parts)

		if len(parts) < 1 {
			sendJSONResponse(w, 404, APIResponse{Success: false, Error: "Invalid database API endpoint"})
			return
		}

		dbName := parts[0]

		// Handle specific database routes
		if dbName == "use" && len(parts) == 1 {
			// This is /api/databases/use - delegate to useDatabaseHandler
			useDatabaseHandler(db)(w, r)
			return
		}

		// Switch to the specified database
		tempDB := dbConn()
		err := UseDatabase(tempDB, dbName)
		if err != nil {
			sendJSONResponse(w, 500, APIResponse{Success: false, Error: err.Error()})
			return
		}
		defer tempDB.Close()

		if len(parts) == 1 {
			// Database-level operations
			switch r.Method {
			case "GET":
				// List tables in database
				tables, err := ListTables(tempDB)
				if err != nil {
					sendJSONResponse(w, 500, APIResponse{Success: false, Error: err.Error()})
					return
				}
				sendJSONResponse(w, 200, APIResponse{
					Success: true,
					Data:    map[string]interface{}{"tables": tables},
				})
			default:
				sendJSONResponse(w, 405, APIResponse{Success: false, Error: "Method not allowed"})
			}
			return
		}

		if len(parts) < 3 || parts[1] != "tables" {
			sendJSONResponse(w, 404, APIResponse{Success: false, Error: "Invalid database table API endpoint"})
			return
		}

		tableName := parts[2]
		action := ""
		if len(parts) > 3 {
			action = parts[3]
		}

		switch action {
		case "columns":
			if r.Method == "POST" {
				addColumnHandler(tempDB)(w, r)
			} else {
				sendJSONResponse(w, 405, APIResponse{Success: false, Error: "Method not allowed"})
			}
		case "schema":
			if r.Method == "GET" {
				getTableSchemaHandler(tempDB)(w, r)
			} else {
				sendJSONResponse(w, 405, APIResponse{Success: false, Error: "Method not allowed"})
			}
		case "records":
			switch r.Method {
			case "GET":
				getRecordsHandler(tempDB)(w, r)
			case "POST":
				insertRecordHandler(tempDB)(w, r)
			case "PUT":
				updateRecordHandler(tempDB)(w, r)
			case "DELETE":
				deleteRecordHandler(tempDB)(w, r)
			default:
				sendJSONResponse(w, 405, APIResponse{Success: false, Error: "Method not allowed"})
			}
		default:
			sendJSONResponse(w, 404, APIResponse{Success: false, Error: "Invalid table action"})
		}
	})

	fmt.Println("🚀 Dynamic Database Management System API Server")
	fmt.Println("📡 Server running on http://localhost:8080")
	fmt.Println("🌐 Frontend available at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
