package time_util

import (
	"fmt"
	"testing"
)

func TestExtractTimeRange(t *testing.T) {
	now := GetCurrentTime()
	ExtractTimeRange(&now, "S")
	ExtractTimeRange(&now, "M")
	ExtractTimeRange(&now, "H")
}

func TestGetIntervalTime(t *testing.T) {
	now := GetCurrentTime()
	fmt.Println(now)
	newTime := GetIntervalTime(10, &now)
	fmt.Println(newTime)
}
