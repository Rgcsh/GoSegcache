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
	"math"
)

// Get
//
//	@Description: 根据缓存key获取缓存值
//	@receiver s:
//	@param _:
//	@param r:
//	@return *proto.GetResponse:
//	@return error:
func (s *Service) Get(_ context.Context, r *proto.GetReq) (*proto.GetResponse, error) {
	key := r.Key

	//	从 KeyHashMap 检测key是否存在,不存在报错 message=no_exist
	//keyHashMapValue, ok := segcache_service.KeyHashMap[key]
	keyHashMapV, ok := segcache_service.KeyHashMap.Load(key)
	if ok != true {
		return &proto.GetResponse{Message: "no exist"}, nil
	}
	keyHashMapValue := keyHashMapV.(segcache_service.KeyHashMapValue)

	//获取当前时间
	currentTime := time_util.GetCurrentTime()
	currentUnixFullMinutes := time_util.GetUnixMinutes(&currentTime)

	// 根据 KeyHashMap的值 获取 从segment取具体数据的相关数据
	// 获取到 存数据所在的segment
	segment := *keyHashMapValue.SegmentPoint
	segmentBody := *segment.Body
	ttlMapValue := *segment.TTLMapValuePoint

	//获取到存储到segment里的此key对应的缓存数据
	segmentItem, ok := segcache_service.ExtractSegmentItem(keyHashMapValue.SegmentPoint, keyHashMapValue.StartIndex)
	if !ok {
		return &proto.GetResponse{Message: "no exist"}, nil
	}

	// 获取到 key在segment索引位置,并找到对应值
	if segmentItem.Key != key {
		// 从segment找的key和查询的key不一样,出bug了
		glog.Log.Error("find value in segment has error,key is not equal to query key,is a bug!")
		return &proto.GetResponse{Message: "find value error"}, nil
	}

	// 获取到 VisitCount在segment索引位置,并找到对应值
	currentUnixMinutesByte, newVisitCountByte := CalVisitCount(*segmentItem.VisitFrequencyByte, currentUnixFullMinutes, key)

	// 存储 访问次数 相关数据
	visitFrequencyByteStartIndex := segmentItem.VisitFrequencyByteStartIndex
	segmentBody[visitFrequencyByteStartIndex] = currentUnixMinutesByte[0]
	segmentBody[visitFrequencyByteStartIndex+1] = currentUnixMinutesByte[1]
	segmentBody[visitFrequencyByteStartIndex+2] = newVisitCountByte[0]

	// 返回数据
	getResponse := proto.GetResponse{Message: "ok", Value: *segmentItem.ValueByte}
	// 如果超时时间是 0,表示永不过期,不用返回 过期时间,否则就要计算过期时间
	if time_util.TimeSubSeconds(time_util.UnixToTime(ttlMapValue.ExpireStartTime), time_util.GetZeroTime()) != 0 {
		expireStartTime := time_util.TimeSubSeconds(&currentTime, time_util.UnixToTime(ttlMapValue.ExpireStartTime))
		expireEndTime := time_util.TimeSubSeconds(&currentTime, time_util.UnixToTime(ttlMapValue.ExpireEndTime))
		getResponse.ExpireStartTime = &expireStartTime
		getResponse.ExpireEndTime = &expireEndTime
	}
	return &getResponse, nil
}

