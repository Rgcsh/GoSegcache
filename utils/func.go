package utils

import (
	"fmt"
	"github.com/shirou/gopsutil/process"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// GenRandDecimal
//
//	@Description: 获取随机float32类型小数
//	@return float32:
func GenRandDecimal() float32 {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.Float32()
}

// ToBytes
//
//	@Description: 将KB,MB,GB,TB格式转为Byte类型的数据大小
//	@param size:
//	@param uint:
//	@return int64:
func ToBytes(size float64, unit string) int64 {
	switch strings.ToUpper(unit) {
	case "K":
		return int64(size * KB)
	case "M":
		return int64(size * MB)
	case "G":
		return int64(size * GB)
	case "T":
		return int64(size * TB)
	default:
		return int64(size)
	}
}

// ExtractStoreUnit
//
//	@Description: 抽取出 字符串格式的 存储数据表达式 为 数据量 和 单位
//	@param s: 如 "3K"
//	@return size: 返回值如 3
//	@return unit: 返回值如 K
func ExtractStoreUnit(s string) (size float64, unit string, err error) {
	size, err = strconv.ParseFloat(s[0:len(s)-1], 64)
	if err != nil {
		return 0, "", err
	}
	unit = s[len(s)-1 : len(s)]
	return
}

// GetRuntimeMemoryStats
//
//	@Description: 获取当前程序 内存使用情况(通过调用go自身的runtime里的函数获取)
func GetRuntimeMemoryStats() uint64 {
	m := runtime.MemStats{}
	runtime.ReadMemStats(&m)
	return m.Alloc
}

// GetProcessMemoryInfo
//
//	@Description: 获取当前运行进程的内存信息(通过系统调用获取,如底层执行ps命令)
func GetProcessMemoryInfo() {
	pid := os.Getpid()
	p, err := process.NewProcess(int32(pid))
	if err != nil {
		fmt.Println("获取进程信息", err)
	}
	//RSS:实际占用的物理内存,可以依此作为 内存是否达到阈值的判断依据
	//VMS:实际使用的虚拟内存
	fmt.Println(p.MemoryInfo())
	fmt.Println(p.MemoryPercent())
}

// GetProcessId
//
//	@Description: 获取当前进程ID
//	@return int:
func GetProcessId() int {
	return os.Getpid()
}

// GetProcessPhysicalMemory
//
//	@Description: 获取进程消耗的物理内存
func GetProcessPhysicalMemory(pid int) (uint64, error) {
	p, err := process.NewProcess(int32(pid))
	if err != nil {
		return 0, err
	}
	//RSS:实际占用的物理内存,可以依此作为 内存是否达到阈值的判断依据
	//VMS:实际使用的虚拟内存
	memoryInfoStat, err := p.MemoryInfo()
	return memoryInfoStat.RSS, err
}

// GetMaxSize
//
//	@Description: 获取最大值
func GetMaxSize(a, b int) int {
	if a < b {
		return b
	}
	return a
}
