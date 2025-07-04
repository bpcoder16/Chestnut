package signauth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"sort"
)

func buildSortedParamString(params map[string]any, toStringFunc func(any) string) string {
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	pairs := url.Values{}
	for _, key := range keys {
		pairs.Set(key, toStringFunc(params[key]))
	}

	return pairs.Encode()
}

func Signature(secretKey string, reqBody map[string]any, timestamp int64, toStringFunc func(any) string) (signStr string) {
	// 加入时间戳
	reqBody["timestamp"] = timestamp
	sortedParamStr := buildSortedParamString(reqBody, toStringFunc)

	// 计算 HMAC-SHA256 签名
	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(sortedParamStr))
	signStr = hex.EncodeToString(mac.Sum(nil))

	return
}
