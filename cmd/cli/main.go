package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"dbmanager/internal/config"
	"dbmanager/internal/models"
	"dbmanager/internal/repository"
	"dbmanager/internal/service"
)

// ═══════════════════════════════════════════════════════
//  Dynamic DB Manager — Interactive CLI
// ═══════════════════════════════════════════════════════

var (
	svc       *service.DBService
	currentDB string
	reader    *bufio.Reader
)

func main() {
	fmt.Println("╔══════════════════════════════════════════════╗")
	fmt.Println("║   ⚡ Dynamic DB Manager — CLI Interface      ║")
	fmt.Println("╚══════════════════════════════════════════════╝")
	fmt.Println()

	cfg := config.Load()
	repo := repository.NewMySQLRepository(cfg.DSN)

	// Wait for MySQL
	for i := 0; i < 30; i++ {
		dbs, err := repo.ListDatabases()
		if err == nil && len(dbs) > 0 {
			fmt.Println("✅ Connected to MySQL")
			break
		}
		if i == 29 {
			fmt.Println("❌ Could not connect to MySQL after 60 seconds")
			os.Exit(1)
		}
		fmt.Printf("⏳ Waiting for MySQL... (%d/30)\n", i+1)
		time.Sleep(2 * time.Second)
	}

	svc = service.NewDBService(repo)
	reader = bufio.NewReader(os.Stdin)

	fmt.Println()
	mainMenu()
}

func mainMenu() {
	for {
		clearScreen()
		fmt.Println("┌──────────────────────────────────────────────────────────────┐")
		fmt.Println("│                      ⚡ MAIN MENU                            │")
		fmt.Println("├──────────────────────────────────────────────────────────────┤")
		fmt.Printf("│  Selected DB: %-46s │\n", currentDBName())
		fmt.Println("├──────────────────────────────────────────────────────────────┤")
		fmt.Println("│  1. Database Operations (List, Create, Use)                  │")
		fmt.Println("│  2. Table Operations    (List, Create, Add Column)          │")
		fmt.Println("│  3. Record Operations   (Select, Insert, Update, Delete)    │")
		fmt.Println("│  4. Help / Command Info                                      │")
		fmt.Println("│  5. Exit                                                     │")
		fmt.Println("└──────────────────────────────────────────────────────────────┘")

		choice := prompt("Select an option (1-5)")

		switch choice {
		case "1":
			databaseMenu()
		case "2":
			tableMenu()
		case "3":
			recordMenu()
		case "4":
			printHelp()
			prompt("Press Enter to continue...")
		case "5", "exit", "quit":
			fmt.Println("Goodbye! 👋")
			os.Exit(0)
		default:
			printError("Invalid choice. Please enter a number between 1 and 5.")
			time.Sleep(1 * time.Second)
		}
	}
}

func databaseMenu() {
	for {
		clearScreen()
		fmt.Println("┌──────────────────────────────────────────────────────────────┐")
		fmt.Println("│                  📂 DATABASE OPERATIONS                      │")
		fmt.Println("├──────────────────────────────────────────────────────────────┤")
		fmt.Printf("│  Selected DB: %-46s │\n", currentDBName())
		fmt.Println("├──────────────────────────────────────────────────────────────┤")
		fmt.Println("│  1. List All Databases                                       │")
		fmt.Println("│  2. Create New Database                                      │")
		fmt.Println("│  3. Switch to Database (Use)                                 │")
		fmt.Println("│  4. Create Sample Database (RealEstate)                      │")
		fmt.Println("│  5. Back to Main Menu                                        │")
		fmt.Println("└──────────────────────────────────────────────────────────────┘")

		choice := prompt("Select an option (1-5)")

		switch choice {
		case "1":
			handleList([]string{"list", "db"})
			prompt("Press Enter to continue...")
		case "2":
			name := prompt("Enter database name")
			if name != "" {
				handleCreate([]string{"create", "db", name})
			}
			prompt("Press Enter to continue...")
		case "3":
			name := prompt("Enter database name to use")
			if name != "" {
				handleUse([]string{"use", name})
			}
			prompt("Press Enter to continue...")
		case "4":
			handleCreate([]string{"create", "sample"})
			prompt("Press Enter to continue...")
		case "5":
			return
		default:
			printError("Invalid choice.")
			time.Sleep(1 * time.Second)
		}
	}
}

