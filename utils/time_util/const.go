package time_util

// 1s
const S1 float32 = 1

// 1小时
const H1 float32 = 60 * 60

// 1天
const D1 float32 = H1 * 24

// 时间范围枚举结构体
type TimeRangeEnumStruct struct {
	Second, Minute, Hour, Null string
}

// 时间范围枚举变量
var TimeRangeEnum = TimeRangeEnumStruct{
	Second: "S",
	Minute: "M",
	Hour:   "H",
	//表示永不过期
	Null: "N",
}
