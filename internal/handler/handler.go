package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"dbmanager/internal/models"
	"dbmanager/internal/service"
)

// Handler holds the HTTP handlers and a reference to the service layer.
type Handler struct {
	svc *service.DBService
}

// NewHandler creates a new Handler.
func NewHandler(svc *service.DBService) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes wires all HTTP routes to the given mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	// Static files
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Root -> serve SPA
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.ServeFile(w, r, "static/index.html")
			return
		}
		http.NotFound(w, r)
	})

	// API endpoints
	mux.HandleFunc("/api/databases", h.handleDatabases)
	mux.HandleFunc("/api/databases/", h.handleDatabaseSub)
	mux.HandleFunc("/api/sample", h.handleSample)
}

// ---------- JSON helpers ----------

func jsonOK(w http.ResponseWriter, data interface{}, msg string) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.APIResponse{Success: true, Data: data, Message: msg})
}

func jsonErr(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(models.APIResponse{Success: false, Error: msg})
}

// ---------- /api/databases ----------

func (h *Handler) handleDatabases(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		databases, err := h.svc.ListDatabases()
		if err != nil {
			jsonErr(w, 500, err.Error())
			return
		}
		jsonOK(w, databases, "")

	case http.MethodPost:
		var req struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonErr(w, 400, "Invalid JSON")
			return
		}
		if err := h.svc.CreateDatabase(req.Name); err != nil {
			jsonErr(w, 500, err.Error())
			return
		}
		jsonOK(w, nil, fmt.Sprintf("Database '%s' created", req.Name))

	default:
		jsonErr(w, 405, "Method not allowed")
	}
}

// ---------- /api/databases/{db}/... ----------

