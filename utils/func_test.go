package utils

import (
	"fmt"
	"testing"
)

func TestExtractStoreUnit(t *testing.T) {
	size, unit, _ := ExtractStoreUnit("3K")
	fmt.Println(ToBytes(size, unit))
	pid := GetProcessId()
	_, err := GetProcessPhysicalMemory(pid)
	if err != nil {
		println(err)
	}
}
