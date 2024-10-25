package crawler

import (
	"crypto/sha1"
	"encoding/hex"
	randV2 "math/rand/v2"
	"os"
	"time"
)

// pseudoRandFloat64 [0.0,1.0)范围随机数
func pseudoRandFloat64() float64 {
	return randV2.Float64()
}

// RandFloat64 生成 [0,max) 范围内的float64类型的随机数
func RandFloat64(max int64) float64 {
	return float64(randV2.Int64N(max)) + randV2.Float64()
}

// RandFloat64MinMax 生成 [min,max) 范围内的float64类型的随机数
func RandFloat64MinMax(min, max int64) float64 {
	// 最大最小一致，永远返回一致值
	if min == max {
		return float64(min)
	}

	// 最大最小颠倒，调换一下
	if min > max {
		mid := max
		max = min
		min = mid
	}
	return RandFloat64(max-min) + float64(min)
}

// RandInt64 生成 [0,max) 范围内的int64类型随机数
func RandInt64(max int64) int64 {
	return randV2.Int64N(max)
}

// RandInt64MinMax 区间 [min,max)
func RandInt64MinMax(min, max int64) int64 {
	// 最大最小一致，永远返回一致值
	if min == max {
		return min
	}

	// 最大最小颠倒，调换一下
	if min > max {
		mid := max
		max = min
		min = mid
	}

	return RandInt64(max-min) + min
}

// RandInt64MinMaxDuration 区间 [min,max)时长随机秒
func RandInt64MinMaxDuration(min, max int64) time.Duration {
	return time.Duration(RandFloat64MinMax(min, max)) * time.Second
}

// 判断文件夹是否存在
func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

// hashSha1 比特计算sha1的哈希值
func hashSha1(b []byte) string {
	d := sha1.Sum(b)
	return hex.EncodeToString(d[0:])
}
