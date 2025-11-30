package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/ryoito218/attendance-gin-app/internal/db"
	"github.com/ryoito218/attendance-gin-app/internal/handler"
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

	attendanceHandler := handler.NewAttendanceHandler(d)

	r.GET("/", func(c *gin.Context) {
		c.File("./public/index.html")
	})

	r.Static("/static", "./public")

	r.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	r.POST("/api/clock-in", attendanceHandler.ClockIn)
	r.POST("/api/clock-out", attendanceHandler.ClockOut)
	r.GET("/api/attendance", attendanceHandler.ListAttendance)

	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}

func setDefaultEnvIfEmpty(key, value string) {
	if os.Getenv(key) == "" {
		os.Setenv(key, value)
	}
}