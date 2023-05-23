package rpc_api

import (
	"GoSegcache/config"
	"GoSegcache/pkg/glog"
	"GoSegcache/proto"
	"GoSegcache/segcache_service"
	"GoSegcache/utils"
	"GoSegcache/utils/time_util"
	"GoSegcache/utils/transform"
	"context"
	"fmt"
)

// Set
//
//	@Description:存储缓存key,value
//	@receiver s:
//	@param _:
//	@param r:
//	@return *proto.SetResponse:
//	@return error:
func (s *Service) Set(_ context.Context, r *proto.SetReq) (*proto.SetResponse, error) {
	key := r.Key
	keyByte := transform.StringToByte(key)
	lenKeyByte := len(keyByte)

	valueByte := r.Value
	lenValueByte := len(valueByte)

	//获取当前时间
	currentTime := time_util.GetCurrentTime()
	var expireTime float32 = 0

	//表示byte类型的访问频率,值为0,因为固定,所以直接给出byte类型值即可,省去int转byte过程
	//16bits(前2个字节)的unix上次访问的分钟精度的时间戳 +8bits(最后一个字节)的访问次数(最大值为255)
	currentUnixFullMinutes := time_util.GetUnixMinutes(&currentTime)
	currentUnixMinutesByte := transform.ByteTailSplit(transform.Uint32ToByte(currentUnixFullMinutes), 2)
	visitFrequencyByte := []byte{currentUnixMinutesByte[0], currentUnixMinutesByte[1], 0}

	//表示byte类型的访问频率的长度
	lenVisitFrequencyByte := 3
	//表示byte类型的value值的长度 所占用byte类型的长度(如 value值占用byte长度为 10485760,那么这个长度数据用byte存储就是4个Byte)
	lenValueByteLen := 4
	lenKeyByteLen := 4
	storeByteLen := lenKeyByteLen + lenValueByteLen + lenKeyByte + lenValueByte + lenVisitFrequencyByte

	//需要存储的byte数据,按照 key,value,访问频率的数据拼接
	mergeSegmentByte := GenMergeByte(storeByteLen, lenKeyByte, lenValueByte, &keyByte, &valueByte, &visitFrequencyByte)

	if r.ExpireTime != nil {
		expireTime = *r.ExpireTime
	}
	glog.Log.Debug(fmt.Sprintf("recive Set request,key:%s,expireTime:%v", key, expireTime))

	// 存储数据流程
	// 判断超时时间应该放在哪个TTL级别,如果 没设置超时时间(永不过时),则 将其放入n级别的TTL map中,key为 0-0-0-0 0:0:0:0 即可,表示 永不过期
	TimeRange, TTLMapClassify := segcache_service.TimeRangeClassify(expireTime)
	glog.Log.Debug(fmt.Sprintf("key:%s time range is :%v", key, TimeRange))

	//获取超时时间
	realExpireTime := time_util.GetIntervalTime(expireTime, &currentTime)

	//获取过期时间对应 过期时间范围的开始/结束时间
	// 将超时时间 根据TTL级别 获取对应过期的开始时间;如 在"2022-02-02 12:30:30"过期,过期开始时间为 "2022-02-02 12:30:00"
	expireStartTime, expireEndTime := time_util.ExtractTimeRange(realExpireTime, TimeRange)

	//	判断 过期开始时间 在 对应TTL 级别的map中是否有对应的key
	//if ttlMapValueObj, ok := TTLMapClassify[*time_util.TimeToString(*expireStartTime, time_util.DateTimeFormat)]; ok {
	if ttlMapValueInterface, ok := TTLMapClassify.Load((*expireStartTime).Unix()); ok {
		//在 对应TTL 级别的map中有对应的key时的处理流程
		glog.Log.Debug("key has existed in TTLMap")
		//	有TTL map,就取指向segment的尾指针
		ttlMapValueObj := ttlMapValueInterface.(*segcache_service.TTLMapValue)
		tailSegment := ttlMapValueObj.TailSegment
		body := tailSegment.Body
		lenBody := len(*body)
		var segmentPoint *segcache_service.Segment
		var startIndex = 0
		//	然后判断 segment剩余空间是否够存新数据
		if int(config.Conf.Core.SegmentSizeVal)-lenBody >= storeByteLen {
			glog.Log.Debug("segment body is enough to store new cache")
			//segment剩余空间够用,直接存新数据即可
			*body = append(*body, mergeSegmentByte...)
			startIndex = lenBody
			segmentPoint = tailSegment
		} else {
			//segment剩余空间不够,新建一个,然后与老的segment 指针链接
			segmentPoint = newSegmentAndPoint(mergeSegmentByte, ttlMapValueObj, storeByteLen)
			tailSegment.NextSegment = segmentPoint
		}
		// 将key和偏移量放入hash table
		glog.Log.Debug(fmt.Sprintf("key stored in KeyHashMap,startIndex:%v", startIndex))
		keyHashMapValue := segcache_service.KeyHashMapValue{
			SegmentPoint: segmentPoint,
			StartIndex:   uint32(startIndex),
		}
		//segcache_service.KeyHashMap[key] = keyHashMapValue
		segcache_service.KeyHashMap.Store(key, keyHashMapValue)
	} else {
		// 	没有就新增一个TTLMap的key,再新建segment,然后将数据先计算好长度(偏移量)后放入segment,将segment指针访问TTL map的value中; 再将key和偏移量放入hash table
		glog.Log.Debug("key has not existed in TTLMap,now create a new key/value in TTLMap")
		storeByte := make([]byte, 0, utils.GetMaxSize(int(config.Conf.Core.SegmentSizeVal), storeByteLen))
		storeByte = append(storeByte, mergeSegmentByte...)
		ttlMapValue := segcache_service.TTLMapValue{ExpireStartTime: (*expireStartTime).Unix(), ExpireEndTime: (*expireEndTime).Unix()}
		segment := segcache_service.Segment{TTLMapValuePoint: &ttlMapValue, NextSegment: nil, Body: &storeByte}
		ttlMapValue.HeadSegment = &segment
		ttlMapValue.TailSegment = &segment
		//TTLMapClassify[*time_util.TimeToString(*expireStartTime, time_util.DateTimeFormat)] = &ttlMapValue
		TTLMapClassify.Store((*expireStartTime).Unix(), &ttlMapValue)
		keyHashMapValue := segcache_service.KeyHashMapValue{
			SegmentPoint: &segment,
			StartIndex:   0,
		}
		//segcache_service.KeyHashMap[key] = keyHashMapValue
		segcache_service.KeyHashMap.Store(key, keyHashMapValue)
	}
	glog.Log.Debug(fmt.Sprintf("Handler Set request success,key:%s,expireTime:%v", key, expireTime))
	return &proto.SetResponse{Message: "ok"}, nil
}

