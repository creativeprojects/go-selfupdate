package selfupdate

import (
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/stretchr/testify/assert"
)

func TestReleaseLessThan(t *testing.T) {
	testData := []struct {
		current  string
		other    string
		lessThan bool
	}{
		{"1.0", "1.0.0", false},
		{"1.0", "1.0.1", true},
	}
	for _, testItem := range testData {
		release := Release{
			version: semver.MustParse(testItem.current),
		}
		assert.Equal(t, testItem.lessThan, release.LessThan(testItem.other))
	}
}

func TestReleaseGreaterThan(t *testing.T) {
	testData := []struct {
		current     string
		other       string
		greaterThan bool
	}{
		{"1.0", "1.0.0", false},
		{"1.0", "0.9", true},
	}
	for _, testItem := range testData {
		release := Release{
			version: semver.MustParse(testItem.current),
		}
		assert.Equal(t, testItem.greaterThan, release.GreaterThan(testItem.other))
	}
}

func TestReleaseLessOrEqual(t *testing.T) {
	testData := []struct {
		current     string
		other       string
		lessOrEqual bool
	}{
		{"1.0", "1.0.0", true},
		{"1.0", "1.0.1", true},
		{"1.0", "0.9", false},
		{"1.0", "1.0.0-beta", false},
	}
	for _, testItem := range testData {
		release := Release{
			version: semver.MustParse(testItem.current),
		}
		assert.Equal(t, testItem.lessOrEqual, release.LessOrEqual(testItem.other))
	}
}

func TestReleasePointerGreaterOrEqual(t *testing.T) {
	testData := []struct {
		current        string
		other          string
		greaterOrEqual bool
	}{
		{"1.0", "1.0.0", true},
		{"1.0", "0.9", true},
		{"1.0", "1.0.0-beta", true},
	}
	for _, testItem := range testData {
		release := &Release{
			version: semver.MustParse(testItem.current),
		}
		assert.Equal(t, testItem.greaterOrEqual, release.GreaterOrEqual(testItem.other))
	}
}
