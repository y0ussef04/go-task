package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"dbmanager/internal/config"
	"dbmanager/internal/handler"
	"dbmanager/internal/repository"
	"dbmanager/internal/service"
)

func main() {
	fmt.Println("🚀 Starting Dynamic Database Management System...")

	cfg := config.Load()

	// Wait for MySQL to be reachable (Docker healthcheck handles most of this,
	// but an extra retry loop makes startup more robust).
	repo := repository.NewMySQLRepository(cfg.DSN)
	waitForDB(repo)

	svc := service.NewDBService(repo)
	h := handler.NewHandler(svc)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	addr := ":" + cfg.AppPort
	fmt.Printf("📡 Server running on http://localhost%s\n", addr)
	fmt.Printf("🌐 Open your browser at http://localhost%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, handler.LoggingMiddleware(mux)))
}

// waitForDB retries until MySQL is reachable.
func waitForDB(repo *repository.MySQLRepository) {
	for i := 0; i < 30; i++ {
		dbs, err := repo.ListDatabases()
		if err == nil && len(dbs) > 0 {
			fmt.Println("✅ Connected to MySQL")
			return
		}
		fmt.Printf("⏳ Waiting for MySQL... (%d/30)\n", i+1)
		time.Sleep(2 * time.Second)
	}
	log.Fatal("❌ Could not connect to MySQL after 60 seconds")
}
