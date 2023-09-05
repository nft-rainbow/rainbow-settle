package models

import (
	"testing"
	"time"

	"github.com/go-playground/assert/v2"
)

func TestPeroidEndTime(t *testing.T) {
	endTime := PEROID_TYPE_DAY.EndTime(time.Date(2000, 1, 1, 2, 0, 0, 0, time.Local))
	assert.Equal(t, time.Date(2000, 1, 2, 2, 0, 0, 0, time.Local), endTime)
	endTime = PEROID_TYPE_MONTH.EndTime(time.Date(2000, 12, 1, 0, 0, 0, 0, time.Local))
	assert.Equal(t, time.Date(2001, 1, 1, 0, 0, 0, 0, time.Local), endTime)
	endTime = PEROID_TYPE_YEAR.EndTime(time.Date(2000, 1, 1, 0, 0, 0, 0, time.Local))
	assert.Equal(t, time.Date(2001, 1, 1, 0, 0, 0, 0, time.Local), endTime)
}
