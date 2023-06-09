package crontab

import (
	"GoSegcache/config"
	"GoSegcache/pkg/glog"
	"GoSegcache/segcache_service"
	"GoSegcache/utils"
	"GoSegcache/utils/time_util"
	"GoSegcache/utils/transform"
	"fmt"
	"sync"
	"time"
)

//此定时任务 主要负责 TTLMapM,TTLMapH中 处于过期时间范围内的数据进行清理,根据 访问频率,只保留频率高的一部分数据,其余数据删除;
//具体流程 大概为 新建segment,将 需要保存的数据 放入新的里面,并同时修改 KeyHashMap中key对应的值,并修改 TTLMap中首尾segment的指针值即可; 老的segment让GC处理

//无需加额外的锁
//首先,对于segment链表,这个操作执行时,不会有新的写入,因为 正在过期的TTLMap不会有新数据写入,因为缓存过期级别不同,所以不用担心 写入与读取共同出现的问题;
//其次,在写入新的segment时,get操作仍然是老的segment,2者不相关,即使将key对应的 数据写入新segment之后,瞬间get操作并修改了老的segment访问频率值,也没事,因为已经保留了; 就算访问频率值错误了,影响也不大; 本身就是一个

// TTLMapNeedExpiring 需要删除 已经过期数据的TTLMap切片
var TTLMapNeedExpiring = []*sync.Map{&segcache_service.TTLMapM, &segcache_service.TTLMapH}

// 当前程序进程ID
var pid = utils.GetProcessId()

// CleanExpiringData
//
//	@Description:
func CleanExpiringData() {
	sleep := time.Second * 10
	glog.Log.Debug(fmt.Sprintf("睡眠 %v秒后 尝试清除访问频率低的数据", sleep))
	time.Sleep(sleep)
	currentTime := time_util.GetCurrentTime()
	glog.Log.Debug(fmt.Sprintf("开始清除处于过期范围内的部分访问频率低的数据 任务"))

	if !MemoryLimitCheck() {
		//未超过限制进入下次循环 睡1分钟循环执行下一次
		CleanExpiringData()
	}

	//对2个TTLMap级别的数据进行遍历
	glog.Log.Debug(fmt.Sprintf("超过内存限制,开始清理部分过期数据"))
	for index, ttlMap := range TTLMapNeedExpiring {
		glog.Log.Debug(fmt.Sprintf("开始处理 第%v个 TTLMap", index+1))
		//检查出过期需要删除的key,找到对应的TTLMap的值,接着找到对应的首segment
		ttlMap.Range(func(key, value any) bool {
			v := value.(*segcache_service.TTLMapValue)
			//检查是否过期
			expireStartTime := *time_util.UnixToTime(v.ExpireStartTime)
			expireEndTime := *time_util.UnixToTime(v.ExpireEndTime)
			glog.Log.Debug(fmt.Sprintf("key:%v,过期时间范围是 %v-->%v", key, expireStartTime, expireEndTime))
			//当前时间 不在 过期时间范围内,就跳过此条
			if !(currentTime.Before(expireEndTime) && currentTime.After(expireStartTime)) {
				glog.Log.Debug("当前时间 不在 key的过期时间范围,跳过此条数据")
				return true
			}
			//对segment链表进行处理
			glog.Log.Debug("当前时间 在key过期时间范围内,开始对segment链表进行处理")
			FilterSegment(v)
			glog.Log.Debug(fmt.Sprintf("当前时间 key:%v 对应segment链表已经处理完成", key))
			return true
		})
	}

	CleanExpiringData()
}

// MemoryLimitCheck
//
//	@Description: 检测 程序消耗内存是否超过 设置的值
//	@return bool: true:触发内存限制 false:未触发
func MemoryLimitCheck() bool {
	//获取此程序消耗物理内存
	usedRss, err := utils.GetProcessPhysicalMemory(pid)
	if err != nil {
		panic(fmt.Sprintf("Get current process memory info fail, error is %v", err))
	}

	// 检测 程序消耗的物理内存是否超过限制
	memLimit := uint64(config.Conf.Core.LFUMemLimitVal)
	glog.Log.Debug(fmt.Sprintf("获取到此程序消耗物理内存为:%v,配置中限制内存为:%v", memLimit, usedRss))
	if usedRss > memLimit {
		return true
	}
	return false
}