// handleDatabaseSub is the main router for all database-scoped operations.
// Routes:
//
//	GET  /api/databases/{db}/tables                -> list tables
//	POST /api/databases/{db}/tables                -> create table
//	GET  /api/databases/{db}/tables/{t}/columns    -> get columns / schema
//	POST /api/databases/{db}/tables/{t}/columns    -> add column
//	GET  /api/databases/{db}/tables/{t}/records    -> select records
//	POST /api/databases/{db}/tables/{t}/records    -> insert record
//	PUT  /api/databases/{db}/tables/{t}/records    -> update record
//	DELETE /api/databases/{db}/tables/{t}/records  -> delete record
func (h *Handler) handleDatabaseSub(w http.ResponseWriter, r *http.Request) {
	// Parse: /api/databases/{db}[/tables[/{table}[/columns|records]]]
	path := strings.TrimPrefix(r.URL.Path, "/api/databases/")
	parts := strings.Split(strings.TrimRight(path, "/"), "/")

	if len(parts) < 1 || parts[0] == "" {
		jsonErr(w, 400, "Database name is required")
		return
	}

	dbName := parts[0]

	// /api/databases/{db}
	if len(parts) == 1 {
		// GET -> list tables (convenience)
		if r.Method == http.MethodGet {
			tables, err := h.svc.ListTables(dbName)
			if err != nil {
				jsonErr(w, 500, err.Error())
				return
			}
			jsonOK(w, map[string]interface{}{"tables": tables}, "")
			return
		}
		jsonErr(w, 405, "Method not allowed")
		return
	}

	// /api/databases/{db}/tables[/...]
	if parts[1] != "tables" {
		jsonErr(w, 404, "Unknown endpoint")
		return
	}

	// /api/databases/{db}/tables
	if len(parts) == 2 {
		switch r.Method {
		case http.MethodGet:
			tables, err := h.svc.ListTables(dbName)
			if err != nil {
				jsonErr(w, 500, err.Error())
				return
			}
			jsonOK(w, map[string]interface{}{"tables": tables}, "")

		case http.MethodPost:
			var req struct {
				Name    string            `json:"name"`
				Columns map[string]string `json:"columns"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				jsonErr(w, 400, "Invalid JSON")
				return
			}
			var cols []models.Column
			for name, typ := range req.Columns {
				cols = append(cols, models.Column{Name: name, Type: typ})
			}
			if err := h.svc.CreateTable(dbName, req.Name, cols); err != nil {
				jsonErr(w, 500, err.Error())
				return
			}
			jsonOK(w, nil, fmt.Sprintf("Table '%s' created", req.Name))

		default:
			jsonErr(w, 405, "Method not allowed")
		}
		return
	}

	tableName := parts[2]

	// /api/databases/{db}/tables/{table}
	if len(parts) == 3 {
		if r.Method == http.MethodGet {
			cols, err := h.svc.GetColumns(dbName, tableName)
			if err != nil {
				jsonErr(w, 500, err.Error())
				return
			}
			jsonOK(w, map[string]interface{}{"columns": cols}, "")
			return
		}
		jsonErr(w, 405, "Method not allowed")
		return
	}

	action := parts[3]

	switch action {
	case "columns":
		h.handleColumns(w, r, dbName, tableName)
	case "records":
		h.handleRecords(w, r, dbName, tableName)
	case "schema":
		// Alias for GET columns
		if r.Method == http.MethodGet {
			cols, err := h.svc.GetColumns(dbName, tableName)
			if err != nil {
				jsonErr(w, 500, err.Error())
				return
			}
			jsonOK(w, map[string]interface{}{"columns": cols}, "")
			return
		}
		jsonErr(w, 405, "Method not allowed")
	default:
		jsonErr(w, 404, "Unknown action: "+action)
	}
}

// handleColumns handles column operations on a specific table.
func (h *Handler) handleColumns(w http.ResponseWriter, r *http.Request, dbName, tableName string) {
	switch r.Method {
	case http.MethodGet:
		cols, err := h.svc.GetColumns(dbName, tableName)
		if err != nil {
			jsonErr(w, 500, err.Error())
			return
		}
		jsonOK(w, map[string]interface{}{"columns": cols}, "")

	case http.MethodPost:
		var req struct {
			Name string `json:"name"`
			Type string `json:"type"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonErr(w, 400, "Invalid JSON")
			return
		}
		if err := h.svc.AddColumn(dbName, tableName, req.Name, req.Type); err != nil {
			jsonErr(w, 500, err.Error())
			return
		}
		jsonOK(w, nil, fmt.Sprintf("Column '%s' added to '%s'", req.Name, tableName))

	default:
		jsonErr(w, 405, "Method not allowed")
	}
}

// handleRecords handles CRUD operations on records.
func (h *Handler) handleRecords(w http.ResponseWriter, r *http.Request, dbName, tableName string) {
	switch r.Method {
	case http.MethodGet:
		cols, records, err := h.svc.SelectRecords(dbName, tableName)
		if err != nil {
			jsonErr(w, 500, err.Error())
			return
		}
		jsonOK(w, map[string]interface{}{"columns": cols, "records": records}, "")

	case http.MethodPost:
		var data map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			jsonErr(w, 400, "Invalid JSON")
			return
		}
		if err := h.svc.InsertRecord(dbName, tableName, data); err != nil {
			jsonErr(w, 500, err.Error())
			return
		}
		jsonOK(w, nil, "Record inserted")

	case http.MethodPut:
		var req struct {
			Data          map[string]interface{} `json:"data"`
			ConditionCol  string                 `json:"condition_col"`
			ConditionVal  interface{}            `json:"condition_val"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonErr(w, 400, "Invalid JSON")
			return
		}
		if err := h.svc.UpdateRecord(dbName, tableName, req.Data, req.ConditionCol, req.ConditionVal); err != nil {
			jsonErr(w, 500, err.Error())
			return
		}
		jsonOK(w, nil, "Record updated")

	case http.MethodDelete:
		var req struct {
			ConditionCol string      `json:"condition_col"`
			ConditionVal interface{} `json:"condition_val"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonErr(w, 400, "Invalid JSON")
			return
		}
		if err := h.svc.DeleteRecord(dbName, tableName, req.ConditionCol, req.ConditionVal); err != nil {
			jsonErr(w, 500, err.Error())
			return
		}
		jsonOK(w, nil, "Record deleted")

	default:
		jsonErr(w, 405, "Method not allowed")
	}
}

// handleSample creates the sample RealEstate database.
func (h *Handler) handleSample(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonErr(w, 405, "Method not allowed")
		return
	}
	if err := h.svc.CreateSampleDatabase(); err != nil {
		jsonErr(w, 500, err.Error())
		return
	}
	jsonOK(w, nil, "RealEstate sample database created")
}

// LoggingMiddleware logs each request.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("→ %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