func tableMenu() {
	for {
		clearScreen()
		fmt.Println("┌──────────────────────────────────────────────────────────────┐")
		fmt.Println("│                    🏗️  TABLE OPERATIONS                        │")
		fmt.Println("├──────────────────────────────────────────────────────────────┤")
		fmt.Printf("│  Selected DB: %-46s │\n", currentDBName())
		fmt.Println("├──────────────────────────────────────────────────────────────┤")
		fmt.Println("│  1. List Tables in Current DB                                │")
		fmt.Println("│  2. Create New Table (Interactive)                           │")
		fmt.Println("│  3. Add Column to Table                                      │")
		fmt.Println("│  4. Back to Main Menu                                        │")
		fmt.Println("└──────────────────────────────────────────────────────────────┘")

		choice := prompt("Select an option (1-4)")

		switch choice {
		case "1":
			handleList([]string{"list", "tables"})
			prompt("Press Enter to continue...")
		case "2":
			name := prompt("Enter new table name")
			if name != "" {
				handleCreate([]string{"create", "table", name})
			}
			prompt("Press Enter to continue...")
		case "3":
			table := prompt("Enter table name")
			if table == "" {
				continue
			}
			col := prompt("Enter column name")
			if col == "" {
				continue
			}
			typ := prompt("Enter column type (e.g., VARCHAR(255))")
			if typ == "" {
				continue
			}
			handleAdd([]string{"add", "column", table, col, typ})
			prompt("Press Enter to continue...")
		case "4":
			return
		default:
			printError("Invalid choice.")
			time.Sleep(1 * time.Second)
		}
	}
}

func recordMenu() {
	for {
		clearScreen()
		fmt.Println("┌──────────────────────────────────────────────────────────────┐")
		fmt.Println("│                    📄 RECORD OPERATIONS                      │")
		fmt.Println("├──────────────────────────────────────────────────────────────┤")
		fmt.Printf("│  Selected DB: %-46s │\n", currentDBName())
		fmt.Println("├──────────────────────────────────────────────────────────────┤")
		fmt.Println("│  1. View All Records (Select)                                │")
		fmt.Println("│  2. Insert New Record                                        │")
		fmt.Println("│  3. Update Existing Record                                   │")
		fmt.Println("│  4. Delete Record                                            │")
		fmt.Println("│  5. Back to Main Menu                                        │")
		fmt.Println("└──────────────────────────────────────────────────────────────┘")

		choice := prompt("Select an option (1-5)")

		switch choice {
		case "1":
			table := prompt("Enter table name to view")
			if table != "" {
				handleSelect([]string{"select", table})
			}
			prompt("Press Enter to continue...")
		case "2":
			table := prompt("Enter table name to insert into")
			if table != "" {
				handleInsert([]string{"insert", table})
			}
			prompt("Press Enter to continue...")
		case "3":
			table := prompt("Enter table name")
			if table == "" {
				continue
			}
			pkCol := prompt("Enter primary key column name")
			pkVal := prompt("Enter primary key value to update")
			if pkCol != "" && pkVal != "" {
				handleUpdate([]string{"update", table, pkCol, pkVal})
			}
			prompt("Press Enter to continue...")
		case "4":
			table := prompt("Enter table name")
			if table == "" {
				continue
			}
			pkCol := prompt("Enter primary key column name")
			pkVal := prompt("Enter primary key value to delete")
			if pkCol != "" && pkVal != "" {
				handleDelete([]string{"delete", table, pkCol, pkVal})
			}
			prompt("Press Enter to continue...")
		case "5":
			return
		default:
			printError("Invalid choice.")
			time.Sleep(1 * time.Second)
		}
	}
}

func currentDBName() string {
	if currentDB == "" {
		return "None (Please use 'Database Operations' to select one)"
	}
	return currentDB
}

func clearScreen() {
	// A simple way to "clear" terminal scrolling
	fmt.Print("\033[H\033[2J")
}

// ─── Help ──────────────────────────────────────────────

