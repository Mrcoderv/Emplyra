package auth

import (
	"testing"
	"time"
)

func TestJWTGenerateParseRoundTrip(t *testing.T) {
	m := NewJWTManager("test-secret", time.Hour)
	token, err := m.Generate("u1", "jdoe", "HR_ADMIN", "r1", "tenant", "t1")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	claims, err := m.Parse(token)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if claims.UserID != "u1" || claims.Username != "jdoe" || claims.Role != "HR_ADMIN" || claims.RoleID != "r1" || claims.Scope != "tenant" || claims.TenantID != "t1" {
		t.Fatalf("claims mismatch: %+v", claims)
	}
}

func TestJWTParseExpired(t *testing.T) {
	m := NewJWTManager("test-secret", -time.Minute)
	token, err := m.Generate("u1", "jdoe", "EMPLOYEE", "r2", "tenant", "t1")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if _, err := m.Parse(token); err == nil {
		t.Fatal("expected expired token to be rejected")
	}
}

func TestJWTParseWrongSecret(t *testing.T) {
	m := NewJWTManager("test-secret", time.Hour)
	other := NewJWTManager("other-secret", time.Hour)
	token, err := m.Generate("u1", "jdoe", "EMPLOYEE", "r2", "tenant", "t1")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if _, err := other.Parse(token); err == nil {
		t.Fatal("expected token signed with different secret to be rejected")
	}
}

func TestJWTParserRejectsGarbage(t *testing.T) {
	m := NewJWTManager("test-secret", time.Hour)
	if _, err := m.Parse("not.a.jwt"); err == nil {
		t.Fatal("expected garbage token to be rejected")
	}
	if _, err := m.Parse(""); err == nil {
		t.Fatal("expected empty token to be rejected")
	}
}
