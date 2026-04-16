package repository

import (
	"database/sql"
	"fmt"
	"strings"

	"dbmanager/internal/models"

	mysqldriver "github.com/go-sql-driver/mysql"
)

// MySQLRepository handles all direct database interactions.
// It opens short-lived connections scoped to the target database.
type MySQLRepository struct {
	baseDSN func(dbName string) string
}

// NewMySQLRepository creates a repository. dsnFunc builds a DSN for a given db name.
func NewMySQLRepository(dsnFunc func(dbName string) string) *MySQLRepository {
	return &MySQLRepository{baseDSN: dsnFunc}
}

// open returns a *sql.DB connected to the given database (empty string = server-level).
func (r *MySQLRepository) open(dbName string) (*sql.DB, error) {
	db, err := sql.Open("mysql", r.baseDSN(dbName))
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}
	return db, nil
}

// ---------- Database-level operations ----------

// CreateDatabase creates a new MySQL database.
func (r *MySQLRepository) CreateDatabase(name string) error {
	db, err := r.open("")
	if err != nil {
		return err
	}
	defer db.Close()

	// Database names cannot be parameterized, but we validate the identifier
	if !isValidIdentifier(name) {
		return fmt.Errorf("invalid database name: %s", name)
	}
	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE `%s`", name))
	if err != nil {
		if isDuplicateError(err) {
			return fmt.Errorf("database '%s' already exists", name)
		}
		return err
	}
	return nil
}

// ListDatabases returns all database names.
func (r *MySQLRepository) ListDatabases() ([]string, error) {
	db, err := r.open("")
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("SHOW DATABASES")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		list = append(list, name)
	}
	return list, nil
}

// ---------- Table-level operations ----------

// ListTables returns all table names in the given database.
func (r *MySQLRepository) ListTables(dbName string) ([]string, error) {
	db, err := r.open(dbName)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("SHOW TABLES")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		list = append(list, name)
	}
	return list, nil
}

// CreateTable creates a table with the given columns inside dbName.
func (r *MySQLRepository) CreateTable(dbName, tableName string, columns []models.Column) error {
	if !isValidIdentifier(tableName) {
		return fmt.Errorf("invalid table name: %s", tableName)
	}
	if len(columns) == 0 {
		return fmt.Errorf("at least one column is required")
	}

	db, err := r.open(dbName)
	if err != nil {
		return err
	}
	defer db.Close()

	var parts []string
	for _, col := range columns {
		if !isValidIdentifier(col.Name) {
			return fmt.Errorf("invalid column name: %s", col.Name)
		}
		if !isValidType(col.Type) {
			return fmt.Errorf("invalid column type: %s", col.Type)
		}
		parts = append(parts, fmt.Sprintf("`%s` %s", col.Name, col.Type))
	}

	query := fmt.Sprintf("CREATE TABLE `%s` (%s)", tableName, strings.Join(parts, ", "))
	_, err = db.Exec(query)
	if err != nil {
		if isDuplicateError(err) {
			return fmt.Errorf("table '%s' already exists in database '%s'", tableName, dbName)
		}
		return err
	}
	return nil
}

