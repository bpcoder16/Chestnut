package utils

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"strings"
)

func MD5String(plaintext string) string {
	// 创建一个 MD5 哈希对象
	hash := md5.New()
	// 写入数据到哈希对象
	hash.Write([]byte(plaintext))
	// 计算哈希值
	md5sum := hash.Sum(nil)
	// 将哈希值转换为十六进制字符串
	return hex.EncodeToString(md5sum)
}

func UniqueID() string {
	return uuid.New().String()
}

func SHA265String(plaintext string) string {
	hash := sha256.New()
	hash.Write([]byte(plaintext))
	hashedSum := hash.Sum(nil)
	// 将哈希值转换为十六进制字符串
	return hex.EncodeToString(hashedSum)
}

func ToString(val any) string {
	switch v := val.(type) {
	case nil:
		return ""
	case string:
		return v
	case bool:
		if v {
			return "true"
		}
		return "false"
	case float32, float64:
		// 去掉无意义小数点
		return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.6f", v), "0"), ".")
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", v)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v)
	case []interface{}, map[string]interface{}:
		bs, _ := json.Marshal(v)
		return string(bs)
	default:
		// 兜底处理
		bs, _ := json.Marshal(v)
		return string(bs)
	}
}
