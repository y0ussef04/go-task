# Dynamic Database Management System in Go

A comprehensive database management system built in Go that connects to MySQL, creates databases and tables programmatically, and performs full CRUD operations with a command-line interface.

## Features

- **MySQL Integration**: Connects to MySQL server using Go's database/sql package
- **Database Creation**: Programmatically creates databases
- **Dynamic Table Creation**: Flexible `CreateTable` function that can create tables with 2, 3, or n columns
- **Full CRUD Operations**: Create, Read, Update, Delete operations for any table
- **Dynamic Column Addition**: Add new columns to existing tables
- **CLI Interface**: Interactive command-line interface for all operations
- **RealEstate Use Case**: Includes sample database creation with Campaign, Agent, and Properties tables

## Prerequisites

- Docker and Docker Compose
- Go 1.22+ (for local development)

## Quick Start with Docker

1. Clone or navigate to the project directory
2. Run the application:

```bash
docker-compose up --build
```

This will:
- Start a MySQL 8 container
- Build the Go application
- Run the CLI application

## Accessing the CLI

The application provides an interactive menu:

```
Menu:
1. Create Database
2. Use Database
3. Create Table
4. Add Column to Table
5. Insert Record
6. View Records
7. Update Record
8. Delete Record
9. Create Sample RealEstate Database
0. Exit
```

Choose options to perform various database operations.

## Sample Usage

1. **Create Sample Database**: Choose option 9 to create the RealEstate database with sample tables
2. **Create Custom Table**: Choose option 3, enter table name and column details
3. **Insert Data**: Choose option 5, enter table name and provide values for each column
4. **View Data**: Choose option 6, enter table name to see all records

## Project Structure

```
.
├── main.go              # Main application code with CLI
├── Dockerfile           # Docker build configuration
├── docker-compose.yml   # Multi-container setup
└── README.md           # This file
```

## Core Functions

### Database Operations
- `CreateDatabase(dbName string)` - Creates a new database
- `UseDatabase(dbName string)` - Switches to a database

### Table Operations
- `CreateTable(tableName string, columns map[string]string)` - Creates table with specified columns
- `AddColumn(tableName string, columnName string, dataType string)` - Adds column to existing table

### CRUD Operations
- `Insert(tableName string, data map[string]interface{})` - Inserts new record
- `Select(tableName string)` - Displays all records from table
- `Update(tableName string, data map[string]interface{}, condition string)` - Updates records
- `Delete(tableName string, condition string)` - Deletes records

## Database Schema (Sample)

### Campaign Table
- `camp_id` (INT, AUTO_INCREMENT, PRIMARY KEY)
- `Camp_name` (VARCHAR(255))
- `Cost` (FLOAT)

### Agent Table
- `Agent_id` (INT, AUTO_INCREMENT, PRIMARY KEY)
- `Agent_name` (VARCHAR(255))
- `Agent_salary` (FLOAT)
- `Agent_Address` (VARCHAR(255))

### Properties Table
- `Prop_id` (INT, AUTO_INCREMENT, PRIMARY KEY)
- `Prop_location` (VARCHAR(255))

## Local Development

If you prefer to run locally:

1. Install Go and MySQL
2. Start MySQL server
3. Update connection string in `main.go` if needed
4. Run:

```bash
go mod tidy
go run main.go
```

## Dependencies

- `github.com/go-sql-driver/mysql` - MySQL driver for Go

## Docker Configuration

- **MySQL**: Version 8, root password: `secret`, port 3306
- **Go App**: Version 1.22-alpine
- **Health Checks**: MySQL health check ensures database is ready

## Extending the System

The system is designed to be fully dynamic:
- Table names and column names are user-provided
- Data types are specified at runtime
- SQL queries are built dynamically
- All operations work with any valid table structure