package time_util

import "testing"

func TestExtractTimeRange(t *testing.T) {
	now := GetCurrentTime()
	ExtractTimeRange(&now, "S")
	ExtractTimeRange(&now, "M")
	ExtractTimeRange(&now, "H")
}
