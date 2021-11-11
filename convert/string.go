package convert

import (
	"strconv"
	"strings"
)

// String 能转换的string
type String string

// Int 转换为int型
// 示例：
//   a := "1"
//   i := convert.String(a).Int()
func (s String) Int() int {
	if s == "" {
		return 0
	}
	i, _ := strconv.Atoi(string(s))
	return i
}

// UInt8 字符串转换为uint8
// 示例：
//   a := "1"
//   i := convert.String(a).UInt8()
func (s String) UInt8() uint8 {
	return uint8(s.Int64())
}

// UInt32 字符串转换为uint32
// 示例：
//   a := "1"
//   i := convert.String(a).UInt32()
func (s String) UInt32() uint32 {
	return uint32(s.Int64())
}

// Int64 字符串转换为int64
// 示例：
//   a := "1"
//   i := convert.String(a).Int64()
func (s String) Int64() int64 {
	if s == "" {
		return 0
	}
	i, _ := strconv.ParseInt(string(s), 10, 64)
	return i
}

// UInt64 字符串转换为uint64
// 示例：
//   a := "1"
//   i := convert.String(a).UInt64()
func (s String) UInt64() uint64 {
	if s == "" {
		return 0
	}
	i, _ := strconv.ParseInt(string(s), 10, 64)
	return uint64(i)
}

// Float64 字符串转换为float64
// 示例：
//   a := "1.1"
//   i := convert.String(a).Float64()
func (s String) Float64() float64 {
	f, _ := strconv.ParseFloat(string(s), 64)
	return f
}

// IntSlice 字符串转换为int切片
// 示例：
//   a := "1,2,4"
//   i := convert.String(a).IntSlice(",")
func (s String) IntSlice(sep string) []int {
	if s == "" {
		return nil
	}
	ss := strings.Split(string(s), sep)
	ret := make([]int, 0, len(ss))
	for i := range ss {
		ret = append(ret, String(ss[i]).Int())
	}
	return ret
}

// UInt32Slice 字符串转换为uint32切片
// 示例：
//   a := "1,2,4"
//   i := convert.String(a).UInt32Slice(",")
func (s String) UInt32Slice(sep string) []uint32 {
	if s == "" {
		return nil
	}
	ss := strings.Split(string(s), sep)
	ret := make([]uint32, 0, len(ss))
	for i := range ss {
		ret = append(ret, String(ss[i]).UInt32())
	}
	return ret
}

// UInt64Slice 字符串转换为uint64切片
// 示例：
//   a := "1,2,4"
//   i := convert.String(a).UInt64Slice(",")
func (s String) UInt64Slice(sep string) []uint64 {
	if s == "" {
		return nil
	}
	ss := strings.Split(string(s), sep)
	ret := make([]uint64, 0, len(ss))
	for i := range ss {
		ret = append(ret, String(ss[i]).UInt64())
	}
	return ret
}

// Int64Slice 字符串转换为int64切片
// 示例：
//   a := "1,2,4"
//   i := convert.String(a).UInt64Slice(",")
func (s String) Int64Slice(sep string) []int64 {
	if s == "" {
		return nil
	}
	ss := strings.Split(string(s), sep)
	ret := make([]int64, 0, len(ss))
	for i := range ss {
		ret = append(ret, String(ss[i]).Int64())
	}
	return ret
}
