package utils

// 将KB,MB,GB,TB对应的字节大小
const (
	_          = iota
	KB float64 = 1 << (10 * iota)
	MB
	GB
	TB
)
