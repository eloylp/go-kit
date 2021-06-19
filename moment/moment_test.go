package moment_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"go.eloylp.dev/kit/moment"
)

const fixture = "2021-01-01 23:59:59"

func TestNewFakedNow(t *testing.T) {
	expected, err := time.Parse("2006-01-02 15:04:05", fixture)
	if err != nil {
		panic(err)
	}
	now := moment.NewFakedNow(t, fixture)
	assert.Equal(t, expected, now())
}

func TestNewFakedNowWithLocation(t *testing.T) {
	expected, err := time.ParseInLocation("2006-01-02 15:04:05", fixture, time.UTC)
	if err != nil {
		panic(err)
	}
	now := moment.NewFakedNowWithLoc(t, fixture, time.UTC)
	assert.Equal(t, expected, now())
}
