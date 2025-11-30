package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ryoito218/attendance-gin-app/internal/db"
)

func newTestDB(t *testing.T) *sql.DB {
	t.Helper()

	setEnvIfEmpty("DB_HOST", "127.0.0.1")
	setEnvIfEmpty("DB_USER", "appuser")
	setEnvIfEmpty("DB_PASSWORD", "apppass")
	setEnvIfEmpty("DB_NAME", "attendance_db")

	d, err := db.NewMySQL()
	if err != nil {
		t.Fatalf("failed to connect test db: %v", err)
	}

	if _, err := d.Exec("DELETE FROM attendance"); err != nil {
		t.Fatalf("failed to clean attendance table: %v", err)
	}

	return d
}

func setEnvIfEmpty(key, value string) {
	if os.Getenv(key) == "" {
		os.Setenv(key, value)
	}
}

func TestClockIn_FirstTime_Success(t *testing.T) {
	d := newTestDB(t)
	defer d.Close()

	gin.SetMode(gin.TestMode)
	r := gin.Default()

	h := NewAttendanceHandler(d)
	r.POST("/api/clock-in", h.ClockIn)

	req := httptest.NewRequest(http.MethodPost, "/api/clock-in", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status code want %d, got %d", http.StatusOK, w.Code)
	}

	var cnt int
	err := d.QueryRow("SELECT COUNT(*) FROM attendance").Scan(&cnt)
	if err != nil {
		t.Fatalf("failed to count attendance: %v", err)
	}
	if cnt != 1 {
		t.Fatalf("attendance count want 1, got %d", cnt)
	}
}

func TestClockOut_AfterClockIn_Success(t *testing.T) {
	d := newTestDB(t)
	defer d.Close()

	workDate := time.Now().Format("2006-01-02")
	clockIn := time.Now().Add(-9 * time.Hour)

	_, err := d.Exec(
		`INSERT INTO attendance (user_id, work_date, clock_in)
		 VALUES (?, ?, ?)`,
		fixedUserID, workDate, clockIn,
	)
	if err != nil {
		t.Fatalf("failed to insert test clock-in: %v", err)
	}

	gin.SetMode(gin.TestMode)
	r := gin.Default()

	h := NewAttendanceHandler(d)
	r.POST("/api/clock-out", h.ClockOut)

	req := httptest.NewRequest(http.MethodPost, "/api/clock-out", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status code want %d, got %d", http.StatusOK, w.Code)
	}

	var cnt int
	err = d.QueryRow("SELECT COUNT(*) FROM attendance WHERE clock_out IS NOT NULL").Scan(&cnt)
	if err != nil {
		t.Fatalf("failed to check clock_out: %v", err)
	}
	if cnt != 1 {
		t.Fatalf("want clock_out set count = 1, got %d", cnt)
	}
}

func TestClockOut_WithoutClockIn_ReturnsBadRequest(t *testing.T) {
	d := newTestDB(t)
	defer d.Close()

	gin.SetMode(gin.TestMode)
	r := gin.Default()

	h := NewAttendanceHandler(d)
	r.POST("/api/clock-out", h.ClockOut)

	req := httptest.NewRequest(http.MethodPost, "/api/clock-out", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status code want %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestClockOut_Double_ReturnsBadRequest(t *testing.T) {
	d := newTestDB(t)
	defer d.Close()

	workDate := time.Now().Format("2006-01-02")
	clockIn := time.Now().Add(-9 * time.Hour)
	clockOut := time.Now()

	// 事前に退勤済みレコードを挿入
	_, err := d.Exec(
		`INSERT INTO attendance (user_id, work_date, clock_in, clock_out)
		 VALUES (?, ?, ?, ?)`,
		fixedUserID, workDate, clockIn, clockOut,
	)
	if err != nil {
		t.Fatalf("failed to insert test record: %v", err)
	}

	gin.SetMode(gin.TestMode)
	r := gin.Default()

	h := NewAttendanceHandler(d)
	r.POST("/api/clock-out", h.ClockOut)

	req := httptest.NewRequest(http.MethodPost, "/api/clock-out", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status code want %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestListAttendance_InvalidDays_ReturnBadRequest(t *testing.T) {
	d := newTestDB(t)
	defer d.Close()

	gin.SetMode(gin.TestMode)
	r := gin.Default()

	h := NewAttendanceHandler(d)
	r.GET("/api/attendance", h.ListAttendance)

	req := httptest.NewRequest(http.MethodGet, "/api/attendance?days=-1", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status code want %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestListAttendance_Normal(t *testing.T) {
	d := newTestDB(t)
	defer d.Close()

	today := time.Now().Format("2006-01-02")
	clockIn := time.Now().Add(-9 * time.Hour)
	clockOut := time.Now()

	_, err := d.Exec(
		`INSERT INTO attendance (user_id, work_date, clock_in, clock_out)
		 VALUES (?, ?, ?, ?)`,
		fixedUserID, today, clockIn, clockOut,
	)
	if err != nil {
		t.Fatalf("failed to insert attendance: %v", err)
	}

	gin.SetMode(gin.TestMode)
	r := gin.Default()

	h := NewAttendanceHandler(d)
	r.GET("/api/attendance", h.ListAttendance)

	req := httptest.NewRequest(http.MethodGet, "/api/attendance?days=7", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status code want %d, got %d", http.StatusOK, w.Code)
	}

	var resp []AttendanceRecordResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(resp) != 1 {
		t.Fatalf("response length want 1, got %d", len(resp))
	}

	if resp[0].WorkDate != today {
		t.Fatalf("work_date want %s, got %s", today, resp[0].WorkDate)
	}

	if resp[0].WorkDurationMinutes <= 0 {
		t.Fatalf("work_duration_minutes should be positive, got %d", resp[0].WorkDurationMinutes)
	}
}