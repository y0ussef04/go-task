package service

import (
	"dbmanager/internal/models"
	"dbmanager/internal/repository"
	"fmt"
	"strings"
)

// DBService is the business logic layer between handlers and the repository.
type DBService struct {
	repo *repository.MySQLRepository
}

// NewDBService creates a new service instance.
func NewDBService(repo *repository.MySQLRepository) *DBService {
	return &DBService{repo: repo}
}

// ---------- Database operations ----------

func (s *DBService) CreateDatabase(name string) error {
	if name == "" {
		return fmt.Errorf("database name is required")
	}
	return s.repo.CreateDatabase(name)
}

func (s *DBService) ListDatabases() ([]string, error) {
	return s.repo.ListDatabases()
}

// ---------- Table operations ----------

func (s *DBService) ListTables(dbName string) ([]string, error) {
	if dbName == "" {
		return nil, fmt.Errorf("database name is required")
	}
	return s.repo.ListTables(dbName)
}

func (s *DBService) CreateTable(dbName, tableName string, columns []models.Column) error {
	if dbName == "" {
		return fmt.Errorf("database name is required")
	}
	if tableName == "" {
		return fmt.Errorf("table name is required")
	}
	if len(columns) == 0 {
		return fmt.Errorf("at least one column is required")
	}
	return s.repo.CreateTable(dbName, tableName, columns)
}

func (s *DBService) GetColumns(dbName, tableName string) ([]models.Column, error) {
	if dbName == "" || tableName == "" {
		return nil, fmt.Errorf("database and table names are required")
	}
	return s.repo.GetColumns(dbName, tableName)
}

func (s *DBService) AddColumn(dbName, tableName, colName, colType string) error {
	if dbName == "" || tableName == "" || colName == "" || colType == "" {
		return fmt.Errorf("all fields are required")
	}
	return s.repo.AddColumn(dbName, tableName, colName, colType)
}

// ---------- Record CRUD ----------

func (s *DBService) InsertRecord(dbName, tableName string, data map[string]interface{}) error {
	if dbName == "" || tableName == "" {
		return fmt.Errorf("database and table names are required")
	}
	return s.repo.InsertRecord(dbName, tableName, data)
}

func (s *DBService) SelectRecords(dbName, tableName string) ([]models.Column, []models.Record, error) {
	if dbName == "" || tableName == "" {
		return nil, nil, fmt.Errorf("database and table names are required")
	}
	return s.repo.SelectRecords(dbName, tableName)
}

func (s *DBService) UpdateRecord(dbName, tableName string, data map[string]interface{}, condCol string, condVal interface{}) error {
	if dbName == "" || tableName == "" || condCol == "" {
		return fmt.Errorf("database, table, and condition column are required")
	}
	return s.repo.UpdateRecord(dbName, tableName, data, condCol, condVal)
}

func (s *DBService) DeleteRecord(dbName, tableName, condCol string, condVal interface{}) error {
	if dbName == "" || tableName == "" || condCol == "" {
		return fmt.Errorf("database, table, and condition column are required")
	}
	return s.repo.DeleteRecord(dbName, tableName, condCol, condVal)
}

// ---------- Sample data ----------

func (s *DBService) CreateSampleDatabase() error {
	if err := s.CreateDatabase("RealEstate"); err != nil {
		// Ignore "already exists" errors for sample data
		if !strings.Contains(err.Error(), "already exists") {
			return err
		}
	}

	columns := []struct {
		table string
		cols  []models.Column
	}{
		{
			table: "Campaign",
			cols: []models.Column{
				{Name: "camp_id", Type: "INT AUTO_INCREMENT PRIMARY KEY"},
				{Name: "Camp_name", Type: "VARCHAR(255)"},
				{Name: "Cost", Type: "FLOAT"},
			},
		},
		{
			table: "Agent",
			cols: []models.Column{
				{Name: "Agent_id", Type: "INT AUTO_INCREMENT PRIMARY KEY"},
				{Name: "Agent_name", Type: "VARCHAR(255)"},
				{Name: "Agent_salary", Type: "FLOAT"},
				{Name: "Agent_Address", Type: "VARCHAR(255)"},
			},
		},
		{
			table: "Properties",
			cols: []models.Column{
				{Name: "Prop_id", Type: "INT AUTO_INCREMENT PRIMARY KEY"},
				{Name: "Prop_location", Type: "VARCHAR(255)"},
			},
		},
	}

	for _, t := range columns {
		if err := s.repo.CreateTable("RealEstate", t.table, t.cols); err != nil {
			// Ignore "already exists" errors for sample data
			if !strings.Contains(err.Error(), "already exists") {
				return fmt.Errorf("create table %s: %w", t.table, err)
			}
		}
	}
	return nil
}
