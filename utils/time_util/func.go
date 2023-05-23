package time_util

import (
	"fmt"
	"time"
)

type Time time.Time

const (
	zone           = "Asia/Shanghai"
	DateTimeFormat = "2006-01-02 15:04:05"
	DateFormat     = "2006-01-02"
	TimeFormat     = "15:04:05"
)

// UnmarshalJson implements json unmarshal interface.
func (t *Time) UnmarshalJson(data []byte) (err error) {
	now, err := time.ParseInLocation(`"`+DateTimeFormat+`"`, string(data), time.Local)
	*t = Time(now)
	return
}

// MarshalJSON implements json marshal interface.
func (t Time) MarshalJSON() ([]byte, error) {
	b := make([]byte, 0, len(DateTimeFormat)+2)
	b = append(b, '"')
	b = time.Time(t).AppendFormat(b, DateTimeFormat)
	b = append(b, '"')
	return b, nil
}

func (t Time) String() string {
	return time.Time(t).Format(DateTimeFormat)
}

func (t Time) Local() time.Time {
	loc, _ := time.LoadLocation(zone)
	return time.Time(t).In(loc)
}
func GetLocation() *time.Location {
	loc, _ := time.LoadLocation(zone)
	return loc
}

func (t Time) Unix() int64 {
	return time.Time(t).Unix()
}

// Scan valueof time.Time 注意是指针类型 method
func (t *Time) Scan(v interface{}) error {
	value, ok := v.(time.Time)
	if ok {
		*t = Time(value)
		return nil
	}
	return fmt.Errorf("can not convert %v to timestamp", v)
}

// ParseLocal
//
//	@Description: 修改时间的时区为上海时区,且返回值为字符串格式
//	@param inputTime:
//	@return string:
func ParseLocal(inputTime time.Time) string {
	loc, _ := time.LoadLocation(zone)
	return inputTime.In(loc).Format(DateTimeFormat)
}

func UnixToTime(unixTime int64) *time.Time {
	t := time.Unix(unixTime, 0)

	return &t
}

// FormatTimeLocal
//
//	@Description: 将时间修改为上海时区,并 返回 DateTimeFormat 字符串格式
//	@param inputTime:
//	@return string:
func FormatTimeLocal(inputTime time.Time) string {
	return inputTime.Format(DateTimeFormat)
}

// StringToTime
//
//	@Description: 字符串格式的时间 转为 Time对象的 时间
//	@param s:
//	@return time.Time:
//	@return error:
func StringToTime(s string, format string) (*time.Time, error) {
	// 设置时区
	location, _ := time.LoadLocation("Asia/Shanghai")
	timeObj, err := time.ParseInLocation(format, s, location)
	return &timeObj, err
}

// @Description: 执行时间周期触发器
// @param seconds: 周期秒数
// @return <-chan:
func TimeCycle(seconds time.Duration) <-chan time.Time {
	timer := time.NewTimer(seconds * time.Second)
	return timer.C
}

// TimeToString
//
//	@Description: 时间转为字符串
//	@param t:
//	@param format:
//	@return string:
func TimeToString(t time.Time, format string) *string {
	stringTime := t.Format(format)
	return &stringTime
}

// @Description: 获取当前时间
// @return string: "2022-12-07 18:14:30"
func GetCurrentStringTime() string {
	return *TimeToString(time.Now(), DateTimeFormat)
}

// GetCurrentTime
//
//	@Description: 获取当前时间
//	@return time.Time:
func GetCurrentTime() time.Time {
	return time.Now()
}

// TimeSubSeconds
//
//	@Description: 获取2个时间 的 秒级别的 时间差
//	@param startTime:
//	@param endTime:
//	@return float64:
func TimeSubSeconds(startTime *time.Time, endTime *time.Time) float64 {
	return endTime.Sub(*startTime).Seconds()
}

// @Description: 获取过去某天的 时间
// @param days:
// @return *time.Time:
func GetPastDate(days int) *time.Time {
	currentTime := time.Now()
	pastTime := currentTime.AddDate(0, 0, days)
	return &pastTime
}

// GetIntervalTime
//
//	@Description: 获取 给定时间 间隔后的时间
//	@param seconds:
//	@param t:
//	@return *time.Time:
func GetIntervalTime(Seconds float32, t *time.Time) *time.Time {
	//将Seconds(秒)转为纳秒
	nano := time.Duration(Seconds * 1000 * 1000 * 1000)
	intervalTime := t.Add(nano)
	return &intervalTime
}

// GetZeroTime
//
//	@Description: 获取0值时间
//	@return *time.Time:
func GetZeroTime() *time.Time {
	t := time.Date(0, 0, 0, 0, 0, 0, 0, GetLocation())
	return &t
}

// GetUnixMinutes
//
//	@Description: 获取unix分钟,返回uint32,4个字节长度的数据结构就够用,不用int64
//	@return uint32:
func GetUnixMinutes(t *time.Time) uint32 {
	unix := t.Unix()
	return uint32(unix / 60)
}

// UnixMinutesSub
//
//	@Description: 2个分钟级别的时间 相减 得到分钟级别的时间差
//	@param firstMinutes:
//	@param lastMinutes: 此值最好大于firstMinutes,否则值为负数
//	@return int32:
func UnixMinutesSub(firstMinutes, lastMinutes uint16) uint16 {
	return lastMinutes - firstMinutes
}

// ExtractTimeRange
//
//	@Description: 根据时间级别抽取对应的时间范围
//	@param time:
//	@param timeClassify: S:秒级 M:分钟级 H:小时级
func ExtractTimeRange(t *time.Time, timeClassify string) (*time.Time, *time.Time) {
	var startDate, endDate time.Time
	if timeClassify == TimeRangeEnum.Second {
		//对秒 进行范围抽取
		startDate = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 0, GetLocation())
		endDate = startDate.Add(time.Second)
	} else if timeClassify == TimeRangeEnum.Minute {
		//对分钟进行范围抽取
		startDate = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 0, 0, GetLocation())
		endDate = startDate.Add(time.Minute)
	} else if timeClassify == TimeRangeEnum.Hour {
		//对小时进行范围抽取
		startDate = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, GetLocation())
		endDate = startDate.Add(time.Hour)
	} else {
		//永不过期 设置时间都是 0
		startDate = *GetZeroTime()
		endDate = startDate
	}
	return &startDate, &endDate
}
