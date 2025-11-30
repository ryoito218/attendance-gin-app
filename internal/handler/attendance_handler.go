package handler

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const fixedUserID = 1

type AttendanceHandler struct {
	DB *sql.DB
}

func NewAttendanceHandler(db *sql.DB) *AttendanceHandler {
	return &AttendanceHandler{DB: db}
}

func (h *AttendanceHandler) ClockIn(c *gin.Context) {
	now := time.Now()
	workDate := now.Format("2006-01-02")

	var id int
	var clockIn, clockOut sql.NullTime

	err := h.DB.QueryRow(
		`SELECT id, clock_in, clock_out
		 FROM attendance
		 WHERE user_id = ? AND work_date = ?`,
		fixedUserID, workDate,
	).Scan(&id, &clockIn, &clockOut)

	switch {
	case err == sql.ErrNoRows:
		_, err := h.DB.Exec(
			`INSERT INTO attendance (user_id, work_date, clock_in)
			 VALUES (?, ?, ?)`,
			fixedUserID, workDate, now,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to insert clock_in"})
			return 
		}
	case err != nil:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	default:
		if clockIn.Valid {
			c.JSON(http.StatusBadRequest, gin.H{"error": "already clocked in"})
			return 
		}

		// なぜ存在
		_, err := h.DB.Exec(
			`UPDATE attendance
			 SET clock_in = ?
			 WHERE id = ?`,
			now, id,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update clock_in"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"work_date": workDate,
		"clock_in": now,
	})
}

func (h *AttendanceHandler) ClockOut(c *gin.Context) {
	now := time.Now()
	workDate := now.Format("2006-01-02")

	var id int
	var clockIn, clockOut sql.NullTime

	err := h.DB.QueryRow(
		`SELECT id, clock_in, clock_out
		 FROM attendance
		 WHERE user_id = ? AND work_date = ?`,
		fixedUserID, workDate,
	).Scan(&id, &clockIn, &clockOut)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no attendance record for today"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}

	if !clockIn.Valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "not clocked in yet"})
	}

	if clockOut.Valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "already clocked out"})
	}

	_, err = h.DB.Exec(
		`UPDATE attendance
		 SET clock_out = ?
		 WHERE id = ?`,
		now, id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update clock_out"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"work_date": workDate,
		"clock_out": now,
	})
}