// GetColumns returns ordered column info for a table.
func (r *MySQLRepository) GetColumns(dbName, tableName string) ([]models.Column, error) {
	if !isValidIdentifier(tableName) {
		return nil, fmt.Errorf("invalid table name: %s", tableName)
	}

	db, err := r.open(dbName)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(fmt.Sprintf("SHOW COLUMNS FROM `%s`", tableName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cols []models.Column
	for rows.Next() {
		var field, typ, null, key sql.NullString
		var defaultVal, extra sql.NullString
		if err := rows.Scan(&field, &typ, &null, &key, &defaultVal, &extra); err != nil {
			return nil, err
		}
		if field.Valid && typ.Valid {
			cols = append(cols, models.Column{Name: field.String, Type: typ.String})
		}
	}
	return cols, nil
}

// AddColumn adds a new column to an existing table.
func (r *MySQLRepository) AddColumn(dbName, tableName, colName, colType string) error {
	if !isValidIdentifier(tableName) || !isValidIdentifier(colName) {
		return fmt.Errorf("invalid identifier")
	}
	if !isValidType(colType) {
		return fmt.Errorf("invalid column type: %s", colType)
	}

	db, err := r.open(dbName)
	if err != nil {
		return err
	}
	defer db.Close()

	query := fmt.Sprintf("ALTER TABLE `%s` ADD COLUMN `%s` %s", tableName, colName, colType)
	_, err = db.Exec(query)
	return err
}

// ---------- Record CRUD ----------

// InsertRecord inserts a row using prepared statements.
func (r *MySQLRepository) InsertRecord(dbName, tableName string, data map[string]interface{}) error {
	if !isValidIdentifier(tableName) {
		return fmt.Errorf("invalid table name")
	}
	if len(data) == 0 {
		return fmt.Errorf("no data provided")
	}

	db, err := r.open(dbName)
	if err != nil {
		return err
	}
	defer db.Close()

	var colNames []string
	var placeholders []string
	var values []interface{}
	for col, val := range data {
		if !isValidIdentifier(col) {
			return fmt.Errorf("invalid column name: %s", col)
		}
		colNames = append(colNames, fmt.Sprintf("`%s`", col))
		placeholders = append(placeholders, "?")
		values = append(values, val)
	}

	query := fmt.Sprintf("INSERT INTO `%s` (%s) VALUES (%s)",
		tableName, strings.Join(colNames, ", "), strings.Join(placeholders, ", "))

	stmt, err := db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(values...)
	if err != nil {
		if isDuplicateError(err) {
			return fmt.Errorf("duplicate entry: a record with the same key already exists in '%s'", tableName)
		}
		return err
	}
	return nil
}

// SelectRecords retrieves all records from a table.
func (r *MySQLRepository) SelectRecords(dbName, tableName string) ([]models.Column, []models.Record, error) {
	if !isValidIdentifier(tableName) {
		return nil, nil, fmt.Errorf("invalid table name")
	}

	cols, err := r.GetColumns(dbName, tableName)
	if err != nil {
		return nil, nil, err
	}
	if len(cols) == 0 {
		return cols, []models.Record{}, nil
	}

	db, err := r.open(dbName)
	if err != nil {
		return nil, nil, err
	}
	defer db.Close()

	var quotedCols []string
	for _, c := range cols {
		quotedCols = append(quotedCols, fmt.Sprintf("`%s`", c.Name))
	}

	query := fmt.Sprintf("SELECT %s FROM `%s`", strings.Join(quotedCols, ", "), tableName)
	rows, err := db.Query(query)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var records []models.Record
	for rows.Next() {
		values := make([]interface{}, len(cols))
		scanArgs := make([]interface{}, len(cols))
		for i := range values {
			scanArgs[i] = &values[i]
		}
		if err := rows.Scan(scanArgs...); err != nil {
			return nil, nil, err
		}

		record := make(models.Record)
		for i, col := range cols {
			switch v := values[i].(type) {
			case []byte:
				record[col.Name] = string(v)
			case nil:
				record[col.Name] = nil
			default:
				record[col.Name] = v
			}
		}
		records = append(records, record)
	}

	return cols, records, nil
}

// UpdateRecord updates rows that match conditionCol = conditionVal.
func (r *MySQLRepository) UpdateRecord(dbName, tableName string, data map[string]interface{}, condCol string, condVal interface{}) error {
	if !isValidIdentifier(tableName) || !isValidIdentifier(condCol) {
		return fmt.Errorf("invalid identifier")
	}
	if len(data) == 0 {
		return fmt.Errorf("no data provided")
	}

	db, err := r.open(dbName)
	if err != nil {
		return err
	}
	defer db.Close()

	var setParts []string
	var values []interface{}
	for col, val := range data {
		if !isValidIdentifier(col) {
			return fmt.Errorf("invalid column name: %s", col)
		}
		setParts = append(setParts, fmt.Sprintf("`%s` = ?", col))
		values = append(values, val)
	}
	values = append(values, condVal)

	query := fmt.Sprintf("UPDATE `%s` SET %s WHERE `%s` = ?",
		tableName, strings.Join(setParts, ", "), condCol)

	stmt, err := db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	result, err := stmt.Exec(values...)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("record not found: no rows matched %s = %v in '%s'", condCol, condVal, tableName)
	}
	return nil
}

// DeleteRecord deletes rows that match conditionCol = conditionVal.
func (r *MySQLRepository) DeleteRecord(dbName, tableName, condCol string, condVal interface{}) error {
	if !isValidIdentifier(tableName) || !isValidIdentifier(condCol) {
		return fmt.Errorf("invalid identifier")
	}

	db, err := r.open(dbName)
	if err != nil {
		return err
	}
	defer db.Close()

	query := fmt.Sprintf("DELETE FROM `%s` WHERE `%s` = ?", tableName, condCol)

	stmt, err := db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	result, err := stmt.Exec(condVal)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("record not found: no rows matched %s = %v in '%s'", condCol, condVal, tableName)
	}
	return nil
}

// ---------- Helpers ----------

// isValidIdentifier checks that a MySQL identifier is safe (alphanumeric + underscore).
func isValidIdentifier(name string) bool {
	if name == "" || len(name) > 64 {
		return false
	}
	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_') {
			return false
		}
	}
	return true
}

// isValidType checks that a column type string is safe.
// Allows alphanumeric, spaces, parentheses, commas – covers types like VARCHAR(255), DECIMAL(10,2).
func isValidType(t string) bool {
	if t == "" || len(t) > 100 {
		return false
	}
	for _, c := range t {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') ||
			c == '(' || c == ')' || c == ',' || c == ' ' || c == '_') {
			return false
		}
	}
	return true
}

// isDuplicateError checks if a MySQL error is a duplicate entry or "already exists" error.
func isDuplicateError(err error) bool {
	if mysqlErr, ok := err.(*mysqldriver.MySQLError); ok {
		// 1062 = Duplicate entry, 1007 = Database already exists, 1050 = Table already exists
		return mysqlErr.Number == 1062 || mysqlErr.Number == 1007 || mysqlErr.Number == 1050
	}
	return false
}
