package convert

import (
	"bytes"
	"strconv"
	"strings"
	"unicode"
)

// String 能转换的string
type String string

// IsEmpty 是否为空字符串：空字符串、空白符等
func (s String) IsEmpty() bool {
	if len(string(s)) == 0 {
		return true
	}
	if strings.TrimSpace(string(s)) == "" {
		return true
	}
	return false
}

// Lower calls the strings.ToLower
func (s String) Lower() string {
	return strings.ToLower(string(s))
}

// Upper calls the strings.ToUpper
func (s String) Upper() string {
	return strings.ToUpper(string(s))
}

// Title calls the strings.Title
func (s String) Title() string {
	if s.IsEmpty() {
		return string(s)
	}
	return strings.Title(string(s))
}

// UnTitle return the original string if rune is not letter at index 0
func (s String) UnTitle() string {
	ss := string(s)
	if s.IsEmpty() {
		return ss
	}
	r := rune(ss[0])
	if !unicode.IsUpper(r) && !unicode.IsLower(r) {
		return ss
	}
	return string(unicode.ToLower(r)) + ss[1:]
}

// ToCamel converts the input text into camel case
func (s String) ToCamel() string {
	list := s.splitBy(func(r rune) bool {
		return r == '_'
	}, true)
	var target []string
	for _, item := range list {
		target = append(target, String(item).Title())
	}
	return strings.Join(target, "")
}

// ToSnake converts the input text into snake case
func (s String) ToSnake() string {
	list := s.splitBy(unicode.IsUpper, false)
	var target []string
	for _, item := range list {
		target = append(target, String(item).Lower())
	}
	return strings.Join(target, "_")
}

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

// it will not ignore spaces
func (s String) splitBy(fn func(r rune) bool, remove bool) []string {
	if s.IsEmpty() {
		return nil
	}
	var list []string
	buffer := new(bytes.Buffer)
	for _, r := range s {
		if fn(r) {
			if buffer.Len() != 0 {
				list = append(list, buffer.String())
				buffer.Reset()
			}
			if !remove {
				buffer.WriteRune(r)
			}
			continue
		}
		buffer.WriteRune(r)
	}
	if buffer.Len() != 0 {
		list = append(list, buffer.String())
	}
	return list
}