func printHelp() {
	fmt.Println("┌──────────────────────────────────────────────────────────────┐")
	fmt.Println("│                     Available Commands                       │")
	fmt.Println("├──────────────────────────────────────────────────────────────┤")
	fmt.Println("│  list db                 — List all databases               │")
	fmt.Println("│  list tables             — List tables in current database  │")
	fmt.Println("│  create db <name>        — Create a new database           │")
	fmt.Println("│  create table <name>     — Create a table (interactive)    │")
	fmt.Println("│  create sample           — Create sample RealEstate DB     │")
	fmt.Println("│  use <db_name>           — Switch to a database            │")
	fmt.Println("│  add column <table> <col> <type> — Add column to table     │")
	fmt.Println("│  select <table>          — View all records                │")
	fmt.Println("│  insert <table>          — Insert a record (interactive)   │")
	fmt.Println("│  update <table> <pk_col> <pk_val> — Update record          │")
	fmt.Println("│  delete <table> <pk_col> <pk_val> — Delete record          │")
	fmt.Println("│  help                    — Show this help                  │")
	fmt.Println("│  exit / quit             — Exit the CLI                    │")
	fmt.Println("└──────────────────────────────────────────────────────────────┘")
}

// ─── Output helpers ────────────────────────────────────

func printSuccess(msg string) {
	fmt.Printf("  ✅ %s\n", msg)
}

func printError(msg string) {
	fmt.Printf("  ❌ %s\n", msg)
}

func printInfo(msg string) {
	fmt.Printf("  ℹ️  %s\n", msg)
}

func prompt(label string) string {
	fmt.Printf("  %s: ", label)
	line, _ := reader.ReadString('\n')
	return strings.TrimSpace(line)
}

// ─── List ──────────────────────────────────────────────

func handleList(parts []string) {
	if len(parts) < 2 {
		printError("Usage: list db | list tables")
		return
	}

	switch strings.ToLower(parts[1]) {
	case "db", "databases":
		dbs, err := svc.ListDatabases()
		if err != nil {
			printError(err.Error())
			return
		}
		if len(dbs) == 0 {
			printInfo("No databases found.")
			return
		}
		fmt.Println()
		fmt.Println("  ┌─────────────────────────────────┐")
		fmt.Println("  │         Databases                │")
		fmt.Println("  ├─────────────────────────────────┤")
		for _, db := range dbs {
			marker := "  "
			if db == currentDB {
				marker = "▸ "
			}
			fmt.Printf("  │  %s%-29s│\n", marker, db)
		}
		fmt.Println("  └─────────────────────────────────┘")

	case "tables":
		if currentDB == "" {
			printError("No database selected. Use: use <db_name>")
			return
		}
		tables, err := svc.ListTables(currentDB)
		if err != nil {
			printError(err.Error())
			return
		}
		if len(tables) == 0 {
			printInfo("No tables in '" + currentDB + "'.")
			return
		}
		fmt.Println()
		fmt.Printf("  ┌─────────────────────────────────┐\n")
		fmt.Printf("  │  Tables in %-20s │\n", currentDB)
		fmt.Printf("  ├─────────────────────────────────┤\n")
		for _, t := range tables {
			fmt.Printf("  │  🏗️  %-27s│\n", t)
		}
		fmt.Println("  └─────────────────────────────────┘")

	default:
		printError("Usage: list db | list tables")
	}
}

// ─── Create ────────────────────────────────────────────

