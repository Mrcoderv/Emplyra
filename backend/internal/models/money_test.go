package models

import (
	"encoding/json"
	"testing"
)

func TestFromDecimalNormalizes(t *testing.T) {
	cases := map[string]string{
		"100":       "100.00",
		" 12.5 ":    "12.50",
		"0.1":       "0.10",
		"1/2":       "0.50",
		"3/2":       "1.50",
		"-5":        "-5.00",
		"":          "0.00",
		"0.123":     "", // too many decimals -> error
		"not-a-num": "",
	}
	for in, want := range cases {
		d, err := FromDecimal(in)
		if want == "" {
			if err == nil {
				t.Errorf("FromDecimal(%q): expected error, got %q", in, d)
			}
			continue
		}
		if err != nil {
			t.Errorf("FromDecimal(%q): %v", in, err)
			continue
		}
		if d.FloatString() != want {
			t.Errorf("FromDecimal(%q) = %s, want %s", in, d.FloatString(), want)
		}
	}
}

func TestDecimalArithmetic(t *testing.T) {
	a := MustDecimal("1000.50")
	b := MustDecimal("250.25")
	if got := Add(a, b).FloatString(); got != "1250.75" {
		t.Errorf("Add = %s", got)
	}
	if got := Sub(a, b).FloatString(); got != "750.25" {
		t.Errorf("Sub = %s", got)
	}
	if got := Mul(MustDecimal("10.50"), MustDecimal("2.5")).FloatString(); got != "26.25" {
		t.Errorf("Mul = %s", got)
	}
	if got := PercentOf(MustDecimal("2000"), MustDecimal("10")).FloatString(); got != "200.00" {
		t.Errorf("PercentOf = %s", got)
	}
	if got := Sub(MustDecimal("100.00"), Add(MustDecimal("30.00"), MustDecimal("15.00"))).FloatString(); got != "55.00" {
		t.Errorf("net = %s", got)
	}
}

func TestDecimalJSON(t *testing.T) {
	b, err := json.Marshal(MustDecimal("12.30"))
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != `"12.30"` {
		t.Fatalf("marshal = %s", b)
	}
	var d Decimal
	if err := json.Unmarshal([]byte(`"8.25"`), &d); err != nil {
		t.Fatal(err)
	}
	if d.FloatString() != "8.25" {
		t.Fatalf("unmarshal = %s", d.FloatString())
	}
	if err := json.Unmarshal([]byte(`"1.234"`), &d); err == nil {
		t.Fatal("expected error for more than 2 decimal places")
	}
}

func TestDecimalScan(t *testing.T) {
	var d Decimal
	if err := d.Scan([]byte("42.10")); err != nil {
		t.Fatal(err)
	}
	if d.FloatString() != "42.10" {
		t.Fatalf("scan = %s", d.FloatString())
	}
	if err := d.Scan(nil); err != nil {
		t.Fatal(err)
	}
	if d.FloatString() != "0.00" {
		t.Fatalf("scan(nil) = %s", d.FloatString())
	}
}

func TestDecimalCents(t *testing.T) {
	if got := MustDecimal("123.45").Cents(); got != 12345 {
		t.Fatalf("Cents = %d", got)
	}
}
