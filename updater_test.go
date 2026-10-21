package selfupdate_test

import (
	"strings"
	"testing"

	"github.com/creativeprojects/go-selfupdate"
	"github.com/creativeprojects/go-selfupdate/memory"
)

func TestCompileRegexForFiltering(t *testing.T) {
	filters := []string{
		"^hello$",
		"^(\\d\\.)+\\d$",
	}
	up, err := selfupdate.NewUpdater(selfupdate.Config{
		Source:  &memory.Source{},
		Filters: filters,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(up.Filters()) != 2 {
		t.Fatalf("Wanted 2 regexes but got %d", len(up.Filters()))
	}
	for i, r := range up.Filters() {
		want := filters[i]
		got := r.String()
		if want != got {
			t.Errorf("Compiled regex is %q but specified was %q", got, want)
		}
	}
}

func TestFilterRegexIsBroken(t *testing.T) {
	_, err := selfupdate.NewUpdater(selfupdate.Config{
		Source:  &memory.Source{},
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
