package tcrypt

import (
	"crypto/md5"
	"encoding/hex"
)

// Md5Encrypt 计算给定字符串的 MD5 哈希值
// 输入: data string - 需要加密的字符串
// 输出: string - MD5 哈希值的十六进制表示
func Md5Encrypt(data string) string {
	hasher := md5.New()
	hasher.Write([]byte(data))
	return hex.EncodeToString(hasher.Sum(nil))
}