// CalVisitCount
//
//	@Description:使用LFU算法 计算新的访问频率相关数据
//	@param visitFrequencyByte:存储的访问频率的byte数据切片
//	@param currentUnixFullMinutes:当前unix分钟级的时间戳
//	@param key:缓存的key
//	@return currentUnixMinutesByte:当前unix时间戳的byte切片的尾部16bits数据
//	@return newVisitCountByte:新计算的访问次数的1字节数据
func CalVisitCount(visitFrequencyByte []byte, currentUnixFullMinutes uint32, key string) (currentUnixMinutesByte []byte, newVisitCountByte []byte) {
	// 开始对访问次数进行操作
	//数据结构存 16bits的unix上次访问的分钟精度的时间戳(只截取低16位数据) +8bits的访问次数(最大值为255) 共3个字节
	// 切分visitFrequencyByte,分成前半部分unix 16bits时间戳,和后半部分 8bits 访问次数;
	//只存16bits长度的分钟数,数据最大值为uint类型的65535,对应业务中的分钟单位,即45天左右; 也就是说 16bits长度只能存储45天数据,若>45天后,随无法准确判断过了几个周期,默认当做下个周期计算; 但是 对于此算法来讲问题不大,因为 只影响这一次的衰减计算,只要后面访问频率<45天,就可以正确计算;
	//存储的切分后的unix 16bits时间戳
	fmt.Println("获取visitFrequencyByte值", visitFrequencyByte)
	storeUnixMinutes := transform.ByteToUint16(visitFrequencyByte[0:2])
	//当前时间的切分后的16bits分钟时间戳
	currentUnixMinutesByte = transform.ByteTailSplit(transform.Uint32ToByte(currentUnixFullMinutes), 2)
	currentUnixMinutes := transform.ByteToUint16(currentUnixMinutesByte)
	//访问次数
	visitCount := transform.ByteToUint8(visitFrequencyByte[2:3])
	glog.Log.Debug(fmt.Sprintf("get key:%v visit count is :%v", key, visitCount))
	//每次get取数时,先 根据 上次 访问时间,算出距离当前时间的 分钟差距(当 存储的时间戳>当前时间戳,表示过了一个或多个周期,则算法为:2**16-存储的时间戳+当前时间戳;反之 就是 当前时间戳-存储的时间戳),根据配置项(lfu-decay-time) 计算出衰减值; 公式: 分钟差距/lfu-decay-time ; 也就是lfu-decay-time值越大,衰减越慢
	var unitMinutesSub uint16
	if currentUnixMinutes < storeUnixMinutes {
		unitMinutesSub = math.MaxUint16 - storeUnixMinutes + currentUnixMinutes
	} else {
		unitMinutesSub = currentUnixMinutes - storeUnixMinutes
	}
	damping := float64(unitMinutesSub) / config.Conf.Core.LFUDecayTime
	glog.Log.Debug(fmt.Sprintf("get key:%v damping is :%v", key, damping))
	//根据 访问次数-衰减值  算出最新的值;
	var newVisitCount uint8
	// 最终值,不能为负数,最少为0
	if damping > float64(visitCount) {
		newVisitCount = 0
	} else {
		newVisitCount = visitCount - uint8(damping)
	}
	glog.Log.Debug(fmt.Sprintf("get key:%v damping result is :%v", key, newVisitCount))

	//再计算 这次访问对应的更新次数;具体算法如下
	// 有2个配置参数: lfu-log-factor:数值越大,则阈值p越小,越不容易新增访问次数; >=0
	//				baseval: 用上面算出最新的 访问次数的值即可;
	// 算法:p= 1/(lfu-log-factor * baseval +1); 也就是 baseval对应的访问次数越大,则想新增一次计数的概率就越小,从而达到 LFU算法的目的,用 少量数据 通过概率的方式 表示大量数据;
	// 使用时, 生成一个随机值r(范围为0-1),如果r<p,就 将访问次数+1,如果访问次数当前值就是255,则跳过这一步 即可;
	if newVisitCount < math.MaxUint8 {
		p := 1 / float32(config.Conf.Core.LFULogFactor*uint16(newVisitCount)+1)
		randomFloat := utils.GenRandDecimal()
		glog.Log.Debug(fmt.Sprintf("get key:%v accum probability is :%v", key, p))

		if randomFloat < p {
			newVisitCount += 1
			glog.Log.Debug(fmt.Sprintf("key:%v add a visit count success,so lucky,result visist count is:%v", key, newVisitCount))
		}
	}
	newVisitCountByte = transform.Uint8ToByte(newVisitCount)
	fmt.Println("newVisitCountByte最终修改后值为:", newVisitCountByte)
	return currentUnixMinutesByte, newVisitCountByte
}
