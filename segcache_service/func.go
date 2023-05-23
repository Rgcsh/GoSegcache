package segcache_service

import (
	"GoSegcache/utils/time_util"
	"GoSegcache/utils/transform"
	"sync"
)

// TimeRangeClassify
//
//	@Description: 时间范围分类得到对应的
//	@param num:
func TimeRangeClassify(num float32) (string, *sync.Map) {
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

// SegmentBodyItem
// @Description: segment的body中 一个缓存的相关数据结构体
type SegmentBodyItem struct {
	LenKeyByte                   uint32
	LenValueByte                 uint32
	Key                          string
	KeyByte                      *[]byte
	ValueByte                    *[]byte
	VisitFrequencyByte           *[]byte
	NextItemStartIndex           uint32
	VisitFrequencyByteStartIndex uint32
}

// ExtractSegmentItem
//
//	@Description: 抽取出 存入segment.body中的某条缓存数据
//
// 每个缓存数据的存储顺序为  key占用多少字节的值(固定4字节)|value占用多少字节的值(固定4个字节)|key值|value值|访问频率(固定3个字节)
// 根据如上 存储方式抽取对应数据
//
//	@param segment:
//	@param startIndex:
//	@return *SegmentBodyItem:
//	@return bool:
func ExtractSegmentItem(segment *Segment, startIndex uint32) (*SegmentBodyItem, bool) {
	b := *segment.Body
	//切片索引越界,表示 没有数据
	if startIndex+4 > uint32(len(b)) {
		return nil, false
	}
	//获取 key占用多少字节的值(固定4字节)
	lenKeyByte := transform.ByteToUint32(b[startIndex : startIndex+4])
	// 如果key值长度为0,表示 没有数据
	if lenKeyByte == uint32(0) {
		return nil, false
	}
	//获取 value占用多少字节的值(固定4个字节)
	lenValueByte := transform.ByteToUint32(b[startIndex+4 : startIndex+8])
	//获取 key值
	keyByte := b[startIndex+8 : startIndex+8+lenKeyByte]
	key := transform.ByteToString(keyByte)
	//获取 value值
	valueByte := b[startIndex+8+lenKeyByte : startIndex+8+lenKeyByte+lenValueByte]
	//获取 访问频率(固定3个字节)
	visitFrequencyByte := b[startIndex+8+lenKeyByte+lenValueByte : startIndex+8+lenKeyByte+lenValueByte+3]
	//访问频率开始索引
	visitFrequencyByteStartIndex := startIndex + 8 + lenKeyByte + lenValueByte
	//获取下条数据的开始索引
	nextItemStartIndex := startIndex + 8 + lenKeyByte + lenValueByte + 3
	//todo:对此结构体 进行sync.pool操作
	return &SegmentBodyItem{LenKeyByte: lenKeyByte,
		LenValueByte:                 lenValueByte,
		Key:                          key,
		KeyByte:                      &keyByte,
		ValueByte:                    &valueByte,
		VisitFrequencyByte:           &visitFrequencyByte,
		NextItemStartIndex:           nextItemStartIndex,
		VisitFrequencyByteStartIndex: visitFrequencyByteStartIndex,
	}, true
}