// newSegmentAndPoint
//
//	@Description: 新建一个segment,填入数据,然后 修改TTLMap的TailSegment指向新的segment
//	@param mergeSegmentByte:
//	@param ttlMapValueObj:
//	@return segmentPoint:
func newSegmentAndPoint(mergeSegmentByte []byte, ttlMapValueObj *segcache_service.TTLMapValue, storeByteLen int) (segmentPoint *segcache_service.Segment) {
	glog.Log.Debug("segment body is not enough,now will create a new segment")
	//新建一个segment,填入数据
	storeByte := make([]byte, 0, utils.GetMaxSize(int(config.Conf.Core.SegmentSizeVal), storeByteLen))
	storeByte = append(storeByte, mergeSegmentByte...)
	segmentPoint = &segcache_service.Segment{TTLMapValuePoint: ttlMapValueObj, NextSegment: nil, Body: &storeByte}
	//修改TTLMap的TailSegment指向新的segment
	ttlMapValueObj.TailSegment = segmentPoint
	return segmentPoint
}

// GenMergeByte
//
//	@Description:将用户输入的key,value和访问频率 数据 按照顺序合并为一条byte数据
//	@param storeByteLen:
//	@param keyByte:
//	@param value:
//	@param visitFrequencyByte:
//	@return []byte:
func GenMergeByte(storeByteLen int, lenKeyByte, lenValueByte int, keyByte, valueByte, visitFrequencyByte *[]byte) []byte {
	mergeSegmentByte := make([]byte, 0, storeByteLen)

	mergeSegmentByte = append(mergeSegmentByte, transform.Uint32ToByte(uint32(lenKeyByte))...)
	mergeSegmentByte = append(mergeSegmentByte, transform.Uint32ToByte(uint32(lenValueByte))...)
	mergeSegmentByte = append(mergeSegmentByte, *keyByte...)
	mergeSegmentByte = append(mergeSegmentByte, *valueByte...)
	mergeSegmentByte = append(mergeSegmentByte, *visitFrequencyByte...)
	return mergeSegmentByte
}
