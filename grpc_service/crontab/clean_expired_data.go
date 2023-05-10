package crontab

import (
	"GoSegcache/pkg/glog"
	"GoSegcache/segcache_service"
	"GoSegcache/utils/time_util"
	"fmt"
	"sync"
	"time"
)

// TTLMapNeedExpired 需要删除 已经过期数据的TTLMap切片
var TTLMapNeedExpired = []*sync.Map{&segcache_service.TTLMapS, &segcache_service.TTLMapM, &segcache_service.TTLMapH}

// CleanExpiredData
//
//	@Description: 此定时任务 主要负责  对已经过期的TTLMap对应的 segment里的key 对应在 KeyHashMap 中的数据删除,再删除TTLMap中的数据即可
//
// segment中的数据不用删,只要引用被删除,GC会自动处理
// 无需加额外的锁,sync.map够用; 在删除过程中 遇到get操作时,只要在删除对应 KeyHashMap的key之前都能get操作获取到数据,不会报错,因为segment中的数据不删除
func CleanExpiredData() {
	currentTime := time_util.GetCurrentTime()
	glog.Log.Debug(fmt.Sprintf("start clean expired data,current time is %v", currentTime))

	//对3个TTLMap级别的数据进行遍历
	for _, ttlMap := range TTLMapNeedExpired {
		//检查出过期需要删除的key,找到对应的TTLMap的值,接着找到对应的首segment
		ttlMap.Range(func(key, value any) bool {
			v := value.(*segcache_service.TTLMapValue)
			//检查是否过期
			expireEndTime := v.ExpireEndTime
			if currentTime.Before(expireEndTime) {
				return true
			}
			//沿着 segment链表,将对应存储的key在 KeyHashMap中找到,并删除即可
			CleanSegmentLink(v.HeadSegment)
			//删完后,再删除 TTLMap中此key数据即可
			ttlMap.Delete(key)
			return true
		})
	}

	//睡500毫秒循环执行下一次
	time.Sleep(time.Millisecond * 500)
	CleanExpiredData()
}

// CleanSegmentLink
//
//	@Description: 沿着 segment链表,将对应存储的key在 KeyHashMap中找到并删除
//	@param segment:
func CleanSegmentLink(segment *segcache_service.Segment) {
	//处理当前segment数据
	startIndex := uint32(0)
	for {
		segmentItem, ok := segcache_service.ExtractSegmentItem(segment, startIndex)
		if !ok {
			break
		}
		key := segmentItem.Key
		//删除KeyHashMap中的key
		segcache_service.KeyHashMap.Delete(key)
		startIndex = segmentItem.NextItemStartIndex
	}

	//到下一个segment继续执行
	nextSegment := segment.NextSegment
	if nextSegment != nil {
		CleanSegmentLink(nextSegment)
	} else {
		return
	}
}
