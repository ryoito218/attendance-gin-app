package handler

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ryoito218/attendance-gin-app/internal/domain"
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
		"status":    "ok",
		"work_date": workDate,
		"clock_in":  now,
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
		return
	}

	if clockOut.Valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "already clocked out"})
		return
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
		"status":    "ok",
		"work_date": workDate,
		"clock_out": now,
	})
}

type AttendanceRecordResponse struct {
	WorkDate            string  `json:"work_date"`
	ClockIn             *string `json:"clock_in,omitempty"`  // なぜポインタ型？
	ClockOut            *string `json:"clock_out,omitempty"` // なぜポインタ型？
	WorkDurationMinutes int     `json:"work_duration_minutes"`
}

func (h *AttendanceHandler) ListAttendance(c *gin.Context) {
	daysStr := c.DefaultQuery("days", "7")

	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid days"})
		return
	}

	now := time.Now()
	startDate := now.AddDate(0, 0, -(days - 1)).Format("2006-01-02")
	endDate := now.Format("2006-01-02")

	rows, err := h.DB.Query(
		`SELECT work_date, clock_in, clock_out
		 FROM attendance
		 WHERE user_id = ?
		   AND work_date BETWEEN ? AND ?
		 ORDER BY work_date DESC
		`,
		fixedUserID, startDate, endDate,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}
	defer rows.Close()

	var result []AttendanceRecordResponse

	for rows.Next() {
		var workDate time.Time
		var clockIn, clockOut sql.NullTime

		if err := rows.Scan(&workDate, &clockIn, &clockOut); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to scan"})
			return
		}

		var clockInStr *string
		var clockOutStr *string
		var durationMinutes int

		if clockIn.Valid && clockOut.Valid {
			d := domain.CalcWorkDuration(clockIn.Time, clockOut.Time)
			durationMinutes = int(d.Minutes())

			s := clockIn.Time.Format(time.RFC3339)
			clockInStr = &s
			e := clockOut.Time.Format(time.RFC3339)
			clockOutStr = &e
		} else {
			if clockIn.Valid {
				s := clockIn.Time.Format(time.RFC3339)
				clockInStr = &s
			}
			if clockOut.Valid {
				e := clockOut.Time.Format(time.RFC3339)
				clockOutStr = &e
			}
			durationMinutes = 0

		}

		result = append(result, AttendanceRecordResponse{
			WorkDate:            workDate.Format("2006-01-02"),
			ClockIn:             clockInStr,
			ClockOut:            clockOutStr,
			WorkDurationMinutes: durationMinutes,
		})
	}

	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "rows error"})
		return
	}

	c.JSON(http.StatusOK, result)
}
