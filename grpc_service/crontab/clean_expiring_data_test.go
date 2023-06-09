package crontab

import (
	"GoSegcache/config"
	"GoSegcache/proto"
	"bou.ke/monkey"
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

// TestCleanExpiringData
//
//		@Description:
//		设置假时间为 2023-01-01 12:00:30
//		缓存 N个M级数据(缓存时间为2小时),对 4个key进行get操作(使其访问次数>1,成为热点数据),
//		时间流逝 到 2023-01-01 14:00:10
//	 monkey patch MemoryLimitCheck 函数 为true
//	 睡眠2秒,期间让程序 自动执行
//		@param t:
func TestCleanExpiringData(t *testing.T) {
	FakeTimeNow("2023-01-01 12:00:30")

	config.Conf.Core.LFUVisitCountLimit = 1
	config.Conf.Core.SegmentSizeVal = 60
	config.Conf.Core.LFUMemLimitVal = 10
	var key string
	var valStr string
	expireTime := float32(2 * 60 * 60)

	c := proto.NewGoSegcacheApiClient(Connect())
	//设置缓存
	loopCount := 300
	for i := 0; i < loopCount; i++ {
		key = fmt.Sprintf("key:%v", i)
		valStr = fmt.Sprintf("value is %v", key)
		value := []byte(valStr)
		setReq := &proto.SetReq{Key: key, Value: value, ExpireTime: &expireTime}
		if expireTime == 0 {
			setReq.ExpireTime = nil
		}
		r, err := c.Set(context.Background(), setReq)
		assert.Equal(t, err, nil)
		assert.Equal(t, r.Message, "ok")
	}
	CheckSetGet(t, c, "too long", "val", expireTime+2*24*60*60)

	//造假时间 及 触发内存限制
	FakeTimeNow("2023-01-01 14:00:10")

	//部分数据生成热key
	for i := 0; i < loopCount; i++ {
		if i%3 == 0 {
			key = fmt.Sprintf("key:%v", i)
			GetCheck(t, c, key, "ok")
		}
	}

	//运行被测试的函数
	if config.Conf.Core.LFUEnable == 1 {
		go CleanExpiringData()
	}
	time.Sleep(time.Second * 12)

	//结果检测
	for i := 0; i < loopCount; i++ {
		if i%3 == 0 {
			//检查热key数据应该仍存在
			key = fmt.Sprintf("key:%v", i)
			GetCheck(t, c, key, "ok")
		} else {
			//	检测 非热key数据应该被删除
			key = fmt.Sprintf("key:%v", i)
			GetCheck(t, c, key, "no exist")
		}
	}
}

// FakeMemoryLimitCheck
//
//	@Description: 通过 monkey patch伪造触发 内存限制
//	@param now:
func FakeMemoryLimitCheck() {
	monkey.Patch(MemoryLimitCheck, func() bool {
		return true
	})
}
