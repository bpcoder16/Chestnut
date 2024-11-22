package lru

type Config struct {
	Default struct {
		Size int `json:"size"`
	} `json:"default"`
	Expire struct {
		Size           int   `json:"size"`
		TTLMillisecond int64 `json:"ttlMillisecond"`
	} `json:"expire"`
}
