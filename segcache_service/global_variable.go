package segcache_service

import (
	"GoSegcache/utils/time_util"
	"sync"
)

const SegmentBodyLen = 1024 * 1024 * 10

// Segment
// @Description: 存数据的结构体
type Segment struct {
	//对应TTLMapValue的指针 用来get操作时找过期时间
	TTLMapValuePoint *TTLMapValue
	//执行下个segment的指针
	NextSegment *Segment
	//存数据的byte类型数组,目前长度固定为10M,可以改小点,或根据配置动态修改
	// 每个缓存数据的存储顺序为  key占用多少字节的值(固定4字节)|value占用多少字节的值(固定4个字节)|key值|value值|访问频率(固定3个字节)
	Body *[]byte
}

// TTLMapValue
// @Description: ttl map的值的结构体,存执行首尾segment的指针 和 超时时间范围的 结束时间
type TTLMapValue struct {
	HeadSegment *Segment
	TailSegment *Segment
	//get操作及删除过期数据时需要
	//时间格式使用 unix时间戳,省内存
	ExpireStartTime int64
	ExpireEndTime   int64
}

// TTLMap key值为 过期时间范围的起始时间(格式使用 unix时间戳,省内存)
type TTLMap map[int64]*TTLMapValue

// TTLMapS 秒级的ttl map
// var TTLMapS TTLMap = make(TTLMap)
var TTLMapS sync.Map

// TTLMapM 分钟级的ttl map
// var TTLMapM TTLMap = make(TTLMap)
var TTLMapM sync.Map

// TTLMapH 小时级的ttl map
// var TTLMapH TTLMap = make(TTLMap)
var TTLMapH sync.Map

// TTLMapN 永不过期的ttl map
// var TTLMapN TTLMap = make(TTLMap)
var TTLMapN sync.Map

// KeyHashMap 存放 key到 segment中找到具体缓存数据的 相关值
// var KeyHashMap = make(map[string]KeyHashMapValue)
var KeyHashMap sync.Map

var TTLMapClassify = map[string]*sync.Map{
	time_util.TimeRangeEnum.Second: &TTLMapS,
	time_util.TimeRangeEnum.Minute: &TTLMapM,
	time_util.TimeRangeEnum.Hour:   &TTLMapH,
	time_util.TimeRangeEnum.Null:   &TTLMapN,
}

// LenVisitFrequencyByte 访问频率的byte类型的长度 固定为3,前2个字节存 最后访问的unix分钟级时间戳,最后一个字节存 访问次数
const LenVisitFrequencyByte = 3

// KeyHashMapValue
// @Description: 找到segment的路径
type KeyHashMapValue struct {
	//对应segment的指针
	SegmentPoint *Segment
	//segment中存放对应缓存数据的其实索引位置
	StartIndex uint32
}
