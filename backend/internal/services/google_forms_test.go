package services

import (
	"errors"
	"testing"

	"github.com/emplyra/backend/internal/models"
)

func TestResolveFieldMappingHeaders(t *testing.T) {
	headers := []string{"Timestamp", "Full Name", "Email Address", "Phone", "Skills"}

	idx, err := resolveFieldMapping(headers, defaultFieldRules, false)
	if err != nil {
		t.Fatalf("default resolve: %v", err)
	}
	if _, ok := idx["email"]; !ok {
		t.Fatal("expected email mapped from 'Email Address'")
	}
	if idx["skills"] != 4 {
		t.Fatalf("expected skills at col 4, got %d", idx["skills"])
	}
	if _, ok := idx["phone"]; !ok {
		t.Fatal("expected phone mapped")
	}
}

func TestResolveFieldMappingStrictMissingColumn(t *testing.T) {
	headers := []string{"Email", "Phone"}
	rules := []FieldMap{{Source: "Full Name", Target: "first_name"}, {Source: "Email", Target: "email"}}
	_, err := resolveFieldMapping(headers, rules, true)
	if !errors.Is(err, ErrGoogleMissingHeader) {
		t.Fatalf("expected ErrGoogleMissingHeader, got %v", err)
	}
}

func TestResolveFieldMappingStrictInvalidTarget(t *testing.T) {
	headers := []string{"Email", "SSN"}
	rules := []FieldMap{{Source: "SSN", Target: "social_security"}}
	_, err := resolveFieldMapping(headers, rules, true)
	if !errors.Is(err, ErrGoogleTargetInvalid) {
		t.Fatalf("expected ErrGoogleTargetInvalid, got %v", err)
	}
}

func TestValuesFromRow(t *testing.T) {
	head := map[string]int{"email": 1, "phone": 3, "first_name": 0}
	row := []string{" Jane Doe ", "jane@example.com", "extra", "555-1234", ""}
	got := valuesFromRow(head, row)
	if got["email"] != "jane@example.com" || got["phone"] != "555-1234" {
		t.Fatalf("unexpected values: %+v", got)
	}
	if got["first_name"] != "Jane Doe" {
		t.Fatalf("expected trimmed first_name, got %q", got["first_name"])
	}
	if _, ok := got["skills"]; ok {
		t.Fatal("blank skills should be omitted")
	}
}

func TestApplyNameSplit(t *testing.T) {
	m := map[string]string{"first_name": "Jane Marie Doe"}
	applyNameSplit(m)
	if m["first_name"] != "Jane Marie" {
		t.Fatalf("first_name = %q", m["first_name"])
	}
	if m["last_name"] != "Doe" {
		t.Fatalf("last_name = %q", m["last_name"])
	}
}

func TestApplyNameSplitRespectsExplicitLastName(t *testing.T) {
	m := map[string]string{"first_name": "Jane Marie Doe", "last_name": "Doe-McGee"}
	applyNameSplit(m)
	if m["first_name"] != "Jane Marie Doe" || m["last_name"] != "Doe-McGee" {
		t.Fatalf("explicit last_name must not be overwritten: %+v", m)
	}
}

func TestExternalResponseID(t *testing.T) {
	integ := &models.GoogleFormIntegration{ID: "abc-123"}
	if got := externalResponseID(integ, 5, ""); got != "gsrow:abc-123:5" {
		t.Fatalf("fallback id = %q", got)
	}
	if got := externalResponseID(integ, 5, "987"); got != "gsr:987" {
		t.Fatalf("explicit id = %q", got)
	}
}

func TestValidEmail(t *testing.T) {
	cases := map[string]bool{
		"jane@example.com": true,
		"a@b.co":           true,
		"":                 false,
		"plaintext":        false,
		"jane@":            false,
		"@example.com":     false,
		"jane@example":     false,
		"ja ne@x.com":      false,
	}
	for email, want := range cases {
		if got := validEmail(email); got != want {
			t.Errorf("validEmail(%q) = %v, want %v", email, got, want)
		}
	}
}

func TestParseSubmittedAt(t *testing.T) {
	if parseSubmittedAt("") != nil {
		t.Fatal("empty should map to nil")
	}
	layouts := []string{"2026-08-29T10:00:00Z", "2026-08-29 10:00:00", "8/29/2026 10:00:00", "2026-08-29"}
	for _, l := range layouts {
		if parseSubmittedAt(l) == nil {
			t.Errorf("expected parse for %q", l)
		}
	}
	if parseSubmittedAt("nonsense") != nil {
		t.Fatal("nonsense should not parse")
	}
}

func TestIsBlankRow(t *testing.T) {
	if !isBlankRow([]string{"", " ", "\t"}) {
		t.Fatal("expected blank row")
	}
	if isBlankRow([]string{"", "Jane", ""}) {
		t.Fatal("expected non-blank row")
	}
}