func handleCreate(parts []string) {
	if len(parts) < 2 {
		printError("Usage: create db <name> | create table <name> | create sample")
		return
	}

	switch strings.ToLower(parts[1]) {
	case "db", "database":
		if len(parts) < 3 {
			printError("Usage: create db <name>")
			return
		}
		name := parts[2]
		if err := svc.CreateDatabase(name); err != nil {
			printError(err.Error())
			return
		}
		printSuccess("Database '" + name + "' created.")

	case "table":
		if currentDB == "" {
			printError("No database selected. Use: use <db_name>")
			return
		}
		if len(parts) < 3 {
			printError("Usage: create table <name>")
			return
		}
		tableName := parts[2]
		fmt.Println()
		printInfo("Define columns for table '" + tableName + "'. Type 'done' when finished.")
		fmt.Println("  Available types: INT, VARCHAR(255), VARCHAR(100), BIGINT, FLOAT,")
		fmt.Println("                   DOUBLE, DECIMAL(10,2), TEXT, DATE, DATETIME, BOOLEAN")
		fmt.Println("                   INT AUTO_INCREMENT PRIMARY KEY")
		fmt.Println()

		var columns []models.Column
		for {
			colName := prompt("Column name (or 'done')")
			if strings.ToLower(colName) == "done" {
				break
			}
			if colName == "" {
				continue
			}
			colType := prompt("Column type")
			if colType == "" {
				printError("Type cannot be empty.")
				continue
			}
			columns = append(columns, models.Column{Name: colName, Type: colType})
			fmt.Printf("    ➕ Added: %s %s\n", colName, colType)
		}

		if len(columns) == 0 {
			printError("No columns defined. Table not created.")
			return
		}

		if err := svc.CreateTable(currentDB, tableName, columns); err != nil {
			printError(err.Error())
			return
		}
		printSuccess(fmt.Sprintf("Table '%s' created with %d column(s).", tableName, len(columns)))

	case "sample":
		if err := svc.CreateSampleDatabase(); err != nil {
			printError(err.Error())
			return
		}
		printSuccess("RealEstate sample database created with Campaign, Agent, Properties tables.")

	default:
		printError("Usage: create db <name> | create table <name> | create sample")
	}
}

// ─── Use ───────────────────────────────────────────────

func handleUse(parts []string) {
	if len(parts) < 2 {
		printError("Usage: use <db_name>")
		return
	}
	dbName := parts[1]

	// Verify it exists
	dbs, err := svc.ListDatabases()
	if err != nil {
		printError(err.Error())
		return
	}
	found := false
	for _, db := range dbs {
		if db == dbName {
			found = true
			break
		}
	}
	if !found {
		printError("Database '" + dbName + "' does not exist.")
		return
	}

	currentDB = dbName
	printSuccess("Switched to database '" + dbName + "'.")
}

// ─── Add Column ────────────────────────────────────────

func handleAdd(parts []string) {
	if len(parts) < 5 || strings.ToLower(parts[1]) != "column" {
		printError("Usage: add column <table> <col_name> <col_type>")
		return
	}
	if currentDB == "" {
		printError("No database selected. Use: use <db_name>")
		return
	}

	table := parts[2]
	colName := parts[3]
	colType := strings.Join(parts[4:], " ")

	if err := svc.AddColumn(currentDB, table, colName, colType); err != nil {
		printError(err.Error())
		return
	}
	printSuccess(fmt.Sprintf("Column '%s' (%s) added to '%s'.", colName, colType, table))
}

// ─── Select ────────────────────────────────────────────

func handleSelect(parts []string) {
	if len(parts) < 2 {
		printError("Usage: select <table>")
		return
	}
	if currentDB == "" {
		printError("No database selected. Use: use <db_name>")
		return
	}

	table := parts[1]
	cols, records, err := svc.SelectRecords(currentDB, table)
	if err != nil {
		printError(err.Error())
		return
	}

	if len(records) == 0 {
		printInfo("📭 No records in '" + table + "'.")
		return
	}

	// Calculate column widths
	widths := make([]int, len(cols))
	for i, col := range cols {
		widths[i] = len(col.Name)
	}
	for _, rec := range records {
		for i, col := range cols {
			val := fmt.Sprintf("%v", rec[col.Name])
			if rec[col.Name] == nil {
				val = "NULL"
			}
			if len(val) > widths[i] {
				widths[i] = len(val)
			}
		}
	}
	// Cap at 30 chars
	for i := range widths {
		if widths[i] > 30 {
			widths[i] = 30
		}
		if widths[i] < 4 {
			widths[i] = 4
		}
	}

	// Print header
	fmt.Println()
	fmt.Printf("  %d record(s) in '%s'\n\n", len(records), table)

	printSeparator(widths)
	fmt.Print("  │")
	for i, col := range cols {
		fmt.Printf(" %-*s │", widths[i], col.Name)
	}
	fmt.Println()
	printSeparator(widths)

	// Print rows
	for _, rec := range records {
		fmt.Print("  │")
		for i, col := range cols {
			val := "NULL"
			if rec[col.Name] != nil {
				val = fmt.Sprintf("%v", rec[col.Name])
			}
			if len(val) > widths[i] {
				val = val[:widths[i]-1] + "…"
			}
			fmt.Printf(" %-*s │", widths[i], val)
		}
		fmt.Println()
	}
	printSeparator(widths)
}

