package services

import (
	"testing"
	"time"

	"github.com/emplyra/backend/internal/models"
)

func TestBusinessDays(t *testing.T) {
	mon := time.Date(2026, time.August, 24, 12, 0, 0, 0, time.UTC) // Monday
	fri := time.Date(2026, time.August, 28, 12, 0, 0, 0, time.UTC) // Friday
	nextMon := time.Date(2026, time.August, 31, 12, 0, 0, 0, time.UTC)

	if got := businessDays(mon, mon); got != 1 {
		t.Errorf("single day = %d", got)
	}
	if got := businessDays(mon, fri); got != 5 {
		t.Errorf("mon-fri = %d", got)
	}
	// Mon..Sun should skip the weekend -> 5 business days.
	if got := businessDays(mon, time.Date(2026, time.August, 30, 12, 0, 0, 0, time.UTC)); got != 5 {
		t.Errorf("mon-sun = %d", got)
	}
	if got := businessDays(fri, nextMon); got != 2 {
		t.Errorf("fri-mon = %d", got)
	}
	if got := businessDays(mon, nextMon); got != 6 {
		t.Errorf("mon-next-mon = %d", got)
	}
}

func TestPayrollGrossAndNet(t *testing.T) {
	p := &models.Payroll{
		BasicSalary: models.MustDecimal("5000.00"),
		Allowances:  models.MustDecimal("1000.00"),
		Bonus:       models.MustDecimal("500.00"),
		Overtime:    models.MustDecimal("250.50"),
	}
	gross := payrollGross(p)
	if gross.FloatString() != "6750.50" {
		t.Fatalf("gross = %s", gross.FloatString())
	}
	tax := models.MustDecimal("450.00")
	ded := models.MustDecimal("200.25")
	net := models.Sub(gross, models.Add(tax, ded))
	if net.FloatString() != "6100.25" {
		t.Fatalf("net = %s", net.FloatString())
	}
}

func TestComputeLate(t *testing.T) {
	base := time.Date(2026, time.August, 25, 0, 0, 0, 0, time.Local)
	onTime := base.Add(9*time.Hour + 5*time.Minute) // 09:05 within grace (15m)
	if late, mins := computeLate(onTime); late || mins != 0 {
		t.Fatalf("09:05 should not be late, got late=%v mins=%d", late, mins)
	}
	late := base.Add(9*time.Hour + 20*time.Minute) // 09:20 -> 20 min late
	if ok, mins := computeLate(late); !ok || mins != 20 {
		t.Fatalf("09:20 should be 20 min late, got ok=%v mins=%d", ok, mins)
	}
	before := base.Add(8 * time.Hour) // 08:00 early
	if ok, mins := computeLate(before); ok || mins != 0 {
		t.Fatalf("08:00 should not be late, got ok=%v mins=%d", ok, mins)
	}
}

func TestDecOrZero(t *testing.T) {
	if got := decOrZero("12.50").FloatString(); got != "12.50" {
		t.Fatalf("decOrZero valid = %s", got)
	}
	if got := decOrZero("").FloatString(); got != "0.00" {
		t.Fatalf("decOrZero empty = %s", got)
	}
	if got := decOrZero("garbage").FloatString(); got != "0.00" {
		t.Fatalf("decOrZero garbage = %s", got)
	}
}
