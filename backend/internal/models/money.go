package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
)

// Decimal is an exact fixed-point currency value backed by a string.
// It maps to Postgres `numeric` and never uses floating point, so currency
// calculations are precise.
type Decimal string

func (d Decimal) GormDataType() string { return "numeric" }

func (d Decimal) Value() (driver.Value, error) {
	s := strings.TrimSpace(string(d))
	if s == "" {
		s = "0.00"
	}
	return s, nil
}

func (d *Decimal) Scan(v interface{}) error {
	switch x := v.(type) {
	case []byte:
		*d = Decimal(string(x))
	case string:
		*d = Decimal(x)
	case nil:
		*d = "0.00"
	default:
		return fmt.Errorf("cannot scan %T into Decimal", v)
	}
	return nil
}

func (d Decimal) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.FloatString())
}

func (d *Decimal) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	parsed, err := FromDecimal(s)
	if err != nil {
		return err
	}
	*d = parsed
	return nil
}

func (d Decimal) Rat() *big.Rat {
	r, ok := new(big.Rat).SetString(strings.TrimSpace(string(d)))
	if !ok {
		return new(big.Rat)
	}
	return r
}

func (d Decimal) FloatString() string {
	return d.Rat().FloatString(2)
}

func (d Decimal) IsZero() bool {
	return d.Rat().Sign() == 0
}

// FromDecimal parses and normalizes a decimal string (max 2 decimal places).
func FromDecimal(s string) (Decimal, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Decimal("0.00"), nil
	}
	r, ok := new(big.Rat).SetString(s)
	if !ok {
		return "0.00", fmt.Errorf("invalid decimal: %q", s)
	}
	f := new(big.Rat).Mul(r, big.NewRat(100, 1))
	if f.Denom().Cmp(big.NewInt(1)) != 0 {
		return "0.00", fmt.Errorf("invalid decimal: %q (max 2 decimal places)", s)
	}
	return Decimal(r.FloatString(2)), nil
}

func MustDecimal(s string) Decimal {
	d, err := FromDecimal(s)
	if err != nil {
		return Decimal("0.00")
	}
	return d
}

func Add(a, b Decimal) Decimal {
	return Decimal(new(big.Rat).Add(a.Rat(), b.Rat()).FloatString(2))
}

func Sub(a, b Decimal) Decimal {
	return Decimal(new(big.Rat).Sub(a.Rat(), b.Rat()).FloatString(2))
}

func Mul(a Decimal, factor Decimal) Decimal {
	return Decimal(new(big.Rat).Mul(a.Rat(), factor.Rat()).FloatString(2))
}

// PercentOf computes a * (pct/100).
func PercentOf(a Decimal, pct Decimal) Decimal {
	ratio := new(big.Rat).Quo(pct.Rat(), big.NewRat(100, 1))
	return Decimal(new(big.Rat).Mul(a.Rat(), ratio).FloatString(2))
}

// Cents returns the amount as an integer count of the smallest unit (cent).
func (d Decimal) Cents() int64 {
	r := new(big.Rat).Mul(d.Rat(), big.NewRat(100, 1))
	q := new(big.Int).Quo(r.Num(), r.Denom())
	return q.Int64()
}
