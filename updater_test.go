package selfupdate

import (
	"strings"
	"testing"
)

func TestCompileRegexForFiltering(t *testing.T) {
	filters := []string{
		"^hello$",
		"^(\\d\\.)+\\d$",
	}
	up, err := NewUpdater(Config{
		Filters: filters,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(up.filters) != 2 {
		t.Fatalf("Wanted 2 regexes but got %d", len(up.filters))
	}
	for i, r := range up.filters {
		want := filters[i]
		got := r.String()
		if want != got {
			t.Errorf("Compiled regex is %q but specified was %q", got, want)
		}
	}
}

func TestFilterRegexIsBroken(t *testing.T) {
	_, err := NewUpdater(Config{
		Filters: []string{"(foo"},
	})
	if err == nil {
		t.Fatal("Error unexpectedly did not occur")
	}
	msg := err.Error()
	if !strings.Contains(msg, "could not compile regular expression \"(foo\" for filtering releases") {
		t.Fatalf("Error message is unexpected: %q", msg)
	}
}
