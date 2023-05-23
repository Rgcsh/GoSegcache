# 基于Go语言的分段缓存系统

## 设计思路

基于 [Segcache: a memory-efficient and scalable in-memory key-value cache for small objects](https://www.usenix.org/system/files/nsdi21-yang.pdf)
2021年的论文提供的缓存思路实现的一种缓存方式;

### 缓存设计思路

### 行业痛点

* 缓存一个缓存数据需要耗费很多 额外的数据空间;
* 缓存过期时,无法高效主动的回收内存;

### 优点

* 节约存储数据时耗费的空间:将每个缓存中可公用的元数据 集中存储,如超时时间,创建时间;
* 主动及时删除过期缓存,做到及时回收内存空间;

### 缺点

* 淡化了缓存的过期时间精度,可能会使缓存提前过期或稍后过期(在缓存超过1天时间的,过期最大误差为1h),但是在缓存系统中,这种问题可以忽略;
* todo:后续 尝试优化 TTL buckets的存储结构 达到 存取效率 及 缓存时间问题的平衡;

### 与其他缓存平台的对比

* 与redis,memcached相比,如上所述优点仍然存在;

### 实现功能

* set命令存数据 key:string/int 类型 val:任何类型 expire:float/int类型
* get命令取数据,并根据LFU算法设置其热度值
* 每秒定时遍历TTL buckets 删除已经完全过期的segment数据,及根据LFU算法的热度值删除刚开始过期的数据,并重新分配内存;

### 相关数据结构及存储的数据概述

#### TTL(有效时间)级别分为3种,具体如下

* S(秒)级 TTL范围为 1s~1h
* Min(分钟)级 TTL范围为 1h~1d
* h(小时)级 TTL范围为 1d~无穷大;
* n级 TTL范围为 永不过期;

#### TTL数据结构

* 3种级别 均对应3个map结构,key为 TTL范围的起始时间(如过期时间为 1h,在"2022-02-02 12:30:30"过期,则TTL级别为Min级,
  对应的key值取过期时间所在分钟的开始时间,即"2022-02-02 12:30:00"),若为n级,则值为 "0-0-0-0 0:0:0:0",
  value为list结构,对应 首/尾segment的指针地址 及 TTL范围的过期时间(示例为 "2022-02-02 12:31:00",作用为 如果当前时间>
  过期时间,则直接删除相关数据即可);

#### Segment数据结构

* segment之间以单链表的形式相连接; 目的:防止只用一个segment存储所有数据导致 连续的物理内存块无法分配;
  链表形式 使segment之间不用分配物理上连续内存;
* segment自身以list结构,值为byte类型 存储数据; 目的:固定list长度,存储数据超出,就新建一个segment放新数据;
  值为byte类型:因为所有数据结构都可以转为byte类型,所以存byte类型就是存所有数据结构,且
  go语言是强类型语言,即使值为interface{}类型,也可能造成list内部内存空洞,利用率低,如 interface{}
  为16B,但是只存一个int8类型数据,只占用1B,则浪费了15B内存空间;
  byte类型可以保证list内部不会有内存空洞,数据边界通过 记录偏移量 存入 hashtable中实现取数逻辑;
  值存储 key,value,访问次数;偏移量记录 起始位置/key长度/value长度/访问次数长度(int32即可); 这样在访问时,根据偏移量即可进行操作

#### hashtable数据结构

* map数据结构,key为string类型,存缓存的名字,value为 结构体类型,存 对应存储数据的segment的指针地址,及具体数据的偏移量;
* 客户端自行实现 原始数据类型到byte类型的相互转换; 这样 缓存系统可以存储任意类型的数据,如map,slice等等,反正最终存的都是byte类型;

### 客户端服务端交互方式

* 使用grpc实现;其基于http2.0实现多路复用(多个请求共用同一个连接),且支持全双工通信方式(双方同时发送数据),支持安全认证
* 使用时 基于TSL/SSL+TOKEN认证方式 实现 传输时安全加密 及 账户密码认证(
  防止非法用户访问); [链接](https://zhuanlan.zhihu.com/p/375573984)

### 多核CPU,多线程处理方式

* 使用go语言天生的多核并发出现优势

### 多机器分布式缓存实现方式

* 使用一致性hash算法实现

# 相关链接

* [redis LFU算法概览](https://blog.csdn.net/u010887744/article/details/110357096)
* [redis LFU算法实现整体步骤](https://blog.csdn.net/u013277209/article/details/126754778)
* [redis LFU算法实现具体到代码层细节](https://blog.csdn.net/m0_69745415/article/details/124370410)
* [psutil使用](https://blog.csdn.net/haiming0415/article/details/125313441)
* [VSS、RSS、PSS、USS 内存使用分析](https://blog.csdn.net/m0_51504545/article/details/119685325)
* [gomonkey permission denied](https://blog.csdn.net/D1124615130/article/details/121660126)
* 
# todo

* go性能优化方法 https://tehub.com/a/c2qgqWywfl

