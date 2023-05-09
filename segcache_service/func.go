package segcache_service

import "GoSegcache/utils/time_util"

// TimeRangeClassify
//
//	@Description: 时间范围分类得到对应的
//	@param num:
func TimeRangeClassify(num float32) (string, TTLMap) {
	if num == 0 {
		return time_util.TimeRangeEnum.Null, TTLMapClassify[time_util.TimeRangeEnum.Null]
	}
	if time_util.S1 <= num && num < time_util.H1 {
		return time_util.TimeRangeEnum.Second, TTLMapClassify[time_util.TimeRangeEnum.Second]
	} else if time_util.H1 <= num && num < time_util.D1 {
		return time_util.TimeRangeEnum.Minute, TTLMapClassify[time_util.TimeRangeEnum.Minute]
	} else {
		return time_util.TimeRangeEnum.Hour, TTLMapClassify[time_util.TimeRangeEnum.Hour]
	}
}
