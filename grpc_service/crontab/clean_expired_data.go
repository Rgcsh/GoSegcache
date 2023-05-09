package crontab

//todo:map全部改为 sync.map,且相关调用也要修改

//todo:main函数生成一个子协程, 循环调用,直至进程结束,期间执行完一次,休眠 配置中给定时间后,再进入下一次循环

//todo:此定时任务 主要负责 对已经过期的TTLMap对应的 segment里的key 对应在 KeyHashMap 中的数据删除,再删除TTLMap中的数据即可,segment中的数据不用删,只要引用被删除,GC会自动处理
