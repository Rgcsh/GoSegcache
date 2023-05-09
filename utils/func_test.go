package utils

import (
	"fmt"
	"testing"
)

func TestExtractStoreUnit(t *testing.T) {
	size, unit, _ := ExtractStoreUnit("3K")
	fmt.Println(ToBytes(size, unit))
	x := GetMemoryStats()
	fmt.Println("dfdfdf", x, "dfdf")
	y := GetMemPercent()
	fmt.Println(y)
}
