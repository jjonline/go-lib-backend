package crawler

import (
	"crypto/sha1"
	"encoding/hex"
	"os"
)

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
