package segcache_service

import (
	"GoSegcache/utils/time_util"
	"time"
)

const SegmentBodyLen = 1024 * 1024 * 10

// Segment
// @Description: 存数据的结构体
type Segment struct {
	//对应TTLMapValue的指针
	TTLMapValuePoint *TTLMapValue
	//执行下个segment的指针
	NextSegment *Segment
	//存数据的byte类型数组,目前长度固定为10M,可以改小点,或根据配置动态修改
	Body *[]byte
}

// TTLMapValue
// @Description: ttl map的值的结构体,存执行首尾segment的指针 和 超时时间范围的 结束时间
type TTLMapValue struct {
	HeadSegment     *Segment
	TailSegment     *Segment
	ExpireStartTime time.Time
	ExpireEndTime   time.Time
}

// TTLMap key值为 过期时间范围的起始时间
type TTLMap map[string]*TTLMapValue

// TTLMapS 秒级的ttl map
var TTLMapS TTLMap = make(TTLMap)

// TTLMapM 分钟级的ttl map
var TTLMapM TTLMap = make(TTLMap)

// TTLMapH 小时级的ttl map
var TTLMapH TTLMap = make(TTLMap)

// TTLMapN 永不过期的ttl map
var TTLMapN TTLMap = make(TTLMap)

// KeyHashMap 存放 key到 segment中找到具体缓存数据的 相关值
var KeyHashMap = make(map[string]KeyHashMapValue)

var TTLMapClassify = map[string]TTLMap{
	time_util.TimeRangeEnum.Second: TTLMapS,
	time_util.TimeRangeEnum.Minute: TTLMapM,
	time_util.TimeRangeEnum.Hour:   TTLMapH,
	time_util.TimeRangeEnum.Null:   TTLMapN,
}

// LenVisitFrequencyByte 访问频率的byte类型的长度 固定为3,前2个字节存 最后访问的unix分钟级时间戳,最后一个字节存 访问次数
const LenVisitFrequencyByte = 3

// KeyHashMapValue
// @Description: 找到segment的路径
type KeyHashMapValue struct {
	//对应segment的指针
	SegmentPoint *Segment
	//segment中存放对应缓存数据的其实索引位置
	StartIndex int
	//key的byte类型的长度
	KeyLength int
	//value的byte类型的长度
	ValueLength int
}
