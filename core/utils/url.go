package utils

import (
	"fmt"
	"net/url"
	"sort"
	"time"
)

func BuildSign(params map[string]interface{}, secretKey string) string {
	params["secretKey"] = secretKey
	keys := make([]string, 0, len(params))
	for key := range params {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	p := url.Values{}
	for _, key := range keys {
		p.Set(key, fmt.Sprint(params[key]))
	}
	return MD5String(p.Encode())
}

func CheckSign(params map[string]interface{}, secretKey string) bool {
	nowTime := time.Now().Unix()
	timestamp := params["timeStamp"].(int64)
	if nowTime+5 < timestamp || nowTime-5 > timestamp {
		return false
	}
	paramSign := params["sign"].(string)
	checkParams := make(map[string]any)
	for k, v := range params {
		if k == "sign" {
			continue
		}
		checkParams[k] = v
	}
	return paramSign == BuildSign(checkParams, secretKey)
}
