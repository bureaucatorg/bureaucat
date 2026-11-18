package handlers

import (
	"reflect"
	"testing"
)

func TestNormalizeEmailDomains(t *testing.T) {
	got, err := normalizeEmailDomains([]string{" Example.com ", "@zerodha.com", "example.com", ""})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"example.com", "zerodha.com"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}

	for _, bad := range []string{"localhost", "user@example.com", "exa mple.com", ".example.com", "example.com."} {
		if _, err := normalizeEmailDomains([]string{bad}); err == nil {
			t.Errorf("expected error for %q", bad)
		}
	}
}

func TestIsEmailDomainAllowed(t *testing.T) {
	allowed := []string{"example.com"}
	cases := []struct {
		email   string
		allowed []string
		want    bool
	}{
		{"user@other.org", nil, true},
		{"user@example.com", allowed, true},
		{"User@EXAMPLE.COM", allowed, true},
		{"user@sub.example.com", allowed, false},
		{"user@example.com.evil.org", allowed, false},
		{"user@notexample.com", allowed, false},
		{"no-at-sign", allowed, false},
	}
	for _, tc := range cases {
		if got := isEmailDomainAllowed(tc.email, tc.allowed); got != tc.want {
			t.Errorf("isEmailDomainAllowed(%q, %v) = %v, want %v", tc.email, tc.allowed, got, tc.want)
		}
	}
}
