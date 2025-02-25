package utils

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"github.com/google/uuid"
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
