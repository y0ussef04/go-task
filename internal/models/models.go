package models

// Column represents a table column with its name and MySQL type
type Column struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// Record is a dynamic row: column name -> value
type Record map[string]interface{}

// APIResponse is the standard JSON envelope for all API responses
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Error   string      `json:"error,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}
