package crontab

import (
	"GoSegcache/config"
	"GoSegcache/proto"
	"fmt"
	"testing"
	"time"
)

// TestCleanExpiredData
//
//	@Description:新建多个M级别的数据,在超过 过期时间后,尝试获取缓存值,应该显示 不存在
//	@param t:
func TestCleanExpiredData(t *testing.T) {
	//启动主动删除缓存功能 相关子协程
	go CleanExpiredData()
	var key string
	var valStr string
	expireTime := float32(2 * 60 * 60)

	FakeTimeNow("2023-01-01 12:10:00")
	c := proto.NewGoSegcacheApiClient(Connect())
	config.Conf.Core.SegmentSizeVal = 60
	for i := 0; i < 6; i++ {
		key = fmt.Sprintf("key:%v", i)
		valStr = fmt.Sprintf("value is %v", key)
		CheckSetGet(t, c, key, valStr, expireTime)
	}
	time.Sleep(time.Second * 1)

	FakeTimeNow("2023-01-01 14:10:00")
	time.Sleep(time.Second * 1)

	for i := 0; i < 100; i++ {
		key = fmt.Sprintf("key:%v", i)
		GetCheck(t, c, key, "no exist")
	}
}
