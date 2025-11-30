package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/ryoito218/attendance-gin-app/internal/db"
)

func main() {
	setDefaultEnvIfEmpty("DB_HOST", "127.0.0.1")
	setDefaultEnvIfEmpty("DB_USER", "appuser")
	setDefaultEnvIfEmpty("DB_PASSWORD", "apppass")
	setDefaultEnvIfEmpty("DB_NAME", "attendance_db")

	d, err := db.NewMySQL()
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer d.Close()

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}

func setDefaultEnvIfEmpty(key, value string) {
	if os.Getenv(key) == "" {
		os.Setenv(key, value)
	}
}