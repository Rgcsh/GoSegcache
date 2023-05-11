package transform

import (
	"bytes"
	"encoding/binary"
	"time"
)

// StringToByte
//
//	@Description: 字符串转为byte类型
//	@param s:
//	@return []byte:
func StringToByte(s string) []byte {
	return []byte(s)
}

// ByteToString
//
//	@Description: byte转字符串
//	@param b:
//	@return string:
func ByteToString(b []byte) string {
	return string(b)
}

// Int32ToByte
//
//	@Description: 整形转换成字节
//	@param x:
//	@return []byte:
func Int32ToByte(x int32) []byte {
	bytesBuffer := bytes.NewBuffer([]byte{})
	_ = binary.Write(bytesBuffer, binary.BigEndian, x)
	return bytesBuffer.Bytes()
}

// ByteToInt32
//
//	@Description: 字节转换成整形
//	@param b:
//	@return x:
func ByteToInt32(b []byte) (x int32) {
	bytesBuffer := bytes.NewBuffer(b)
	_ = binary.Read(bytesBuffer, binary.BigEndian, &x)
	return x
}

// ByteToUint8
//
//	@Description: 字节转换成整形
//	@param b:
//	@return x:
func ByteToUint8(b []byte) (x uint8) {
	bytesBuffer := bytes.NewBuffer(b)
	_ = binary.Read(bytesBuffer, binary.BigEndian, &x)
	return x
}

// ByteToUint32
//
//	@Description: 字节转换成整形
//	@param b:
//	@return x:
func ByteToUint32(b []byte) (x uint32) {
	bytesBuffer := bytes.NewBuffer(b)
	_ = binary.Read(bytesBuffer, binary.BigEndian, &x)
	return x
}

// Int64ToByte
//
//	@Description: 整形转换成字节
//	@param x:
//	@return []byte:
func Int64ToByte(x int64) []byte {
	bytesBuffer := bytes.NewBuffer([]byte{})
	_ = binary.Write(bytesBuffer, binary.BigEndian, x)
	return bytesBuffer.Bytes()
}

// Uint8ToByte
//
//	@Description: 整形转换成字节
//	@param x:
//	@return []byte:
func Uint8ToByte(x uint8) []byte {
	bytesBuffer := bytes.NewBuffer([]byte{})
	_ = binary.Write(bytesBuffer, binary.BigEndian, x)
	return bytesBuffer.Bytes()
}

// ByteTailSplit
//
//		@Description: 对byte切片类型 从尾部向前的tailSplitNum个数据返回
//		@param b:
//		@param tailSplitNum:
//	 @return []byte:
//	 demo: >>> ByteTailSplit([]byte{1, 2, 3, 4, 5, 6},2)
//	 demo: >>> []byte{5,6}
func ByteTailSplit(b []byte, tailSplitNum int) []byte {
	startIndex := len(b) - tailSplitNum
	return b[startIndex:]
}

// ByteToUint16
//
//	@Description: Byte类型转为uint16类型
//	@param b:
//	@return x:
func ByteToUint16(b []byte) (x uint16) {
	bytesBuffer := bytes.NewBuffer(b)
	_ = binary.Read(bytesBuffer, binary.BigEndian, &x)
	return x
}

func Int64ToByteToUint16(x int64) (y uint16) {
	xBuffer := bytes.NewBuffer([]byte{})
	_ = binary.Write(xBuffer, binary.BigEndian, x)
	xByte := xBuffer.Bytes()

	yBuffer := bytes.NewBuffer(xByte[6:8])
	_ = binary.Read(yBuffer, binary.BigEndian, &y)
	return
}

// GetUnixMinutes
//
//	@Description: 获取unix分钟,返回int32,4个字节长度的数据结构就够用,不用int64
//	@return int32:
func GetUnixMinutes(t *time.Time) uint32 {
	unix := t.Unix()
	return uint32(unix / 60)
}

// Uint32ToByte
//
//	@Description: 整形转换成字节
//	@param x:
//	@return []byte:
func Uint32ToByte(x uint32) []byte {
	bytesBuffer := bytes.NewBuffer([]byte{})
	_ = binary.Write(bytesBuffer, binary.BigEndian, x)
	return bytesBuffer.Bytes()
}