func printSeparator(widths []int) {
	fmt.Print("  ├")
	for i, w := range widths {
		fmt.Print(strings.Repeat("─", w+2))
		if i < len(widths)-1 {
			fmt.Print("┼")
		}
	}
	fmt.Println("┤")
}

// ─── Insert ────────────────────────────────────────────

func handleInsert(parts []string) {
	if len(parts) < 2 {
		printError("Usage: insert <table>")
		return
	}
	if currentDB == "" {
		printError("No database selected. Use: use <db_name>")
		return
	}

	table := parts[1]
	cols, err := svc.GetColumns(currentDB, table)
	if err != nil {
		printError(err.Error())
		return
	}
	if len(cols) == 0 {
		printError("Table '" + table + "' has no columns.")
		return
	}

	fmt.Println()
	printInfo("Enter values for each column (leave blank to skip):")
	fmt.Println()

	data := make(map[string]interface{})
	for _, col := range cols {
		val := prompt(fmt.Sprintf("%s (%s)", col.Name, col.Type))
		if val != "" {
			data[col.Name] = val
		}
	}

	if len(data) == 0 {
		printError("No values provided. Record not inserted.")
		return
	}

	if err := svc.InsertRecord(currentDB, table, data); err != nil {
		printError(err.Error())
		return
	}
	printSuccess("Record inserted into '" + table + "'.")
}

// ─── Update ────────────────────────────────────────────

func handleUpdate(parts []string) {
	if len(parts) < 4 {
		printError("Usage: update <table> <pk_col> <pk_val>")
		return
	}
	if currentDB == "" {
		printError("No database selected. Use: use <db_name>")
		return
	}

	table := parts[1]
	pkCol := parts[2]
	pkVal := parts[3]

	// Get the schema to prompt for values
	cols, err := svc.GetColumns(currentDB, table)
	if err != nil {
		printError(err.Error())
		return
	}

	fmt.Println()
	printInfo(fmt.Sprintf("Updating record where %s = %s", pkCol, pkVal))
	printInfo("Enter new values (leave blank to keep current):")
	fmt.Println()

	data := make(map[string]interface{})
	for _, col := range cols {
		if col.Name == pkCol {
			continue // skip primary key
		}
		val := prompt(fmt.Sprintf("%s (%s)", col.Name, col.Type))
		if val != "" {
			data[col.Name] = val
		}
	}

	if len(data) == 0 {
		printError("No changes made.")
		return
	}

	if err := svc.UpdateRecord(currentDB, table, data, pkCol, pkVal); err != nil {
		printError(err.Error())
		return
	}
	printSuccess("Record updated successfully.")
}

// ─── Delete ────────────────────────────────────────────

func handleDelete(parts []string) {
	if len(parts) < 4 {
		printError("Usage: delete <table> <pk_col> <pk_val>")
		return
	}
	if currentDB == "" {
		printError("No database selected. Use: use <db_name>")
		return
	}

	table := parts[1]
	pkCol := parts[2]
	pkVal := parts[3]

	// Confirm
	confirm := prompt(fmt.Sprintf("Delete record where %s = %s? (yes/no)", pkCol, pkVal))
	if strings.ToLower(confirm) != "yes" && strings.ToLower(confirm) != "y" {
		printInfo("Cancelled.")
		return
	}

	if err := svc.DeleteRecord(currentDB, table, pkCol, pkVal); err != nil {
		printError(err.Error())
		return
	}
	printSuccess("Record deleted.")
}

// ─── Argument parser ──────────────────────────────────

func splitArgs(line string) []string {
	var args []string
	var current strings.Builder
	inQuote := false

	for _, c := range line {
		switch {
		case c == '"':
			inQuote = !inQuote
		case c == ' ' && !inQuote:
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(c)
		}
	}
	if current.Len() > 0 {
		args = append(args, current.String())
	}
	return args
}
