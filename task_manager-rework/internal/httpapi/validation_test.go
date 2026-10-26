package httpapi

import "testing"

func TestValidStatus(t *testing.T) {
	cases := []struct {
		value string
		want  bool
	}{
		{"new", true},
		{"in_progress", true},
		{"completed", true},
		{"blocked", false},
		{"", false},
	}
	for _, tc := range cases {
		if got := validStatus(tc.value); got != tc.want {
			t.Errorf("validStatus(%q) = %v, want %v", tc.value, got, tc.want)
		}
	}
}

func TestValidPriority(t *testing.T) {
	cases := []struct {
		value string
		want  bool
	}{
		{"low", true},
		{"medium", true},
		{"high", true},
		{"urgent", false},
		{"", false},
	}
	for _, tc := range cases {
		if got := validPriority(tc.value); got != tc.want {
			t.Errorf("validPriority(%q) = %v, want %v", tc.value, got, tc.want)
		}
	}
}
