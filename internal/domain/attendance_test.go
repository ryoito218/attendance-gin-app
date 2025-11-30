package domain

import (
	"testing"
	"time"
)

func TestCalcWorkDuration_Normal(t *testing.T) {
	in := time.Date(2025, 1, 1, 9, 0, 0, 0, time.Local)
	out := time.Date(2025, 1, 1, 18, 0, 0, 0, time.Local)

	got := CalcWorkDuration(in, out)
	want := 9 * time.Hour

	if got != want {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestCalcWorkDuration_Negative(t *testing.T) {
	in := time.Date(2025, 1, 1, 18, 0, 0, 0, time.Local)
	out := time.Date(2025, 1, 1, 9, 0, 0, 0, time.Local)

	got := CalcWorkDuration(in, out)
	if got != 0 {
		t.Fatalf("want 0, got %v", got)
	}
}