// FilterSegment
//
//	@Description: 将 大于配置文件中的访问次数保留 到新的segment中,并修改相关数据
//	@param segment:
func FilterSegment(ttlMapValue *segcache_service.TTLMapValue) {
	segment := ttlMapValue.HeadSegment
	//处理当前segment数据
	startIndex := uint32(0)
	newSegment := &segcache_service.Segment{TTLMapValuePoint: ttlMapValue}
	//新的头segment
	newHeadSegment := newSegment
	//对segment链表循环
	for {
		//对单个segment里的item循环
		for {
			glog.Log.Debug("开始循环单个segment里的数据")
			var ok bool
			newSegment, startIndex, ok = HandlerSegmentItem(segment, newSegment, startIndex, ttlMapValue)
			if !ok {
				break
			}
		}

		glog.Log.Debug("到下一个segment继续执行")
		segment = segment.NextSegment
		if segment == nil {
			glog.Log.Debug("segment链表已经全部处理完成")
			break
		}
	}
	glog.Log.Debug("将新的首尾segment指针 赋值给ttlMapValue,从而放弃老segment,让GC处理")
	ttlMapValue.HeadSegment = newHeadSegment
	ttlMapValue.TailSegment = newSegment
}

// HandlerSegmentItem
//
//	@Description:
//	@param oldSegment:老的(目前正在使用的)segment指针
//	@param newSegment:新的segment链表中 最新的一个segment指针,作用是 将oldSegment中的符合条件的item放入此segment中
//	@param startIndex:老的segment中,访问item的开始索引
//	@param ttlMapValue:
//	@return *segcache_service.Segment: 新的segment链表中 最新的一个segment指针
//	@return uint32:oldSegment下一个item的开始index
//	@return bool: oldSegment是否继续访问下一个item false:表示将oldSegment访问到底,没有新的item true:继续访问下个item
func HandlerSegmentItem(oldSegment, newSegment *segcache_service.Segment, startIndex uint32, ttlMapValue *segcache_service.TTLMapValue) (*segcache_service.Segment, uint32, bool) {
	segmentItem, ok := segcache_service.ExtractSegmentItem(oldSegment, startIndex)
	if !ok {
		glog.Log.Debug("当前segment的Body已经访问到尾部,没有新数据,直接返回")
		return newSegment, 0, false
	}
	key := segmentItem.Key
	//获取访问次数,并与 配置的 LFUVisitCountLimit对比
	visitCount := transform.ByteToUint8((*oldSegment.Body)[segmentItem.VisitFrequencyByteStartIndex+2 : segmentItem.VisitFrequencyByteStartIndex+3])
	// < LFUVisitCountLimit时,删除KeyHashMap中的key即可
	glog.Log.Debug(fmt.Sprintf("缓存key:%v 访问次数:%v,配置访问次数阈值:%v", key, visitCount, config.Conf.Core.LFUVisitCountLimit))
	if visitCount < config.Conf.Core.LFUVisitCountLimit {
		glog.Log.Debug(fmt.Sprintf("未超过配置访问次数阈值,从KeyHashMap中删除key:%v对应相关数据", key))
		segcache_service.KeyHashMap.Delete(key)
		return newSegment, segmentItem.NextItemStartIndex, true
	}

	// >= LFUVisitCountLimit时,将现在的segment中对应的item byte切片数据copy到 newSegment中(注意空间是否足够的处理)
	glog.Log.Debug(fmt.Sprintf("超过配置访问次数阈值,key:%v 正常进行后续任务", key))
	itemByte := (*oldSegment.Body)[startIndex:segmentItem.NextItemStartIndex]
	storeByteLen := len(itemByte)

	//没有Body,新建一个 的情况
	if newSegment.Body == nil {
		//新建一个segment,填入数据
		storeByte := make([]byte, 0, utils.GetMaxSize(int(config.Conf.Core.SegmentSizeVal), storeByteLen))
		storeByte = append(storeByte, itemByte...)
		newSegment.Body = &storeByte
		return newSegment, segmentItem.NextItemStartIndex, true
	}

	newSegmentBodyLen := len(*newSegment.Body)
	var newSegmentBodyStartIndex = 0

	//	判断 segment剩余空间是否够存新数据
	if int(config.Conf.Core.SegmentSizeVal)-newSegmentBodyLen >= storeByteLen {
		//segment剩余空间够用,直接存新数据即可
		glog.Log.Debug("segment存储数据的Body够存储一条新数据")
		newSegmentBodyStartIndex = newSegmentBodyLen
		*newSegment.Body = append(*newSegment.Body, itemByte...)
		return newSegment, segmentItem.NextItemStartIndex, true
	}

	//segment剩余空间不够
	glog.Log.Debug("segment存储数据的Body不够用,创建一个新的的segment")
	//新建一个segment,填入数据
	storeByte := make([]byte, 0, utils.GetMaxSize(int(config.Conf.Core.SegmentSizeVal), storeByteLen))
	storeByte = append(storeByte, itemByte...)
	createNewSegment := &segcache_service.Segment{TTLMapValuePoint: ttlMapValue, NextSegment: nil, Body: &storeByte}
	//将2个segment 指针链接
	newSegment.NextSegment = createNewSegment
	//修改 KeyHashMap中此key对应的相关值
	segcache_service.KeyHashMap.Store(key, segcache_service.KeyHashMapValue{SegmentPoint: createNewSegment, StartIndex: uint32(newSegmentBodyStartIndex)})
	return createNewSegment, segmentItem.NextItemStartIndex, true
}
